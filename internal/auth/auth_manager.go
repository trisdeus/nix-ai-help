package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nix-ai-help/internal/collaboration/team"
	"nix-ai-help/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

// AuthManager handles user authentication and session management
type AuthManager struct {
	users       map[string]*AuthUser
	sessions    map[string]*Session
	teamManager *team.TeamManager
	logger      *logger.Logger
	storePath   string
}

// AuthUser represents a user in the authentication system
type AuthUser struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	DisplayName  string    `json:"display_name"`
	Role         string    `json:"role"` // admin, user
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	LastLogin    time.Time `json:"last_login"`
}

// Session represents an active user session
type Session struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Active    bool      `json:"active"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Token   string      `json:"token,omitempty"`
	User    *PublicUser `json:"user,omitempty"`
}

// PublicUser represents user data safe for public consumption
type PublicUser struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	DisplayName string   `json:"display_name"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(teamManager *team.TeamManager, logger *logger.Logger) (*AuthManager, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	storePath := filepath.Join(configDir, "nixai", "auth")
	if err := os.MkdirAll(storePath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create auth directory: %w", err)
	}

	am := &AuthManager{
		users:       make(map[string]*AuthUser),
		sessions:    make(map[string]*Session),
		teamManager: teamManager,
		logger:      logger,
		storePath:   storePath,
	}

	// Load existing users
	if err := am.loadUsers(); err != nil {
		am.logger.Warn(fmt.Sprintf("Failed to load users: %v", err))
	}

	// Create default admin user if no users exist
	if len(am.users) == 0 {
		if err := am.createDefaultAdmin(); err != nil {
			return nil, fmt.Errorf("failed to create default admin: %w", err)
		}
	}

	return am, nil
}

// createDefaultAdmin creates a default admin user
func (am *AuthManager) createDefaultAdmin() error {
	username := "admin"
	password := "nixai-admin-2024" // Should be changed on first login
	email := "admin@nixai.local"
	displayName := "NixAI Administrator"

	_, err := am.CreateUser(username, email, displayName, password, "admin")
	if err != nil {
		return err
	}

	am.logger.Info("Created default admin user (username: admin, password: nixai-admin-2024)")
	am.logger.Warn("Please change the default admin password immediately!")

	return nil
}

// CreateUser creates a new user
func (am *AuthManager) CreateUser(username, email, displayName, password, role string) (*AuthUser, error) {
	// Validate input
	if username == "" || password == "" || email == "" {
		return nil, fmt.Errorf("username, email, and password are required")
	}

	if role == "" {
		role = "user"
	}

	// Check if username already exists
	for _, user := range am.users {
		if strings.ToLower(user.Username) == strings.ToLower(username) {
			return nil, fmt.Errorf("username '%s' already exists", username)
		}
		if strings.ToLower(user.Email) == strings.ToLower(email) {
			return nil, fmt.Errorf("email '%s' already exists", email)
		}
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate user ID
	userID := am.generateID("user")

	user := &AuthUser{
		ID:           userID,
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
		DisplayName:  displayName,
		Role:         role,
		Active:       true,
		CreatedAt:    time.Now(),
	}

	am.users[userID] = user

	// Create corresponding team user
	if am.teamManager != nil {
		_, err := am.teamManager.CreateUser(nil, username, email, displayName)
		if err != nil {
			am.logger.Warn(fmt.Sprintf("Failed to create team user: %v", err))
		}
	}

	// Save users
	if err := am.saveUsers(); err != nil {
		am.logger.Error(fmt.Sprintf("Failed to save users: %v", err))
	}

	am.logger.Info(fmt.Sprintf("Created user: %s (ID: %s)", username, userID))
	return user, nil
}

// Authenticate validates user credentials and creates a session
func (am *AuthManager) Authenticate(username, password string) (*LoginResponse, error) {
	// Find user
	var user *AuthUser
	for _, u := range am.users {
		if strings.ToLower(u.Username) == strings.ToLower(username) && u.Active {
			user = u
			break
		}
	}

	if user == nil {
		return &LoginResponse{
			Success: false,
			Message: "Invalid username or password",
		}, nil
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return &LoginResponse{
			Success: false,
			Message: "Invalid username or password",
		}, nil
	}

	// Create session
	token, err := am.createSession(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login
	user.LastLogin = time.Now()
	if err := am.saveUsers(); err != nil {
		am.logger.Error(fmt.Sprintf("Failed to update last login: %v", err))
	}

	publicUser := &PublicUser{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        user.Role,
		Permissions: am.getRolePermissions(user.Role),
	}

	return &LoginResponse{
		Success: true,
		Message: "Login successful",
		Token:   token,
		User:    publicUser,
	}, nil
}

// ValidateSession validates a session token
func (am *AuthManager) ValidateSession(token string) (*PublicUser, error) {
	session, exists := am.sessions[token]
	if !exists || !session.Active || time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("invalid or expired session")
	}

	user, exists := am.users[session.UserID]
	if !exists || !user.Active {
		return nil, fmt.Errorf("user not found or inactive")
	}

	return &PublicUser{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        user.Role,
		Permissions: am.getRolePermissions(user.Role),
	}, nil
}

// Logout invalidates a session
func (am *AuthManager) Logout(token string) error {
	if session, exists := am.sessions[token]; exists {
		session.Active = false
		am.logger.Info(fmt.Sprintf("User %s logged out", session.UserID))
	}
	return nil
}

// createSession creates a new session for a user
func (am *AuthManager) createSession(userID string) (string, error) {
	token := am.generateToken()
	session := &Session{
		Token:     token,
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour sessions
		Active:    true,
	}

	am.sessions[token] = session
	return token, nil
}

// generateToken generates a secure random token
func (am *AuthManager) generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateID generates a unique ID
func (am *AuthManager) generateID(prefix string) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), len(am.users))))
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(hash[:8]))
}

// getRolePermissions returns permissions for a role
func (am *AuthManager) getRolePermissions(role string) []string {
	switch role {
	case "admin":
		return []string{"read", "write", "admin", "fleet", "teams", "builder", "user_management"}
	case "user":
		return []string{"read", "write", "builder"}
	default:
		return []string{"read"}
	}
}

// loadUsers loads users from disk
func (am *AuthManager) loadUsers() error {
	usersPath := filepath.Join(am.storePath, "users.json")
	data, err := os.ReadFile(usersPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No users file exists yet
		}
		return err
	}

	var users map[string]*AuthUser
	if err := json.Unmarshal(data, &users); err != nil {
		return err
	}

	am.users = users
	return nil
}

// saveUsers saves users to disk
func (am *AuthManager) saveUsers() error {
	usersPath := filepath.Join(am.storePath, "users.json")
	data, err := json.MarshalIndent(am.users, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(usersPath, data, 0600)
}

// ListUsers returns all users (admin only)
func (am *AuthManager) ListUsers() []*PublicUser {
	var users []*PublicUser
	for _, user := range am.users {
		if user.Active {
			users = append(users, &PublicUser{
				ID:          user.ID,
				Username:    user.Username,
				Email:       user.Email,
				DisplayName: user.DisplayName,
				Role:        user.Role,
				Permissions: am.getRolePermissions(user.Role),
			})
		}
	}
	return users
}

// ChangePassword changes a user's password
func (am *AuthManager) ChangePassword(userID, currentPassword, newPassword string) error {
	user, exists := am.users[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	user.PasswordHash = string(passwordHash)
	return am.saveUsers()
}
