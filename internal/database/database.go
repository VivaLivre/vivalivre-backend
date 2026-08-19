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
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(50) DEFAULT 'user'`,
		`ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL`,
		`ALTER TABLE bathrooms ADD COLUMN IF NOT EXISTS operating_hours JSONB DEFAULT '{"type": "unknown"}'::jsonb`,
		`ALTER TABLE bathrooms ADD COLUMN IF NOT EXISTS observations TEXT`,
		`ALTER TABLE bathrooms ADD COLUMN IF NOT EXISTS user_id INTEGER REFERENCES users(id) ON DELETE SET NULL`,
		`ALTER TABLE bathrooms ADD COLUMN IF NOT EXISTS has_changing_table BOOLEAN DEFAULT false`,
		`ALTER TABLE bathrooms ADD COLUMN IF NOT EXISTS is_free BOOLEAN DEFAULT false`,
		`ALTER TABLE bathrooms ADD COLUMN IF NOT EXISTS photo_url TEXT`,
		`ALTER TABLE bathrooms ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'approved'`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'active'`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS height INTEGER`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS weight DOUBLE PRECISION`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS birth_date DATE`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS cpf VARCHAR(14)`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS date_of_birth DATE`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS gender VARCHAR(50)`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS clinical_condition TEXT`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS comorbidities JSONB DEFAULT '[]'::jsonb`,

		`CREATE TABLE IF NOT EXISTS review_helpful_votes (
			id SERIAL PRIMARY KEY,
			review_id INTEGER NOT NULL REFERENCES bathroom_reviews(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			is_helpful BOOLEAN NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(review_id, user_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_reviews_bathroom ON bathroom_reviews(bathroom_id)`,
		`CREATE INDEX IF NOT EXISTS idx_reviews_user ON bathroom_reviews(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_reviews_created ON bathroom_reviews(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_helpful_votes_review ON review_helpful_votes(review_id)`,

		`CREATE TABLE IF NOT EXISTS bathroom_reports (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			bathroom_id INTEGER NOT NULL REFERENCES bathrooms(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			reason VARCHAR(50) NOT NULL,
			description TEXT,
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS bathroom_suggestions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			bathroom_id INTEGER NOT NULL REFERENCES bathrooms(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			suggested_updates JSONB NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_bathroom_reports_bathroom ON bathroom_reports(bathroom_id)`,
		`CREATE INDEX IF NOT EXISTS idx_bathroom_suggestions_bathroom ON bathroom_suggestions(bathroom_id)`,
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
