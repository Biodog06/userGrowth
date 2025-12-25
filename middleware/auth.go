package middleware

import (
	"fmt"
	"net/http"
	"time"
	"usergrowth/configs"
	"usergrowth/redis"

	"github.com/gin-gonic/gin"
)
import "github.com/golang-jwt/jwt/v5"

var jwtSecret []byte
var jwtExpireTime time.Duration

type UserClaims struct {
	UserId string `json:"userid"`
	jwt.RegisteredClaims
}

func InitJWT(cfg *config.Config) {
	jwtSecret = []byte(cfg.JWT.Secret)
	if jwtSecret == nil {
		jwtSecret = []byte("secret")
	}
	jwtExpireTime = cfg.JWT.Expire
	fmt.Println("JWT Secret:", jwtSecret)
	fmt.Println("jwtExpireTime:", jwtExpireTime)
}

func GenerateToken(userid string) (string, error) {
	ExpireTime := time.Now().Add(jwtExpireTime)
	claims := &UserClaims{
		UserId: userid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(ExpireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	} else {
		return signedToken, nil
	}
}

func ValidateToken(tokenString string, err error) (*UserClaims, error) {
	claims := &UserClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if token.Valid {
		return claims, nil
	}

	return nil, err
}

func JWTMiddleware(rdb *redis.MyRedis) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString, _ := ctx.Cookie("jwt-token")
		claims, err := ValidateToken(tokenString, nil)
		if err != nil {
			fmt.Println(err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
		}
		cache, err := rdb.GetCache(tokenString, ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "redis not found",
			})
		}
		if cache == claims.UserId {
			ctx.Set("userid", claims.UserId)
		}

		ctx.Next()
	}
}
