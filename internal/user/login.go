package user

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"usergrowth/middleware"
	"usergrowth/mysql"
	"usergrowth/redis"

	"github.com/gin-gonic/gin"
)

func Login(rdb *redis.MyRedis, msq *mysql.MyDB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		username := ctx.PostForm("username")
		password := ctx.PostForm("password")
		md5Password := md5.Sum([]byte(password))
		hashPass := hex.EncodeToString(md5Password[:])
		
		//if checkPassword, loaded := userMap.Load(username); loaded {
		//	if checkPassword == md5Password {
		//		userVal, _ := idMap.Load(username)
		//		userid := userVal.(string)
		//		token, err := middleware.GenerateToken(userid)
		//		if err != nil {
		//			fmt.Println(err)
		//			return
		//		}
		//		fmt.Println(token)
		//		ctx.SetCookie("jwt-token", token, 15*60, "/", "", false, true)
		//		err = rdb.SetCache(token, userid, redis.JWTExpireTime, ctx)
		//		if err != nil {
		//			fmt.Println(err)
		//			return
		//		}
		//		ctx.JSON(http.StatusOK, gin.H{
		//			"message": username + " login success",
		//			"data": gin.H{
		//				"Name": username,
		//				"Pass": password,
		//			},
		//			"code": 200,
		//		})
		//	} else {
		//		ctx.JSON(http.StatusOK, gin.H{
		//			"message": "check password fail",
		//		})
		//	}
		//} else {
		//	ctx.JSON(http.StatusOK, gin.H{
		//		"message": "user not exist",
		//	})
		//}

		//MYSQL VERSION

		repo := NewUserRepository(msq.DB)
		if user, err := repo.FindUserByUsername(username); err == nil {
			if user.password == hashPass {
				token, err := middleware.GenerateToken(user.userid)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(token)
				ctx.SetCookie("jwt-token", token, 15*60, "/", "", false, true)
				if err := rdb.SetCache(token, user.userid, redis.JWTExpireTime, ctx); err != nil {
					fmt.Println(err)
					return
				}
				ctx.JSON(http.StatusOK, gin.H{
					"message": username + " login success",
					"data": gin.H{
						"Name": username,
						"Pass": password,
					},
					"code": 200,
				})
			} else {
				ctx.JSON(http.StatusOK, gin.H{
					"message": "check password fail",
				})
			}
		} else {
			if errors.Is(err, ErrUserNotFound) {
				ctx.JSON(http.StatusUnauthorized, gin.H{
					"message": "invalid username or password",
				})
				return
			} else {
				fmt.Println(err)
				return
			}
		}
	}
}
