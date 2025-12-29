package user

import (
	"reflect"
	"testing"
	"usergrowth/internal/logs"
	"usergrowth/mysql"
	"usergrowth/redis"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

func TestLogin(t *testing.T) {
	type args struct {
		rdb        *redis.MyRedis
		msq        *mysql.MyDB
		userLogger *logs.MyLogger
	}
	tests := []struct {
		name string
		args args
		want gin.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Login(tt.args.rdb, tt.args.msq, tt.args.userLogger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Login() = %v, want %v", got, tt.want)
			}
		})
	}
}
