// Package company — multi-tenant company + members. Read by tygo.
package company

// Info is the per-user view of a company (with their role + ownership flag).
type CompanyInfo struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	IsOwner   bool   `json:"isOwner"`
	CreatedAt string `json:"createdAt"`
}

type CompanyCreateRequest struct {
	Name string `json:"name"`
}

type CompanyCreateResponse struct {
	Company CompanyInfo `json:"company"`
}

type SelectRequest struct {
	CompanyID int64 `json:"companyId"`
}

type CompanyListResponse struct {
	Companies []CompanyInfo `json:"companies"`
}

type Member struct {
	UserID   int64  `json:"userId"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	JoinedAt string `json:"joinedAt"`
}

type MemberListResponse struct {
	Members []Member `json:"members"`
}

type InviteRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

type InviteResponse struct {
	Member Member `json:"member"`
}

type UpdateRoleBody struct {
	Role string `json:"role"`
}

type Role struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type RolesResponse struct {
	Roles []Role `json:"roles"`
}

// ===== huma Input/Output types =====

type CompanyListOutput struct {
	Body CompanyListResponse
}

type CompanyCreateInput struct {
	Body CompanyCreateRequest
}

type CompanyCreateOutput struct {
	Body CompanyCreateResponse
}

// SelectInput uses a path param for the company ID.
type SelectInput struct {
	ID int64 `path:"id"`
}

type RolesOutput struct {
	Body RolesResponse
}

type MemberListOutput struct {
	Body MemberListResponse
}

type InviteInput struct {
	Body InviteRequest
}

type InviteOutput struct {
	Body InviteResponse
}

// RemoveMemberInput uses a path param for the user ID.
type RemoveMemberInput struct {
	UserID int64 `path:"userId"`
}

// UpdateRoleInput uses a path param for the user ID and a body for the role.
type UpdateRoleInput struct {
	UserID int64          `path:"userId"`
	Body   UpdateRoleBody `json:"body"`
}
