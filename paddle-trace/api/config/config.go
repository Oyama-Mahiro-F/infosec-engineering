// Package config 系统配置管理
// 从环境变量读取所有配置项，支持Docker Compose和本地开发两种模式
package config

import (
	"os"
	"strconv"
)

// Config 全局配置结构体
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	XChain   XChainConfig
	JWT      JWTConfig
	NFC      NFCConfig
}

type ServerConfig struct {
	Port string
	Mode string // debug/release
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host string
	Port string
}

type XChainConfig struct {
	Nodes       []string // 区块链节点地址列表
	ChainName   string
	CryptoType  string // sm2/ecdsa
}

type JWTConfig struct {
	Secret     string
	ExpireHour int
}

type NFCConfig struct {
	DefaultAESKey string // 默认AES-128密钥（生产环境应从HSM读取）
}

// Load 从环境变量加载配置
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "paddle"),
			Password: getEnv("DB_PASSWORD", "paddle_trace_2026"),
			DBName:   getEnv("DB_NAME", "paddle_trace"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: getEnv("REDIS_PORT", "6379"),
		},
		XChain: XChainConfig{
			Nodes:      parseNodeList(),
			ChainName:  getEnv("XCHAIN_NAME", "xuper"),
			CryptoType: getEnv("XCHAIN_CRYPTO", "sm2"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "paddle-trace-jwt-secret-key-2026"),
			ExpireHour: getEnvInt("JWT_EXPIRE_HOUR", 24),
		},
		NFC: NFCConfig{
			DefaultAESKey: getEnv("NFC_DEFAULT_AES_KEY", ""),
		},
	}
}

// DSN 返回PostgreSQL连接字符串
func (d *DatabaseConfig) DSN() string {
	return "host=" + d.Host +
		" port=" + d.Port +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.DBName +
		" sslmode=" + d.SSLMode
}

// Addr 返回Redis连接地址
func (r *RedisConfig) Addr() string {
	return r.Host + ":" + r.Port
}

// ============================================================================
// 辅助函数
// ============================================================================

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func parseNodeList() []string {
	// 从环境变量读取节点地址列表
	var nodes []string
	for i := 1; i <= 7; i++ {
		key := "XCHAIN_NODE" + strconv.Itoa(i)
		if addr := os.Getenv(key); addr != "" {
			nodes = append(nodes, addr)
		}
	}
	if len(nodes) == 0 {
		// 默认本地开发地址
		nodes = []string{
			"172.25.0.11:37101",
			"172.25.0.12:37101",
			"172.25.0.13:37101",
			"172.25.0.14:37101",
			"172.25.0.15:37101",
			"172.25.0.16:37101",
			"172.25.0.17:37101",
		}
	}
	return nodes
}
