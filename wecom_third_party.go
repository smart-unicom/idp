// 企业微信第三方应用登录提供者实现
// 支持企业微信第三方应用授权登录和用户信息获取
package idp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// WeComIdProvider 企业微信第三方应用登录提供者
// 实现企业微信第三方应用OAuth2登录功能
type WeComIdProvider struct {
	Client *http.Client   // HTTP客户端
	Config *oauth2.Config // OAuth2配置
}

// NewWeComIdProvider 创建企业微信第三方应用登录提供者实例
// 参数:
//   - clientId: 企业微信第三方应用的CorpId
//   - clientSecret: 企业微信第三方应用的ProviderSecret
//   - redirectUrl: OAuth2重定向URL
// 返回:
//   - *WeComIdProvider: 企业微信第三方应用登录提供者实例
func NewWeComIdProvider(clientId string, clientSecret string, redirectUrl string) *WeComIdProvider {
	idp := &WeComIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

// SetHttpClient 设置HTTP客户端
// 参数:
//   - client: HTTP客户端实例
func (idp *WeComIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig 获取企业微信第三方应用OAuth2配置
// 参数:
//   - clientId: 企业微信第三方应用的CorpId
//   - clientSecret: 企业微信第三方应用的ProviderSecret
//   - redirectUrl: OAuth2重定向URL
// 返回:
//   - *oauth2.Config: OAuth2配置实例
func (idp *WeComIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
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

// WeComProviderToken 企业微信第三方应用访问令牌结构体
type WeComProviderToken struct {
	Errcode             int    `json:"errcode"`              // 错误码
	Errmsg              string `json:"errmsg"`               // 错误信息
	ProviderAccessToken string `json:"provider_access_token"` // 第三方应用访问令牌
	ExpiresIn           int    `json:"expires_in"`           // 访问令牌的有效时间，单位是秒
}

// GetToken 通过授权码获取企业微信第三方应用访问令牌
// 参数:
//   - code: 企业微信返回的授权码
// 返回:
//   - *oauth2.Token: OAuth2访问令牌
//   - error: 错误信息
// 详细文档: https://work.weixin.qq.com/api/doc/90001/90143/91125
func (idp *WeComIdProvider) GetToken(code string) (*oauth2.Token, error) {
	pTokenParams := &struct {
		CorpId         string `json:"corpid"`
		ProviderSecret string `json:"provider_secret"`
	}{idp.Config.ClientID, idp.Config.ClientSecret}
	data, err := idp.postWithBody(pTokenParams, "https://qyapi.weixin.qq.com/cgi-bin/service/get_provider_token")
	if err != nil {
		return nil, err
	}
	pToken := &WeComProviderToken{}
	err = json.Unmarshal(data, pToken)
	if err != nil {
		return nil, err
	}
	if pToken.Errcode != 0 {
		return nil, fmt.Errorf("pToken.Errcode = %d, pToken.Errmsg = %s", pToken.Errcode, pToken.Errmsg)
	}

	token := &oauth2.Token{
		AccessToken: pToken.ProviderAccessToken,
		Expiry:      time.Unix(time.Now().Unix()+int64(pToken.ExpiresIn), 0),
	}

	raw := make(map[string]interface{})
	raw["code"] = code
	token = token.WithExtra(raw)

	return token, nil
}

type WeComUserInfo struct {
	Errcode  int    `json:"errcode"`
	Errmsg   string `json:"errmsg"`
	Usertype int    `json:"usertype"`
	UserInfo struct {
		Userid     string `json:"userid"`
		OpenUserid string `json:"open_userid"`
		Name       string `json:"name"`
		Avatar     string `json:"avatar"`
	} `json:"user_info"`
	CorpInfo struct {
		Corpid string `json:"corpid"`
	} `json:"corp_info"`
	Agent []struct {
		Agentid  int `json:"agentid"`
		AuthType int `json:"auth_type"`
	} `json:"agent"`
	AuthInfo struct {
		Department []struct {
			Id       int  `json:"id"`
			Writable bool `json:"writable"`
		} `json:"department"`
	} `json:"auth_info"`
}

// GetUserInfo use WeComProviderToken gotten before return WeComUserInfo
// get more detail via: https://work.weixin.qq.com/api/doc/90001/90143/91125
func (idp *WeComIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	accessToken := token.AccessToken
	code := token.Extra("code").(string)

	requestBody := &struct {
		AuthCode string `json:"auth_code"`
	}{code}
	data, err := idp.postWithBody(requestBody, fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/service/get_login_info?access_token=%s", accessToken))
	if err != nil {
		return nil, err
	}

	wecomUserInfo := WeComUserInfo{}
	err = json.Unmarshal(data, &wecomUserInfo)
	if err != nil {
		return nil, err
	}
	if wecomUserInfo.Errcode != 0 {
		return nil, fmt.Errorf("wecomUserInfo.Errcode = %d, wecomUserInfo.Errmsg = %s", wecomUserInfo.Errcode, wecomUserInfo.Errmsg)
	}

	userInfo := UserInfo{
		Id:          wecomUserInfo.UserInfo.OpenUserid,
		Username:    wecomUserInfo.UserInfo.Name,
		DisplayName: wecomUserInfo.UserInfo.Name,
		AvatarUrl:   wecomUserInfo.UserInfo.Avatar,
	}
	return &userInfo, nil
}

func (idp *WeComIdProvider) postWithBody(body interface{}, url string) ([]byte, error) {
	bs, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	r := strings.NewReader(string(bs))
	resp, err := idp.Client.Post(url, "application/json;charset=UTF-8", r)
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
