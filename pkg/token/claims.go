package token

import (
	"time"
)

type Claims interface {
	GetIssuer() string
	GetIssuedAt() time.Time
	GetNotBeforeAt() time.Time
	GetExpirationAt() time.Time
	GetPayload() any
}

type UniversalClaims struct {
	Iss     string
	IssAt   time.Time
	ExpAt   time.Time
	NbfAt   time.Time
	Payload any // generic data
}

type UserDataClaims struct {
	// UserID is the ID of the token owner.
	UserID string `json:"userId"`
}

func (claims UniversalClaims) GetIssuer() string {
	return claims.Iss
}

func (claims UniversalClaims) GetIssuedAt() time.Time {
	return claims.IssAt
}

func (claims UniversalClaims) GetExpirationAt() time.Time {
	return claims.ExpAt
}

func (claims UniversalClaims) GetNotBeforeAt() time.Time {
	return claims.NbfAt
}

// GetPayload will return always map[string]interface{} type because
// of JSON unmarshal.
func (claims UniversalClaims) GetPayload() any {
	return claims.Payload
}
