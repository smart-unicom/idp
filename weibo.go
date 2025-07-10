// 新浪微博OAuth2登录提供者实现
// 支持新浪微博网页登录和用户信息获取
package idp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/oauth2"
)

// WeiBoIdProvider 新浪微博登录提供者
// 实现新浪微博OAuth2登录功能
type WeiBoIdProvider struct {
	Client *http.Client   // HTTP客户端
	Config *oauth2.Config // OAuth2配置
}

// NewWeiBoIdProvider 创建新浪微博登录提供者实例
// 参数:
//   - clientId: 新浪微博应用的App Key
//   - clientSecret: 新浪微博应用的App Secret
//   - redirectUrl: OAuth2重定向URL
//
// 返回:
//   - *WeiBoIdProvider: 新浪微博登录提供者实例
func NewWeiBoIdProvider(clientId string, clientSecret string, redirectUrl string) *WeiBoIdProvider {
	idp := &WeiBoIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

// SetHttpClient 设置HTTP客户端
// 参数:
//   - client: HTTP客户端实例
func (idp *WeiBoIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig 获取新浪微博OAuth2配置
// 参数:
//   - clientId: 新浪微博应用的App Key
//   - clientSecret: 新浪微博应用的App Secret
//   - redirectUrl: OAuth2重定向URL
//
// 返回:
//   - *oauth2.Config: OAuth2配置实例
func (idp *WeiBoIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL: "https://api.weibo.com/oauth2/access_token",
	}

	config := &oauth2.Config{
		Scopes:       []string{""},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

// WeiboAccessToken 新浪微博访问令牌响应结构体
type WeiboAccessToken struct {
	AccessToken string `json:"access_token"` // 访问令牌
	ExpiresIn   int    `json:"expires_in"`   // 访问令牌的有效时间，单位是秒
	RemindIn    string `json:"remind_in"`    // 提醒时间（即将废弃，请使用expires_in）
	Uid         string `json:"uid"`          // 用户UID
}

// GetToken 通过授权码获取新浪微博访问令牌
// 参数:
//   - code: 新浪微博返回的授权码
//
// 返回:
//   - *oauth2.Token: OAuth2访问令牌
//   - error: 错误信息
//
// 详细文档: https://open.weibo.com/wiki/Oauth2/access_token
func (idp *WeiBoIdProvider) GetToken(code string) (*oauth2.Token, error) {
	params := url.Values{}
	params.Add("grant_type", "authorization_code")
	params.Add("client_id", idp.Config.ClientID)
	params.Add("client_secret", idp.Config.ClientSecret)
	params.Add("code", code)
	params.Add("redirect_uri", idp.Config.RedirectURL)

	// accessTokenUrl := fmt.Sprintf("%s?%s", idp.Config.Endpoint.TokenURL, params.Encode())
	resp, err := idp.Client.PostForm(idp.Config.Endpoint.TokenURL, params)
	// resp, err := idp.GetUrlResp(accessTokenUrl)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var weiboAccessToken WeiboAccessToken
	if err = json.Unmarshal(bs, &weiboAccessToken); err != nil {
		return nil, err
	}

	token := oauth2.Token{
		AccessToken: weiboAccessToken.AccessToken,
		TokenType:   "WeiboAccessToken",
		Expiry:      time.Unix(time.Now().Unix()+int64(weiboAccessToken.ExpiresIn), 0),
	}

	idp.Config.Scopes[0] = weiboAccessToken.Uid
	return &token, nil
}

// WeiboUserinfo 新浪微博用户信息结构体
// 包含新浪微博用户的详细信息
type WeiboUserinfo struct {
	Id              int    `json:"id"`                // 用户ID
	ScreenName      string `json:"screen_name"`       // 用户昵称
	Name            string `json:"name"`              // 用户真实姓名
	Province        string `json:"province"`          // 省份
	City            string `json:"city"`              // 城市
	Location        string `json:"location"`          // 地理位置
	Description     string `json:"description"`       // 个人描述
	Url             string `json:"url"`               // 用户博客地址
	ProfileImageUrl string `json:"profile_image_url"` // 头像地址
	Domain          string `json:"domain"`            // 用户个性化域名
	Gender          string `json:"gender"`            // 性别，m：男、f：女、n：未知
	FollowersCount  int    `json:"followers_count"`   // 粉丝数
	FriendsCount    int    `json:"friends_count"`     // 关注数
	StatusesCount   int    `json:"statuses_count"`    // 微博数
	FavouritesCount int    `json:"favourites_count"`  // 收藏数
	CreatedAt       string `json:"created_at"`        // 用户创建时间
	Following       bool   `json:"following"`         // 是否已关注
	AllowAllActMsg  bool   `json:"allow_all_act_msg"` // 是否允许所有人给我发私信
	GeoEnabled      bool   `json:"geo_enabled"`       // 是否允许标识用户的地理位置
	Verified        bool   `json:"verified"`          // 是否是微博认证用户
	Status          struct {
		CreatedAt           string        `json:"created_at"`              // 微博创建时间
		Id                  int64         `json:"id"`                      // 微博ID
		Text                string        `json:"text"`                    // 微博内容
		Source              string        `json:"source"`                  // 微博来源
		Favorited           bool          `json:"favorited"`               // 是否已收藏
		Truncated           bool          `json:"truncated"`               // 是否被截断
		InReplyToStatusId   string        `json:"in_reply_to_status_id"`   // 回复的微博ID
		InReplyToUserId     string        `json:"in_reply_to_user_id"`     // 回复的用户ID
		InReplyToScreenName string        `json:"in_reply_to_screen_name"` // 回复的用户昵称
		Geo                 interface{}   `json:"geo"`                     // 地理信息
		Mid                 string        `json:"mid"`                     // 微博MID
		Annotations         []interface{} `json:"annotations"`             // 元数据
		RepostsCount        int           `json:"reposts_count"`           // 转发数
		CommentsCount       int           `json:"comments_count"`          // 评论数
	} `json:"status"` // 用户最新一条微博信息
	AllowAllComment  bool   `json:"allow_all_comment"`  // 是否允许所有人对我的微博进行评论
	AvatarLarge      string `json:"avatar_large"`       // 用户大头像地址
	VerifiedReason   string `json:"verified_reason"`    // 认证原因
	FollowMe         bool   `json:"follow_me"`          // 该用户是否关注当前登录用户
	OnlineStatus     int    `json:"online_status"`      // 用户的在线状态
	BiFollowersCount int    `json:"bi_followers_count"` // 用户的互粉数
}

// GetUserInfo 通过访问令牌获取新浪微博用户信息
// 参数:
//   - token: OAuth2访问令牌
//
// 返回:
//   - *UserInfo: 标准化用户信息
//   - error: 错误信息
//
// 详细文档: https://open.weibo.com/wiki/2/users/show
func (idp *WeiBoIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	var weiboUserInfo WeiboUserinfo
	accessToken := token.AccessToken
	uid := idp.Config.Scopes[0]
	id, _ := strconv.Atoi(uid)

	userInfoUrl := fmt.Sprintf("https://api.weibo.com/2/users/show.json?access_token=%s&uid=%d", accessToken, id)
	resp, err := idp.GetUrlResp(userInfoUrl)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal([]byte(resp), &weiboUserInfo); err != nil {
		return nil, err
	}

	// weibo user email need to get separately through this url, need user authorization.
	e := struct {
		Email string `json:"email"`
	}{}
	emailUrl := fmt.Sprintf("https://api.weibo.com/2/account/profile/email.json?access_token=%s", accessToken)
	resp, err = idp.GetUrlResp(emailUrl)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal([]byte(resp), &e); err != nil {
		return nil, err
	}

	userInfo := UserInfo{
		Id:          strconv.Itoa(weiboUserInfo.Id),
		Username:    weiboUserInfo.Name,
		DisplayName: weiboUserInfo.Name,
		AvatarUrl:   weiboUserInfo.AvatarLarge,
		Email:       e.Email,
	}
	return &userInfo, nil
}

// GetUrlResp 发送HTTP GET请求并获取响应内容
// 参数:
//   - url: 请求URL
//
// 返回:
//   - string: 响应内容
//   - error: 错误信息
func (idp *WeiBoIdProvider) GetUrlResp(url string) (string, error) {
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
