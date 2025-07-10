// GitHub OAuth2 登录提供者实现
package idp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// GithubIdProvider GitHub登录提供者
// 实现GitHub OAuth2登录功能
type GithubIdProvider struct {
	Client *http.Client   // HTTP客户端
	Config *oauth2.Config // OAuth2配置
}

// NewGithubIdProvider 创建GitHub登录提供者实例
// 参数:
//   - clientId: GitHub应用的客户端ID
//   - clientSecret: GitHub应用的客户端密钥
//   - redirectUrl: OAuth2重定向URL
//
// 返回:
//   - *GithubIdProvider: GitHub登录提供者实例
func NewGithubIdProvider(clientId string, clientSecret string, redirectUrl string) *GithubIdProvider {
	idp := &GithubIdProvider{}

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
func (idp *GithubIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig 获取GitHub OAuth2配置
// 返回:
//   - *oauth2.Config: OAuth2配置实例
func (idp *GithubIdProvider) getConfig() *oauth2.Config {
	endpoint := oauth2.Endpoint{
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
	}

	config := &oauth2.Config{
		Scopes:   []string{"user:email", "read:user"},
		Endpoint: endpoint,
	}

	return config
}

// GithubToken GitHub访问令牌响应结构体
type GithubToken struct {
	AccessToken string `json:"access_token"` // 访问令牌
	TokenType   string `json:"token_type"`   // 令牌类型
	Scope       string `json:"scope"`        // 授权范围
	Error       string `json:"error"`        // 错误信息
}

// GetToken 通过授权码获取GitHub访问令牌
// 参数:
//   - code: GitHub返回的授权码
//
// 返回:
//   - *oauth2.Token: OAuth2访问令牌
//   - error: 错误信息
func (idp *GithubIdProvider) GetToken(code string) (*oauth2.Token, error) {
	params := &struct {
		Code         string `json:"code"`
		ClientId     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}{code, idp.Config.ClientID, idp.Config.ClientSecret}
	data, err := idp.postWithBody(params, idp.Config.Endpoint.TokenURL)
	if err != nil {
		return nil, err
	}
	pToken := &GithubToken{}
	if err = json.Unmarshal(data, pToken); err != nil {
		return nil, err
	}
	if pToken.Error != "" {
		return nil, fmt.Errorf("err: %s", pToken.Error)
	}

	token := &oauth2.Token{
		AccessToken: pToken.AccessToken,
		TokenType:   "Bearer",
	}

	return token, nil
}

// GitHubUserInfo GitHub用户信息响应结构体
// 包含从GitHub API获取的完整用户信息
type GitHubUserInfo struct {
	Login                   string      `json:"login"`                     // 用户名
	Id                      int         `json:"id"`                        // 用户ID
	NodeId                  string      `json:"node_id"`                   // 节点ID
	AvatarUrl               string      `json:"avatar_url"`                // 头像URL
	GravatarId              string      `json:"gravatar_id"`               // Gravatar ID
	Url                     string      `json:"url"`                       // API URL
	HtmlUrl                 string      `json:"html_url"`                  // 主页URL
	FollowersUrl            string      `json:"followers_url"`             // 关注者URL
	FollowingUrl            string      `json:"following_url"`             // 关注URL
	GistsUrl                string      `json:"gists_url"`                 // Gists URL
	StarredUrl              string      `json:"starred_url"`               // 星标URL
	SubscriptionsUrl        string      `json:"subscriptions_url"`         // 订阅URL
	OrganizationsUrl        string      `json:"organizations_url"`         // 组织URL
	ReposUrl                string      `json:"repos_url"`                 // 仓库URL
	EventsUrl               string      `json:"events_url"`                // 事件URL
	ReceivedEventsUrl       string      `json:"received_events_url"`       // 接收事件URL
	Type                    string      `json:"type"`                      // 用户类型
	SiteAdmin               bool        `json:"site_admin"`                // 是否为站点管理员
	Name                    string      `json:"name"`                      // 真实姓名
	Company                 string      `json:"company"`                   // 公司
	Blog                    string      `json:"blog"`                      // 博客
	Location                string      `json:"location"`                  // 位置
	Email                   string      `json:"email"`                     // 邮箱
	Hireable                bool        `json:"hireable"`                  // 是否可雇佣
	Bio                     string      `json:"bio"`                       // 个人简介
	TwitterUsername         interface{} `json:"twitter_username"`          // Twitter用户名
	PublicRepos             int         `json:"public_repos"`              // 公开仓库数
	PublicGists             int         `json:"public_gists"`              // 公开Gists数
	Followers               int         `json:"followers"`                 // 关注者数
	Following               int         `json:"following"`                 // 关注数
	CreatedAt               time.Time   `json:"created_at"`                // 创建时间
	UpdatedAt               time.Time   `json:"updated_at"`                // 更新时间
	PrivateGists            int         `json:"private_gists"`             // 私有Gists数
	TotalPrivateRepos       int         `json:"total_private_repos"`       // 总私有仓库数
	OwnedPrivateRepos       int         `json:"owned_private_repos"`       // 拥有的私有仓库数
	DiskUsage               int         `json:"disk_usage"`                // 磁盘使用量
	Collaborators           int         `json:"collaborators"`             // 协作者数
	TwoFactorAuthentication bool        `json:"two_factor_authentication"` // 是否启用双因子认证
	Plan                    struct {    // 订阅计划
		Name          string `json:"name"`          // 计划名称
		Space         int    `json:"space"`         // 存储空间
		Collaborators int    `json:"collaborators"` // 协作者数量
		PrivateRepos  int    `json:"private_repos"` // 私有仓库数量
	} `json:"plan"`
}

// GetUserInfo 通过访问令牌获取GitHub用户信息
// 参数:
//   - token: OAuth2访问令牌
//
// 返回:
//   - *UserInfo: 标准化的用户信息
//   - error: 错误信息
func (idp *GithubIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "token "+token.AccessToken)
	resp, err := idp.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var githubUserInfo GitHubUserInfo
	err = json.Unmarshal(body, &githubUserInfo)
	if err != nil {
		return nil, err
	}

	userInfo := UserInfo{
		Id:          strconv.Itoa(githubUserInfo.Id),
		Username:    githubUserInfo.Login,
		DisplayName: githubUserInfo.Name,
		Email:       githubUserInfo.Email,
		AvatarUrl:   githubUserInfo.AvatarUrl,
	}
	return &userInfo, nil
}

// postWithBody 发送POST请求
// 参数:
//   - body: 请求体数据
//   - url: 请求URL
//
// 返回:
//   - []byte: 响应数据
//   - error: 错误信息
func (idp *GithubIdProvider) postWithBody(body interface{}, url string) ([]byte, error) {
	bs, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	r := strings.NewReader(string(bs))
	req, _ := http.NewRequest("POST", url, r)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := idp.Client.Do(req)
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
