// Package apikey — programmatic auth via long-lived bearer keys ("sk_<hex>").
// Stored as SHA-256 hash; full token shown once at create time.
package apikey

// APIKeyInfo is the public view of a key. Token is set only on Create.
type APIKeyInfo struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Prefix     string `json:"prefix"`
	Token      string `json:"token,omitempty"` // only on create
	CreatedAt  string `json:"createdAt"`
	LastUsedAt string `json:"lastUsedAt,omitempty"`
	RevokedAt  string `json:"revokedAt,omitempty"`
}

type APIKeyCreateRequest struct {
	Name string `json:"name"`
}

type APIKeyCreateResponse struct {
	APIKey APIKeyInfo `json:"apiKey"`
}

type APIKeyListResponse struct {
	APIKeys []APIKeyInfo `json:"apiKeys"`
}

// ===== huma Input/Output types =====

type APIKeyListOutput struct {
	Body APIKeyListResponse
}

type APIKeyCreateInput struct {
	Body APIKeyCreateRequest
}

type APIKeyCreateOutput struct {
	Body APIKeyCreateResponse
}

type RevokeInput struct {
	ID int64 `path:"id"`
}
