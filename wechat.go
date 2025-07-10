// 微信OAuth2登录提供者实现
// 支持微信网页登录和公众号扫码登录
package idp

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/skip2/go-qrcode"
	"golang.org/x/oauth2"
)

// 微信登录缓存相关变量
var (
	WechatCacheMap map[string]WechatCacheMapValue // 微信登录缓存映射
	Lock           sync.RWMutex                   // 读写锁
)

// WeChatIdProvider 微信登录提供者
// 实现微信OAuth2登录功能
type WeChatIdProvider struct {
	Client *http.Client   // HTTP客户端
	Config *oauth2.Config // OAuth2配置
}

// WechatCacheMapValue 微信缓存值结构体
// 用于临时存储扫码登录状态
type WechatCacheMapValue struct {
	IsScanned     bool   // 是否已扫码
	WechatUnionId string // 微信UnionId
}

// NewWeChatIdProvider 创建微信登录提供者实例
// 参数:
//   - clientId: 微信应用的AppId
//   - clientSecret: 微信应用的AppSecret
//   - redirectUrl: OAuth2重定向URL
//
// 返回:
//   - *WeChatIdProvider: 微信登录提供者实例
func NewWeChatIdProvider(clientId string, clientSecret string, redirectUrl string) *WeChatIdProvider {
	idp := &WeChatIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

// SetHttpClient 设置HTTP客户端
// 参数:
//   - client: HTTP客户端实例
func (idp *WeChatIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig 获取微信OAuth2配置
// 参数:
//   - clientId: 微信应用的AppId
//   - clientSecret: 微信应用的AppSecret
//   - redirectUrl: OAuth2重定向URL
//
// 返回:
//   - *oauth2.Config: OAuth2配置实例
func (idp *WeChatIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL: "https://graph.qq.com/oauth2.0/token",
	}

	config := &oauth2.Config{
		Scopes:       []string{"snsapi_login"},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

// WechatAccessToken 微信访问令牌响应结构体
type WechatAccessToken struct {
	AccessToken  string `json:"access_token"`  // 接口调用凭证
	ExpiresIn    int64  `json:"expires_in"`    // access_token接口调用凭证超时时间，单位（秒）
	RefreshToken string `json:"refresh_token"` // 用户刷新access_token
	Openid       string `json:"openid"`        // 授权用户唯一标识
	Scope        string `json:"scope"`         // 用户授权的作用域，使用逗号（,）分隔
	Unionid      string `json:"unionid"`       // 只有在用户将公众号绑定到微信开放平台帐号后，才会出现该字段
}

// GetToken 通过授权码获取微信访问令牌
// 参数:
//   - code: 微信返回的授权码或特殊格式的票据
//
// 返回:
//   - *oauth2.Token: OAuth2访问令牌
//   - error: 错误信息
//
// 详细文档: https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html
func (idp *WeChatIdProvider) GetToken(code string) (*oauth2.Token, error) {
	if strings.HasPrefix(code, "wechat_oa:") {
		token := oauth2.Token{
			AccessToken: code,
			TokenType:   "WeChatAccessToken",
			Expiry:      time.Time{},
		}
		return &token, nil
	}

	params := url.Values{}
	params.Add("grant_type", "authorization_code")
	params.Add("appid", idp.Config.ClientID)
	params.Add("secret", idp.Config.ClientSecret)
	params.Add("code", code)

	accessTokenUrl := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?%s", params.Encode())
	tokenResponse, err := idp.Client.Get(accessTokenUrl)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(tokenResponse.Body)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(tokenResponse.Body)
	if err != nil {
		return nil, err
	}

	// {"errcode":40163,"errmsg":"code been used, rid: 6206378a-793424c0-2e4091cc"}
	if strings.Contains(buf.String(), "errcode") {
		return nil, fmt.Errorf(buf.String())
	}

	var wechatAccessToken WechatAccessToken
	if err = json.Unmarshal(buf.Bytes(), &wechatAccessToken); err != nil {
		return nil, err
	}

	token := oauth2.Token{
		AccessToken:  wechatAccessToken.AccessToken,
		TokenType:    "WeChatAccessToken",
		RefreshToken: wechatAccessToken.RefreshToken,
		Expiry:       time.Time{},
	}

	raw := make(map[string]string)
	raw["Openid"] = wechatAccessToken.Openid
	token.WithExtra(raw)

	return &token, nil
}

// WechatUserInfo 微信用户信息结构体
type WechatUserInfo struct {
	Openid     string   `json:"openid"`     // 用户的唯一标识
	Nickname   string   `json:"nickname"`   // 用户昵称
	Sex        int      `json:"sex"`        // 用户的性别，值为1时是男性，值为2时是女性，值为0时是未知
	Language   string   `json:"language"`   // 用户的语言，简体中文为zh_CN
	City       string   `json:"city"`       // 普通用户个人资料填写的城市
	Province   string   `json:"province"`   // 用户个人资料填写的省份
	Country    string   `json:"country"`    // 国家，如中国为CN
	Headimgurl string   `json:"headimgurl"` // 用户头像，最后一个数值代表正方形头像大小（有0、46、64、96、132数值可选，0代表640*640正方形头像），用户没有头像时该项为空
	Privilege  []string `json:"privilege"`  // 用户特权信息，json 数组，如微信沃卡用户为（chinaunicom）
	Unionid    string   `json:"unionid"`    // 只有在用户将公众号绑定到微信开放平台帐号后，才会出现该字段
}

// GetUserInfo 通过访问令牌获取微信用户信息
// 参数:
//   - token: OAuth2访问令牌
//
// 返回:
//   - *UserInfo: 标准化用户信息
//   - error: 错误信息
//
// 详细文档: https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Authorized_Interface_Calling_UnionId.html
func (idp *WeChatIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	var wechatUserInfo WechatUserInfo
	accessToken := token.AccessToken

	if strings.HasPrefix(accessToken, "wechat_oa:") {
		Lock.RLock()
		mapValue, ok := WechatCacheMap[accessToken[10:]]
		Lock.RUnlock()

		if !ok || mapValue.WechatUnionId == "" {
			return nil, fmt.Errorf("error ticket")
		}

		Lock.Lock()
		delete(WechatCacheMap, accessToken[10:])
		Lock.Unlock()

		userInfo := UserInfo{
			Id:          mapValue.WechatUnionId,
			Username:    "wx_user_" + mapValue.WechatUnionId,
			DisplayName: "wx_user_" + mapValue.WechatUnionId,
			AvatarUrl:   "",
		}
		return &userInfo, nil
	}

	openid := token.Extra("Openid")

	userInfoUrl := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s", accessToken, openid)
	resp, err := idp.Client.Get(userInfoUrl)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	if err = json.Unmarshal(buf.Bytes(), &wechatUserInfo); err != nil {
		return nil, err
	}

	id := wechatUserInfo.Unionid
	if id == "" {
		id = wechatUserInfo.Openid
	}

	extra := make(map[string]string)
	extra["wechat_unionid"] = wechatUserInfo.Openid
	// For WeChat, different appId corresponds to different openId
	extra[BuildWechatOpenIdKey(idp.Config.ClientID)] = wechatUserInfo.Openid
	userInfo := UserInfo{
		Id:          id,
		Username:    wechatUserInfo.Nickname,
		DisplayName: wechatUserInfo.Nickname,
		AvatarUrl:   wechatUserInfo.Headimgurl,
		Extra:       extra,
	}
	return &userInfo, nil
}

// BuildWechatOpenIdKey 构建微信OpenId键
// 参数:
//   - appId: 微信应用AppId
//
// 返回:
//   - string: 格式化的键值
func BuildWechatOpenIdKey(appId string) string {
	return fmt.Sprintf("wechat_openid_%s", appId)
}

// GetWechatOfficialAccountAccessToken 获取微信公众号访问令牌
// 参数:
//   - clientId: 微信公众号AppId
//   - clientSecret: 微信公众号AppSecret
//
// 返回:
//   - string: 访问令牌
//   - string: 错误消息
//   - error: 错误信息
func GetWechatOfficialAccountAccessToken(clientId string, clientSecret string) (string, string, error) {
	accessTokenUrl := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", clientId, clientSecret)
	request, err := http.NewRequest("GET", accessTokenUrl, nil)
	if err != nil {
		return "", "", err
	}

	client := new(http.Client)
	resp, err := client.Do(request)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var data struct {
		ExpireIn    int    `json:"expires_in"`
		AccessToken string `json:"access_token"`
		ErrCode     int    `json:"errcode"`
		Errmsg      string `json:"errmsg"`
	}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		return "", "", err
	}

	return data.AccessToken, data.Errmsg, nil
}

// GetWechatOfficialAccountQRCode 获取微信公众号二维码
// 参数:
//   - clientId: 微信公众号AppId
//   - clientSecret: 微信公众号AppSecret
//   - providerId: 提供者ID，用作场景字符串
//
// 返回:
//   - string: Base64编码的二维码图片
//   - string: 二维码票据
//   - error: 错误信息
func GetWechatOfficialAccountQRCode(clientId string, clientSecret string, providerId string) (string, string, error) {
	accessToken, errMsg, err := GetWechatOfficialAccountAccessToken(clientId, clientSecret)
	if err != nil {
		return "", "", err
	}

	if errMsg != "" {
		return "", "", fmt.Errorf("Fail to fetch WeChat QRcode: %s", errMsg)
	}

	client := new(http.Client)

	weChatEndpoint := "https://api.weixin.qq.com/cgi-bin/qrcode/create"
	qrCodeUrl := fmt.Sprintf("%s?access_token=%s", weChatEndpoint, accessToken)
	params := fmt.Sprintf(`{"expire_seconds": 3600, "action_name": "QR_STR_SCENE", "action_info": {"scene": {"scene_str": "%s"}}}`, providerId)

	bodyData := bytes.NewReader([]byte(params))
	requeset, err := http.NewRequest("POST", qrCodeUrl, bodyData)
	if err != nil {
		return "", "", err
	}

	resp, err := client.Do(requeset)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	var data struct {
		Ticket        string `json:"ticket"`
		ExpireSeconds int    `json:"expire_seconds"`
		URL           string `json:"url"`
	}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		return "", "", err
	}

	var png []byte
	png, err = qrcode.Encode(data.URL, qrcode.Medium, 256)
	base64Image := base64.StdEncoding.EncodeToString(png)
	return base64Image, data.Ticket, nil
}

// VerifyWechatSignature 验证微信签名
// 参数:
//   - token: 微信公众号配置的Token
//   - nonce: 随机数
//   - timestamp: 时间戳
//   - signature: 微信加密签名
//
// 返回:
//   - bool: 验证结果
func VerifyWechatSignature(token string, nonce string, timestamp string, signature string) bool {
	// verify the signature
	tmpArr := sort.StringSlice{token, timestamp, nonce}
	sort.Sort(tmpArr)

	tmpStr := ""
	for _, str := range tmpArr {
		tmpStr = tmpStr + str
	}

	b := sha1.Sum([]byte(tmpStr))
	res := hex.EncodeToString(b[:])
	return res == signature
}
