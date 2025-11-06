package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/moabdelazem/k8s-app/pkg/logger"
	"go.uber.org/zap"
)

// DB holds the database connection pool
var DB *sql.DB

// Config represents database configuration
type Config struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	MaxRetries      int           // Maximum number of connection retry attempts
	RetryDelay      time.Duration // Initial delay between retries
}

// NewConnection creates a new database connection pool with retry logic
func NewConnection(cfg *Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	// Set default retry values if not provided
	maxRetries := cfg.MaxRetries
	if maxRetries == 0 {
		maxRetries = 5 // Default to 5 retries
	}

	retryDelay := cfg.RetryDelay
	if retryDelay == 0 {
		retryDelay = 2 * time.Second // Default to 2 seconds initial delay
	}

	var db *sql.DB
	var err error

	// Try to open database connection
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Retry logic with exponential backoff
	for attempt := 1; attempt <= maxRetries; attempt++ {
		logger.Info("Attempting database connection",
			zap.Int("attempt", attempt),
			zap.Int("max_retries", maxRetries),
			zap.String("host", cfg.Host),
			zap.String("database", cfg.DBName),
		)

		err = db.Ping()
		if err == nil {
			// Connection successful
			logger.Info("Database connection established",
				zap.String("host", cfg.Host),
				zap.String("port", cfg.Port),
				zap.String("database", cfg.DBName),
				zap.Int("attempts", attempt),
			)
			DB = db
			return db, nil
		}

		// Connection failed
		logger.Warn("Database connection attempt failed",
			zap.Int("attempt", attempt),
			zap.Int("max_retries", maxRetries),
			zap.Error(err),
		)

		// If this was the last attempt, return the error
		if attempt == maxRetries {
			db.Close()
			return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
		}

		// Calculate backoff delay with exponential increase
		backoffDelay := retryDelay * time.Duration(attempt)
		logger.Info("Retrying database connection",
			zap.Duration("delay", backoffDelay),
			zap.Int("next_attempt", attempt+1),
		)

		time.Sleep(backoffDelay)
	}

	// This should never be reached, but just in case
	db.Close()
	return nil, fmt.Errorf("failed to connect to database after %d attempts", maxRetries)
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// Ping checks if the database connection is alive
func Ping() error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	return DB.Ping()
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	return DB
}

// Stats returns database connection pool statistics
func Stats() sql.DBStats {
	if DB == nil {
		return sql.DBStats{}
	}
	return DB.Stats()
}
