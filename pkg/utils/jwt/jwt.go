package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenType  = "access"
	RefreshTokenType = "refresh"
)

type JwtAuthenticator struct {
	secret         string
	accessExpires  int
	refreshExpires int
}

func NewJwtAuthenticator(sec string, aExp, rExp int) *JwtAuthenticator {
	return &JwtAuthenticator{secret: sec, accessExpires: aExp, refreshExpires: rExp}
}

func (a *JwtAuthenticator) NewToken(uid string, tokenType string) (string, error) {
	nowTime := time.Now()
	var token *jwt.Token
	claim := jwt.MapClaims{
		"uid": uid,
		"nbf": jwt.NewNumericDate(nowTime),
		"iat": jwt.NewNumericDate(nowTime),
	}

	if tokenType == AccessTokenType {
		claim["type"] = AccessTokenType
		claim["exp"] = jwt.NewNumericDate(nowTime.Add(time.Minute * time.Duration(a.accessExpires)))
	} else {
		claim["type"] = RefreshTokenType
		claim["exp"] = jwt.NewNumericDate(nowTime.Add(time.Hour * time.Duration(a.accessExpires)))
	}

	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	ss, err := token.SignedString([]byte(a.secret))
	if err != nil {
		return "", err
	}

	return ss, nil
}

func (a *JwtAuthenticator) VerifyToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.secret), nil
	})

	switch {
	case token.Valid:
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return "", errors.New("invalidate access token: no user info")
		}

		if claims["type"].(string) != AccessTokenType {
			return "", errors.New("invalidate access token type")
		}

		return claims["uid"].(string), nil
	case errors.Is(err, jwt.ErrTokenMalformed):
		return "", errors.New("invalidate access token formate")
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		return "", errors.New("invalidate signature of access token")
	case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
		return "", errors.New("access token expired")
	default:
		return "", errors.New("invalidate access token")
	}
}

func (a *JwtAuthenticator) RefreshToken(tokenStr string) (string, error) {
	rt, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.secret), nil
	})

	if err != nil {
		return "", err
	}

	if !rt.Valid {
		return "", errors.New("invalidate refresh token")
	}

	claims, ok := rt.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalidate refresh token, no user info")
	}

	if claims["type"].(string) != RefreshTokenType {
		return "", errors.New("invalidate refresh token type")
	}

	at, err := a.NewToken(claims["uid"].(string), AccessTokenType)
	if err != nil {
		return "", err
	}

	return at, nil
}
