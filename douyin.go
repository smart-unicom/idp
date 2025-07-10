// 抖音OAuth2登录提供者实现
// 支持抖音网页登录和用户信息获取
package idp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

// DouyinIdProvider 抖音登录提供者
// 实现抖音OAuth2登录功能
type DouyinIdProvider struct {
	Client *http.Client    // HTTP客户端
	Config *oauth2.Config // OAuth2配置
}

// NewDouyinIdProvider 创建抖音登录提供者实例
// 参数:
//   - clientId: 抖音应用的Client Key
//   - clientSecret: 抖音应用的Client Secret
//   - redirectUrl: OAuth2重定向URL
// 返回:
//   - *DouyinIdProvider: 抖音登录提供者实例
func NewDouyinIdProvider(clientId string, clientSecret string, redirectUrl string) *DouyinIdProvider {
	idp := &DouyinIdProvider{}
	idp.Config = idp.getConfig(clientId, clientSecret, redirectUrl)
	return idp
}

// SetHttpClient 设置HTTP客户端
// 参数:
//   - client: HTTP客户端实例
func (idp *DouyinIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig 获取抖音OAuth2配置
// 参数:
//   - clientId: 抖音应用的Client Key
//   - clientSecret: 抖音应用的Client Secret
//   - redirectUrl: OAuth2重定向URL
// 返回:
//   - *oauth2.Config: OAuth2配置实例
func (idp *DouyinIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL: "https://open.douyin.com/oauth/access_token",
		AuthURL:  "https://open.douyin.com/platform/oauth/connect",
	}

	config := &oauth2.Config{
		Scopes:       []string{"user_info"},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

// get more details via: https://open.douyin.com/platform/doc?doc=docs/openapi/account-permission/get-access-token
/*
{
  "data": {
    "access_token": "access_token",
    "description": "",
    "error_code": "0",
    "expires_in": "86400",
    "open_id": "aaa-bbb-ccc",
    "refresh_expires_in": "86400",
    "refresh_token": "refresh_token",
    "scope": "user_info"
  },
  "message": "<nil>"
}
*/

// DouyinTokenResp 抖音令牌响应结构体
type DouyinTokenResp struct {
	Data struct {
		AccessToken  string `json:"access_token"`  // 访问令牌
		ExpiresIn    int64  `json:"expires_in"`    // 访问令牌的有效时间，单位是秒
		OpenId       string `json:"open_id"`       // 用户OpenID
		RefreshToken string `json:"refresh_token"` // 刷新令牌
		Scope        string `json:"scope"`         // 授权范围
	} `json:"data"`    // 响应数据
	Message string `json:"message"` // 响应消息
}

// GetToken 通过授权码获取抖音访问令牌
// 参数:
//   - code: 抖音返回的授权码
// 返回:
//   - *oauth2.Token: OAuth2访问令牌
//   - error: 错误信息
// 详细文档: https://open.douyin.com/platform/doc?doc=docs/openapi/account-permission/get-access-token
func (idp *DouyinIdProvider) GetToken(code string) (*oauth2.Token, error) {
	payload := url.Values{}
	payload.Set("code", code)
	payload.Set("grant_type", "authorization_code")
	payload.Set("client_key", idp.Config.ClientID)
	payload.Set("client_secret", idp.Config.ClientSecret)
	resp, err := idp.Client.PostForm(idp.Config.Endpoint.TokenURL, payload)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	tokenResp := &DouyinTokenResp{}
	err = json.Unmarshal(data, tokenResp)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal token response: %s", err.Error())
	}

	token := &oauth2.Token{
		AccessToken:  tokenResp.Data.AccessToken,
		RefreshToken: tokenResp.Data.RefreshToken,
		Expiry:       time.Unix(time.Now().Unix()+tokenResp.Data.ExpiresIn, 0),
	}

	raw := make(map[string]interface{})
	raw["open_id"] = tokenResp.Data.OpenId
	token = token.WithExtra(raw)

	return token, nil
}

// get more details via: https://open.douyin.com/platform/doc?doc=docs/openapi/account-management/get-account-open-info
/*
{
  "data": {
    "avatar": "https://example.com/x.jpeg",
    "city": "上海",
    "country": "中国",
    "description": "",
    "e_account_role": "<nil>",
    "error_code": "0",
    "gender": "<nil>",
    "nickname": "张伟",
    "open_id": "0da22181-d833-447f-995f-1beefea5bef3",
    "province": "上海",
    "union_id": "1ad4e099-4a0c-47d1-a410-bffb4f2f64a4"
  }
}
*/

// DouyinUserInfo 抖音用户信息结构体
type DouyinUserInfo struct {
	Data struct {
		Avatar  string `json:"avatar"`   // 用户头像URL
		City    string `json:"city"`     // 城市
		Country string `json:"country"`  // 国家
		// 性别：0->未知, 1->男, 2->女
		Gender   int64  `json:"gender"`   // 性别
		Nickname string `json:"nickname"` // 用户昵称
		OpenId   string `json:"open_id"`  // 用户OpenID
		Province string `json:"province"` // 省份
	} `json:"data"` // 用户信息数据
}

// GetUserInfo 通过访问令牌获取抖音用户信息
// 参数:
//   - token: OAuth2访问令牌
// 返回:
//   - *UserInfo: 标准化用户信息
//   - error: 错误信息
func (idp *DouyinIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	body := &struct {
		AccessToken string `json:"access_token"`
		OpenId      string `json:"open_id"`
	}{token.AccessToken, token.Extra("open_id").(string)}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", "https://open.douyin.com/oauth/userinfo/", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Add("access-token", token.AccessToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := idp.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var douyinUserInfo DouyinUserInfo
	err = json.Unmarshal(respBody, &douyinUserInfo)
	if err != nil {
		return nil, err
	}

	userInfo := UserInfo{
		Id:          douyinUserInfo.Data.OpenId,
		Username:    douyinUserInfo.Data.Nickname,
		DisplayName: douyinUserInfo.Data.Nickname,
		AvatarUrl:   douyinUserInfo.Data.Avatar,
	}
	return &userInfo, nil
}
