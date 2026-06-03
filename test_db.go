package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	dbURL := os.Getenv("DB_URL")
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	query := `
		SELECT id, name, address, photo_url, is_accessible, has_changing_table, is_free, comment, created_at
		FROM bathrooms
		WHERE status = 'pending'
		ORDER BY created_at ASC
	`
	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, address string
		var photoURL *string
		var isAccessible bool
		var hasChangingTable *bool
		var isFree *bool
		var comment *string
		var createdAt string

		err := rows.Scan(&id, &name, &address, &photoURL, &isAccessible, &hasChangingTable, &isFree, &comment, &createdAt)
		if err != nil {
			fmt.Printf("Scan error for row: %v\n", err)
		} else {
			fmt.Printf("Scanned ID: %d\n", id)
		}
	}
	
	if err := rows.Err(); err != nil {
		fmt.Printf("Rows error: %v\n", err)
	}
	fmt.Println("Done")
}
