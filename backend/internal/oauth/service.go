package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mahendrakalkura/saas-blueprint/internal/config"
	"github.com/mahendrakalkura/saas-blueprint/internal/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type Service struct {
	googleConfig *oauth2.Config
	githubConfig *oauth2.Config
}

// UserInfo represents user information from OAuth provider
type UserInfo struct {
	ProviderUserID string
	Email          string
	Name           string
	AvatarURL      string
}

func NewService(cfg *config.OAuthConfig) *Service {
	return &Service{
		googleConfig: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
		githubConfig: &oauth2.Config{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			RedirectURL:  cfg.GitHubRedirectURL,
			Scopes: []string{
				"user:email",
				"read:user",
			},
			Endpoint: github.Endpoint,
		},
	}
}

// GetAuthURL returns the OAuth authorization URL for the specified provider
func (s *Service) GetAuthURL(provider, state string) (string, error) {
	switch provider {
	case models.OAuthProviderGoogle:
		return s.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
	case models.OAuthProviderGitHub:
		return s.githubConfig.AuthCodeURL(state), nil
	default:
		return "", fmt.Errorf("unsupported OAuth provider: %s", provider)
	}
}

// ExchangeCode exchanges an authorization code for an access token
func (s *Service) ExchangeCode(ctx context.Context, provider, code string) (*oauth2.Token, error) {
	switch provider {
	case models.OAuthProviderGoogle:
		return s.googleConfig.Exchange(ctx, code)
	case models.OAuthProviderGitHub:
		return s.githubConfig.Exchange(ctx, code)
	default:
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}
}

// GetUserInfo fetches user information from the OAuth provider
func (s *Service) GetUserInfo(ctx context.Context, provider string, token *oauth2.Token) (*UserInfo, error) {
	switch provider {
	case models.OAuthProviderGoogle:
		return s.getGoogleUserInfo(ctx, token)
	case models.OAuthProviderGitHub:
		return s.getGitHubUserInfo(ctx, token)
	default:
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}
}

func (s *Service) getGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := s.googleConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var data struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %w", err)
	}

	return &UserInfo{
		ProviderUserID: data.ID,
		Email:          data.Email,
		Name:           data.Name,
		AvatarURL:      data.Picture,
	}, nil
}

func (s *Service) getGitHubUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := s.githubConfig.Client(ctx, token)

	// Get user profile
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var data struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %w", err)
	}

	// If email is not public, fetch it from emails endpoint
	email := data.Email
	if email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailResp.Body.Close()
			emailBody, err := io.ReadAll(emailResp.Body)
			if err == nil {
				var emails []struct {
					Email   string `json:"email"`
					Primary bool   `json:"primary"`
				}
				if json.Unmarshal(emailBody, &emails) == nil {
					for _, e := range emails {
						if e.Primary {
							email = e.Email
							break
						}
					}
				}
			}
		}
	}

	name := data.Name
	if name == "" {
		name = data.Login
	}

	return &UserInfo{
		ProviderUserID: fmt.Sprintf("%d", data.ID),
		Email:          email,
		Name:           name,
		AvatarURL:      data.AvatarURL,
	}, nil
}
