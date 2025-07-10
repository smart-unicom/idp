// 支付宝OAuth2登录提供者实现
// 支持支付宝网页登录和RSA签名验证
package idp

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// AlipayIdProvider 支付宝登录提供者
// 实现支付宝OAuth2登录功能
type AlipayIdProvider struct {
	Client *http.Client   // HTTP客户端
	Config *oauth2.Config // OAuth2配置
}

// NewAlipayIdProvider 创建支付宝登录提供者实例
// 参数:
//   - clientId: 支付宝应用的AppId
//   - clientSecret: 支付宝应用的私钥
//   - redirectUrl: OAuth2重定向URL
//
// 返回:
//   - *AlipayIdProvider: 支付宝登录提供者实例
func NewAlipayIdProvider(clientId string, clientSecret string, redirectUrl string) *AlipayIdProvider {
	idp := &AlipayIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

// SetHttpClient 设置HTTP客户端
// 参数:
//   - client: HTTP客户端实例
func (idp *AlipayIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig 获取支付宝OAuth2配置
// 参数:
//   - clientId: 支付宝应用的AppId
//   - clientSecret: 支付宝应用的私钥
//   - redirectUrl: OAuth2重定向URL
//
// 返回:
//   - *oauth2.Config: OAuth2配置实例
func (idp *AlipayIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		AuthURL:  "https://openauth.alipay.com/oauth2/publicAppAuthorize.htm",
		TokenURL: "https://openapi.alipay.com/gateway.do",
	}

	config := &oauth2.Config{
		Scopes:       []string{"", ""},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

// AlipayAccessToken 支付宝访问令牌响应结构体
type AlipayAccessToken struct {
	Response AlipaySystemOauthTokenResponse `json:"alipay_system_oauth_token_response"` // 令牌响应数据
	Sign     string                         `json:"sign"`                               // 签名
}

// AlipaySystemOauthTokenResponse 支付宝系统OAuth令牌响应
type AlipaySystemOauthTokenResponse struct {
	AccessToken  string `json:"access_token"`   // 访问令牌
	AlipayUserId string `json:"alipay_user_id"` // 支付宝用户ID
	ExpiresIn    int    `json:"expires_in"`     // 访问令牌有效期，单位秒
	ReExpiresIn  int    `json:"re_expires_in"`  // 刷新令牌有效期，单位秒
	RefreshToken string `json:"refresh_token"`  // 刷新令牌
	UserId       string `json:"user_id"`        // 用户ID
}

// GetToken 通过授权码获取支付宝访问令牌
// 参数:
//   - code: 支付宝返回的授权码
//
// 返回:
//   - *oauth2.Token: OAuth2访问令牌
//   - error: 错误信息
//
// 详细文档: https://opendocs.alipay.com/apis/api_9/alipay.system.oauth.token
func (idp *AlipayIdProvider) GetToken(code string) (*oauth2.Token, error) {
	pTokenParams := &struct {
		ClientId  string `json:"app_id"`
		CharSet   string `json:"charset"`
		Code      string `json:"code"`
		GrantType string `json:"grant_type"`
		Method    string `json:"method"`
		SignType  string `json:"sign_type"`
		TimeStamp string `json:"timestamp"`
		Version   string `json:"version"`
	}{idp.Config.ClientID, "utf-8", code, "authorization_code", "alipay.system.oauth.token", "RSA2", time.Now().Format("2006-01-02 15:04:05"), "1.0"}

	data, err := idp.postWithBody(pTokenParams, idp.Config.Endpoint.TokenURL)
	if err != nil {
		return nil, err
	}

	pToken := &AlipayAccessToken{}
	err = json.Unmarshal(data, pToken)
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken: pToken.Response.AccessToken,
		Expiry:      time.Unix(time.Now().Unix()+int64(pToken.Response.ExpiresIn), 0),
	}
	return token, nil
}

// AlipayUserResponse 支付宝用户信息响应结构体
type AlipayUserResponse struct {
	AlipayUserInfoShareResponse AlipayUserInfoShareResponse `json:"alipay_user_info_share_response"` // 用户信息响应数据
	Sign                        string                      `json:"sign"`                            // 签名
}

// AlipayUserInfoShareResponse 支付宝用户信息共享响应
type AlipayUserInfoShareResponse struct {
	Code     string `json:"code"`      // 网关返回码
	Msg      string `json:"msg"`       // 网关返回码描述
	Avatar   string `json:"avatar"`    // 用户头像地址
	NickName string `json:"nick_name"` // 用户昵称
	UserId   string `json:"user_id"`   // 支付宝用户的userId
}

// GetUserInfo 通过访问令牌获取支付宝用户信息
// 参数:
//   - token: OAuth2访问令牌
//
// 返回:
//   - *UserInfo: 标准化用户信息
//   - error: 错误信息
func (idp *AlipayIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	atUserInfo := &AlipayUserResponse{}
	accessToken := token.AccessToken

	pTokenParams := &struct {
		ClientId  string `json:"app_id"`
		CharSet   string `json:"charset"`
		AuthToken string `json:"auth_token"`
		Method    string `json:"method"`
		SignType  string `json:"sign_type"`
		TimeStamp string `json:"timestamp"`
		Version   string `json:"version"`
	}{idp.Config.ClientID, "utf-8", accessToken, "alipay.user.info.share", "RSA2", time.Now().Format("2006-01-02 15:04:05"), "1.0"}
	data, err := idp.postWithBody(pTokenParams, idp.Config.Endpoint.TokenURL)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, atUserInfo)
	if err != nil {
		return nil, err
	}

	userInfo := UserInfo{
		Id:          atUserInfo.AlipayUserInfoShareResponse.UserId,
		Username:    atUserInfo.AlipayUserInfoShareResponse.NickName,
		DisplayName: atUserInfo.AlipayUserInfoShareResponse.NickName,
		AvatarUrl:   atUserInfo.AlipayUserInfoShareResponse.Avatar,
	}

	return &userInfo, nil
}

// postWithBody 发送带请求体的POST请求并进行RSA签名
// 参数:
//   - body: 请求参数结构体
//   - targetUrl: 目标URL
//
// 返回:
//   - []byte: 响应数据
//   - error: 错误信息
func (idp *AlipayIdProvider) postWithBody(body interface{}, targetUrl string) ([]byte, error) {
	bs, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	bodyJson := make(map[string]interface{})
	err = json.Unmarshal(bs, &bodyJson)
	if err != nil {
		return nil, err
	}

	formData := url.Values{}
	for k := range bodyJson {
		formData.Set(k, bodyJson[k].(string))
	}

	sign, err := rsaSignWithRSA256(getStringToSign(formData), idp.Config.ClientSecret)
	if err != nil {
		return nil, err
	}

	formData.Set("sign", sign)

	resp, err := idp.Client.PostForm(targetUrl, formData)
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

// getStringToSign 获取待签名字符串
// 按照支付宝签名规则对参数进行排序和拼接
// 详细文档: https://opendocs.alipay.com/common/02kf5q
// 参数:
//   - formData: 表单数据
//
// 返回:
//   - string: 待签名字符串
func getStringToSign(formData url.Values) string {
	keys := make([]string, 0, len(formData))
	for k := range formData {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	str := ""
	for _, k := range keys {
		if k == "sign" || formData[k][0] == "" {
			continue
		} else {
			str += "&" + k + "=" + formData[k][0]
		}
	}
	str = strings.Trim(str, "&")
	return str
}

// rsaSignWithRSA256 使用RSA私钥对内容进行SHA256签名
// 参数:
//   - signContent: 待签名内容
//   - privateKey: RSA私钥字符串
//
// 返回:
//   - string: Base64编码的签名结果
//   - error: 错误信息
func rsaSignWithRSA256(signContent string, privateKey string) (string, error) {
	privateKey = formatPrivateKey(privateKey)
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		panic("fail to parse privateKey")
	}

	h := sha256.New()
	h.Write([]byte(signContent))
	hashed := h.Sum(nil)

	privateKeyRSA, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKeyRSA.(*rsa.PrivateKey), crypto.SHA256, hashed)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// formatPrivateKey 格式化私钥字符串为PEM格式
// 将数据库中存储的私钥字符串转换为标准PEM格式
// 参数:
//   - privateKey: 原始私钥字符串
//
// 返回:
//   - string: PEM格式的私钥
func formatPrivateKey(privateKey string) string {
	// each line length is 64
	preFmtPrivateKey := ""
	for i := 0; ; {
		if i+64 <= len(privateKey) {
			preFmtPrivateKey = preFmtPrivateKey + privateKey[i:i+64] + "\n"
			i += 64
		} else {
			preFmtPrivateKey = preFmtPrivateKey + privateKey[i:]
			break
		}
	}
	privateKey = strings.Trim(preFmtPrivateKey, "\n")

	// add pkcs#8 BEGIN and END
	PemBegin := "-----BEGIN PRIVATE KEY-----\n"
	PemEnd := "\n-----END PRIVATE KEY-----"
	if !strings.HasPrefix(privateKey, PemBegin) {
		privateKey = PemBegin + privateKey
	}
	if !strings.HasSuffix(privateKey, PemEnd) {
		privateKey = privateKey + PemEnd
	}
	return privateKey
}
