package user

import (
	"reflect"
	"testing"

	"gorm.io/gorm"
)

func TestNewUserRepository(t *testing.T) {
	type args struct {
		db *gorm.DB
	}
	tests := []struct {
		name string
		args args
		want UserRepository
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewUserRepository(tt.args.db); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUserRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_userRepository_CreateUser(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		user *Users
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &userRepository{
				db: tt.fields.db,
			}
			if err := repo.CreateUser(tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("userRepository.CreateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_userRepository_FindUserByUsername(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		username string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Users
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &userRepository{
				db: tt.fields.db,
			}
			got, err := repo.FindUserByUsername(tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Fatalf("userRepository.FindUserByUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("userRepository.FindUserByUsername() = %v, want %v", got, tt.want)
			}
		})
	}
}
