package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sanketh-sg/prost/services/users/auth"
	"github.com/sanketh-sg/prost/services/users/models"
	"github.com/sanketh-sg/prost/services/users/repository"
)

type OAuthHandler struct {
	oauthManager	*auth.OAuthManager
	jwtManager		*auth.JWTManager
	oauthProviderRepo *repository.OAuthProviderRepository
	userRepo 		repository.UserRepositoryInterface
}

func NewOAuthHandler(
    oauthManager *auth.OAuthManager, 
    jwtManager *auth.JWTManager, 
    oauthProviderRepo *repository.OAuthProviderRepository,
    userRepo repository.UserRepositoryInterface,
) *OAuthHandler {
    return &OAuthHandler{
        oauthManager: oauthManager,
        jwtManager: jwtManager,
        oauthProviderRepo: oauthProviderRepo,
        userRepo: userRepo,
    }
}

// InitiateOAuth initiates OAuth login flow
// @Summary Initiate OAuth login
// @Description Start OAuth authentication with Auth0
// @Tags auth
// @Produce json
// @Success 302 "Redirects to Auth0"
// @Router /oauth/login [get]
func (oh *OAuthHandler) InitiateOAuth(ctx *gin.Context) {
	// Generate state for CSRF protection
    state := uuid.New().String()

	log.Printf("Initiating OAuth with Auth0, state: %s",state)

	ctx.SetCookie("oauth_state",state,600,"/","",false,true)
    // log.Println("Cookie Set")
	authURL := oh.oauthManager.GetAuthorizationURL(state)
    // log.Println(authURL)
	ctx.Redirect(http.StatusTemporaryRedirect, authURL)	
}

func (oh *OAuthHandler) InitiateGmailOAuth(ctx *gin.Context){
    state := uuid.New().String()

    log.Printf("Initiating direct Gmail OAuth, state: %s", state)

    ctx.SetCookie("oauth_state", state, 600, "/", "", false, true)

    // Build Google authorization URL directly (not through Auth0)
    // This requires Auth0 to have Google as a connection
    params := url.Values{
        "client_id":     {oh.oauthManager.GetClientID()},
        "redirect_uri":  {oh.oauthManager.GetRedirectURI()},
        "response_type": {"code"},
        "scope":         {"openid email profile"},
        "state":         {state},
        "connection":    {"google-oauth2"},
    }.Encode()

    authURL := fmt.Sprintf("%s/authorize?%s", oh.oauthManager.GetAuth0Domain(), params)

    log.Printf("Gmail OAuth URL: %s", authURL)
    ctx.Redirect(http.StatusTemporaryRedirect, authURL)
}


// OAuthCallback handles OAuth callback from Auth0
// @Summary OAuth callback
// @Description Handle OAuth callback and generate JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param code query string true "Authorization code"
// @Param state query string true "State parameter"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Router /oauth/callback [get]
func (oh *OAuthHandler) OAuthCallback(c *gin.Context) {
    log.Printf("OAuth callback received:")
    log.Printf("  URL: %s", c.Request.URL.String())
    log.Printf("  Query params: %v", c.Request.URL.Query())

    code := c.Query("code")
    state := c.Query("state")
    ctx := c.Request.Context()

    if errorParam := c.Query("error"); errorParam != "" {
        errorDesc := c.Query("error_description")
        log.Printf("Auth0 error: %s - %s", errorParam, errorDesc)
        c.JSON(http.StatusBadRequest, gin.H{
            "error": errorParam,
            "message": errorDesc,
        })
        return
    }
    if code == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "authorization code missing"})
        return
    }

    // Validate state (CSRF check)
    savedState, err := c.Cookie("oauth_state")
    if err != nil || savedState != state {
        log.Printf("State validation failed: saved=%s, received=%s, err=%v", savedState, state, err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state parameter"})
        return
    }

    log.Printf("OAuth callback received with code: %s...", code[:20])

    // Step 1: Exchange authorization code for OAuth token
    token, err := oh.oauthManager.ExchangeCodeForToken(ctx, code)
    if err != nil {
        log.Printf("Token exchange failed: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "token exchange failed"})
        return
    }

    log.Printf("Token exchanged successfully, expires in: %d seconds", token.ExpiresIn)

    // Step 2: Get user info from Auth0
    userInfo, err := oh.oauthManager.GetUserInfo(ctx, *token)
    if err != nil {
        log.Printf("Failed to get user info: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
        return
    }

    log.Printf("User info retrieved from Auth0: %s (%s)", userInfo.Name, userInfo.Email)

    // Step 3: Check if OAuth provider already exists for this user
    existingProvider, err := oh.oauthProviderRepo.GetByProviderSub(ctx, "auth0", userInfo.Sub)
    var user *models.User

    if err == nil && existingProvider != nil {
        // OAuth provider exists, fetch user
        log.Printf("OAuth provider found for user: %s", existingProvider.UserID)
        user, err = oh.userRepo.GetUserByID(ctx, existingProvider.UserID)
        if err != nil {
            log.Printf("Failed to get user: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "user lookup failed"})
            return
        }
    } else {
        // Step 4a: Check if email exists (user might have registered with password before)
        existingUser, err := oh.userRepo.GetUserByEmail(ctx, userInfo.Email)
        if err == nil && existingUser != nil {
            // Email exists, link OAuth to existing user
            log.Printf("Linking OAuth provider to existing user: %s", existingUser.ID)
            user = existingUser
        } else {
            // Step 4b: Create new user
            log.Printf("Creating new user for OAuth: %s", userInfo.Email)
            user = &models.User{
                ID:       uuid.New().String(),
                Email:    userInfo.Email,
                Username: userInfo.Name,
                CreatedAt: time.Now().UTC(),
                UpdatedAt: time.Now().UTC(),
            }
            
            err := oh.userRepo.CreateUser(ctx, user)
            if err != nil {
                log.Printf("Failed to create user: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "user creation failed"})
                return
            }
            log.Printf("User created successfully: %s", user.ID)
        }
    }

    // Step 5: Link OAuth provider to user (if not already linked)
    if existingProvider == nil {
        oauthProvider := &models.OAuthProvider{
            UserID:        user.ID,
            Provider:      "auth0",
            ProviderSub:   userInfo.Sub,
            ProviderEmail: userInfo.Email,
            PictureURL:    userInfo.Picture,
        }

        err := oh.oauthProviderRepo.CreateOAuthProvider(ctx, oauthProvider)
        if err != nil {
            log.Printf("Failed to link OAuth provider: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to link OAuth provider"})
            return
        }
        log.Printf("OAuth provider linked to user: %s", user.ID)
    }
    // Step 6: Generate JWT access token
    accessToken, expiresAt, err := oh.jwtManager.GenerateToken(
        user.ID,
        user.Email,
        user.Username,
        24*time.Hour,
    )
    if err != nil {
        log.Printf("Failed to generate access token: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
        return
    }

    log.Printf("Access token generated, expires at: %v", expiresAt)

    // Step 7: Generate JWT refresh token
    refreshToken, _, err := oh.jwtManager.GenerateRefreshToken(user.ID, 7*24*time.Hour)
    if err != nil {
        log.Printf("Failed to generate refresh token: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "refresh token generation failed"})
        return
    }

    log.Printf("✓ OAuth login successful for user: %s", user.Email)

    // Return tokens and user info
    // c.JSON(http.StatusOK, models.LoginResponse{
    //     User: models.User{
    //         ID:        user.ID,
    //         Email:     user.Email,
    //         Username:  user.Username,
    //         CreatedAt: user.CreatedAt,
    //     },
    //     AccessToken:  accessToken,
    //     RefreshToken: refreshToken,
    //     ExpiresIn:    3600,
    //     TokenType:    "Bearer",
    // })

    // Redirect to frontend with tokens in URL
    frontendURL := os.Getenv("FRONTEND_URL")
    if frontendURL == "" {
        frontendURL = "http://localhost:5173"
    }

    // Build redirect URL with tokens as query parameters
    redirectURL := fmt.Sprintf(
        "%s/oauth/callback?access_token=%s&refresh_token=%s&user_id=%s&email=%s&username=%s",
        frontendURL,
        url.QueryEscape(accessToken),
        url.QueryEscape(refreshToken),
        user.ID,
        url.QueryEscape(user.Email),
        url.QueryEscape(user.Username),
    )
    
    log.Printf("Redirecting to frontend: %s", redirectURL)
    c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// RefreshToken handles token refresh using refresh token
// @Summary Refresh access token
// @Description Generate a new access token using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh_token query string true "Refresh token"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /oauth/refresh [post]
func (oh *OAuthHandler) RefreshToken(c *gin.Context) {
    refreshToken := c.Query("refresh_token")
    if refreshToken == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
        return
    }

    // Validate refresh token
    claims, err := oh.jwtManager.ValidateRefreshToken(refreshToken)
    if err != nil {
        log.Printf("Refresh token validation failed: %v", err)
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
        return
    }

    // Get user details
    ctx := c.Request.Context()
    user, err := oh.userRepo.GetUserByID(ctx, claims.UserID)
    if err != nil {
        log.Printf("User not found: %v", err)
        c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
        return
    }

    // Generate new access token
    accessToken, expiresAt, err := oh.jwtManager.GenerateToken(
        user.ID,
        user.Email,
        user.Username,
        24*time.Hour,
    )
    if err != nil {
        log.Printf("Failed to generate access token: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
        return
    }

    log.Printf("Access token refreshed for user: %s, expires at: %v", user.ID, expiresAt)

    // Return new access token
    c.JSON(http.StatusOK, gin.H{
        "access_token": accessToken,
        "expires_in":   3600,
        "token_type":   "Bearer",
    })
}