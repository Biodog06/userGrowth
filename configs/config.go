package config

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"
)

type ConfigManager struct {
	Config *Config
	cfg    *gcfg.Config
}

type Config struct {
	App           AppConfig           `yaml:"app"`
	MySQL         MySQLConfig         `yaml:"mysql"`
	Redis         RedisConfig         `yaml:"redis"`
	Elasticsearch ElasticsearchConfig `yaml:"elasticsearch"`
	JWT           JWTConfig           `yaml:"jwt"`
	Middleware    MiddlewareConfig    `yaml:"middleware"`
	Tracing       TracingConfig       `yaml:"tracing"`
}

type MiddlewareConfig struct {
	Error  bool `yaml:"error" default:"true"`
	Access bool `yaml:"access" default:"true"`
	JWT    bool `yaml:"jwt" default:"true"`
}

type AppConfig struct {
	Name    string `yaml:"name"`
	Port    string `yaml:"port" env:"APP_PORT" default:"8080" required:"true"`
	LogPath string `yaml:"logPath" env:"APP_LOG_PATH" default:"./logs"`
}

type MySQLConfig struct {
	Host string `yaml:"host" default:"localhost" required:"true"`
	Port int    `yaml:"port" default:"3306" required:"true"`
	User string `yaml:"user" required:"true"`
	Pass string `yaml:"pass" required:"true"`
	DB   string `yaml:"db" required:"true"`
}

type RedisConfig struct {
	Host string `yaml:"host" default:"localhost" required:"true"`
	Port int    `yaml:"port" default:"6379" required:"true"`
	Pass string `yaml:"pass"`
}

type ElasticsearchConfig struct {
	Host         string `yaml:"host" default:"localhost" required:"true"`
	Port         int    `yaml:"port" default:"9200" required:"true"`
	MaxQueueSize int    `yaml:"maxQueueSize" default:"100"`
	Workers      int    `yaml:"workers" default:"3"`
	LogIndex     string `yaml:"logIndex" required:"true"`
	MaxBatchSize int    `yaml:"maxBatchSize" default:"20"`
}

type JWTConfig struct {
	Secret string        `yaml:"secret" default:"test"`
	Expire time.Duration `yaml:"expire" default:"1h"`
}

type TracingConfig struct {
	Endpoint    string `yaml:"endpoint" required:"true"`
	Path        string `yaml:"path" default:"/v1/traces"`
	ServiceName string `yaml:"serviceName" default:"gf-growth"`
}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		Config: &Config{},
		cfg:    g.Cfg(),
	}
}

func (c *ConfigManager) LoadConfigWithReflex(filePath string) {

	adapter := g.Cfg().GetAdapter().(*gcfg.AdapterFile)

	adapter.SetFileName(filePath)

	val, err := c.cfg.Get(gctx.GetInitCtx(), ".")
	if err != nil {
		fmt.Println("get config error:", err)
		return
	}
	if err = val.Scan(c.Config); err != nil {
		fmt.Println("scan config error:", err)
		return
	}
	c.ConfigReflexIterator(reflect.ValueOf(c.Config).Elem())
}

func (c *ConfigManager) ConfigReflexIterator(v reflect.Value) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Kind() == reflect.Struct {
			c.ConfigReflexIterator(v.Field(i))
		} else {
			field := v.FieldByName(t.Field(i).Name)
			env := t.Field(i).Tag.Get("env")

			if env != "" {
				value := os.Getenv(env)
				if value != "" {
					c.setFieldValue(value, field)
				}
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

func (c *ConfigManager) setFieldValue(value string, field reflect.Value) {
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
		if err != nil {
			fmt.Println(err)
		}
		return
	default:
		panic("unknown type")
	}
}
func (c *ConfigManager) StartWatcher(path string) {
	if adapter, ok := c.cfg.GetAdapter().(gcfg.WatcherAdapter); ok {
		adapter.AddWatcher(path, func(ctx context.Context) {
			c.LoadConfigWithReflex(path)
		})
	}
}
func (c *ConfigManager) PrintConfig() {
	c.ConfigReflexIterator(reflect.ValueOf(c.Config).Elem())
	fmt.Println("App Name:", c.Config.App.Name)
	fmt.Println("App Port:", c.Config.App.Port)
	fmt.Println("App LogPath:", c.Config.App.LogPath)
	fmt.Println("MySQL Host:", c.Config.MySQL.Host)
	fmt.Println("MySQL Port:", c.Config.MySQL.Port)
	fmt.Println("MySQL User:", c.Config.MySQL.User)
	fmt.Println("MySQL Pass:", c.Config.MySQL.Pass)
	fmt.Println("MySQL DB:", c.Config.MySQL.DB)
	fmt.Println("Redis Host:", c.Config.Redis.Host)
	fmt.Println("Redis Port:", c.Config.Redis.Port)
	fmt.Println("Redis Pass:", c.Config.Redis.Pass)
	fmt.Println("Elasticsearch Host:", c.Config.Elasticsearch.Host)
	fmt.Println("JWT Secret:", c.Config.JWT.Secret)
	fmt.Println("JWT Expire:", c.Config.JWT.Expire)
	fmt.Println("Tracing Endpoint:", c.Config.Tracing.Endpoint)
	fmt.Println("Tracing Path:", c.Config.Tracing.Path)
	fmt.Println("Tracing ServiceName:", c.Config.Tracing.ServiceName)
}
