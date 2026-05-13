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

// EnsureRatingsSchema creates ratings tables/indexes if missing.
func EnsureRatingsSchema() error {
	if pool == nil {
		return fmt.Errorf("database pool is not initialized")
	}

	ctx := context.Background()
	statements := []string{
		`CREATE TABLE IF NOT EXISTS bathroom_reviews (
			id SERIAL PRIMARY KEY,
			bathroom_id INTEGER NOT NULL REFERENCES bathrooms(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
			title VARCHAR(100),
			comment TEXT,
			cleanliness_rating INTEGER CHECK (cleanliness_rating >= 1 AND cleanliness_rating <= 5),
			accessibility_rating INTEGER CHECK (accessibility_rating >= 1 AND accessibility_rating <= 5),
			spaciousness_rating INTEGER CHECK (spaciousness_rating >= 1 AND spaciousness_rating <= 5),
			helpful_count INTEGER DEFAULT 0,
			unhelpful_count INTEGER DEFAULT 0,
			status VARCHAR(20) DEFAULT 'approved',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(bathroom_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS review_helpful_votes (
			id SERIAL PRIMARY KEY,
			review_id INTEGER NOT NULL REFERENCES bathroom_reviews(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			is_helpful BOOLEAN NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(review_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS review_photos (
			id SERIAL PRIMARY KEY,
			review_id INTEGER NOT NULL REFERENCES bathroom_reviews(id) ON DELETE CASCADE,
			photo_url TEXT NOT NULL,
			uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_reviews_bathroom ON bathroom_reviews(bathroom_id)`,
		`CREATE INDEX IF NOT EXISTS idx_reviews_user ON bathroom_reviews(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_reviews_created ON bathroom_reviews(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_helpful_votes_review ON review_helpful_votes(review_id)`,
	}

	for _, stmt := range statements {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return err
		}
	}

	return nil
}

// CloseDB closes the database connection pool
func CloseDB() {
	if pool != nil {
		pool.Close()
	}
}
