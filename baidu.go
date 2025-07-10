// 百度OAuth2登录提供者实现
// 支持百度网页登录和用户信息获取
package idp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

// BaiduIdProvider 百度登录提供者
// 实现百度OAuth2登录功能
type BaiduIdProvider struct {
	Client *http.Client    // HTTP客户端
	Config *oauth2.Config // OAuth2配置
}

// NewBaiduIdProvider 创建百度登录提供者实例
// 参数:
//   - clientId: 百度应用的API Key
//   - clientSecret: 百度应用的Secret Key
//   - redirectUrl: OAuth2重定向URL
// 返回:
//   - *BaiduIdProvider: 百度登录提供者实例
func NewBaiduIdProvider(clientId string, clientSecret string, redirectUrl string) *BaiduIdProvider {
	idp := &BaiduIdProvider{}

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
func (idp *BaiduIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig 获取百度OAuth2配置
// 返回:
//   - *oauth2.Config: OAuth2配置实例
func (idp *BaiduIdProvider) getConfig() *oauth2.Config {
	endpoint := oauth2.Endpoint{
		AuthURL:  "https://openapi.baidu.com/oauth/2.0/authorize",
		TokenURL: "https://openapi.baidu.com/oauth/2.0/token",
	}

	config := &oauth2.Config{
		Scopes:   []string{"email"},
		Endpoint: endpoint,
	}

	return config
}

// GetToken 通过授权码获取百度访问令牌
// 参数:
//   - code: 百度返回的授权码
// 返回:
//   - *oauth2.Token: OAuth2访问令牌
//   - error: 错误信息
func (idp *BaiduIdProvider) GetToken(code string) (*oauth2.Token, error) {
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, idp.Client)
	return idp.Config.Exchange(ctx, code)
}

/*
{
    "userid":"2097322476",
    "username":"wl19871011",
    "realname":"阳光",
    "userdetail":"喜欢自由",
    "birthday":"1987-01-01",
    "marriage":"恋爱",
    "sex":"男",
    "blood":"O",
    "constellation":"射手",
    "figure":"小巧",
    "education":"大学/专科",
    "trade":"计算机/电子产品",
    "job":"未知",
    "birthday_year":"1987",
    "birthday_month":"01",
    "birthday_day":"01",
}
*/

// BaiduUserInfo 百度用户信息结构体
type BaiduUserInfo struct {
	OpenId   string `json:"openid"`   // 用户的OpenID
	Username string `json:"username"` // 用户名
	Portrait string `json:"portrait"` // 用户头像标识
}

// GetUserInfo 通过访问令牌获取百度用户信息
// 参数:
//   - token: OAuth2访问令牌
// 返回:
//   - *UserInfo: 标准化用户信息
//   - error: 错误信息
func (idp *BaiduIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	resp, err := idp.Client.Get(fmt.Sprintf("https://openapi.baidu.com/rest/2.0/passport/users/getInfo?access_token=%s", token.AccessToken))
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	baiduUser := BaiduUserInfo{}
	if err = json.Unmarshal(data, &baiduUser); err != nil {
		return nil, err
	}

	userInfo := UserInfo{
		Id:          baiduUser.OpenId,
		Username:    baiduUser.Username,
		DisplayName: baiduUser.Username,
		AvatarUrl:   fmt.Sprintf("https://himg.bdimg.com/sys/portrait/item/%s", baiduUser.Portrait),
	}
	return &userInfo, nil
}
