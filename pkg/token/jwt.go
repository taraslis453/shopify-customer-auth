package token

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func makeTimestamp(t time.Time) float64 {
	return float64(t.Unix())
}

// SignJWTToken generates signed string from claims based on the secret key.
func SignJWTToken(claims Claims, secret string) (string, error) {
	jwtMapClaims := make(jwt.MapClaims)
	jwtMapClaims["iss"] = claims.GetIssuer()
	jwtMapClaims["exp"] = makeTimestamp(claims.GetExpirationAt())

	iatTimestamp := makeTimestamp(claims.GetIssuedAt())
	if iatTimestamp != 0 {
		jwtMapClaims["iat"] = iatTimestamp // set issued at only if UNIX timestamp != 0
	}
	nbfTimestamp := makeTimestamp(claims.GetNotBeforeAt())
	if nbfTimestamp != 0 {
		jwtMapClaims["nbf"] = nbfTimestamp // set not before only if UNIX timestamp != 0
	}
	if claims.GetPayload() != nil {
		jwtMapClaims["payload"] = claims.GetPayload() // set only if payload is presented
	}

	// Sign and get the complete encoded token as a string using the secret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtMapClaims)
	return token.SignedString([]byte(secret))
}

type VerifyJWTTokenOptions struct {
	Token                string
	Secret               string
	NotToCheckExpiration bool
}

// VerifyJWTToken verifies JWT token signature and returns it's claims.
// VerifyJWTToken verifies JWT token signature and returns it's claims without checking expiration.
func VerifyJWTToken(opt VerifyJWTTokenOptions) (*UniversalClaims, error) {
	token, err := jwt.Parse(opt.Token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(opt.Secret), nil
	})

	// Skip checking error for expiration, but still check for other types of errors.
	if err != nil && !strings.Contains(err.Error(), "token is expired") {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	if !opt.NotToCheckExpiration {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return nil, fmt.Errorf("token is expired")
			}
		}
	}

	return parseJWTClaims(claims)
}

func parseJWTClaims(claims jwt.MapClaims) (*UniversalClaims, error) {
	uniClaims := new(UniversalClaims)
	iss, err := claims.GetIssuer()
	if err != nil {
		return nil, fmt.Errorf("invalid issuer")
	}
	uniClaims.Iss = iss // set issuer

	issAt, err := claims.GetIssuedAt()
	if err != nil {
		return nil, fmt.Errorf("invalid issued time")
	}
	if issAt != nil {
		uniClaims.IssAt = issAt.Time // issued at time
	}

	nbfAt, err := claims.GetNotBefore()
	if err != nil {
		return nil, fmt.Errorf("invalid not before time")
	}
	if nbfAt != nil {
		uniClaims.NbfAt = nbfAt.Time // not before time
	}

	expAt, err := claims.GetExpirationTime()
	if err != nil {
		return nil, fmt.Errorf("invalid expiration time")
	}
	uniClaims.ExpAt = expAt.Time // expire time (always should be present)

	// Parse payload. Payload can be empty, but payload type
	// always will be map[string]interface{} because of JSON unmarshaling. When JSON unmarshal structure
	// it unmarshals into generic map
	if payload, ok := claims["payload"]; ok {
		uniClaims.Payload = payload.(map[string]interface{})
	}

	return uniClaims, nil
}
