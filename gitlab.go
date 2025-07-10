// GitLab OAuth2登录提供者实现
// 支持GitLab网页登录和用户信息获取
package idp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/oauth2"
)

// GitlabIdProvider GitLab登录提供者
// 实现GitLab OAuth2登录功能
type GitlabIdProvider struct {
	Client *http.Client    // HTTP客户端
	Config *oauth2.Config // OAuth2配置
}

// NewGitlabIdProvider 创建GitLab登录提供者实例
// 参数:
//   - clientId: GitLab应用的Client ID
//   - clientSecret: GitLab应用的Client Secret
//   - redirectUrl: OAuth2重定向URL
// 返回:
//   - *GitlabIdProvider: GitLab登录提供者实例
func NewGitlabIdProvider(clientId string, clientSecret string, redirectUrl string) *GitlabIdProvider {
	idp := &GitlabIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

// SetHttpClient 设置HTTP客户端
// 参数:
//   - client: HTTP客户端实例
func (idp *GitlabIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig 获取GitLab OAuth2配置
// 参数:
//   - clientId: GitLab应用的Client ID
//   - clientSecret: GitLab应用的Client Secret
//   - redirectUrl: OAuth2重定向URL
// 返回:
//   - *oauth2.Config: OAuth2配置实例
func (idp *GitlabIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL: "https://gitlab.com/oauth/token",
	}

	config := &oauth2.Config{
		Scopes:       []string{"read_user+profile"},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

// GitlabProviderToken GitLab访问令牌结构体
type GitlabProviderToken struct {
	AccessToken  string `json:"access_token"`  // 访问令牌
	TokenType    string `json:"token_type"`    // 令牌类型
	ExpiresIn    int    `json:"expires_in"`    // 访问令牌的有效时间，单位是秒
	RefreshToken string `json:"refresh_token"` // 刷新令牌
	CreatedAt    int    `json:"created_at"`    // 创建时间
}

// GetToken 通过授权码获取GitLab访问令牌
// 参数:
//   - code: GitLab返回的授权码
// 返回:
//   - *oauth2.Token: OAuth2访问令牌
//   - error: 错误信息
// 详细文档: https://docs.gitlab.com/ee/api/oauth2.html
func (idp *GitlabIdProvider) GetToken(code string) (*oauth2.Token, error) {
	params := url.Values{}
	params.Add("grant_type", "authorization_code")
	params.Add("client_id", idp.Config.ClientID)
	params.Add("client_secret", idp.Config.ClientSecret)
	params.Add("code", code)
	params.Add("redirect_uri", idp.Config.RedirectURL)

	accessTokenUrl := fmt.Sprintf("%s?%s", idp.Config.Endpoint.TokenURL, params.Encode())
	resp, err := idp.Client.Post(accessTokenUrl, "application/json;charset=UTF-8", nil)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	gtoken := &GitlabProviderToken{}
	if err = json.Unmarshal(data, gtoken); err != nil {
		return nil, err
	}

	// gtoken.ExpiresIn always returns 0, so we set Expiry=7200 to avoid verification errors.
	token := &oauth2.Token{
		AccessToken:  gtoken.AccessToken,
		TokenType:    gtoken.TokenType,
		RefreshToken: gtoken.RefreshToken,
		Expiry:       time.Unix(time.Now().Unix()+int64(7200), 0),
	}

	return token, nil
}

// GitlabUserInfo GitLab用户信息响应结构体
// 包含从GitLab API获取的完整用户信息
type GitlabUserInfo struct {
	Id              int         `json:"id"`                // 用户ID
	Name            string      `json:"name"`              // 真实姓名
	Username        string      `json:"username"`          // 用户名
	State           string      `json:"state"`             // 用户状态
	AvatarUrl       string      `json:"avatar_url"`        // 头像URL
	WebUrl          string      `json:"web_url"`           // 用户主页URL
	CreatedAt       time.Time   `json:"created_at"`        // 创建时间
	Bio             string      `json:"bio"`               // 个人简介
	BioHtml         string      `json:"bio_html"`          // 个人简介HTML
	Location        string      `json:"location"`          // 位置
	PublicEmail     string      `json:"public_email"`      // 公开邮箱
	Skype           string      `json:"skype"`             // Skype账号
	Linkedin        string      `json:"linkedin"`          // LinkedIn账号
	Twitter         string      `json:"twitter"`           // Twitter账号
	WebsiteUrl      string      `json:"website_url"`       // 个人网站URL
	Organization    string      `json:"organization"`      // 组织
	JobTitle        string      `json:"job_title"`         // 职位
	Pronouns        interface{} `json:"pronouns"`          // 代词
	Bot             bool        `json:"bot"`               // 是否为机器人
	WorkInformation interface{} `json:"work_information"`  // 工作信息
	Followers       int         `json:"followers"`         // 关注者数
	Following       int         `json:"following"`         // 关注数
	LastSignInAt    time.Time   `json:"last_sign_in_at"`   // 最后登录时间
	ConfirmedAt     time.Time   `json:"confirmed_at"`      // 确认时间
	LastActivityOn  string      `json:"last_activity_on"`  // 最后活动日期
	Email           string      `json:"email"`             // 邮箱地址
	ThemeId         int         `json:"theme_id"`          // 主题ID
	ColorSchemeId   int         `json:"color_scheme_id"`   // 颜色方案ID
	ProjectsLimit   int         `json:"projects_limit"`    // 项目限制
	CurrentSignInAt time.Time   `json:"current_sign_in_at"` // 当前登录时间
	Identities      []struct {  // 身份信息
		Provider       string      `json:"provider"`        // 提供者
		ExternUid      string      `json:"extern_uid"`      // 外部UID
		SamlProviderId interface{} `json:"saml_provider_id"` // SAML提供者ID
	} `json:"identities"`
	CanCreateGroup                 bool        `json:"can_create_group"`                  // 是否可创建组
	CanCreateProject               bool        `json:"can_create_project"`                // 是否可创建项目
	TwoFactorEnabled               bool        `json:"two_factor_enabled"`                // 是否启用双因子认证
	External                       bool        `json:"external"`                          // 是否为外部用户
	PrivateProfile                 bool        `json:"private_profile"`                   // 是否为私有档案
	CommitEmail                    string      `json:"commit_email"`                      // 提交邮箱
	SharedRunnersMinutesLimit      interface{} `json:"shared_runners_minutes_limit"`       // 共享运行器分钟限制
	ExtraSharedRunnersMinutesLimit interface{} `json:"extra_shared_runners_minutes_limit"` // 额外共享运行器分钟限制
}

// GetUserInfo 获取GitLab用户信息
// 参数:
//   - token: OAuth2访问令牌
// 返回:
//   - *UserInfo: 标准化用户信息
//   - error: 错误信息
func (idp *GitlabIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	resp, err := idp.Client.Get("https://gitlab.com/api/v4/user?access_token=" + token.AccessToken)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	guser := GitlabUserInfo{}
	if err = json.Unmarshal(data, &guser); err != nil {
		return nil, err
	}

	userInfo := UserInfo{
		Id:          strconv.Itoa(guser.Id),
		Username:    guser.Username,
		DisplayName: guser.Name,
		AvatarUrl:   guser.AvatarUrl,
		Email:       guser.Email,
	}
	return &userInfo, nil
}
