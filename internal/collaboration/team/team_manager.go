package team

import (
	"context"
	"fmt"
	"time"

	"nix-ai-help/pkg/logger"
)

// TeamManager handles team management for collaborative configuration
type TeamManager struct {
	logger *logger.Logger
	teams  map[string]*Team
	users  map[string]*User
}

// NewTeamManager creates a new team manager
func NewTeamManager(logger *logger.Logger) *TeamManager {
	return &TeamManager{
		logger: logger,
		teams:  make(map[string]*Team),
		users:  make(map[string]*User),
	}
}

// Team represents a configuration team
type Team struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	CreatedAt   time.Time          `json:"created_at"`
	CreatedBy   string             `json:"created_by"`
	Members     map[string]*Member `json:"members"`
	Settings    *TeamSettings      `json:"settings"`
	Workspaces  []string           `json:"workspaces"`
	Active      bool               `json:"active"`
}

// User represents a system user
type User struct {
	ID          string           `json:"id"`
	Username    string           `json:"username"`
	Email       string           `json:"email"`
	DisplayName string           `json:"display_name"`
	CreatedAt   time.Time        `json:"created_at"`
	LastActive  time.Time        `json:"last_active"`
	Teams       []string         `json:"teams"`
	Preferences *UserPreferences `json:"preferences"`
	Profile     *UserProfile     `json:"profile"`
	Active      bool             `json:"active"`
}

// Member represents a team member with role and permissions
type Member struct {
	UserID      string          `json:"user_id"`
	Role        Role            `json:"role"`
	Permissions map[string]bool `json:"permissions"`
	JoinedAt    time.Time       `json:"joined_at"`
	InvitedBy   string          `json:"invited_by"`
	LastActive  time.Time       `json:"last_active"`
	Active      bool            `json:"active"`
}

// Role represents user roles within a team
type Role string

const (
	RoleOwner      Role = "owner"
	RoleAdmin      Role = "admin"
	RoleMaintainer Role = "maintainer"
	RoleDeveloper  Role = "developer"
	RoleViewer     Role = "viewer"
	RoleGuest      Role = "guest"
)

// TeamSettings contains team configuration settings
type TeamSettings struct {
	AllowGuestAccess  bool   `json:"allow_guest_access"`
	RequireApproval   bool   `json:"require_approval"`
	MaxMembers        int    `json:"max_members"`
	DefaultRole       Role   `json:"default_role"`
	BranchProtection  bool   `json:"branch_protection"`
	RequireReviews    bool   `json:"require_reviews"`
	MinReviewers      int    `json:"min_reviewers"`
	AllowForcePush    bool   `json:"allow_force_push"`
	NotificationLevel string `json:"notification_level"`
}

// UserPreferences contains user preferences
type UserPreferences struct {
	Theme                string `json:"theme"`
	Language             string `json:"language"`
	Timezone             string `json:"timezone"`
	EmailNotifications   bool   `json:"email_notifications"`
	DesktopNotifications bool   `json:"desktop_notifications"`
	DefaultBranch        string `json:"default_branch"`
}

// UserProfile contains user profile information
type UserProfile struct {
	Avatar      string `json:"avatar"`
	Bio         string `json:"bio"`
	Location    string `json:"location"`
	Website     string `json:"website"`
	Company     string `json:"company"`
	PublicEmail bool   `json:"public_email"`
}

// CreateTeam creates a new team
func (tm *TeamManager) CreateTeam(ctx context.Context, name, description, createdBy string) (*Team, error) {
	if name == "" {
		return nil, fmt.Errorf("team name is required")
	}

	// Check if team name already exists
	for _, team := range tm.teams {
		if team.Name == name {
			return nil, fmt.Errorf("team with name %s already exists", name)
		}
	}

	teamID := tm.generateID("team")
	team := &Team{
		ID:          teamID,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		CreatedBy:   createdBy,
		Members:     make(map[string]*Member),
		Settings: &TeamSettings{
			AllowGuestAccess:  false,
			RequireApproval:   true,
			MaxMembers:        50,
			DefaultRole:       RoleDeveloper,
			BranchProtection:  true,
			RequireReviews:    true,
			MinReviewers:      1,
			AllowForcePush:    false,
			NotificationLevel: "normal",
		},
		Workspaces: []string{},
		Active:     true,
	}

	// Add creator as owner
	if createdBy != "" {
		team.Members[createdBy] = &Member{
			UserID:      createdBy,
			Role:        RoleOwner,
			Permissions: tm.getRolePermissions(RoleOwner),
			JoinedAt:    time.Now(),
			InvitedBy:   "",
			LastActive:  time.Now(),
			Active:      true,
		}
	}

	tm.teams[teamID] = team
	tm.logger.Info(fmt.Sprintf("Created team: %s (ID: %s)", name, teamID))

	return team, nil
}

// AddMember adds a member to a team
func (tm *TeamManager) AddMember(ctx context.Context, teamID, userID string, role Role, invitedBy string) error {
	team, exists := tm.teams[teamID]
	if !exists {
		return fmt.Errorf("team %s not found", teamID)
	}

	if !team.Active {
		return fmt.Errorf("team %s is inactive", teamID)
	}

	// Check if user already exists in team
	if _, exists := team.Members[userID]; exists {
		return fmt.Errorf("user %s is already a member of team %s", userID, teamID)
	}

	// Check team member limit
	if len(team.Members) >= team.Settings.MaxMembers {
		return fmt.Errorf("team %s has reached maximum member limit (%d)", teamID, team.Settings.MaxMembers)
	}

	// Validate role
	if !tm.isValidRole(role) {
		return fmt.Errorf("invalid role: %s", role)
	}

	member := &Member{
		UserID:      userID,
		Role:        role,
		Permissions: tm.getRolePermissions(role),
		JoinedAt:    time.Now(),
		InvitedBy:   invitedBy,
		LastActive:  time.Now(),
		Active:      true,
	}

	team.Members[userID] = member

	// Update user's team list
	if user, exists := tm.users[userID]; exists {
		user.Teams = tm.appendUnique(user.Teams, teamID)
	}

	tm.logger.Info(fmt.Sprintf("Added user %s to team %s with role %s", userID, teamID, role))
	return nil
}

// RemoveMember removes a member from a team
func (tm *TeamManager) RemoveMember(ctx context.Context, teamID, userID string, removedBy string) error {
	team, exists := tm.teams[teamID]
	if !exists {
		return fmt.Errorf("team %s not found", teamID)
	}

	member, exists := team.Members[userID]
	if !exists {
		return fmt.Errorf("user %s is not a member of team %s", userID, teamID)
	}

	// Prevent removing the last owner
	if member.Role == RoleOwner {
		ownerCount := 0
		for _, m := range team.Members {
			if m.Role == RoleOwner && m.Active {
				ownerCount++
			}
		}
		if ownerCount <= 1 {
			return fmt.Errorf("cannot remove the last owner from team %s", teamID)
		}
	}

	delete(team.Members, userID)

	// Update user's team list
	if user, exists := tm.users[userID]; exists {
		user.Teams = tm.removeFromSlice(user.Teams, teamID)
	}

	tm.logger.Info(fmt.Sprintf("Removed user %s from team %s", userID, teamID))
	return nil
}

// UpdateMemberRole updates a member's role
func (tm *TeamManager) UpdateMemberRole(ctx context.Context, teamID, userID string, newRole Role, updatedBy string) error {
	team, exists := tm.teams[teamID]
	if !exists {
		return fmt.Errorf("team %s not found", teamID)
	}

	member, exists := team.Members[userID]
	if !exists {
		return fmt.Errorf("user %s is not a member of team %s", userID, teamID)
	}

	// Validate new role
	if !tm.isValidRole(newRole) {
		return fmt.Errorf("invalid role: %s", newRole)
	}

	// Prevent demoting the last owner
	if member.Role == RoleOwner && newRole != RoleOwner {
		ownerCount := 0
		for _, m := range team.Members {
			if m.Role == RoleOwner && m.Active {
				ownerCount++
			}
		}
		if ownerCount <= 1 {
			return fmt.Errorf("cannot demote the last owner of team %s", teamID)
		}
	}

	oldRole := member.Role
	member.Role = newRole
	member.Permissions = tm.getRolePermissions(newRole)

	tm.logger.Info(fmt.Sprintf("Updated user %s role from %s to %s in team %s", userID, oldRole, newRole, teamID))
	return nil
}

// GetTeam retrieves a team by ID
func (tm *TeamManager) GetTeam(ctx context.Context, teamID string) (*Team, error) {
	team, exists := tm.teams[teamID]
	if !exists {
		return nil, fmt.Errorf("team %s not found", teamID)
	}
	return team, nil
}

// ListTeams returns all teams
func (tm *TeamManager) ListTeams(ctx context.Context) ([]*Team, error) {
	var teams []*Team
	for _, team := range tm.teams {
		if team.Active {
			teams = append(teams, team)
		}
	}
	return teams, nil
}

// ListUserTeams returns teams for a specific user
func (tm *TeamManager) ListUserTeams(ctx context.Context, userID string) ([]*Team, error) {
	var userTeams []*Team

	for _, team := range tm.teams {
		if member, exists := team.Members[userID]; exists && member.Active && team.Active {
			userTeams = append(userTeams, team)
		}
	}

	return userTeams, nil
}

// CreateUser creates a new user
func (tm *TeamManager) CreateUser(ctx context.Context, username, email, displayName string) (*User, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	// Check if username already exists
	for _, user := range tm.users {
		if user.Username == username {
			return nil, fmt.Errorf("username %s already exists", username)
		}
	}

	userID := tm.generateID("user")
	user := &User{
		ID:          userID,
		Username:    username,
		Email:       email,
		DisplayName: displayName,
		CreatedAt:   time.Now(),
		LastActive:  time.Now(),
		Teams:       []string{},
		Preferences: &UserPreferences{
			Theme:                "default",
			Language:             "en",
			Timezone:             "UTC",
			EmailNotifications:   true,
			DesktopNotifications: true,
			DefaultBranch:        "main",
		},
		Profile: &UserProfile{
			PublicEmail: false,
		},
		Active: true,
	}

	tm.users[userID] = user
	tm.logger.Info(fmt.Sprintf("Created user: %s (ID: %s)", username, userID))

	return user, nil
}

// GetUser retrieves a user by ID
func (tm *TeamManager) GetUser(ctx context.Context, userID string) (*User, error) {
	user, exists := tm.users[userID]
	if !exists {
		return nil, fmt.Errorf("user %s not found", userID)
	}
	return user, nil
}

// CheckPermission checks if a user has a specific permission in a team
func (tm *TeamManager) CheckPermission(ctx context.Context, teamID, userID, permission string) (bool, error) {
	team, exists := tm.teams[teamID]
	if !exists {
		return false, fmt.Errorf("team %s not found", teamID)
	}

	member, exists := team.Members[userID]
	if !exists {
		return false, fmt.Errorf("user %s is not a member of team %s", userID, teamID)
	}

	if !member.Active {
		return false, fmt.Errorf("user %s is inactive in team %s", userID, teamID)
	}

	hasPermission, exists := member.Permissions[permission]
	return exists && hasPermission, nil
}

// getRolePermissions returns the permissions for a given role
func (tm *TeamManager) getRolePermissions(role Role) map[string]bool {
	permissions := make(map[string]bool)

	switch role {
	case RoleOwner:
		permissions["admin"] = true
		permissions["manage_team"] = true
		permissions["manage_members"] = true
		permissions["delete_team"] = true
		permissions["manage_settings"] = true
		permissions["read"] = true
		permissions["write"] = true
		permissions["deploy"] = true
		permissions["merge"] = true
		permissions["force_push"] = true

	case RoleAdmin:
		permissions["manage_members"] = true
		permissions["manage_settings"] = true
		permissions["read"] = true
		permissions["write"] = true
		permissions["deploy"] = true
		permissions["merge"] = true

	case RoleMaintainer:
		permissions["read"] = true
		permissions["write"] = true
		permissions["deploy"] = true
		permissions["merge"] = true

	case RoleDeveloper:
		permissions["read"] = true
		permissions["write"] = true

	case RoleViewer:
		permissions["read"] = true

	case RoleGuest:
		permissions["read"] = true
	}

	return permissions
}

// isValidRole checks if a role is valid
func (tm *TeamManager) isValidRole(role Role) bool {
	validRoles := []Role{RoleOwner, RoleAdmin, RoleMaintainer, RoleDeveloper, RoleViewer, RoleGuest}
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

// Helper functions

func (tm *TeamManager) generateID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func (tm *TeamManager) appendUnique(slice []string, item string) []string {
	for _, existing := range slice {
		if existing == item {
			return slice
		}
	}
	return append(slice, item)
}

func (tm *TeamManager) removeFromSlice(slice []string, item string) []string {
	var result []string
	for _, existing := range slice {
		if existing != item {
			result = append(result, existing)
		}
	}
	return result
}
