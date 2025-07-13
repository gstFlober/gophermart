package config

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	PostgresDB = "POSTGRES_DB"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Database DatabaseConfig `mapstructure:"database"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Accural  string         `mapstructure:"accural"`
}

type ServerConfig struct {
	Address         string        `mapstructure:"address"`
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Type             string                 `mapstructure:"type"`
	FileDatabase     FileDatabaseConfig     `mapstructure:"file"`
	PostgresDatabase PostgresDatabaseConfig `mapstructure:"postgres"`
}

type FileDatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type PostgresDatabaseConfig struct {
	URI          string        `mapstructure:"uri"`
	Host         string        `mapstructure:"host"`
	Port         string        `mapstructure:"port"`
	User         string        `mapstructure:"user"`
	Password     string        `mapstructure:"password"`
	DBName       string        `mapstructure:"dbname"`
	SSLMode      string        `mapstructure:"sslmode"`
	MaxOpenConns int           `mapstructure:"max_open_conns"`
	MaxIdleConns int           `mapstructure:"max_idle_conns"`
	MaxLifetime  time.Duration `mapstructure:"max_lifetime"`
	LogLevel     string        `mapstructure:"log_level"` // silent, error, warn, info
}

type AuthConfig struct {
	JWTSecret string `mapstructure:"jwt_secret"`
}
type LoggerConfig struct {
	Level  string `mapstructure:"level"`
	format string `mapstructure:"format"`
}

var (
	configInstance *Config
	configOnce     sync.Once
	configMutex    sync.RWMutex
)

func MustLoad() *Config {
	configOnce.Do(func() {
		cfg, err := loadConfig()
		if err != nil {
			panic(fmt.Sprintf("failed to load config: %v", err))
		}
		configInstance = cfg
	})
	return configInstance
}
func loadConfig() (*Config, error) {
	// Объявление флагов
	gophemartHost := flag.String("gophermart-host", "", "Server host")
	gophemartPort := flag.String("gophermart-port", "", "Server port")
	gophemartDatabaseURI := flag.String("gophermart-database-uri", "", "Database URI")
	jwtSecret := flag.String("jwt-secret", "", "JWT secret key")
	accrualHost := flag.String("accrual-host", "", "Accrual system address")
	accrualPort := flag.String("accrual-port", "", "Accrual system address")
	flag.Parse()
	accrualAddr := *accrualHost + ":" + *accrualPort

	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 1. Установка значений по умолчанию
	setDefaults(v)

	// 2. Загрузка конфигурационного файла
	configPath := os.Getenv("CONFIG_PATH")
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
		v.AddConfigPath("/etc/gophermart")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// 3. Применение аргументов командной строки (наивысший приоритет)
	if *gophemartHost != "" {
		v.Set("server.host", *gophemartHost)
	}
	if *gophemartPort != "" {
		v.Set("server.port", *gophemartPort)
	}
	if *gophemartDatabaseURI != "" {
		v.Set("database.postgres.uri", *gophemartDatabaseURI)
	}
	if *jwtSecret != "" {
		v.Set("auth.jwt_secret", *jwtSecret)
	}
	//if *accrualAddr != "" { // Обработка нового флага
	//	v.Set("accural", *accrualAddr)
	//}

	// Автоматическое обновление адреса сервера
	v.Set("server.address", net.JoinHostPort(
		v.GetString("server.host"),
		v.GetString("server.port"),
	))

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	cfg.Accural = accrualAddr
	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.address", ":8080")
	v.SetDefault("server.shutdown_timeout", 15*time.Second)

	v.SetDefault("auth.jwt_secret", "supersecretkey")

	v.SetDefault("database.type", PostgresDB)
	v.SetDefault("database.file.path", "./data/app.db")

	v.SetDefault("database.postgres.uri", "postgresql://admin:secret@127.0.0.1:5432/mydb?sslmode=disable")
	v.SetDefault("database.postgres.host", "localhost")
	v.SetDefault("database.postgres.port", "5432")
	v.SetDefault("database.postgres.user", "admin")
	v.SetDefault("database.postgres.password", "secret")
	v.SetDefault("database.postgres.dbname", "mydb")
	v.SetDefault("database.postgres.sslmode", "disable")
	v.SetDefault("database.postgres.max_open_conns", 25)
	v.SetDefault("database.postgres.max_idle_conns", 5)
	v.SetDefault("database.postgres.max_lifetime", 5*time.Minute)
	v.SetDefault("database.postgres.log_level", "warn")

	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
}
