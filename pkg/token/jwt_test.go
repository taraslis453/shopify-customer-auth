package token

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignJWTToken(t *testing.T) {
	type args struct {
		claims *UniversalClaims
		secret string
	}

	timeNow := time.Now()
	testCases := []struct {
		name string
		args args
	}{
		{
			name: "positive: sign with null data",
			args: args{
				claims: &UniversalClaims{
					Iss:   "test",
					IssAt: timeNow,
					NbfAt: timeNow,
					ExpAt: time.Now().Add(time.Minute),
				},
				secret: "secret",
			},
		},
	}

	for _, tc := range testCases {
		token, err := SignJWTToken(tc.args.claims, tc.args.secret)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	}
}

func TestVerifyJWTToken(t *testing.T) {
	type args struct {
		signSecret           string
		verifySecret         string
		claims               *UniversalClaims
		notToCheckExpiration bool
	}

	timeNow := time.Now()
	testCases := []struct {
		name               string
		args               args
		expectedIss        string
		expectedIssAt      time.Time
		expectedNbfAt      time.Time
		expectedExpAt      time.Time
		expectedClaimsData interface{}
		expectedError      error
	}{
		{
			name: "positive: verify with standard",
			args: args{
				claims: &UniversalClaims{
					Iss:   "test",
					IssAt: timeNow,
					NbfAt: timeNow.Add(-time.Minute),
					ExpAt: timeNow.Add(time.Minute),
				},
				signSecret:           "signSecret",
				verifySecret:         "signSecret",
				notToCheckExpiration: false,
			},
			expectedIss:   "test",
			expectedIssAt: timeNow,
			expectedNbfAt: timeNow.Add(-time.Minute),
			expectedExpAt: timeNow.Add(time.Minute),
		},
		{
			name: "positive: wrong secret",
			args: args{
				claims: &UniversalClaims{
					Iss:   "test",
					IssAt: timeNow,
					NbfAt: timeNow.Add(-time.Second),
					ExpAt: timeNow.Add(time.Minute),
				},
				signSecret:           "secret",
				verifySecret:         "wrong",
				notToCheckExpiration: false,
			},
			expectedError: fmt.Errorf("failed to parse token: token signature is invalid: signature is invalid"),
		},
		{
			name: "positive: expired",
			args: args{
				claims: &UniversalClaims{
					Iss:   "test",
					IssAt: timeNow,
					NbfAt: timeNow.Add(-time.Second),
					ExpAt: timeNow.Add(-time.Minute),
				},
				signSecret:           "secret",
				verifySecret:         "secret",
				notToCheckExpiration: true,
			},
			expectedError: nil,
			expectedIss:   "test",
			expectedIssAt: timeNow,
			expectedNbfAt: timeNow.Add(-time.Second),
			expectedExpAt: timeNow.Add(-time.Minute),
		},
		{
			name: "positive: not valid yet (nbf)",
			args: args{
				claims: &UniversalClaims{
					Iss:   "test",
					IssAt: timeNow,
					NbfAt: timeNow.Add(time.Minute),
					ExpAt: timeNow.Add(time.Minute * 2),
				},
				signSecret:           "secret",
				verifySecret:         "secret",
				notToCheckExpiration: false,
			},
			expectedError: fmt.Errorf("failed to parse token: token has invalid claims: token is not valid yet"),
		},
	}

	for _, tc := range testCases {
		token, err := SignJWTToken(tc.args.claims, tc.args.signSecret)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		decodedClaims, err := VerifyJWTToken(VerifyJWTTokenOptions{
			Token:                token,
			Secret:               tc.args.verifySecret,
			NotToCheckExpiration: tc.args.notToCheckExpiration,
		})
		if tc.expectedError != nil {
			require.EqualValues(t, tc.expectedError.Error(), err.Error())
			continue
		}
		assert.Equalf(t, tc.expectedIss, decodedClaims.GetIssuer(), "issuer mismatched")
		assert.Equalf(t, true, equalTimeWithoutNanoseconds(tc.expectedIssAt, decodedClaims.GetIssuedAt()), "issued time mismatched")
		assert.Equalf(t, true, equalTimeWithoutNanoseconds(tc.expectedNbfAt, decodedClaims.GetNotBeforeAt()), "not before time mismatched")
		assert.Equalf(t, true, equalTimeWithoutNanoseconds(tc.expectedExpAt, decodedClaims.GetExpirationAt()), "expiration time mismatched")
	}

	// Additional test for expired token with checkExpiration set to true.
	expiredClaims := &UniversalClaims{
		Iss:   "test",
		IssAt: timeNow,
		NbfAt: timeNow.Add(-time.Second),
		ExpAt: timeNow.Add(-time.Minute),
	}
	token, err := SignJWTToken(expiredClaims, "secret")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	_, err = VerifyJWTToken(VerifyJWTTokenOptions{
		Token:                token,
		Secret:               "secret",
		NotToCheckExpiration: false,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token is expired")
}

func equalTimeWithoutNanoseconds(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.Month() == t2.Month() &&
		t1.Day() == t2.Day() && t1.Hour() == t2.Hour() &&
		t1.Minute() == t2.Minute() && t1.Second() == t2.Second()
}
