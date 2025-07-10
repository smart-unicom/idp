// QQ OAuth2登录提供者实现
// 支持QQ网页登录和用户信息获取
package idp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"

	"golang.org/x/oauth2"
)

// QqIdProvider QQ登录提供者
// 实现QQ OAuth2登录功能
type QqIdProvider struct {
	Client *http.Client    // HTTP客户端
	Config *oauth2.Config // OAuth2配置
}

// NewQqIdProvider 创建QQ登录提供者实例
// 参数:
//   - clientId: QQ应用的AppId
//   - clientSecret: QQ应用的AppKey
//   - redirectUrl: OAuth2重定向URL
// 返回:
//   - *QqIdProvider: QQ登录提供者实例
func NewQqIdProvider(clientId string, clientSecret string, redirectUrl string) *QqIdProvider {
	idp := &QqIdProvider{}

	config := idp.getConfig()
	config.ClientID = clientId
	config.ClientSecret = clientSecret
	config.RedirectURL = redirectUrl
	idp.Config = config

	return idp
}

// SetHttpClient 设置HTTP客户端
// 参数:
//   - client: HTTP客户端实例
func (idp *QqIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig 获取QQ OAuth2配置
// 返回:
//   - *oauth2.Config: OAuth2配置实例
func (idp *QqIdProvider) getConfig() *oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL: "https://graph.qq.com/oauth2.0/token",
	}

	config := &oauth2.Config{
		Scopes:   []string{"get_user_info"},
		Endpoint: endpoint,
	}

	return config
}

// GetToken 通过授权码获取QQ访问令牌
// 参数:
//   - code: QQ返回的授权码
// 返回:
//   - *oauth2.Token: OAuth2访问令牌
//   - error: 错误信息
func (idp *QqIdProvider) GetToken(code string) (*oauth2.Token, error) {
	params := url.Values{}
	params.Add("grant_type", "authorization_code")
	params.Add("client_id", idp.Config.ClientID)
	params.Add("client_secret", idp.Config.ClientSecret)
	params.Add("code", code)
	params.Add("redirect_uri", idp.Config.RedirectURL)

	accessTokenUrl := fmt.Sprintf("https://graph.qq.com/oauth2.0/token?%s", params.Encode())
	resp, err := idp.Client.Get(accessTokenUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	tokenContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile("token=(.*?)&")
	matched := re.FindAllStringSubmatch(string(tokenContent), -1)
	accessToken := matched[0][1]
	token := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}
	return token, nil
}

// QqUserInfo QQ用户信息结构体
type QqUserInfo struct {
	Ret             int    `json:"ret"`               // 返回码
	Msg             string `json:"msg"`               // 返回信息
	IsLost          int    `json:"is_lost"`           // 是否丢失
	Nickname        string `json:"nickname"`          // 用户昵称
	Gender          string `json:"gender"`            // 性别
	GenderType      int    `json:"gender_type"`       // 性别类型
	Province        string `json:"province"`          // 省份
	City            string `json:"city"`              // 城市
	Year            string `json:"year"`              // 出生年份
	Constellation   string `json:"constellation"`     // 星座
	Figureurl       string `json:"figureurl"`         // 头像URL
	Figureurl1      string `json:"figureurl_1"`       // 头像URL1
	Figureurl2      string `json:"figureurl_2"`       // 头像URL2
	FigureurlQq1    string `json:"figureurl_qq_1"`    // QQ头像URL1
	FigureurlQq2    string `json:"figureurl_qq_2"`    // QQ头像URL2
	FigureurlQq     string `json:"figureurl_qq"`      // QQ头像URL
	FigureurlType   string `json:"figureurl_type"`    // 头像类型
	IsYellowVip     string `json:"is_yellow_vip"`     // 是否黄钻用户
	Vip             string `json:"vip"`               // 是否VIP用户
	YellowVipLevel  string `json:"yellow_vip_level"`  // 黄钻等级
	Level           string `json:"level"`             // 用户等级
	IsYellowYearVip string `json:"is_yellow_year_vip"` // 是否年费黄钻用户
}

// GetUserInfo 通过访问令牌获取QQ用户信息
// 参数:
//   - token: OAuth2访问令牌
// 返回:
//   - *UserInfo: 标准化用户信息
//   - error: 错误信息
func (idp *QqIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	openIdUrl := fmt.Sprintf("https://graph.qq.com/oauth2.0/me?access_token=%s", token.AccessToken)
	resp, err := idp.Client.Get(openIdUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	openIdBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile("\"openid\":\"(.*?)\"}")
	matched := re.FindAllStringSubmatch(string(openIdBody), -1)
	openId := matched[0][1]
	if openId == "" {
		return nil, errors.New("openId is empty")
	}

	userInfoUrl := fmt.Sprintf(
		"https://graph.qq.com/user/get_user_info?access_token=%s&oauth_consumer_key=%s&openid=%s",
		token.AccessToken, idp.Config.ClientID, openId)
	resp, err = idp.Client.Get(userInfoUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	userInfoBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var qqUserInfo QqUserInfo
	err = json.Unmarshal(userInfoBody, &qqUserInfo)
	if err != nil {
		return nil, err
	}

	if qqUserInfo.Ret != 0 {
		return nil, fmt.Errorf("ret expected 0, got %d", qqUserInfo.Ret)
	}

	userInfo := UserInfo{
		Id:          openId,
		Username:    qqUserInfo.Nickname,
		DisplayName: qqUserInfo.Nickname,
		AvatarUrl:   qqUserInfo.FigureurlQq1,
	}
	return &userInfo, nil
}
