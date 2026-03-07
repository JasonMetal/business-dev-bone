// Package auth encrypt and compare password string.
package auth

import (
	"business-dev-bone/internal/pkg/code"
	"business-dev-bone/pkg/component-base/core"
	"business-dev-bone/pkg/component-base/errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"time"
)

const JWTHeaderName, Salt = "authorization", "hell"

// Encrypt encrypts the plain text with bcrypt.
func Encrypt(source string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(source), bcrypt.DefaultCost)
	return string(hashedBytes), err
}

// Compare compares the encrypted text with the plain text if it's the same.
func Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Sign issue a jwt token based on secretID, secretKey, iss and aud.
func Sign(secretID string, secretKey string, iss, aud string) string {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Minute).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Add(0).Unix(),
		"aud": aud,
		"iss": iss,
	}

	// create a new token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = secretID

	// Sign the token with the specified secret.
	tokenString, _ := token.SignedString([]byte(secretKey))

	return tokenString
}

func JWTVerify() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader(JWTHeaderName)
		claims, err := CheckJWT(tokenString)
		if err != nil {
			core.WriteResponse(c, err, nil)
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Set("uid", claims["uid"])
		c.Set("rid", claims["rid"])
		c.Set("userType", claims["userType"])
		c.Next()
	}
}

func CheckJWT(jwtToken string) (jwt.MapClaims, error) {
	if jwtToken == "" {
		return nil, errors.WithCode(code.ErrMissingHeader, "未提供Token")
	}

	// 解析并验证Token
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名算法: %v", token.Header["alg"])
		}
		return []byte(Salt), nil
	})

	if err != nil || !token.Valid {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, errors.WithCode(code.ErrExpired, "Token已过期")
			}
		}
		return nil, errors.WithCode(code.ErrTokenInvalid, "无效Token")
	}

	// 提取Payload声明
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	return nil, errors.WithCode(code.ErrInvalidAuthHeader, "声明解析失败")
}
