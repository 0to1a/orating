package company

// Roles validate app-side via IsValidRole; the DB column has no CHECK
// constraint so adding a role doesn't require a migration.
var availableRoles = []Role{
	{Value: "admin", Label: "Admin"},
	{Value: "sales", Label: "Sales"},
	{Value: "finance", Label: "Finance"},
	{Value: "member", Label: "Member"},
}

func ListRoles() []Role {
	out := make([]Role, len(availableRoles))
	copy(out, availableRoles)
	return out
}

func IsValidRole(role string) bool {
	for _, r := range availableRoles {
		if r.Value == role {
			return true
		}
	}
	return false
}

const (
	RoleAdmin   = "admin"
	RoleSales   = "sales"
	RoleFinance = "finance"
	RoleMember  = "member"
)
