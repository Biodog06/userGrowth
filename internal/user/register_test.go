package user

import (
	"reflect"
	"testing"
	"usergrowth/internal/logs"
	"usergrowth/mysql"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

func TestRegister(t *testing.T) {
	type args struct {
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
			if got := Register(tt.args.msq, tt.args.userLogger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Register() = %v, want %v", got, tt.want)
			}
		})
	}
}
