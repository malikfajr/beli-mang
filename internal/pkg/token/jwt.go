package token

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtClaim struct {
	Username string `json:"username"`
	Admin    bool   `json:"admin"`
	jwt.RegisteredClaims
}

var secret []byte

func init() {
	secret = []byte(os.Getenv("JWT_SECRET"))
}

func CreateToken(username string, admin bool) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &JwtClaim{
		Admin:    admin,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
		},
	})

	ss, err := token.SignedString(secret)
	if err != nil {
		panic(err)
	}

	return ss
}

// jwt claim token
func ClaimToken(token string) (*JwtClaim, error) {
	parsed, err := jwt.ParseWithClaims(token, &JwtClaim{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if parsed.Method != jwt.SigningMethodHS256 {
		return nil, errors.New("Invalid token")
	}

	if claim, ok := parsed.Claims.(*JwtClaim); ok {
		return claim, nil
	} else {
		return nil, errors.New("Invalid token")
	}
}
