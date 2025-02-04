package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"xyphos/internal/models"
	"xyphos/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// 🌐 GitHubConfig holds the configuration for GitHub OAuth
type GitHubConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// 🔑 GitHubAuth handles GitHub OAuth authentication
type GitHubAuth struct {
	config     *oauth2.Config
	store      store.UserStore
	httpClient *http.Client
}

// 🔐 GitHub OAuth configuration
var githubOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
	ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
	Scopes: []string{
		"user:email",
		"read:org",
	},
	Endpoint: github.Endpoint,
}

// GitHubUser 👤 GitHub user response
type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// CustomClaims 🎟️ Custom claims for JWT
type CustomClaims struct {
	GithubID int    `json:"github_id"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// Handler AuthHandler 🔄 Auth handler with user store
type Handler struct {
	userStore store.UserStore
}

// NewAuthHandler 🆕 Create new auth handler
func NewAuthHandler(userStore store.UserStore) *Handler {
	return &Handler{
		userStore: userStore,
	}
}

// InitGitHubAuthRoutes 🔄 Initialize GitHub auth routes
func (h *Handler) InitGitHubAuthRoutes(r *gin.Engine) {
	auth := r.Group("/auth/github")
	githubOAuthConfig.ClientID = os.Getenv("GITHUB_CLIENT_ID")
	githubOAuthConfig.ClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
	githubOAuthConfig.RedirectURL = os.Getenv("GITHUB_REDIRECT_URL")
	{
		auth.GET("/login", h.handleGitHubLogin)
		auth.GET("/callback", h.handleGitHubCallback)
		auth.GET("/user", h.handleGitHubUser)
	}
}

// 📝 Handle GitHub login
func (h *Handler) handleGitHubLogin(c *gin.Context) {
	url := githubOAuthConfig.AuthCodeURL("state")
	fmt.Println("Redirecting to GitHub login URL:", url)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// ✅ Handle GitHub callback
func (h *Handler) handleGitHubCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if state != "state" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	token, err := githubOAuthConfig.Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	// 🔍 Get user info from GitHub
	githubUser, err := getGitHubUser(token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}

	// 🔍 Check if user exists
	user, err := h.userStore.GetUserByGithubID(context.Background(), githubUser.ID)

	if err != nil {
		if user == nil {
			user = &models.User{
				GithubID:  githubUser.ID,
				Email:     githubUser.Email,
				Name:      githubUser.Name,
				Username:  githubUser.Login,
				AvatarURL: githubUser.AvatarURL,
			}
		}
	} else {
		// 📝 Create or update user
		// 📝 Create or update user
		user.Email = githubUser.Email
		user.Name = githubUser.Name
		user.Username = githubUser.Login
		user.AvatarURL = githubUser.AvatarURL
		user.UpdatedAt = time.Now()
	}

	if err := h.userStore.CreateUser(context.Background(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user"})
		return
	}

	// 🎟️ Generate JWT token
	jwtToken, err := generateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	// return the token
	c.JSON(http.StatusOK, gin.H{"token": jwtToken})
}

// 👤 Handle GitHub user info
func (h *Handler) handleGitHubUser(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
		return
	}

	// Extract token from "Bearer <token>"
	token := authHeader[7:]
	githubUser, err := getGitHubUser(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}

	// 🔍 Get user from database
	user, err := h.userStore.GetUserByGithubID(context.Background(), githubUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// 🔍 Get GitHub user info
func getGitHubUser(token string) (*GitHubUser, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// 🎟️ Generate JWT token
func generateJWT(user *models.User) (string, error) {
	claims := &CustomClaims{
		GithubID: user.GithubID,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// 🔒 AuthMiddleware handles JWT authentication
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		tokenString := authHeader[7:]
		claims := &CustomClaims{}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is not valid"})
			c.Abort()
			return
		}

		// Get user from database
		user, err := h.userStore.GetUserByGithubID(context.Background(), claims.GithubID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			c.Abort()
			return
		}

		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Next()
	}
}

// 🆕 NewGitHubAuth creates a new GitHub authentication handler
func NewGitHubAuth(cfg GitHubConfig, store store.UserStore) *GitHubAuth {
	config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       cfg.Scopes,
		Endpoint:     github.Endpoint,
	}

	return &GitHubAuth{
		config:     config,
		store:      store,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// 🔗 GetAuthURL returns the GitHub OAuth authorization URL
func (g *GitHubAuth) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// 🔄 ExchangeCode exchanges the OAuth code for a token and user information
func (g *GitHubAuth) ExchangeCode(ctx context.Context, code string) (*models.User, error) {
	// Exchange code for token
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get GitHub user information
	githubUser, err := g.getGitHubUser(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub user: %w", err)
	}

	// Check if user exists
	user, err := g.store.GetUserByGithubID(ctx, githubUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if user == nil {
		// Create new user
		user = &models.User{
			ID:        generateID(),
			GithubID:  githubUser.ID,
			Username:  githubUser.Login,
			Name:      githubUser.Name,
			Email:     githubUser.Email,
			AvatarURL: githubUser.AvatarURL,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := g.store.CreateUser(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// Update existing user
		user.Username = githubUser.Login
		user.Name = githubUser.Name
		user.Email = githubUser.Email
		user.AvatarURL = githubUser.AvatarURL
		user.UpdatedAt = time.Now()

		if err := g.store.UpdateUser(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	return user, nil
}

// 👤 getGitHubUser fetches user information from GitHub API
func (g *GitHubAuth) getGitHubUser(ctx context.Context, accessToken string) (*GitHubUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s: %s", resp.Status, string(body))
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &user, nil
}

// 🔒 ValidateState validates the OAuth state parameter
func (g *GitHubAuth) ValidateState(state string, expectedState string) bool {
	return state == expectedState
}

// 🔄 RefreshToken refreshes the OAuth access token
func (g *GitHubAuth) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	newToken, err := g.config.TokenSource(ctx, token).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return newToken, nil
}

// 🆔 generateID generates a unique ID for new users
func generateID() string {
	return fmt.Sprintf("usr_%d", time.Now().UnixNano())
}
