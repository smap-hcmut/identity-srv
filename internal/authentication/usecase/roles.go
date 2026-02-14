package usecase

// MapGroupsToRole maps user groups to a role
// Returns the highest privilege role if user belongs to multiple groups
func (rm *RoleMapper) MapGroupsToRole(userGroups []string) string {
	if len(userGroups) == 0 {
		return rm.defaultRole
	}

	// Create a set of user groups for fast lookup
	userGroupSet := make(map[string]bool)
	for _, group := range userGroups {
		userGroupSet[group] = true
	}

	// Find all matching roles
	matchedRoles := make([]string, 0)
	for role, groups := range rm.roleMapping {
		for _, group := range groups {
			if userGroupSet[group] {
				matchedRoles = append(matchedRoles, role)
				break // Found a match for this role, move to next role
			}
		}
	}

	// If no roles matched, return default role
	if len(matchedRoles) == 0 {
		return rm.defaultRole
	}

	// If only one role matched, return it
	if len(matchedRoles) == 1 {
		return matchedRoles[0]
	}

	// Multiple roles matched - return highest privilege role
	return rm.selectHighestPrivilegeRole(matchedRoles)
}

// selectHighestPrivilegeRole selects the role with highest privilege
func (rm *RoleMapper) selectHighestPrivilegeRole(roles []string) string {
	highestRole := roles[0]
	highestPriority := rolePriority[highestRole]

	for _, role := range roles[1:] {
		priority := rolePriority[role]
		if priority > highestPriority {
			highestRole = role
			highestPriority = priority
		}
	}

	return highestRole
}

// GetRoleMapping returns the current role mapping configuration
func (rm *RoleMapper) GetRoleMapping() map[string][]string {
	return rm.roleMapping
}

// GetDefaultRole returns the default role
func (rm *RoleMapper) GetDefaultRole() string {
	return rm.defaultRole
}
