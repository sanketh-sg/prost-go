package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// OAuthManager handles OAuth 2.0 with Auth0
type OAuthManager struct {
	clientID string
	clientSecret string
	redirectURI string
	auth0Domain string
	httpClient *http.Client
}

func (om *OAuthManager) GetClientID() string {
    return om.clientID
}

func (om *OAuthManager) GetRedirectURI() string {
    return om.redirectURI
}

func (om *OAuthManager) GetAuth0Domain() string {
    return om.auth0Domain
}

// OAuthToken represents OAuth token response from Auth0
type OAuthToken struct {
	AccessToken 	string `json:"access_token"`
	TokenType		string	`json:"token_type"`
	ExpiresIn 		int		`json:"expires_in"`
	RefreshToken	string	`json:"refresh_token,omitempty"`
	IDToken			string	`json:"id_token,omitempty"`
}

type UserInfo struct {
	Sub      string `json:"sub"`
    Email    string `json:"email"`
    Name     string `json:"name"`
    Picture  string `json:"picture,omitempty"`
    Verified bool   `json:"email_verified"`
}

func NewOAuthManager() *OAuthManager {
	return &OAuthManager{
		clientID: os.Getenv("AUTH0_CLIENT_ID"),
		clientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
		redirectURI: os.Getenv("AUTH0_REDIRECT_URI"),
		auth0Domain: os.Getenv("AUTH0_DOMAIN"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
        },
	}
}

// GetAuthorizationURL builds Auth0 authorization URL
func(om *OAuthManager) GetAuthorizationURL(state string) string {
	params := url.Values{
		"client_id": 	{om.clientID},
		"redirect_uri": {om.redirectURI},
		"response_type":{"code"},
		"scope":		{"openid email profile"},
		"state": 		{state},
	}
	// log.Println(params)
	authURL := fmt.Sprintf("%s/authorize?%s", om.auth0Domain,params.Encode())
	return  authURL
}

// ExchangeCodeForToken exchanges authorization code for access token
// This is a backend-to-backend call (never goes through frontend)
// Authorization code is temporary and only valid once
func (om *OAuthManager) ExchangeCodeForToken(ctx context.Context, code string) (*OAuthToken, error) {
	tokenURL := fmt.Sprintf("%s/oauth/token", om.auth0Domain)

	reqBody := url.Values{
        "grant_type":    {"authorization_code"},
        "code":          {code},
        "client_id":     {om.clientID},
        "client_secret": {om.clientSecret}, // This is secret, only on backend
        "redirect_uri":  {om.redirectURI},
    }

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(reqBody.Encode()))
	if err != nil {
        return nil, fmt.Errorf("failed to create token request: %w", err)
	}

    // Set headers
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("Accept", "application/json")

	// Send request to Auth0
    resp, err := om.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to exchange code with Auth0: %w", err)
    }
    defer resp.Body.Close()

	// Check response status
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("token exchange failed: status=%d, body=%s", resp.StatusCode, string(body))
    }

	//Decode Response
	var authToken OAuthToken
	if err := json.NewDecoder(resp.Body).Decode(&authToken); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}
	    
	return &authToken, nil

}

// GetUserInfo sends a GET request to Provider to obtain UserInfo
// Uses the authtoken received from token exchange, token_id field do contain userInfo in JWT encoded format but it varies with provider on what it contains
func (om *OAuthManager) GetUserInfo(ctx context.Context, authToken OAuthToken) (*UserInfo, error) {
	userInfoUrl := fmt.Sprintf("%s/userinfo", om.auth0Domain)

	req, err := http.NewRequestWithContext(ctx,"GET",userInfoUrl,nil)
	if err != nil {
        return nil, fmt.Errorf("failed to create userinfo request: %w", err)
    }

    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken.AccessToken))

	resp, err := om.httpClient.Do(req)
	if err != nil {
        return nil, fmt.Errorf("failed to get user info from Auth0: %w", err)
    }
    defer resp.Body.Close()

	// Check response status
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("userinfo request failed: status=%d, body=%s", resp.StatusCode, string(body))
    }

    // Decode response
    var userInfo UserInfo
    if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
        return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
    }

    return &userInfo, nil
}



