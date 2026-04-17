package memory

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func getMySQLEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func LoadMySQLConfigFromEnv() *MySQLConfig {
	return &MySQLConfig{
		Host:     getMySQLEnvOrDefault("MYSQL_HOST", "localhost"),
		Port:     3306,
		User:     getMySQLEnvOrDefault("MYSQL_USER", "root"),
		Password: getMySQLEnvOrDefault("MYSQL_PASSWORD", "Yl300822!"),
		Database: getMySQLEnvOrDefault("MYSQL_DATABASE", "goagent"),
	}
}

func (c *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Database)
}

var (
	db     *sql.DB
	dbOnce sync.Once
	dbLock sync.Mutex
)

func GetDB() (*sql.DB, error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	if db != nil {
		return db, nil
	}

	cfg := LoadMySQLConfigFromEnv()
	var err error
	db, err = sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping mysql: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func InitMySQLSchema() error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	schemas := []string{
		`CREATE TABLE IF NOT EXISTS memories (
			id VARCHAR(64) PRIMARY KEY,
			content TEXT NOT NULL,
			memory_type VARCHAR(32) NOT NULL,
			user_id VARCHAR(64) NOT NULL,
			timestamp DATETIME NOT NULL,
			importance DOUBLE DEFAULT 0.5,
			metadata TEXT,
			INDEX idx_user_id (user_id),
			INDEX idx_memory_type (memory_type),
			INDEX idx_timestamp (timestamp)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS entities (
			id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(64),
			embedding BLOB,
			properties TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			INDEX idx_name (name),
			INDEX idx_type (type)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS relations (
			id VARCHAR(64) PRIMARY KEY,
			source_id VARCHAR(64) NOT NULL,
			target_id VARCHAR(64) NOT NULL,
			relation_type VARCHAR(64),
			weight DOUBLE DEFAULT 1.0,
			properties TEXT,
			created_at DATETIME NOT NULL,
			INDEX idx_source (source_id),
			INDEX idx_target (target_id),
			FOREIGN KEY (source_id) REFERENCES entities(id) ON DELETE CASCADE,
			FOREIGN KEY (target_id) REFERENCES entities(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS perceptual_memories (
			id VARCHAR(64) PRIMARY KEY,
			content TEXT NOT NULL,
			modality VARCHAR(32) NOT NULL,
			embedding BLOB,
			user_id VARCHAR(64) NOT NULL,
			timestamp DATETIME NOT NULL,
			importance DOUBLE DEFAULT 0.5,
			metadata TEXT,
			INDEX idx_user_id (user_id),
			INDEX idx_modality (modality),
			INDEX idx_timestamp (timestamp)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			return fmt.Errorf("failed to create schema: %w", err)
		}
	}

	return nil
}

func CloseDB() error {
	dbLock.Lock()
	defer dbLock.Unlock()

	if db != nil {
		err := db.Close()
		db = nil
		return err
	}
	return nil
}
