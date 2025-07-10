// Package idp 第三方身份认证提供者组件库
// 支持微信、企业微信、支付宝、GitHub等主流平台的OAuth2登录集成
package idp

import (
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

// UserInfo 用户信息结构体
// 包含从第三方平台获取的用户基本信息
type UserInfo struct {
	Id          string            // 用户唯一标识
	Username    string            // 用户名
	DisplayName string            // 显示名称
	UnionId     string            // 联合ID（如微信UnionId）
	Email       string            // 邮箱地址
	Phone       string            // 手机号码
	CountryCode string            // 国家代码
	AvatarUrl   string            // 头像URL
	Extra       map[string]string // 扩展信息
}

// ProviderInfo 第三方登录提供者配置信息
// 包含OAuth2认证所需的各种配置参数
type ProviderInfo struct {
	Type          string            // 提供者类型（如WeChat、GitHub等）
	SubType       string            // 子类型（如微信公众号、小程序等）
	ClientId      string            // 客户端ID
	ClientSecret  string            // 客户端密钥
	ClientId2     string            // 备用客户端ID
	ClientSecret2 string            // 备用客户端密钥
	AppId         string            // 应用ID
	HostUrl       string            // 主机URL
	RedirectUrl   string            // 重定向URL

	TokenURL    string            // 获取Token的URL
	AuthURL     string            // 授权URL
	UserInfoURL string            // 获取用户信息的URL
	UserMapping map[string]string // 用户字段映射
}

// 支持的第三方登录平台常量定义
const (
	// 国内平台
	IDP_WECHAT           string = "WeChat"          // 微信
	IDP_QQ               string = "QQ"              // QQ
	IDP_BAIDU            string = "Baidu"           // 百度
	IDP_ALIPAY           string = "Alipay"          // 支付宝
	IDP_BILIBILI         string = "Bilibili"        // 哔哩哔哩
	IDP_DOUYIN           string = "Douyin"          // 抖音
	IDP_DING_TALK        string = "DingTalk"        // 钉钉
	IDP_WEIBO            string = "Weibo"           // 微博
	IDP_WECOM            string = "WeCom"           // 企业微信第三方应用
	IDP_WECOM_INTERNAL   string = "WeComInternal"   // 企业微信内部应用

	// 国外平台
	IDP_GITHUB string = "GitHub" // GitHub
	IDP_GITEE  string = "Gitee"  // 码云
	IDP_GITLAB string = "GitLab" // GitLab
)

// IdProvider 第三方身份认证提供者接口
// 定义了所有第三方登录提供者必须实现的方法
type IdProvider interface {
	// SetHttpClient 设置HTTP客户端
	// 参数:
	//   - client: HTTP客户端实例
	SetHttpClient(client *http.Client)
	
	// GetToken 通过授权码获取访问令牌
	// 参数:
	//   - code: 授权码
	// 返回:
	//   - *oauth2.Token: OAuth2访问令牌
	//   - error: 错误信息
	GetToken(code string) (*oauth2.Token, error)
	
	// GetUserInfo 通过访问令牌获取用户信息
	// 参数:
	//   - token: OAuth2访问令牌
	// 返回:
	//   - *UserInfo: 用户信息
	//   - error: 错误信息
	GetUserInfo(token *oauth2.Token) (*UserInfo, error)
}

// GetIdProvider 根据提供者信息创建对应的身份认证提供者实例
// 参数:
//   - idpInfo: 提供者配置信息
//   - redirectUrl: OAuth2重定向URL
// 返回:
//   - IdProvider: 身份认证提供者实例
//   - error: 错误信息
func GetIdProvider(idpInfo *ProviderInfo, redirectUrl string) (IdProvider, error) {
	switch idpInfo.Type {
	case "GitHub":
		return NewGithubIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "QQ":
		return NewQqIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "WeChat":
		return NewWeChatIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "DingTalk":
		return NewDingTalkIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "Weibo":
		return NewWeiBoIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "Gitee":
		return NewGiteeIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "GitLab":
		return NewGitlabIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "Baidu":
		return NewBaiduIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "Alipay":
		return NewAlipayIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "Douyin":
		return NewDouyinIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "Bilibili":
		return NewBilibiliIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "WeCom":
		return NewWeComIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "WeComInternal":
		return NewWeComInternalIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	default:
		return nil, fmt.Errorf("不支持的登录提供者类型: %s", idpInfo.Type)
	}
}
