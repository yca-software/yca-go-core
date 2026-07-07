package chi_types

import "github.com/google/uuid"

type AccessType string

const (
	AccessTypeUser   AccessType = "user"
	AccessTypeAPIKey AccessType = "api_key"
)

type JWTAccessTokenPermissionData struct {
	OrganizationID uuid.UUID `json:"organizationId"`
	Permissions    []string  `json:"permissions"` // e.g., ["members:read", "org:write"]
}

type AccessInfo struct {
	RequestID string
	IPAddress string
	UserAgent string

	Type      AccessType
	SubjectID uuid.UUID // The UserID *or* the ApiKeyID
	Email     string    // Blank for API Keys

	Roles   []JWTAccessTokenPermissionData // Both Users and API Keys use this exact array!
	IsAdmin bool                           // Always false for API Keys

	// Impersonation (Ignored for API Keys)
	ImpersonatedBy      uuid.NullUUID
	ImpersonatedByEmail string
}
