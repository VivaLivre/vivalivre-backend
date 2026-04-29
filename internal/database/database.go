package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pool *pgxpool.Pool
	once sync.Once
)

// GetDB returns a singleton database connection pool
func GetDB() *pgxpool.Pool {
	once.Do(func() {
		dbURL := os.Getenv("DB_URL")
		if dbURL == "" {
			log.Fatal("DB_URL environment variable is not set")
		}

		config, err := pgxpool.ParseConfig(dbURL)
		if err != nil {
			log.Fatalf("Unable to parse database URL: %v", err)
		}

		ctx := context.Background()
		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			log.Fatalf("Unable to connect to database: %v", err)
		}

		// Verify connection
		err = pool.Ping(ctx)
		if err != nil {
			log.Fatalf("Unable to ping database: %v", err)
		}

		fmt.Println("Successfully connected to PostgreSQL/PostGIS")
	})

	return pool
}

// CloseDB closes the database connection pool
func CloseDB() {
	if pool != nil {
		pool.Close()
	}
}
