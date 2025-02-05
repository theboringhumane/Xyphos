package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"xyphos/internal/models"
	"xyphos/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
	ClientSecret string `json:"client_secret"`
	ClientID     string `json:"client_id"`
	ID           string `json:"id"`
	GithubID     int    `json:"github_id"`
	Email        string `json:"email"`
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
				ID:        generateID(),
				GithubID:  githubUser.ID,
				Email:     githubUser.Email,
				Name:      githubUser.Name,
				Username:  githubUser.Login,
				AvatarURL: githubUser.AvatarURL,
			}
		}
	} else {
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
		ID:       user.ID,
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token " + err.Error()})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is not valid"})
			c.Abort()
			return
		}

		if claims.ExpiresAt.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
			c.Abort()
			return
		}

		if claims.NotBefore.After(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is not yet valid"})
			c.Abort()
			return
		}

		// Handle GitHub auth token
		if claims.GithubID != 0 {
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

			c.Set("user", user)
			c.Set("claims", claims)
			c.Next()
			return
		}

		// Handle client config token
		clientConfig, err := h.userStore.GetClientConfigByClientID(context.Background(), claims.ClientID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get client config"})
			c.Abort()
			return
		}

		if clientConfig == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Client config not found"})
			c.Abort()
			return
		}

		if clientConfig.ClientSecret != claims.ClientSecret {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid client secret"})
			c.Abort()
			return
		}

		if clientConfig.ExpiresAt.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Client config has expired"})
			c.Abort()
			return
		}

		c.Set("clientConfig", clientConfig)
		c.Set("claims", claims)
		c.Next()
	}
}

// 🆔 generateID generates a unique ID for new users
func generateID() string {
	return uuid.New().String()
}
