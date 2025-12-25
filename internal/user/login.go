package user

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"usergrowth/middleware"
	"usergrowth/redis"

	"github.com/gin-gonic/gin"
)

func Login(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	md5Password := md5.Sum([]byte(password))

	if checkPassword, loaded := userMap.Load(username); loaded {
		if checkPassword == md5Password {
			userVal, _ := idMap.Load(username)
			userid := userVal.(string)
			token, err := middleware.GenerateToken(userid)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(token)
			ctx.SetCookie("jwt-token", token, 15*60, "/", "", false, true)
			err = redis.SetCache(token, userid, redis.JWTExpireTime, ctx)
			if err != nil {
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
		ctx.JSON(http.StatusOK, gin.H{
			"message": "user not exist",
		})
	}
}
