package user

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strconv"
	"usergrowth/middleware"

	"github.com/gin-gonic/gin"
)

var userid = 1

func Login(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	md5Password := md5.Sum([]byte(password))

	if checkPassword, loaded := userMap.Load(username); loaded {
		if checkPassword == md5Password {
			token, err := middleware.GenerateToken(strconv.Itoa(userid))
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(token)
			if value, err := ctx.Cookie("jwt-token"); err == nil {
				if value == "" {
					userid = userid + 1
					// 在 JSON前 SetCookie 否则不会设置
					ctx.SetCookie("jwt-token", token, 15*60, "/", "", false, true)
				}
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
