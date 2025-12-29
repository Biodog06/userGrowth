package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App   AppConfig           `yaml:"app"`
	MySQL MySQLConfig         `yaml:"mysql"`
	Redis RedisConfig         `yaml:"redis"`
	ES    ElasticsearchConfig `yaml:"elasticsearch"`
	JWT   JWTConfig           `yaml:"jwt"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Port string `yaml:"port" env:"APP_PORT" default:"8080" required:"true"`
}

type MySQLConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port" default:"3306"`
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
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	MaxQueueSize int    `yaml:"maxQueueSize"`
	Workers      int    `yaml:"workers"`
	LogIndex     string `yaml:"logIndex"`
	MaxBatchSize int    `yaml:"maxBatchSize"`
}

type JWTConfig struct {
	Secret string        `yaml:"secret"`
	Expire time.Duration `yaml:"expire"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) LoadConfig(path string) {
	// Implementation for loading configuration from a file
	byteData, _ := os.ReadFile(path)
	err := yaml.Unmarshal(byteData, c)
	if err != nil {
		fmt.Println("yaml Unmarshal err:", err)
		return
	}
}

func (c *Config) LoadConfigWithReflex(path string) {
	byteData, _ := os.ReadFile(path)
	err := yaml.Unmarshal(byteData, c)
	if err != nil {
		fmt.Println("yaml Unmarshal err:", err)
		return
	}
	c.ConfigReflexIterator(reflect.ValueOf(c).Elem())
}

func (c *Config) ConfigReflexIterator(v reflect.Value) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Kind() == reflect.Struct {
			c.ConfigReflexIterator(v.Field(i))
		} else {
			field := v.FieldByName(t.Field(i).Name)
			env := t.Field(i).Tag.Get("env")

			if env != "" {
				value := os.Getenv(env)
				c.setFieldValue(value, field)
			}
			defaultValue := t.Field(i).Tag.Get("default")
			// 比较是否为空
			if reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) && defaultValue != "" {
				c.setFieldValue(defaultValue, field)
			}
			if required := t.Field(i).Tag.Get("required"); required == "true" {
				if reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) {
					err := fmt.Errorf("required field %s is required", t.Field(i).Name)
					fmt.Println(err)
					return
				}
			}
		}
	}
}

func (c *Config) setFieldValue(value string, field reflect.Value) {
	// reflect 不支持将 string 类型转换为其他类型，所以处理一下
	if !field.CanSet() {
		fmt.Println("can not set field:", field)
		return
	}
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			// 处理 time.Duration
			d, err := time.ParseDuration(value)
			if err == nil {
				field.SetInt(int64(d))
				return
			}
		}
		val, err := strconv.Atoi(value)
		if err == nil {
			field.SetInt(int64(val))
		}
		fmt.Println(err)
		return
	default:
		panic("unknown type")
	}
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
	fmt.Println("JWT Secret:", c.JWT.Secret)
	fmt.Println("JWT Expire:", c.JWT.Expire)
}
