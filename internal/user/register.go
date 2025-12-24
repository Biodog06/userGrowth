package user

import (
	"crypto/md5"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

var userMap sync.Map

func Register(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	if _, loaded := userMap.LoadOrStore(username, md5.Sum([]byte(password))); !loaded {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "register success",
			"data": gin.H{
				"Name": username,
				"Pass": password,
			},
			"code": 200,
		})
	}
}
