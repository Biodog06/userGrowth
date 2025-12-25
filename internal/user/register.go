package user

import (
	"crypto/md5"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

var userMap sync.Map

// 给每个 user 一个用户id，比较方便
var idMap sync.Map

var id = 1

func Register(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	if _, loaded := userMap.LoadOrStore(username, md5.Sum([]byte(password))); !loaded {
		idMap.LoadOrStore(username, strconv.Itoa(id))
		id = id + 1
		ctx.JSON(http.StatusOK, gin.H{
			"message": "register success",
			"data": gin.H{
				"Name": username,
				"Pass": password,
			},
			"code": 200,
		})
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "user already exists",
		})
	}
}
