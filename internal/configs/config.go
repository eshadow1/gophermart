// Package configs предоставляет функциональность для инициализации и управления
// конфигурацией приложения. Он поддерживает загрузку параметров через флаги
// командной строки и переменные окружения.
//
// Приоритет значений при инициализации:
// 1. Переменные окружения
// 2. Флаги командной строки
// 3. Значения по умолчанию (константы пакета)
package configs

import (
	"flag"
	"os"
)

const (
	// DefaultEmptyString представляет собой пустую строку
	DefaultEmptyString = ""
	// DefaultAddr — адрес HTTP-сервера приложения по умолчанию.
	DefaultAddr = "localhost:8080"
	// DefaultSystemAddr — адрес внешней системы начислений по умолчанию.
	DefaultSystemAddr = "http://localhost:8090"
	// DefaultLevelLog — уровень логирования по умолчанию.
	DefaultLevelLog = "info"
	// DefaultMigrationPath — путь к директории с миграциями базы данных по умолчанию.
	DefaultMigrationPath = "./migrations"
)

// StorageConfig описывает конфигурацию для работы с хранилищем данных
type StorageConfig struct {
	// PathDB — URI для подключения к базе данных.
	PathDB string
	// PathMigrations — путь к файлам миграций базы данных.
	PathMigrations string
}

// LogConfig описывает конфигурацию подсистемы логирования.
type LogConfig struct {
	// Level — уровень детализации логов (например, info, debug, error).
	Level string
}

// AuthConfig описывает конфигурацию для модуля аутентификации и работы с токенами.
type AuthConfig struct {
	// JWTSecret — секретный ключ для подписи и проверки JWT-токенов.
	JWTSecret []byte
	// TokenIssuer — название издателя (issuer), указываемого в JWT-токенах.
	TokenIssuer string
}

// Config является главной структурой конфигурации приложения,
// агрегирующей все настройки сервера, базы данных, логов и авторизации.
type Config struct {
	// Addr — сетевой адрес (хост:порт), на котором запускается HTTP-сервер приложения.
	Addr string
	// AddrSystem — сетевой адрес (хост:порт) внешней системы начислений.
	AddrSystem string
	// Log содержит настройки логирования.
	Log LogConfig
	// Storage содержит настройки подключения к хранилищу данных.
	Storage StorageConfig
	// Auth содержит настройки аутентификации.
	Auth AuthConfig
}

// NewConfig создает и возвращает указатель на новый экземпляр структуры Config.
func NewConfig() *Config {
	return &Config{}
}

// Init инициализирует конфигурацию.
func (c *Config) Init() {
	c.parseWithFlag()

	c.Addr = c.getEnv("RUN_ADDRESS", c.Addr)
	c.AddrSystem = c.getEnv("ACCRUAL_SYSTEM_ADDRESS", c.AddrSystem)
	c.Log.Level = c.getEnv("LEVEL_LOG", c.Log.Level)
	c.Storage.PathDB = c.getEnv("DATABASE_URI", c.Storage.PathDB)
	c.Storage.PathMigrations = c.getEnv("MIGRATION_PATH", c.Storage.PathMigrations)

	c.Auth.JWTSecret = []byte(c.getEnv("JWT_SECRET", DefaultEmptyString))
	c.Auth.TokenIssuer = c.getEnv("TOKEN_ISSUER", DefaultEmptyString)
}

func (c *Config) parseWithFlag() {
	flag.StringVar(&c.Addr, "a", DefaultAddr, "HTTP server address (host:port)")
	flag.StringVar(&c.AddrSystem, "r", DefaultSystemAddr, "Accrual system address (host:port)")
	flag.StringVar(&c.Log.Level, "l", DefaultLevelLog, "Level log")
	flag.StringVar(&c.Storage.PathDB, "d", DefaultEmptyString, "PostgreSQL URI")
	flag.StringVar(&c.Storage.PathMigrations, "m", DefaultMigrationPath, "Migrations path PostgreSQL")

	flag.Parse()
}

func (*Config) getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
