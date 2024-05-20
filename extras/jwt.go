package extras

import (
	"go-cinema/model"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var JwtSecret = []byte("970f054dbe5343e998214328b8cdaea1e299394dba9ecf01")

type RefreshTokenRequest struct {
	RefreshToken string
}

func CreateToken(user *model.User, t time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["username"] = user.Username
	claims["exp"] = time.Now().Add(t).Unix()

	tokenString, err := token.SignedString(JwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return JwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			expirationTime := time.Unix(int64(exp), 0)
			if time.Now().After(expirationTime) {
				return nil, errors.New("token has expired")
			}
		} else {
			return nil, errors.New("token has no expiration time")
		}
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
