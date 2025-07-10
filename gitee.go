// Gitee OAuth2登录提供者实现
// 支持Gitee网页登录和用户信息获取
package idp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// GiteeIdProvider Gitee登录提供者
// 实现Gitee OAuth2登录功能
type GiteeIdProvider struct {
	Client *http.Client    // HTTP客户端
	Config *oauth2.Config // OAuth2配置
}

// NewGiteeIdProvider 创建Gitee登录提供者实例
// 参数:
//   - clientId: Gitee应用的Client ID
//   - clientSecret: Gitee应用的Client Secret
//   - redirectUrl: OAuth2重定向URL
// 返回:
//   - *GiteeIdProvider: Gitee登录提供者实例
func NewGiteeIdProvider(clientId string, clientSecret string, redirectUrl string) *GiteeIdProvider {
	idp := &GiteeIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

// SetHttpClient 设置HTTP客户端
// 参数:
//   - client: HTTP客户端实例
func (idp *GiteeIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig 获取Gitee OAuth2配置
// 参数:
//   - clientId: Gitee应用的Client ID
//   - clientSecret: Gitee应用的Client Secret
//   - redirectUrl: OAuth2重定向URL
// 返回:
//   - *oauth2.Config: OAuth2配置实例
func (idp *GiteeIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL: "https://gitee.com/oauth/token",
	}

	config := &oauth2.Config{
		Scopes: []string{"user_info emails"},

		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

// GiteeAccessToken Gitee访问令牌结构体
type GiteeAccessToken struct {
	AccessToken  string `json:"access_token"`  // 访问令牌
	TokenType    string `json:"token_type"`    // 令牌类型
	ExpiresIn    int    `json:"expires_in"`    // 访问令牌的有效时间，单位是秒
	RefreshToken string `json:"refresh_token"` // 刷新令牌
	Scope        string `json:"scope"`         // 授权范围
	CreatedAt    int    `json:"created_at"`    // 创建时间
}

// GetToken 通过授权码获取Gitee访问令牌
// 参数:
//   - code: Gitee返回的授权码
// 返回:
//   - *oauth2.Token: OAuth2访问令牌
//   - error: 错误信息
// 详细文档: https://gitee.com/api/v5/oauth_doc#/
func (idp *GiteeIdProvider) GetToken(code string) (*oauth2.Token, error) {
	params := url.Values{}
	params.Add("grant_type", "authorization_code")
	params.Add("client_id", idp.Config.ClientID)
	params.Add("client_secret", idp.Config.ClientSecret)
	params.Add("code", code)
	params.Add("redirect_uri", idp.Config.RedirectURL)

	accessTokenUrl := fmt.Sprintf("%s?%s", idp.Config.Endpoint.TokenURL, params.Encode())
	bs, _ := json.Marshal(params.Encode())
	req, _ := http.NewRequest("POST", accessTokenUrl, strings.NewReader(string(bs)))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.101 Safari/537.36")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	rbs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	tokenResp := GiteeAccessToken{}
	if err = json.Unmarshal(rbs, &tokenResp); err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		RefreshToken: tokenResp.RefreshToken,
		Expiry:       time.Unix(time.Now().Unix()+int64(tokenResp.ExpiresIn), 0),
	}

	return token, nil
}

// GiteeUserResponse Gitee用户信息响应结构体
type GiteeUserResponse struct {
	AvatarUrl         string `json:"avatar_url"`         // 头像URL
	Bio               string `json:"bio"`               // 个人简介
	Blog              string `json:"blog"`              // 博客地址
	CreatedAt         string `json:"created_at"`         // 创建时间
	Email             string `json:"email"`             // 邮箱地址
	EventsUrl         string `json:"events_url"`         // 事件URL
	Followers         int    `json:"followers"`          // 粉丝数
	FollowersUrl      string `json:"followers_url"`      // 粉丝URL
	Following         int    `json:"following"`          // 关注数
	FollowingUrl      string `json:"following_url"`      // 关注URL
	GistsUrl          string `json:"gists_url"`          // Gists URL
	HtmlUrl           string `json:"html_url"`           // HTML URL
	Id                int    `json:"id"`                // 用户ID
	Login             string `json:"login"`             // 登录名
	MemberRole        string `json:"member_role"`        // 成员角色
	Name              string `json:"name"`              // 用户名
	OrganizationsUrl  string `json:"organizations_url"`  // 组织URL
	PublicGists       int    `json:"public_gists"`       // 公开Gists数
	PublicRepos       int    `json:"public_repos"`       // 公开仓库数
	ReceivedEventsUrl string `json:"received_events_url"` // 接收事件URL
	ReposUrl          string `json:"repos_url"`          // 仓库URL
	Stared            int    `json:"stared"`            // 收藏数
	StarredUrl        string `json:"starred_url"`        // 收藏URL
	SubscriptionsUrl  string `json:"subscriptions_url"`  // 订阅URL
	Type              string `json:"type"`              // 用户类型
	UpdatedAt         string `json:"updated_at"`         // 更新时间
	Url               string `json:"url"`               // 用户URL
	Watched           int    `json:"watched"`           // 关注数
	Weibo             string `json:"weibo"`             // 微博地址
}

// GetUserInfo 获取Gitee用户信息
// 参数:
//   - token: OAuth2访问令牌
// 返回:
//   - *UserInfo: 标准化用户信息
//   - error: 错误信息
// 详细文档: https://gitee.com/api/v5/swagger#/getV5User
func (idp *GiteeIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	var gtUserInfo GiteeUserResponse
	accessToken := token.AccessToken

	u := fmt.Sprintf("https://gitee.com/api/v5/user?access_token=%s",
		accessToken)

	userinfoResp, err := idp.GetUrlResp(u)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(userinfoResp), &gtUserInfo); err != nil {
		return nil, err
	}

	userInfo := UserInfo{
		Id:          strconv.Itoa(gtUserInfo.Id),
		Username:    gtUserInfo.Name,
		DisplayName: gtUserInfo.Name,
		Email:       gtUserInfo.Email,
		AvatarUrl:   gtUserInfo.AvatarUrl,
	}

	return &userInfo, nil
}

// GetUrlResp 发送HTTP GET请求并获取响应内容
// 参数:
//   - url: 请求URL
// 返回:
//   - string: 响应内容
//   - error: 错误信息
func (idp *GiteeIdProvider) GetUrlResp(url string) (string, error) {
	resp, err := idp.Client.Get(url)
	if err != nil {
		return "", err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
