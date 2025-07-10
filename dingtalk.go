// 钉钉OAuth2登录提供者实现
// 支持钉钉网页登录和用户信息获取
package idp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// DingTalkIdProvider 钉钉登录提供者
// 实现钉钉OAuth2登录功能
type DingTalkIdProvider struct {
	Client *http.Client    // HTTP客户端
	Config *oauth2.Config // OAuth2配置
}

// NewDingTalkIdProvider 创建钉钉登录提供者实例
// 参数:
//   - clientId: 钉钉应用的Client ID
//   - clientSecret: 钉钉应用的Client Secret
//   - redirectUrl: OAuth2重定向URL
// 返回:
//   - *DingTalkIdProvider: 钉钉登录提供者实例
func NewDingTalkIdProvider(clientId string, clientSecret string, redirectUrl string) *DingTalkIdProvider {
	idp := &DingTalkIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

// SetHttpClient 设置HTTP客户端
// 参数:
//   - client: HTTP客户端实例
func (idp *DingTalkIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig 获取钉钉OAuth2配置
// 参数:
//   - clientId: 钉钉应用的Client ID
//   - clientSecret: 钉钉应用的Client Secret
//   - redirectUrl: OAuth2重定向URL
// 返回:
//   - *oauth2.Config: OAuth2配置实例
func (idp *DingTalkIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		AuthURL:  "https://api.dingtalk.com/v1.0/contact/users/me",
		TokenURL: "https://api.dingtalk.com/v1.0/oauth2/userAccessToken",
	}

	config := &oauth2.Config{
		// DingTalk not allow to set scopes,here it is just a placeholder,
		// convenient to use later
		Scopes: []string{"", ""},

		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

// DingTalkAccessToken 钉钉访问令牌结构体
type DingTalkAccessToken struct {
	ErrCode     int    `json:"code"`        // 错误码
	ErrMsg      string `json:"message"`     // 错误消息
	AccessToken string `json:"accessToken"` // 访问令牌
	ExpiresIn   int64  `json:"expireIn"`    // 访问令牌的有效时间，单位是秒
}

// GetToken 通过授权码获取钉钉访问令牌
// 参数:
//   - code: 钉钉返回的授权码
// 返回:
//   - *oauth2.Token: OAuth2访问令牌
//   - error: 错误信息
// 详细文档: https://open.dingtalk.com/document/orgapp-server/obtain-user-token
func (idp *DingTalkIdProvider) GetToken(code string) (*oauth2.Token, error) {
	pTokenParams := &struct {
		ClientId     string `json:"clientId"`
		ClientSecret string `json:"clientSecret"`
		Code         string `json:"code"`
		GrantType    string `json:"grantType"`
	}{idp.Config.ClientID, idp.Config.ClientSecret, code, "authorization_code"}

	data, err := idp.postWithBody(pTokenParams, idp.Config.Endpoint.TokenURL)
	if err != nil {
		return nil, err
	}

	pToken := &DingTalkAccessToken{}
	err = json.Unmarshal(data, pToken)
	if err != nil {
		return nil, err
	}

	if pToken.ErrCode != 0 {
		return nil, fmt.Errorf("pToken.Errcode = %d, pToken.Errmsg = %s", pToken.ErrCode, pToken.ErrMsg)
	}

	token := &oauth2.Token{
		AccessToken: pToken.AccessToken,
		Expiry:      time.Unix(time.Now().Unix()+pToken.ExpiresIn, 0),
	}
	return token, nil
}

/*
{
{
  "nick" : "zhangsan",
  "avatarUrl" : "https://xxx",
  "mobile" : "150xxxx9144",
  "openId" : "123",
  "unionId" : "z21HjQliSzpw0Yxxxx",
  "email" : "zhangsan@alibaba-inc.com",
  "stateCode" : "86"
}
*/

// DingTalkUserResponse 钉钉用户信息响应结构体
type DingTalkUserResponse struct {
	Nick      string `json:"nick"`      // 用户昵称
	OpenId    string `json:"openId"`    // 用户OpenID
	UnionId   string `json:"unionId"`   // 用户UnionID
	AvatarUrl string `json:"avatarUrl"` // 用户头像URL
	Email     string `json:"email"`     // 用户邮箱
	Mobile    string `json:"mobile"`    // 用户手机号
	StateCode string `json:"stateCode"` // 国家码
}

// GetUserInfo 通过访问令牌获取钉钉用户信息
// 参数:
//   - token: OAuth2访问令牌
// 返回:
//   - *UserInfo: 标准化用户信息
//   - error: 错误信息
// 详细文档: https://open.dingtalk.com/document/orgapp-server/dingtalk-retrieve-user-information
func (idp *DingTalkIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	dtUserInfo := &DingTalkUserResponse{}
	accessToken := token.AccessToken

	reqest, err := http.NewRequest("GET", idp.Config.Endpoint.AuthURL, nil)
	if err != nil {
		return nil, err
	}
	reqest.Header.Add("x-acs-dingtalk-access-token", accessToken)
	resp, err := idp.Client.Do(reqest)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, dtUserInfo)
	if err != nil {
		return nil, err
	}

	countryCode := getCountryCode(dtUserInfo.StateCode, dtUserInfo.Mobile)

	userInfo := UserInfo{
		Id:          dtUserInfo.OpenId,
		Username:    dtUserInfo.Nick,
		DisplayName: dtUserInfo.Nick,
		UnionId:     dtUserInfo.UnionId,
		Email:       dtUserInfo.Email,
		Phone:       dtUserInfo.Mobile,
		CountryCode: countryCode,
		AvatarUrl:   dtUserInfo.AvatarUrl,
	}

	corpAccessToken := idp.getInnerAppAccessToken()
	userId, err := idp.getUserId(userInfo.UnionId, corpAccessToken)
	if err != nil {
		return nil, err
	}

	corpMobile, corpEmail, jobNumber, err := idp.getUserCorpEmail(userId, corpAccessToken)
	if err == nil {
		if corpMobile != "" {
			userInfo.Phone = corpMobile
		}

		if corpEmail != "" {
			userInfo.Email = corpEmail
		}

		if jobNumber != "" {
			userInfo.Username = jobNumber
		}
	}

	return &userInfo, nil
}

// postWithBody 发送POST请求并获取响应内容
// 参数:
//   - body: 请求体数据
//   - url: 请求URL
// 返回:
//   - []byte: 响应内容
//   - error: 错误信息
func (idp *DingTalkIdProvider) postWithBody(body interface{}, url string) ([]byte, error) {
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

// getInnerAppAccessToken 获取企业内部应用访问令牌
// 返回:
//   - string: 企业内部应用访问令牌
func (idp *DingTalkIdProvider) getInnerAppAccessToken() string {
	body := make(map[string]string)
	body["appKey"] = idp.Config.ClientID
	body["appSecret"] = idp.Config.ClientSecret
	respBytes, err := idp.postWithBody(body, "https://api.dingtalk.com/v1.0/oauth2/accessToken")
	if err != nil {
		log.Println(err.Error())
	}

	var data struct {
		ExpireIn    int    `json:"expireIn"`
		AccessToken string `json:"accessToken"`
	}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		log.Println(err.Error())
	}
	return data.AccessToken
}

// getUserId 通过UnionID获取用户ID
// 参数:
//   - unionId: 用户UnionID
//   - accessToken: 企业内部应用访问令牌
// 返回:
//   - string: 用户ID
//   - error: 错误信息
func (idp *DingTalkIdProvider) getUserId(unionId string, accessToken string) (string, error) {
	body := make(map[string]string)
	body["unionid"] = unionId
	respBytes, err := idp.postWithBody(body, "https://oapi.dingtalk.com/topapi/user/getbyunionid?access_token="+accessToken)
	if err != nil {
		return "", err
	}

	var data struct {
		ErrCode    int    `json:"errcode"`
		ErrMessage string `json:"errmsg"`
		Result     struct {
			UserId string `json:"userid"`
		} `json:"result"`
	}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		return "", err
	}
	if data.ErrCode == 60121 {
		return "", fmt.Errorf("该应用只允许本企业内部用户登录，您不属于该企业，无法登录")
	} else if data.ErrCode != 0 {
		return "", fmt.Errorf(data.ErrMessage)
	}
	return data.Result.UserId, nil
}

// getUserCorpEmail 获取用户企业信息
// 参数:
//   - userId: 用户ID
//   - accessToken: 企业内部应用访问令牌
// 返回:
//   - string: 企业手机号
//   - string: 企业邮箱
//   - string: 工号
//   - error: 错误信息
func (idp *DingTalkIdProvider) getUserCorpEmail(userId string, accessToken string) (string, string, string, error) {
	// https://open.dingtalk.com/document/isvapp/query-user-details
	body := make(map[string]string)
	body["userid"] = userId
	respBytes, err := idp.postWithBody(body, "https://oapi.dingtalk.com/topapi/v2/user/get?access_token="+accessToken)
	if err != nil {
		return "", "", "", err
	}

	var data struct {
		ErrMessage string `json:"errmsg"`
		Result     struct {
			Mobile    string `json:"mobile"`
			Email     string `json:"email"`
			JobNumber string `json:"job_number"`
		} `json:"result"`
	}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		return "", "", "", err
	}
	if data.ErrMessage != "ok" {
		return "", "", "", fmt.Errorf(data.ErrMessage)
	}
	return data.Result.Mobile, data.Result.Email, data.Result.JobNumber, nil
}

// getCountryCode 根据国家码和手机号获取国家代码
// 参数:
//   - stateCode: 国家码（如"86"表示中国）
//   - mobile: 手机号码
// 返回:
//   - string: 国家代码（如"CN"表示中国）
func getCountryCode(stateCode, mobile string) string {
	// 常见国家码映射表
	countryCodeMap := map[string]string{
		"86":  "CN", // 中国
		"1":   "US", // 美国
		"44":  "GB", // 英国
		"81":  "JP", // 日本
		"82":  "KR", // 韩国
		"91":  "IN", // 印度
		"33":  "FR", // 法国
		"49":  "DE", // 德国
		"39":  "IT", // 意大利
		"7":   "RU", // 俄罗斯
		"61":  "AU", // 澳大利亚
		"55":  "BR", // 巴西
		"52":  "MX", // 墨西哥
		"34":  "ES", // 西班牙
		"31":  "NL", // 荷兰
		"46":  "SE", // 瑞典
		"47":  "NO", // 挪威
		"45":  "DK", // 丹麦
		"358": "FI", // 芬兰
		"41":  "CH", // 瑞士
		"43":  "AT", // 奥地利
		"32":  "BE", // 比利时
		"351": "PT", // 葡萄牙
		"30":  "GR", // 希腊
		"48":  "PL", // 波兰
		"420": "CZ", // 捷克
		"36":  "HU", // 匈牙利
		"40":  "RO", // 罗马尼亚
		"359": "BG", // 保加利亚
		"385": "HR", // 克罗地亚
		"386": "SI", // 斯洛文尼亚
		"421": "SK", // 斯洛伐克
		"372": "EE", // 爱沙尼亚
		"371": "LV", // 拉脱维亚
		"370": "LT", // 立陶宛
		"353": "IE", // 爱尔兰
		"354": "IS", // 冰岛
		"352": "LU", // 卢森堡
		"377": "MC", // 摩纳哥
		"378": "SM", // 圣马力诺
		"39066": "VA", // 梵蒂冈
		"376": "AD", // 安道尔
		"350": "GI", // 直布罗陀
		"356": "MT", // 马耳他
		"357": "CY", // 塞浦路斯
		"90":  "TR", // 土耳其
		"972": "IL", // 以色列
		"971": "AE", // 阿联酋
		"966": "SA", // 沙特阿拉伯
		"965": "KW", // 科威特
		"974": "QA", // 卡塔尔
		"973": "BH", // 巴林
		"968": "OM", // 阿曼
		"962": "JO", // 约旦
		"961": "LB", // 黎巴嫩
		"963": "SY", // 叙利亚
		"964": "IQ", // 伊拉克
		"98":  "IR", // 伊朗
		"93":  "AF", // 阿富汗
		"92":  "PK", // 巴基斯坦
		"880": "BD", // 孟加拉国
		"94":  "LK", // 斯里兰卡
		"95":  "MM", // 缅甸
		"66":  "TH", // 泰国
		"84":  "VN", // 越南
		"855": "KH", // 柬埔寨
		"856": "LA", // 老挝
		"60":  "MY", // 马来西亚
		"65":  "SG", // 新加坡
		"62":  "ID", // 印度尼西亚
		"63":  "PH", // 菲律宾
		"673": "BN", // 文莱
		"670": "TL", // 东帝汶
		"852": "HK", // 香港
		"853": "MO", // 澳门
		"886": "TW", // 台湾
		"20":  "EG", // 埃及
		"27":  "ZA", // 南非
		"234": "NG", // 尼日利亚
		"254": "KE", // 肯尼亚
		"233": "GH", // 加纳
		"212": "MA", // 摩洛哥
		"213": "DZ", // 阿尔及利亚
		"216": "TN", // 突尼斯
		"218": "LY", // 利比亚
		"249": "SD", // 苏丹
		"251": "ET", // 埃塞俄比亚
		"256": "UG", // 乌干达
		"255": "TZ", // 坦桑尼亚
		"250": "RW", // 卢旺达
		"257": "BI", // 布隆迪
		"243": "CD", // 刚果民主共和国
		"242": "CG", // 刚果共和国
		"236": "CF", // 中非共和国
		"235": "TD", // 乍得
		"237": "CM", // 喀麦隆
		"240": "GQ", // 赤道几内亚
		"241": "GA", // 加蓬
		"239": "ST", // 圣多美和普林西比
		"238": "CV", // 佛得角
		"245": "GW", // 几内亚比绍
		"224": "GN", // 几内亚
		"221": "SN", // 塞内加尔
		"223": "ML", // 马里
		"226": "BF", // 布基纳法索
		"227": "NE", // 尼日尔
		"229": "BJ", // 贝宁
		"228": "TG", // 多哥
		"225": "CI", // 科特迪瓦
		"231": "LR", // 利比里亚
		"232": "SL", // 塞拉利昂
		"220": "GM", // 冈比亚
	}

	// 如果找到对应的国家代码，返回它
	if countryCode, exists := countryCodeMap[stateCode]; exists {
		return countryCode
	}

	// 如果没有找到，默认返回中国
	return "CN"
}
