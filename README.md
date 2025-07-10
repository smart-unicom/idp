# idp

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.16-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/smart-unicom/idp)](https://goreportcard.com/report/github.com/smart-unicom/idp)
[![Coverage Status](https://coveralls.io/repos/github/your-org/idp/badge.svg)](https://coveralls.io/github/your-org/idp)

**一个功能完整、易于使用的 Golang 第三方登录组件库**

支持国内外主流第三方登录平台，提供统一的接口设计和标准化的数据结构

[快速开始](#-快速开始) • [支持平台](#-支持的第三方平台) • [文档](#-核心接口) • [示例](#-使用示例) • [贡献](#-贡献指南)

</div>

## 📋 项目简介

`idp` 是一个基于 Go 语言开发的第三方身份认证提供者（Identity Provider）组件库，旨在为开发者提供统一、简洁的第三方登录接口。通过标准化的接口设计和工厂模式，开发者可以轻松集成多种第三方登录服务，无需关心各平台API的差异性。

## ✨ 核心特性

- **🔌 统一接口设计**：所有第三方登录提供者都实现相同的 `IdProvider` 接口，简化集成复杂度
- **🏭 工厂模式**：通过 `GetIdProvider` 函数统一创建不同类型的登录提供者，支持配置化管理
- **📊 标准化数据结构**：统一的 `UserInfo` 结构体，消除不同平台用户信息格式差异
- **🔐 安全可靠**：支持 OAuth2 标准流程，内置 RSA 签名验证，确保数据传输安全
- **🌐 平台覆盖全面**：支持15+国内外主流第三方登录平台，持续更新
- **📱 多场景支持**：支持网页登录、移动端登录、小程序登录、企业应用登录等多种场景
- **⚙️ 配置灵活**：支持代码配置、JSON配置文件、环境变量等多种配置方式
- **🚀 高性能**：内置连接池、超时控制、错误重试等机制，确保高可用性

## 🚀 支持的第三方平台

### 🇨🇳 国内平台

| 平台 | 类型常量 | 支持功能 | 文档状态 |
|------|----------|----------|----------|
| 微信开放平台 | `IDP_WECHAT` | 网页登录、获取用户信息 | ✅ 完整 |
| 微信小程序 | - | 小程序登录、会话管理 | ✅ 完整 |
| 企业微信第三方应用 | `IDP_WECOM` | 企业登录、用户信息获取 | ✅ 完整 |
| 企业微信内部应用 | `IDP_WECOM_INTERNAL` | 企业内部登录、员工信息获取 | ✅ 完整 |
| 支付宝 | `IDP_ALIPAY` | 网页登录、RSA签名验证 | ✅ 完整 |
| QQ | `IDP_QQ` | 网页登录、用户信息获取 | ✅ 完整 |
| 新浪微博 | `IDP_WEIBO` | 网页登录、用户信息获取 | ✅ 完整 |
| 百度 | `IDP_BAIDU` | 网页登录、用户信息获取 | ✅ 完整 |
| 钉钉 | `IDP_DING_TALK` | 企业登录、员工信息获取、全球国家码 | ✅ 完整 |
| 抖音 | `IDP_DOUYIN` | 网页登录、用户信息获取 | ✅ 完整 |
| 哔哩哔哩 | `IDP_BILIBILI` | 网页登录、用户信息获取 | ✅ 完整 |

### 🌍 国外平台

| 平台 | 类型常量 | 支持功能 | 文档状态 |
|------|----------|----------|----------|
| GitHub | `IDP_GITHUB` | OAuth2登录、用户信息获取 | ✅ 完整 |
| GitLab | `IDP_GITLAB` | OAuth2登录、用户信息获取 | ✅ 完整 |
| Gitee | `IDP_GITEE` | OAuth2登录、用户信息获取 | ✅ 完整 |

## 📦 安装

### 使用 Go Modules（推荐）

```bash
go get github.com/smart-unicom/idp
```

### 版本要求

- Go 1.16 或更高版本
- 支持 Go Modules

## 🛠️ 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "time"
    
    "github.com/smart-unicom/idp"
)

func main() {
    // 创建提供者配置信息
    providerInfo := &idp.ProviderInfo{
        Type:         "WeChat",
        ClientId:     "your_wechat_app_id",
        ClientSecret: "your_wechat_app_secret",
    }
    
    // 通过工厂方法创建登录提供者
    provider, err := idp.GetIdProvider(providerInfo, "https://your-domain.com/callback")
    if err != nil {
        log.Fatal(err)
    }
    
    // 设置自定义HTTP客户端（可选）
    client := &http.Client{Timeout: 30 * time.Second}
    provider.SetHttpClient(client)
    
    // 通过授权码获取访问令牌
    token, err := provider.GetToken("authorization_code")
    if err != nil {
        log.Fatal(err)
    }
    
    // 获取用户信息
    userInfo, err := provider.GetUserInfo(token)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("用户信息: %+v\n", userInfo)
}
```

## 📚 使用示例

### 微信登录示例

```go
// 创建微信登录提供者
wechatProvider := idp.NewWeChatIdProvider(
    "wx1234567890abcdef",  // 微信应用AppId
    "your_app_secret",      // 微信应用AppSecret
    "https://your-domain.com/callback", // 回调地址
)

// 设置自定义HTTP客户端（可选）
client := &http.Client{Timeout: 30 * time.Second}
wechatProvider.SetHttpClient(client)

// 处理回调，获取用户信息
func handleWechatCallback(code string) {
    token, err := wechatProvider.GetToken(code)
    if err != nil {
        log.Printf("获取令牌失败: %v", err)
        return
    }
    
    userInfo, err := wechatProvider.GetUserInfo(token)
    if err != nil {
        log.Printf("获取用户信息失败: %v", err)
        return
    }
    
    fmt.Printf("用户ID: %s\n", userInfo.Id)
    fmt.Printf("用户名: %s\n", userInfo.Username)
    fmt.Printf("显示名: %s\n", userInfo.DisplayName)
    fmt.Printf("邮箱: %s\n", userInfo.Email)
    fmt.Printf("手机号: %s\n", userInfo.Phone)
    fmt.Printf("国家代码: %s\n", userInfo.CountryCode)
    fmt.Printf("头像: %s\n", userInfo.AvatarUrl)
}
```

### 企业微信登录示例

```go
// 企业微信第三方应用登录
wecomProvider := idp.NewWeComIdProvider(
    "your_corp_id",        // 企业ID
    "your_provider_secret", // 第三方应用密钥
    "https://your-domain.com/callback", // 回调地址
)

// 企业微信内部应用登录
wecomInternalProvider := idp.NewWeComInternalIdProvider(
    "your_corp_id",     // 企业ID
    "your_corp_secret", // 企业应用密钥
    "https://your-domain.com/callback", // 回调地址
)

// 处理企业微信登录
func handleWecomLogin(code string, provider idp.IdProvider) {
    token, err := provider.GetToken(code)
    if err != nil {
        log.Printf("获取令牌失败: %v", err)
        return
    }
    
    userInfo, err := provider.GetUserInfo(token)
    if err != nil {
        log.Printf("获取用户信息失败: %v", err)
        return
    }
    
    fmt.Printf("企业微信用户: %+v\n", userInfo)
}
```

### 钉钉登录示例

```go
// 创建钉钉登录提供者
dingtalkProvider := idp.NewDingTalkIdProvider(
    "your_dingtalk_app_id",     // 钉钉应用ID
    "your_dingtalk_app_secret", // 钉钉应用密钥
    "https://your-domain.com/callback", // 回调地址
)

// 处理钉钉登录
func handleDingtalkLogin(code string) {
    token, err := dingtalkProvider.GetToken(code)
    if err != nil {
        log.Printf("获取令牌失败: %v", err)
        return
    }
    
    userInfo, err := dingtalkProvider.GetUserInfo(token)
    if err != nil {
        log.Printf("获取用户信息失败: %v", err)
        return
    }
    
    fmt.Printf("钉钉用户: %+v\n", userInfo)
    fmt.Printf("国家代码: %s\n", userInfo.CountryCode) // 支持全球100+国家地区
}
```

### 支付宝登录示例

```go
// 创建支付宝登录提供者
alipayProvider := idp.NewAlipayIdProvider(
    "2021001234567890",     // 支付宝应用AppId
    "your_private_key",     // 应用私钥
    "https://your-domain.com/callback", // 回调地址
)

// 处理支付宝登录
func handleAlipayLogin(code string) {
    token, err := alipayProvider.GetToken(code)
    if err != nil {
        log.Printf("获取令牌失败: %v", err)
        return
    }
    
    userInfo, err := alipayProvider.GetUserInfo(token)
    if err != nil {
        log.Printf("获取用户信息失败: %v", err)
        return
    }
    
    fmt.Printf("支付宝用户: %+v\n", userInfo)
}
```

### GitHub登录示例

```go
// 创建GitHub登录提供者
githubProvider := idp.NewGithubIdProvider(
    "your_github_client_id",     // GitHub应用Client ID
    "your_github_client_secret", // GitHub应用Client Secret
    "https://your-domain.com/callback", // 回调地址
)

// 处理GitHub登录
func handleGithubLogin(code string) {
    token, err := githubProvider.GetToken(code)
    if err != nil {
        log.Printf("获取令牌失败: %v", err)
        return
    }
    
    userInfo, err := githubProvider.GetUserInfo(token)
    if err != nil {
        log.Printf("获取用户信息失败: %v", err)
        return
    }
    
    fmt.Printf("GitHub用户: %+v\n", userInfo)
}
```

### 微信小程序登录示例

```go
// 创建微信小程序登录提供者
miniprogramProvider := idp.NewWeChatMiniProgramIdProvider(
    "wx1234567890abcdef",  // 小程序AppId
    "your_app_secret",      // 小程序AppSecret
)

// 通过小程序授权码获取会话信息
func handleMiniprogramLogin(jsCode string) {
    session, err := miniprogramProvider.GetSessionByCode(jsCode)
    if err != nil {
        log.Printf("获取会话失败: %v", err)
        return
    }
    
    if session.Errcode != 0 {
        log.Printf("微信接口错误: %s", session.Errmsg)
        return
    }
    
    fmt.Printf("OpenID: %s\n", session.Openid)
    fmt.Printf("SessionKey: %s\n", session.SessionKey)
    if session.Unionid != "" {
        fmt.Printf("UnionID: %s\n", session.Unionid)
    }
}
```

## 🏗️ 核心接口

### IdProvider 接口

所有第三方登录提供者都需要实现以下接口：

```go
type IdProvider interface {
    // SetHttpClient 设置HTTP客户端
    SetHttpClient(client *http.Client)
    
    // GetToken 通过授权码获取访问令牌
    GetToken(code string) (*oauth2.Token, error)
    
    // GetUserInfo 通过访问令牌获取用户信息
    GetUserInfo(token *oauth2.Token) (*UserInfo, error)
}
```

### 标准化用户信息结构

```go
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
```

### 提供者配置信息

```go
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
    TokenURL      string            // 获取Token的URL
    AuthURL       string            // 授权URL
    UserInfoURL   string            // 获取用户信息的URL
    UserMapping   map[string]string // 用户字段映射
}
```

## ⚙️ 配置方式

### 工厂模式创建提供者

```go
// 通过配置信息创建提供者
providerInfo := &idp.ProviderInfo{
    Type:         "GitHub",
    ClientId:     "your_client_id",
    ClientSecret: "your_client_secret",
}

provider, err := idp.GetIdProvider(providerInfo, "https://your-callback-url")
if err != nil {
    log.Fatal(err)
}
```

### JSON配置方式

支持通过JSON配置文件来管理多个第三方登录提供者，便于统一管理和动态配置。

#### 配置文件示例 (config.json)

```json
{
  "providers": {
    "wechat": {
      "type": "WeChat",
      "clientId": "wx1234567890abcdef",
      "clientSecret": "your_wechat_app_secret",
      "redirectUrl": "https://your-domain.com/callback/wechat"
    },
    "github": {
      "type": "GitHub",
      "clientId": "your_github_client_id",
      "clientSecret": "your_github_client_secret",
      "redirectUrl": "https://your-domain.com/callback/github"
    },
    "wecom": {
      "type": "WeCom",
      "clientId": "your_corp_id",
      "clientSecret": "your_provider_secret",
      "redirectUrl": "https://your-domain.com/callback/wecom"
    },
    "dingtalk": {
      "type": "DingTalk",
      "clientId": "your_dingtalk_app_id",
      "clientSecret": "your_dingtalk_app_secret",
      "redirectUrl": "https://your-domain.com/callback/dingtalk"
    },
    "alipay": {
      "type": "Alipay",
      "clientId": "2021001234567890",
      "clientSecret": "your_private_key",
      "redirectUrl": "https://your-domain.com/callback/alipay"
    }
  },
  "defaultRedirectUrl": "https://your-domain.com/callback",
  "httpTimeout": 30
}
```

#### 使用JSON配置的代码示例

```go
package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strings"
    "time"
    
    "github.com/smart-unicom/idp"
)

// 配置结构体
type Config struct {
    Providers map[string]*idp.ProviderInfo `json:"providers"`
    DefaultRedirectUrl string `json:"defaultRedirectUrl"`
    HttpTimeout int `json:"httpTimeout"`
}

// 提供者管理器
type ProviderManager struct {
    config *Config
    providers map[string]idp.IdProvider
    httpClient *http.Client
}

// 从JSON文件加载配置
func LoadConfigFromFile(filename string) (*Config, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var config Config
    err = json.Unmarshal(data, &config)
    if err != nil {
        return nil, err
    }
    
    return &config, nil
}

// 支持从环境变量读取敏感配置
func LoadConfigWithEnv(filename string) (*Config, error) {
    config, err := LoadConfigFromFile(filename)
    if err != nil {
        return nil, err
    }
    
    // 从环境变量覆盖敏感配置
    for name, provider := range config.Providers {
        envPrefix := strings.ToUpper(name)
        
        if clientId := os.Getenv(envPrefix + "_CLIENT_ID"); clientId != "" {
            provider.ClientId = clientId
        }
        
        if clientSecret := os.Getenv(envPrefix + "_CLIENT_SECRET"); clientSecret != "" {
            provider.ClientSecret = clientSecret
        }
        
        if redirectUrl := os.Getenv(envPrefix + "_REDIRECT_URL"); redirectUrl != "" {
            provider.RedirectUrl = redirectUrl
        }
    }
    
    return config, nil
}

// 创建提供者管理器
func NewProviderManager(configFile string) (*ProviderManager, error) {
    config, err := LoadConfigWithEnv(configFile)
    if err != nil {
        return nil, err
    }
    
    // 创建HTTP客户端
    timeout := time.Duration(config.HttpTimeout) * time.Second
    httpClient := &http.Client{Timeout: timeout}
    
    pm := &ProviderManager{
        config: config,
        providers: make(map[string]idp.IdProvider),
        httpClient: httpClient,
    }
    
    // 初始化所有提供者
    err = pm.initProviders()
    if err != nil {
        return nil, err
    }
    
    return pm, nil
}

// 初始化所有提供者
func (pm *ProviderManager) initProviders() error {
    for name, providerInfo := range pm.config.Providers {
        // 如果没有设置重定向URL，使用默认值
        redirectUrl := providerInfo.RedirectUrl
        if redirectUrl == "" {
            redirectUrl = pm.config.DefaultRedirectUrl
        }
        
        provider, err := idp.GetIdProvider(providerInfo, redirectUrl)
        if err != nil {
            return fmt.Errorf("创建提供者 %s 失败: %v", name, err)
        }
        
        // 设置HTTP客户端
        provider.SetHttpClient(pm.httpClient)
        
        pm.providers[name] = provider
    }
    
    return nil
}

// 获取指定的提供者
func (pm *ProviderManager) GetProvider(name string) (idp.IdProvider, error) {
    provider, exists := pm.providers[name]
    if !exists {
        return nil, fmt.Errorf("提供者 %s 不存在", name)
    }
    return provider, nil
}

// 获取所有提供者名称
func (pm *ProviderManager) GetProviderNames() []string {
    names := make([]string, 0, len(pm.providers))
    for name := range pm.providers {
        names = append(names, name)
    }
    return names
}

// 使用示例
func main() {
    // 从配置文件创建提供者管理器
    pm, err := NewProviderManager("config.json")
    if err != nil {
        log.Fatal("加载配置失败:", err)
    }
    
    // 获取微信提供者
    wechatProvider, err := pm.GetProvider("wechat")
    if err != nil {
        log.Fatal("获取微信提供者失败:", err)
    }
    
    // 处理微信登录
    handleLogin("wechat_auth_code", wechatProvider)
    
    // 列出所有可用的提供者
    fmt.Println("可用的登录提供者:", pm.GetProviderNames())
}

// 通用登录处理函数
func handleLogin(code string, provider idp.IdProvider) {
    token, err := provider.GetToken(code)
    if err != nil {
        log.Printf("获取令牌失败: %v", err)
        return
    }
    
    userInfo, err := provider.GetUserInfo(token)
    if err != nil {
        log.Printf("获取用户信息失败: %v", err)
        return
    }
    
    fmt.Printf("用户信息: %+v\n", userInfo)
}
```

### 环境变量配置

```bash
# 微信配置
WECHAT_CLIENT_ID=wx1234567890abcdef
WECHAT_CLIENT_SECRET=your_wechat_app_secret
WECHAT_REDIRECT_URL=https://your-domain.com/callback/wechat

# GitHub配置
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GITHUB_REDIRECT_URL=https://your-domain.com/callback/github
```

## 🔧 高级功能

### 国家代码映射

钉钉登录提供者内置了全球100+国家和地区的电话国家码到ISO国家代码的映射，支持：

- 中国（86 → CN）
- 美国（1 → US）
- 英国（44 → GB）
- 日本（81 → JP）
- 韩国（82 → KR）
- 以及其他90+个国家和地区

### RSA签名验证（支付宝）

支付宝登录提供者内置了RSA256签名验证功能，确保数据传输的安全性。

### HTTP客户端自定义

```go
// 创建自定义HTTP客户端
client := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}

// 应用到提供者
provider.SetHttpClient(client)
```

### 错误处理和重试

```go
// 带重试的登录处理
func handleLoginWithRetry(code string, provider idp.IdProvider, maxRetries int) (*idp.UserInfo, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        token, err := provider.GetToken(code)
        if err != nil {
            lastErr = err
            time.Sleep(time.Duration(i+1) * time.Second)
            continue
        }
        
        userInfo, err := provider.GetUserInfo(token)
        if err != nil {
            lastErr = err
            time.Sleep(time.Duration(i+1) * time.Second)
            continue
        }
        
        return userInfo, nil
    }
    
    return nil, fmt.Errorf("登录失败，已重试 %d 次: %v", maxRetries, lastErr)
}
```

## 📋 依赖项

```go
require (
    golang.org/x/oauth2 v0.0.0-20210323180902-22b0adad7558
)
```

## 🧪 测试

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 基准测试

```bash
# 运行基准测试
go test -bench=. ./...
```

## 🤝 贡献指南

我们欢迎社区贡献！请遵循以下步骤：

### 贡献流程

1. **Fork 本仓库**
2. **创建特性分支** (`git checkout -b feature/AmazingFeature`)
3. **提交更改** (`git commit -m 'Add some AmazingFeature'`)
4. **推送到分支** (`git push origin feature/AmazingFeature`)
5. **开启 Pull Request**

### 代码规范

请遵循项目中的 `代码规范.md` 文件中定义的编码标准：

- ✅ 使用中文注释
- ✅ 遵循 Go 语言命名约定
- ✅ 为所有公开的函数、结构体和接口添加详细注释
- ✅ 保持代码简洁和可读性
- ✅ 每个文件都应包含文件头部注释说明功能
- ✅ 添加单元测试覆盖新功能
- ✅ 确保所有测试通过

### 添加新的登录提供者

1. **创建新文件**：如 `newplatform.go`
2. **实现接口**：实现 `IdProvider` 接口的所有方法
3. **添加常量**：在 `provider.go` 中添加相应的常量定义
4. **更新工厂**：在 `GetIdProvider` 函数中添加新的 case 分支
5. **编写测试**：添加完整的单元测试
6. **更新文档**：添加详细的中文注释和使用示例

### 提交信息规范

```
type(scope): description

[optional body]

[optional footer]
```

类型（type）：
- `feat`: 新功能
- `fix`: 修复bug
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

## 📊 性能优化建议

### 连接池配置

```go
// 优化HTTP传输配置
transport := &http.Transport{
    MaxIdleConns:        100,              // 最大空闲连接数
    MaxIdleConnsPerHost: 10,               // 每个主机最大空闲连接数
    IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
    DisableKeepAlives:   false,            // 启用Keep-Alive
}

client := &http.Client{
    Timeout:   30 * time.Second,
    Transport: transport,
}
```

### 缓存用户信息

```go
// 简单的内存缓存实现
type UserCache struct {
    cache map[string]*idp.UserInfo
    mutex sync.RWMutex
    ttl   time.Duration
}

func (c *UserCache) Get(key string) (*idp.UserInfo, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    user, exists := c.cache[key]
    return user, exists
}

func (c *UserCache) Set(key string, user *idp.UserInfo) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    c.cache[key] = user
    
    // 设置过期时间
    go func() {
        time.Sleep(c.ttl)
        c.mutex.Lock()
        delete(c.cache, key)
        c.mutex.Unlock()
    }()
}
```

## 🔒 安全最佳实践

### 1. 配置安全

```go
// ❌ 错误：硬编码敏感信息
wechatProvider := idp.NewWeChatIdProvider(
    "wx1234567890abcdef",
    "hardcoded_secret", // 不要这样做
    "https://example.com/callback",
)

// ✅ 正确：从环境变量读取
wechatProvider := idp.NewWeChatIdProvider(
    os.Getenv("WECHAT_APP_ID"),
    os.Getenv("WECHAT_APP_SECRET"),
    os.Getenv("WECHAT_REDIRECT_URL"),
)
```

### 2. HTTPS强制

```go
// 确保所有回调URL使用HTTPS
func validateRedirectURL(url string) error {
    if !strings.HasPrefix(url, "https://") {
        return errors.New("回调URL必须使用HTTPS协议")
    }
    return nil
}
```

### 3. 状态参数验证

```go
// 生成和验证state参数
func generateState() string {
    b := make([]byte, 32)
    rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)
}

func validateState(expected, actual string) bool {
    return subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}
```

## 📄 许可证

本项目采用 [Apache License 2.0](LICENSE) 许可证。

## 🆘 支持与反馈

### 获取帮助

- 📖 [Wiki文档](https://github.com/smart-unicom/idp/wiki)
- 💬 [讨论区](https://github.com/smart-unicom/idp/discussions)
- 🐛 [问题反馈](https://github.com/smart-unicom/idp/issues)
- 🔧 [功能请求](https://github.com/smart-unicom/idp/issues/new?template=feature_request.md)

## 🔄 更新日志

### v1.3.0 (2024-01-15)
- 🎉 新增JSON配置文件支持
- 🎉 新增环境变量配置支持
- 🎉 新增提供者管理器（ProviderManager）
- 🔧 优化HTTP客户端配置
- 📚 完善文档和示例代码
- 🧪 增加单元测试覆盖率

### v1.2.0 (2023-12-01)
- 🎉 新增企业微信第三方应用和内部应用支持
- 🎉 新增抖音、哔哩哔哩登录支持
- 🔧 优化钉钉登录，新增全球国家代码映射
- 📚 完善代码注释和文档
- 🐛 修复若干已知问题

### v1.1.0 (2023-10-15)
- 🎉 新增钉钉企业登录支持
- 🎉 新增GitLab、Gitee登录支持
- 🔧 优化错误处理和日志记录
- 🔒 增强安全性验证

### v1.0.0 (2023-08-01)
- 🎉 初始版本发布
- 🎉 支持微信、支付宝、GitHub、QQ、微博、百度等主流第三方登录
- 🏗️ 实现统一的接口设计和标准化数据结构
- 📚 提供完整的文档和示例代码

---

<div align="center">

**如果这个项目对您有帮助，请给我们一个 ⭐️**

[⬆ 回到顶部](#idp)

</div>