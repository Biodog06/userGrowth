package user

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"usergrowth/mysql"
)

//var userMap sync.Map
//
//// 给每个 user 一个用户id，比较方便
//var idMap sync.Map
//
//var id = 1

func Register(msq *mysql.MyDB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		username := ctx.PostForm("username")
		password := ctx.PostForm("password")

		//if _, loaded := userMap.LoadOrStore(username, md5.Sum([]byte(password))); !loaded {
		//	idMap.LoadOrStore(username, strconv.Itoa(id))
		//	id = id + 1
		//	ctx.JSON(http.StatusOK, gin.H{
		//		"message": "register success",
		//		"data": gin.H{
		//			"Name": username,
		//			"Pass": password,
		//		},
		//		"code": 200,
		//	})
		//} else {
		//	ctx.JSON(http.StatusOK, gin.H{
		//		"message": "user already exists",
		//	})
		//}

		//MYSQL VERSION
		data := md5.Sum([]byte(password))
		hashPass := hex.EncodeToString(data[:])
		user := &Users{
			Username: username,
			Password: hashPass,
		}
		repo := NewUserRepository(msq.DB)
		if err := repo.CreateUser(user); err != nil {
			if strings.Contains(err.Error(), "user already exists") {
				ctx.JSON(http.StatusOK, gin.H{
					"message": "user already exists",
				})
			} else {
				fmt.Println(err)
				return
			}
		} else {
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
}
