package middleware

import (
	"fmt"
	"net/http"
	"time"
	config "usergrowth/configs"
	"usergrowth/internal/logs"
	"usergrowth/redis"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel/attribute"
)

var jwtSecret []byte
var jwtExpireTime time.Duration

type UserClaims struct {
	UserId string `json:"userid"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	rdb        redis.Cache
	userLogger logs.Logger
	cfg        *config.MiddlewareConfig
}

func NewJWTManager(rdb redis.Cache, userLogger logs.Logger, cfg *config.MiddlewareConfig) *JWTManager {
	return &JWTManager{
		rdb:        rdb,
		userLogger: userLogger,
		cfg:        cfg,
	}
}

func InitJWT(cfg *config.Config) {
	jwtSecret = []byte(cfg.JWT.Secret)
	if jwtSecret == nil {
		jwtSecret = []byte("secret")
	}
	jwtExpireTime = cfg.JWT.Expire
	fmt.Println("JWT Secret:", string(jwtSecret))
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
	}

	return signedToken, nil
}

func ValidateToken(tokenString string) (*UserClaims, error) {
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
	return nil, fmt.Errorf("token is nil")
}

func ParseTokenUnverified(tokenString string) (*UserClaims, error) {
	claims := &UserClaims{}
	parser := jwt.NewParser()
	_, _, err := parser.ParseUnverified(tokenString, claims)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func (m *JWTManager) JWTHandler(r *ghttp.Request) {
	if !m.cfg.JWT {
		r.Middleware.Next()
		return
	}
	tokenString := r.Cookie.Get("jwt-token").String()
	ctx := r.GetCtx()
	ctx, span := gtrace.NewSpan(ctx, "Middleware.JWTHandler")
	defer span.End()
	r.SetCtx(ctx)

	// 如果没有 Token，直接返回未授权，不要进入验证逻辑
	if tokenString == "" {
		m.userLogger.Info(ctx, "access denied: missing token", "ip", r.GetClientIp())
		r.Response.WriteStatus(http.StatusUnauthorized)
		r.Response.WriteJson(ghttp.DefaultHandlerResponse{
			Code:    http.StatusUnauthorized,
			Message: "未登录或Token已过期",
			Data:    nil,
		})
		r.Exit()
		return
	}

	claims, err := ValidateToken(tokenString)
	if err != nil {
		// 记录无效 Token 尝试（可能是伪造攻击或过期），记录 IP
		m.userLogger.Info(ctx, "access denied: invalid token", "ip", r.GetClientIp(), "error", err.Error())
		r.Response.WriteStatus(http.StatusUnauthorized)
		r.Response.WriteJson(ghttp.DefaultHandlerResponse{
			Code:    http.StatusUnauthorized,
			Message: "无效的Token",
			Data:    nil,
		})
		r.Exit()
		return
	}

	cache, err := m.rdb.GetCache(tokenString, ctx)
	if err != nil {
		// 记录 Redis 验证失败（可能是强制登出或 Redis 故障），记录 UserID 和 IP
		m.userLogger.Info(ctx, "access denied: session expired", "userid", claims.UserId, "ip", r.GetClientIp())
		r.Response.WriteStatus(http.StatusUnauthorized)
		r.Response.WriteJson(ghttp.DefaultHandlerResponse{
			Code:    http.StatusUnauthorized,
			Message: "会话已过期，请重新登录",
			Data:    nil,
		})
		r.Exit()
		return
	}
	if cache == claims.UserId {
		r.SetCtxVar("userid", claims.UserId)
		span.SetAttributes(attribute.String("user.id", claims.UserId))
	}

	r.Middleware.Next()
}
