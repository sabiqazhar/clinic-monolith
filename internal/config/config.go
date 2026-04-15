package config

import "os"

type Config struct {
	ServerPort  string
	PostgresURL string
	RedisPort   string
	RabbitMQURL string
	MysqlURL    string
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func Load() *Config {
	return &Config{
		ServerPort:  getEnv("SERVER", ""),
		PostgresURL: getEnv("PG_URL", ""),
		RedisPort:   getEnv("REDIS_PORT", ""),
		RabbitMQURL: getEnv("RABBITMQ_URL", ""),
		MysqlURL:    getEnv("MYSQL_URL", ""),
	}
}

// BuildMysqlURL constructs MySQL URL from individual env vars
// Use this if MYSQL_URL not set directly
func BuildMysqlURL() string {
	host := getEnv("MYSQL_HOST", "127.0.0.1")
	port := getEnv("MYSQL_PORT", "3306")
	user := getEnv("MYSQL_USER", "")
	pass := getEnv("MYSQL_PASSWORD", "")
	db := getEnv("MYSQL_DATABASE", "")

	if user == "" || db == "" {
		return getEnv("MYSQL_URL", "")
	}
	return user + ":" + pass + "@tcp(" + host + ":" + port + ")/" + db + "?parseTime=true"
}

// BuildRedisAddr constructs Redis address from individual env vars
func BuildRedisAddr() string {
	host := getEnv("REDIS_HOST", "127.0.0.1")
	port := getEnv("REDIS_PORT", "6379")
	return host + ":" + port
}
