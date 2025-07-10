// 微信小程序登录提供者实现
// 支持微信小程序授权登录和会话管理
package idp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

// WeChatMiniProgramIdProvider 微信小程序登录提供者
// 实现微信小程序授权登录功能
type WeChatMiniProgramIdProvider struct {
	Client *http.Client   // HTTP客户端
	Config *oauth2.Config // OAuth2配置
}

// NewWeChatMiniProgramIdProvider 创建微信小程序登录提供者实例
// 参数:
//   - clientId: 微信小程序的AppId
//   - clientSecret: 微信小程序的AppSecret
//
// 返回:
//   - *WeChatMiniProgramIdProvider: 微信小程序登录提供者实例
func NewWeChatMiniProgramIdProvider(clientId string, clientSecret string) *WeChatMiniProgramIdProvider {
	idp := &WeChatMiniProgramIdProvider{}

	config := idp.getConfig(clientId, clientSecret)
	idp.Config = config
	idp.Client = &http.Client{}
	return idp
}

// SetHttpClient 设置HTTP客户端
// 参数:
//   - client: HTTP客户端实例
func (idp *WeChatMiniProgramIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig 获取微信小程序OAuth2配置
// 参数:
//   - clientId: 微信小程序的AppId
//   - clientSecret: 微信小程序的AppSecret
//
// 返回:
//   - *oauth2.Config: OAuth2配置实例
func (idp *WeChatMiniProgramIdProvider) getConfig(clientId string, clientSecret string) *oauth2.Config {
	config := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
	}

	return config
}

// WeChatMiniProgramSessionResponse 微信小程序会话响应结构体
// 包含从微信小程序API获取的会话信息
type WeChatMiniProgramSessionResponse struct {
	Openid     string `json:"openid"`      // 用户唯一标识
	SessionKey string `json:"session_key"` // 会话密钥
	Unionid    string `json:"unionid"`     // 用户在开放平台的唯一标识符，若当前小程序已绑定到微信开放平台帐号下会返回
	Errcode    int    `json:"errcode"`     // 错误码
	Errmsg     string `json:"errmsg"`      // 错误信息
}

// GetSessionByCode 通过授权码获取微信小程序会话信息
// 参数:
//   - code: 微信小程序返回的授权码
//
// 返回:
//   - *WeChatMiniProgramSessionResponse: 会话响应信息
//   - error: 错误信息
//
// 详细文档: https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/user-login/code2Session.html
func (idp *WeChatMiniProgramIdProvider) GetSessionByCode(code string) (*WeChatMiniProgramSessionResponse, error) {
	sessionUri := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		idp.Config.ClientID, idp.Config.ClientSecret, code)
	sessionResponse, err := idp.Client.Get(sessionUri)
	if err != nil {
		return nil, err
	}
	defer sessionResponse.Body.Close()
	data, err := io.ReadAll(sessionResponse.Body)
	if err != nil {
		return nil, err
	}
	var session WeChatMiniProgramSessionResponse
	err = json.Unmarshal(data, &session)
	if err != nil {
		return nil, err
	}
	if session.Errcode != 0 {
		return nil, fmt.Errorf("err: %s", session.Errmsg)
	}
	return &session, nil
}
