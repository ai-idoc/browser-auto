// Package domain 定义核心业务模型
package domain

import "time"

// AuthType 认证类型
type AuthType string

const (
	AuthTypeNone      AuthType = "none"       // 无需认证
	AuthTypeForm      AuthType = "form"       // 表单登录
	AuthTypeSSO       AuthType = "sso"        // SSO 单点登录
	AuthTypeManual    AuthType = "manual"     // 手动登录
	AuthTypeCookie    AuthType = "cookie"     // Cookie 注入
	AuthTypeToken     AuthType = "token"      // Token 注入
)

// AuthConfig 认证配置
type AuthConfig struct {
	Type        AuthType          `json:"type"`
	Credentials *Credentials      `json:"credentials,omitempty"`
	SSOConfig   *SSOConfig        `json:"sso_config,omitempty"`
	SessionID   string            `json:"session_id,omitempty"`
	Cookies     []Cookie          `json:"cookies,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// Credentials 登录凭据
type Credentials struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
	APIKey   string `json:"api_key,omitempty"`
}

// SSOConfig SSO 配置
type SSOConfig struct {
	Provider     SSOProvider `json:"provider"`
	LoginURL     string      `json:"login_url,omitempty"`
	CallbackURL  string      `json:"callback_url,omitempty"`
	ClientID     string      `json:"client_id,omitempty"`
	ClientSecret string      `json:"client_secret,omitempty"`
	TenantID     string      `json:"tenant_id,omitempty"`
	Domain       string      `json:"domain,omitempty"`
}

// SSOProvider SSO 提供商
type SSOProvider string

const (
	SSOProviderGeneric SSOProvider = "generic"
	SSOProviderOAuth2  SSOProvider = "oauth2"
	SSOProviderSAML    SSOProvider = "saml"
	SSOProviderOIDC    SSOProvider = "oidc"
	SSOProviderCAS     SSOProvider = "cas"
)

// Cookie HTTP Cookie
type Cookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Domain   string    `json:"domain"`
	Path     string    `json:"path"`
	Expires  time.Time `json:"expires,omitempty"`
	Secure   bool      `json:"secure"`
	HTTPOnly bool      `json:"http_only"`
}

// Session 认证会话
type Session struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id,omitempty"`
	Cookies   []Cookie          `json:"cookies"`
	Headers   map[string]string `json:"headers,omitempty"`
	ExpiresAt time.Time         `json:"expires_at"`
	CreatedAt time.Time         `json:"created_at"`
}
