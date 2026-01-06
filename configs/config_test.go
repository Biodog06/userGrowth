package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigWithReflex(t *testing.T) {
	cm := NewConfigManager()
	// Assuming config.yaml is in the same directory as this test file
	cm.LoadConfigWithReflex("config.yaml")

	// Assert that config is not nil
	assert.NotNil(t, cm.Config)

	// Assert App config
	assert.Equal(t, "usergrowth", cm.Config.App.Name)
	assert.Equal(t, "8080", cm.Config.App.Port)
	// logPath in yaml is "./logs", check if it matches
	assert.Equal(t, "./logs", cm.Config.App.LogPath)

	// Assert MySQL config
	assert.Equal(t, "localhost", cm.Config.MySQL.Host)
	assert.Equal(t, 3306, cm.Config.MySQL.Port)
	assert.Equal(t, "root", cm.Config.MySQL.User)
	assert.Equal(t, "123456", cm.Config.MySQL.Pass)
	assert.Equal(t, "usergrowth", cm.Config.MySQL.DB)

	// Assert Middleware config
	assert.True(t, *cm.Config.Middleware.Error)
	assert.True(t, *cm.Config.Middleware.Access)
	assert.False(t, *cm.Config.Middleware.JWT)

	// Print config for visual verification
	fmt.Printf("Loaded Config: %+v\n", cm.Config)
}
