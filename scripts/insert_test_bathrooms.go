package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	// Connection string
	connStr := "postgres://vivalivre:vivalivre@localhost:5432/vivalivre_db"

	// Connect to database
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	// Test bathrooms data
	bathrooms := []map[string]interface{}{
		{
			"name":          "Banheiro Adaptado Centro",
			"address":       "Av. Paulista, 1000 - São Paulo",
			"latitude":      -23.5615,
			"longitude":     -46.6558,
			"is_accessible": true,
		},
		{
			"name":          "Banheiro Acessível Pinheiros",
			"address":       "Rua Bandeira, 500 - São Paulo",
			"latitude":      -23.5505,
			"longitude":     -46.7038,
			"is_accessible": true,
		},
		{
			"name":          "Banheiro Vila Mariana",
			"address":       "Av. Imirim, 200 - São Paulo",
			"latitude":      -23.5800,
			"longitude":     -46.6200,
			"is_accessible": true,
		},
		{
			"name":          "Banheiro Consolação",
			"address":       "Rua Augusta, 1500 - São Paulo",
			"latitude":      -23.5450,
			"longitude":     -46.6650,
			"is_accessible": false,
		},
		{
			"name":          "Banheiro Higienópolis",
			"address":       "Av. Higienópolis, 800 - São Paulo",
			"latitude":      -23.5350,
			"longitude":     -46.6450,
			"is_accessible": true,
		},
		{
			"name":          "Banheiro Liberdade",
			"address":       "Rua Galvão Bueno, 300 - São Paulo",
			"latitude":      -23.5650,
			"longitude":     -46.6350,
			"is_accessible": true,
		},
		{
			"name":          "Banheiro Bela Vista",
			"address":       "Av. Paulista, 2000 - São Paulo",
			"latitude":      -23.5550,
			"longitude":     -46.6600,
			"is_accessible": false,
		},
		{
			"name":          "Banheiro Jardins",
			"address":       "Rua Oscar Freire, 500 - São Paulo",
			"latitude":      -23.5650,
			"longitude":     -46.6700,
			"is_accessible": true,
		},
	}

	// Insert bathrooms
	query := `
		INSERT INTO bathrooms (name, address, location, is_accessible, created_at)
		VALUES ($1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326), $5, NOW())
	`

	for _, bathroom := range bathrooms {
		_, err := conn.Exec(context.Background(), query,
			bathroom["name"],
			bathroom["address"],
			bathroom["longitude"],
			bathroom["latitude"],
			bathroom["is_accessible"],
		)
		if err != nil {
			log.Printf("Error inserting bathroom %s: %v\n", bathroom["name"], err)
			continue
		}
		fmt.Printf("✓ Inserted: %s\n", bathroom["name"])
	}

	// Verify insertion
	var count int
	err = conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM bathrooms").Scan(&count)
	if err != nil {
		log.Fatalf("Error counting bathrooms: %v\n", err)
	}

	fmt.Printf("\n✓ Total bathrooms in database: %d\n", count)

	// List all bathrooms
	rows, err := conn.Query(context.Background(), `
		SELECT id, name, address, ST_Y(location::geometry) as latitude, ST_X(location::geometry) as longitude, is_accessible
		FROM bathrooms
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Fatalf("Error querying bathrooms: %v\n", err)
	}
	defer rows.Close()

	fmt.Println("\nBathrooms in database:")
	fmt.Println("─────────────────────────────────────────────────────────────")
	for rows.Next() {
		var id int
		var name, address string
		var latitude, longitude float64
		var isAccessible bool

		err = rows.Scan(&id, &name, &address, &latitude, &longitude, &isAccessible)
		if err != nil {
			log.Printf("Error scanning row: %v\n", err)
			continue
		}

		accessible := "❌"
		if isAccessible {
			accessible = "✓"
		}

		fmt.Printf("ID: %d | %s | %s | Lat: %.4f, Lng: %.4f | Accessible: %s\n",
			id, name, address, latitude, longitude, accessible)
	}
}
