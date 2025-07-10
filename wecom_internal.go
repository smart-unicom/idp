// 企业微信内部应用登录提供者实现
// 支持企业微信内部应用授权登录和用户信息获取
package idp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// WeComInternalIdProvider 企业微信内部应用登录提供者
// 实现企业微信内部应用OAuth2登录功能
type WeComInternalIdProvider struct {
	Client *http.Client   // HTTP客户端
	Config *oauth2.Config // OAuth2配置
}

// NewWeComInternalIdProvider 创建企业微信内部应用登录提供者实例
// 参数:
//   - clientId: 企业微信应用的CorpId
//   - clientSecret: 企业微信应用的CorpSecret
//   - redirectUrl: OAuth2重定向URL
// 返回:
//   - *WeComInternalIdProvider: 企业微信内部应用登录提供者实例
func NewWeComInternalIdProvider(clientId string, clientSecret string, redirectUrl string) *WeComInternalIdProvider {
	idp := &WeComInternalIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

// SetHttpClient 设置HTTP客户端
// 参数:
//   - client: HTTP客户端实例
func (idp *WeComInternalIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig 获取企业微信内部应用OAuth2配置
// 参数:
//   - clientId: 企业微信应用的CorpId
//   - clientSecret: 企业微信应用的CorpSecret
//   - redirectUrl: OAuth2重定向URL
// 返回:
//   - *oauth2.Config: OAuth2配置实例
func (idp *WeComInternalIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	config := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

// WecomInterToken 企业微信内部应用访问令牌结构体
type WecomInterToken struct {
	Errcode     int    `json:"errcode"`      // 错误码
	Errmsg      string `json:"errmsg"`       // 错误信息
	AccessToken string `json:"access_token"` // 访问令牌
	ExpiresIn   int    `json:"expires_in"`   // 访问令牌的有效时间，单位是秒
}

// GetToken 通过授权码获取企业微信内部应用访问令牌
// 参数:
//   - code: 企业微信返回的授权码
// 返回:
//   - *oauth2.Token: OAuth2访问令牌
//   - error: 错误信息
// 详细文档: https://developer.work.weixin.qq.com/document/path/91039
func (idp *WeComInternalIdProvider) GetToken(code string) (*oauth2.Token, error) {
	pTokenParams := &struct {
		CorpId     string `json:"corpid"`
		Corpsecret string `json:"corpsecret"`
	}{idp.Config.ClientID, idp.Config.ClientSecret}
	resp, err := idp.Client.Get(fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", pTokenParams.CorpId, pTokenParams.Corpsecret))
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	pToken := &WecomInterToken{}
	err = json.Unmarshal(data, pToken)
	if err != nil {
		return nil, err
	}
	if pToken.Errcode != 0 {
		return nil, fmt.Errorf("pToken.Errcode = %d, pToken.Errmsg = %s", pToken.Errcode, pToken.Errmsg)
	}

	token := &oauth2.Token{
		AccessToken: pToken.AccessToken,
		Expiry:      time.Unix(time.Now().Unix()+int64(pToken.ExpiresIn), 0),
	}

	raw := make(map[string]interface{})
	raw["code"] = code
	token = token.WithExtra(raw)

	return token, nil
}

// WecomInternalUserResp 企业微信内部用户响应结构体
type WecomInternalUserResp struct {
	Errcode int    `json:"errcode"` // 错误码
	Errmsg  string `json:"errmsg"`  // 错误信息
	UserId  string `json:"UserId"`  // 用户ID
	OpenId  string `json:"OpenId"`  // 开放ID
}

// WecomInternalUserInfo 企业微信内部用户信息结构体
type WecomInternalUserInfo struct {
	Errcode int    `json:"errcode"`     // 错误码
	Errmsg  string `json:"errmsg"`      // 错误信息
	Name    string `json:"name"`        // 用户姓名
	Email   string `json:"email"`       // 邮箱地址
	Avatar  string `json:"avatar"`      // 头像URL
	OpenId  string `json:"open_userid"` // 开放用户ID
	UserId  string `json:"userid"`      // 用户ID
}

// GetUserInfo 获取企业微信内部用户信息
// 参数:
//   - token: OAuth2访问令牌
// 返回:
//   - *UserInfo: 标准化用户信息
//   - error: 错误信息
func (idp *WeComInternalIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	// Get userid first
	accessToken := token.AccessToken
	code := token.Extra("code").(string)
	resp, err := idp.Client.Get(fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/getuserinfo?access_token=%s&code=%s", accessToken, code))
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	userResp := &WecomInternalUserResp{}
	err = json.Unmarshal(data, userResp)
	if err != nil {
		return nil, err
	}
	if userResp.Errcode != 0 {
		return nil, fmt.Errorf("userIdResp.Errcode = %d, userIdResp.Errmsg = %s", userResp.Errcode, userResp.Errmsg)
	}
	if userResp.OpenId != "" {
		return nil, fmt.Errorf("not an internal user")
	}
	// Use userid and accesstoken to get user information
	resp, err = idp.Client.Get(fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/get?access_token=%s&userid=%s", accessToken, userResp.UserId))
	if err != nil {
		return nil, err
	}

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	infoResp := &WecomInternalUserInfo{}
	err = json.Unmarshal(data, infoResp)
	if err != nil {
		return nil, err
	}
	if infoResp.Errcode != 0 {
		return nil, fmt.Errorf("userInfoResp.errcode = %d, userInfoResp.errmsg = %s", infoResp.Errcode, infoResp.Errmsg)
	}
	userInfo := UserInfo{
		Id:          infoResp.UserId,
		Username:    infoResp.Name,
		DisplayName: infoResp.Name,
		Email:       infoResp.Email,
		AvatarUrl:   infoResp.Avatar,
	}

	if userInfo.Id == "" {
		userInfo.Id = userInfo.Username
	}

	return &userInfo, nil
}
