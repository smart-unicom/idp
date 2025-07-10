// 哔哩哔哩OAuth2登录提供者实现
// 支持哔哩哔哩网页登录和用户信息获取
package idp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// BilibiliIdProvider 哔哩哔哩登录提供者
// 实现哔哩哔哩OAuth2登录功能
type BilibiliIdProvider struct {
	Client *http.Client    // HTTP客户端
	Config *oauth2.Config // OAuth2配置
}

// NewBilibiliIdProvider 创建哔哩哔哩登录提供者实例
// 参数:
//   - clientId: 哔哩哔哩应用的Client ID
//   - clientSecret: 哔哩哔哩应用的Client Secret
//   - redirectUrl: OAuth2重定向URL
// 返回:
//   - *BilibiliIdProvider: 哔哩哔哩登录提供者实例
func NewBilibiliIdProvider(clientId string, clientSecret string, redirectUrl string) *BilibiliIdProvider {
	idp := &BilibiliIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

// SetHttpClient 设置HTTP客户端
// 参数:
//   - client: HTTP客户端实例
func (idp *BilibiliIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig 获取哔哩哔哩OAuth2配置
// 参数:
//   - clientId: 哔哩哔哩应用的Client ID
//   - clientSecret: 哔哩哔哩应用的Client Secret
//   - redirectUrl: OAuth2重定向URL
// 返回:
//   - *oauth2.Config: OAuth2配置实例
func (idp *BilibiliIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL: "https://api.bilibili.com/x/account-oauth2/v1/token",
		AuthURL:  "http://member.bilibili.com/arcopen/fn/user/account/info",
	}

	config := &oauth2.Config{
		Scopes:       []string{"", ""},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

// BilibiliProviderToken 哔哩哔哩访问令牌结构体
type BilibiliProviderToken struct {
	AccessToken  string `json:"access_token"`  // 访问令牌
	ExpiresIn    int    `json:"expires_in"`    // 访问令牌的有效时间，单位是秒
	RefreshToken string `json:"refresh_token"` // 刷新令牌
}

// BilibiliIdProviderTokenResponse 哔哩哔哩令牌响应结构体
type BilibiliIdProviderTokenResponse struct {
	Code    int                   `json:"code"`    // 响应码，0表示成功
	Message string                `json:"message"` // 响应消息
	TTL     int                   `json:"ttl"`     // 生存时间
	Data    BilibiliProviderToken `json:"data"`    // 令牌数据
}

// GetToken 通过授权码获取哔哩哔哩访问令牌
// 参数:
//   - code: 哔哩哔哩返回的授权码
// 返回:
//   - *oauth2.Token: OAuth2访问令牌
//   - error: 错误信息
// 详细文档: https://openhome.bilibili.com/doc/4/eaf0e2b5-bde9-b9a0-9be1-019bb455701c
func (idp *BilibiliIdProvider) GetToken(code string) (*oauth2.Token, error) {
	pTokenParams := &struct {
		ClientId     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		GrantType    string `json:"grant_type"`
		Code         string `json:"code"`
	}{
		idp.Config.ClientID,
		idp.Config.ClientSecret,
		"authorization_code",
		code,
	}

	data, err := idp.postWithBody(pTokenParams, idp.Config.Endpoint.TokenURL)
	if err != nil {
		return nil, err
	}

	response := &BilibiliIdProviderTokenResponse{}
	err = json.Unmarshal(data, response)
	if err != nil {
		return nil, err
	}

	if response.Code != 0 {
		return nil, fmt.Errorf("pToken.Errcode = %d, pToken.Errmsg = %s", response.Code, response.Message)
	}

	token := &oauth2.Token{
		AccessToken:  response.Data.AccessToken,
		Expiry:       time.Unix(time.Now().Unix()+int64(response.Data.ExpiresIn), 0),
		RefreshToken: response.Data.RefreshToken,
	}

	return token, nil
}

// BilibiliUserInfo 哔哩哔哩用户信息结构体
type BilibiliUserInfo struct {
	Name   string `json:"name"`   // 用户昵称
	Face   string `json:"face"`   // 用户头像URL
	OpenId string `json:"openid"` // 用户OpenID
}

// BilibiliUserInfoResponse 哔哩哔哩用户信息响应结构体
type BilibiliUserInfoResponse struct {
	Code    int              `json:"code"`    // 响应码，0表示成功
	Message string           `json:"message"` // 响应消息
	TTL     int              `json:"ttl"`     // 生存时间
	Data    BilibiliUserInfo `json:"data"`    // 用户信息数据
}

// GetUserInfo 通过访问令牌获取哔哩哔哩用户信息
// 参数:
//   - token: OAuth2访问令牌
// 返回:
//   - *UserInfo: 标准化用户信息
//   - error: 错误信息
// 详细文档: https://openhome.bilibili.com/doc/4/feb66f99-7d87-c206-00e7-d84164cd701c
func (idp *BilibiliIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	accessToken := token.AccessToken
	clientId := idp.Config.ClientID

	params := url.Values{}
	params.Add("client_id", clientId)
	params.Add("access_token", accessToken)

	userInfoUrl := fmt.Sprintf("%s?%s", idp.Config.Endpoint.AuthURL, params.Encode())

	resp, err := idp.Client.Get(userInfoUrl)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bUserInfoResponse := &BilibiliUserInfoResponse{}
	if err = json.Unmarshal(data, bUserInfoResponse); err != nil {
		return nil, err
	}

	if bUserInfoResponse.Code != 0 {
		return nil, fmt.Errorf("userinfo.Errcode = %d, userinfo.Errmsg = %s", bUserInfoResponse.Code, bUserInfoResponse.Message)
	}

	userInfo := &UserInfo{
		Id:          bUserInfoResponse.Data.OpenId,
		Username:    bUserInfoResponse.Data.Name,
		DisplayName: bUserInfoResponse.Data.Name,
		AvatarUrl:   bUserInfoResponse.Data.Face,
	}

	return userInfo, nil
}

// postWithBody 发送POST请求并获取响应内容
// 参数:
//   - body: 请求体数据
//   - url: 请求URL
// 返回:
//   - []byte: 响应内容
//   - error: 错误信息
func (idp *BilibiliIdProvider) postWithBody(body interface{}, url string) ([]byte, error) {
	bs, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	r := strings.NewReader(string(bs))
	resp, err := idp.Client.Post(url, "application/json;charset=UTF-8", r)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	return data, nil
}
