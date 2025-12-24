package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App   AppConfig           `yaml:"app"`
	MySQL MySQLConfig         `yaml:"mysql"`
	Redis RedisConfig         `yaml:"redis"`
	ES    ElasticsearchConfig `yaml:"elasticsearch"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Port string `yaml:"port"`
}

type MySQLConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
	DB   string `yaml:"db"`
}

type RedisConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Pass string `yaml:"pass"`
}

type ElasticsearchConfig struct {
	Host string `yaml:"host"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) LoadConfig(path string) {
	// Implementation for loading configuration from a file
	byteData, _ := os.ReadFile(path)
	yaml.Unmarshal(byteData, c)
}

func (c *Config) PrintConfig() {
	fmt.Println("App Name:", c.App.Name)
	fmt.Println("App Port:", c.App.Port)
	fmt.Println("MySQL Host:", c.MySQL.Host)
	fmt.Println("MySQL Port:", c.MySQL.Port)
	fmt.Println("MySQL User:", c.MySQL.User)
	fmt.Println("MySQL Pass:", c.MySQL.Pass)
	fmt.Println("MySQL DB:", c.MySQL.DB)
	fmt.Println("Redis Host:", c.Redis.Host)
	fmt.Println("Redis Port:", c.Redis.Port)
	fmt.Println("Redis Pass:", c.Redis.Pass)
	fmt.Println("Elasticsearch Host:", c.ES.Host)
}
