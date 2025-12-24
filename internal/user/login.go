package user

import (
	"crypto/md5"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Login(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	md5Password := md5.Sum([]byte(password))

	if checkPassword, loaded := userMap.Load(username); loaded {
		if checkPassword == md5Password {
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
