// Package auth 提供认证功能
package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/browser-automation/internal/browser"
	"github.com/browser-automation/internal/domain"
	"github.com/google/uuid"
)

// Authenticator 认证器接口
type Authenticator interface {
	Authenticate(ctx context.Context, config *domain.AuthConfig) (*domain.Session, error)
	ValidateSession(ctx context.Context, session *domain.Session) (bool, error)
}

// Service 认证服务
type Service struct {
	browser browser.Controller
}

// NewService 创建认证服务
func NewService(browser browser.Controller) *Service {
	return &Service{browser: browser}
}

// Authenticate 执行认证
func (s *Service) Authenticate(ctx context.Context, config *domain.AuthConfig) (*domain.Session, error) {
	switch config.Type {
	case domain.AuthTypeNone:
		return s.createEmptySession(), nil

	case domain.AuthTypeForm:
		return s.authenticateWithForm(ctx, config)

	case domain.AuthTypeSSO:
		return s.authenticateWithSSO(ctx, config)

	case domain.AuthTypeManual:
		return s.authenticateManually(ctx, config)

	case domain.AuthTypeCookie:
		return s.authenticateWithCookies(ctx, config)

	case domain.AuthTypeToken:
		return s.authenticateWithToken(ctx, config)

	default:
		return nil, fmt.Errorf("unsupported auth type: %s", config.Type)
	}
}

// ValidateSession 验证会话是否有效
func (s *Service) ValidateSession(ctx context.Context, session *domain.Session) (bool, error) {
	if session == nil {
		return false, nil
	}
	if time.Now().After(session.ExpiresAt) {
		return false, nil
	}
	return true, nil
}

func (s *Service) createEmptySession() *domain.Session {
	return &domain.Session{
		ID:        uuid.New().String(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
}

// authenticateWithForm 表单登录
func (s *Service) authenticateWithForm(ctx context.Context, config *domain.AuthConfig) (*domain.Session, error) {
	if config.Credentials == nil {
		return nil, fmt.Errorf("credentials required for form auth")
	}

	// 等待登录表单加载
	if err := s.browser.WaitForSelector(ctx, "input[type='password']", 10*time.Second); err != nil {
		return nil, fmt.Errorf("login form not found: %w", err)
	}

	// 填写用户名（尝试常见选择器）
	usernameSelectors := []string{
		"input[name='username']",
		"input[name='email']",
		"input[type='email']",
		"input[id='username']",
		"input[id='email']",
	}
	for _, sel := range usernameSelectors {
		if err := s.browser.Fill(ctx, sel, config.Credentials.Username); err == nil {
			break
		}
	}

	// 填写密码
	if err := s.browser.Fill(ctx, "input[type='password']", config.Credentials.Password); err != nil {
		return nil, fmt.Errorf("fill password: %w", err)
	}

	// 点击登录按钮
	submitSelectors := []string{
		"button[type='submit']",
		"input[type='submit']",
		"button:has-text('登录')",
		"button:has-text('Login')",
	}
	for _, sel := range submitSelectors {
		if err := s.browser.Click(ctx, sel); err == nil {
			break
		}
	}

	// 等待登录完成
	time.Sleep(3 * time.Second)

	// 提取 cookies
	cookies, err := s.browser.GetCookies(ctx)
	if err != nil {
		return nil, fmt.Errorf("get cookies: %w", err)
	}

	return &domain.Session{
		ID:        uuid.New().String(),
		Cookies:   cookies,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}, nil
}

// authenticateWithSSO SSO 认证
func (s *Service) authenticateWithSSO(ctx context.Context, config *domain.AuthConfig) (*domain.Session, error) {
	if config.SSOConfig == nil {
		return nil, fmt.Errorf("sso config required")
	}

	// 等待 SSO 页面加载
	if err := s.browser.WaitForNavigation(ctx, 10*time.Second); err != nil {
		return nil, fmt.Errorf("wait for sso redirect: %w", err)
	}

	// 获取当前 URL 判断是否在 SSO 页面
	currentURL, _ := s.browser.GetCurrentURL(ctx)

	// 如果在 SSO 登录页，执行登录
	if s.isOnSSOPage(currentURL, config.SSOConfig) {
		if config.Credentials != nil {
			if err := s.performSSOLogin(ctx, config.Credentials, config.SSOConfig); err != nil {
				return nil, fmt.Errorf("sso login: %w", err)
			}
		}
	}

	// 等待回调完成
	time.Sleep(5 * time.Second)

	// 提取 cookies
	cookies, err := s.browser.GetCookies(ctx)
	if err != nil {
		return nil, fmt.Errorf("get cookies: %w", err)
	}

	return &domain.Session{
		ID:        uuid.New().String(),
		Cookies:   cookies,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}, nil
}

// authenticateManually 手动登录
func (s *Service) authenticateManually(ctx context.Context, config *domain.AuthConfig) (*domain.Session, error) {
	// 等待用户手动完成登录（最长 5 分钟）
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("manual login timeout")
		case <-ticker.C:
			// 检查是否已离开登录页
			currentURL, _ := s.browser.GetCurrentURL(ctx)
			if !s.isOnLoginPage(currentURL) {
				// 登录成功
				cookies, err := s.browser.GetCookies(ctx)
				if err != nil {
					return nil, fmt.Errorf("get cookies: %w", err)
				}
				return &domain.Session{
					ID:        uuid.New().String(),
					Cookies:   cookies,
					ExpiresAt: time.Now().Add(24 * time.Hour),
					CreatedAt: time.Now(),
				}, nil
			}
		}
	}
}

// authenticateWithCookies Cookie 注入
func (s *Service) authenticateWithCookies(ctx context.Context, config *domain.AuthConfig) (*domain.Session, error) {
	if len(config.Cookies) == 0 {
		return nil, fmt.Errorf("cookies required")
	}

	if err := s.browser.SetCookies(ctx, config.Cookies); err != nil {
		return nil, fmt.Errorf("set cookies: %w", err)
	}

	return &domain.Session{
		ID:        uuid.New().String(),
		Cookies:   config.Cookies,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}, nil
}

// authenticateWithToken Token 注入
func (s *Service) authenticateWithToken(ctx context.Context, config *domain.AuthConfig) (*domain.Session, error) {
	if config.Credentials == nil || config.Credentials.Token == "" {
		return nil, fmt.Errorf("token required")
	}

	return &domain.Session{
		ID:        uuid.New().String(),
		Headers: map[string]string{
			"Authorization": "Bearer " + config.Credentials.Token,
		},
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}, nil
}

func (s *Service) isOnSSOPage(url string, ssoConfig *domain.SSOConfig) bool {
	if ssoConfig.LoginURL != "" {
		return strings.Contains(url, ssoConfig.LoginURL)
	}
	// 常见 SSO 页面特征
	ssoIndicators := []string{
		"login", "signin", "auth", "sso", "oauth", "saml",
	}
	for _, indicator := range ssoIndicators {
		if strings.Contains(strings.ToLower(url), indicator) {
			return true
		}
	}
	return false
}

func (s *Service) isOnLoginPage(url string) bool {
	loginIndicators := []string{
		"login", "signin", "sign-in", "auth",
	}
	for _, indicator := range loginIndicators {
		if strings.Contains(strings.ToLower(url), indicator) {
			return true
		}
	}
	return false
}

func (s *Service) performSSOLogin(ctx context.Context, creds *domain.Credentials, ssoConfig *domain.SSOConfig) error {
	// 通用 SSO 登录流程
	// 填写用户名
	usernameSelectors := []string{
		"input[name='username']",
		"input[name='email']",
		"input[type='email']",
		"input[id='username']",
	}
	for _, sel := range usernameSelectors {
		if err := s.browser.Fill(ctx, sel, creds.Username); err == nil {
			break
		}
	}

	// 填写密码
	if err := s.browser.Fill(ctx, "input[type='password']", creds.Password); err != nil {
		return fmt.Errorf("fill password: %w", err)
	}

	// 点击登录
	submitSelectors := []string{
		"button[type='submit']",
		"input[type='submit']",
		"#submit",
	}
	for _, sel := range submitSelectors {
		if err := s.browser.Click(ctx, sel); err == nil {
			break
		}
	}

	return nil
}
