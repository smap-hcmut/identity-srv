package usecase

import "strings"

// MapEmailToRole maps user email to a role
// Returns the role if email is found in userRoles, otherwise returns default role
func (rm *RoleMapper) MapEmailToRole(email string) string {
	if role, ok := rm.userRoles[strings.ToLower(strings.TrimSpace(email))]; ok {
		return role
	}
	return rm.defaultRole
}

// GetUserRoles returns the current user roles configuration
func (rm *RoleMapper) GetUserRoles() map[string]string {
	return rm.userRoles
}

// GetDefaultRole returns the default role
func (rm *RoleMapper) GetDefaultRole() string {
	return rm.defaultRole
}
