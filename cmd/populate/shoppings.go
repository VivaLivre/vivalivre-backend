package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gabrieljose2004/vivalivre-backend/internal/database"
)

type RawLocation struct {
	Name             string
	Address          string
	Lat              float64
	Lng              float64
	OpenTime         string
	CloseTime        string
	IsAccessible     bool
	HasChangingTable bool
}

func main() {
	os.Setenv("DB_URL", "postgresql://postgres.leteadhfnmefdmpivcll:ynwhwFQbLLZY0dwJ@aws-1-sa-east-1.pooler.supabase.com:6543/postgres?pgbouncer=true&default_query_exec_mode=exec")
	db := database.GetDB()
	ctx := context.Background()

	rawLocations := []RawLocation{
		{"Grand Plaza Shopping", "Av. Industrial, 600 - Centro, Santo André - SP", -23.648716, -46.531920, "10:00", "22:00", true, true},
		{"Shopping ABC", "Av. Pereira Barreto, 42 - Vila Gilda, Santo André - SP", -23.667739, -46.533405, "10:00", "22:00", true, true},
		{"Atrium Shopping", "R. Giovanni Battista Pirelli, 155 - Vila Homero Thon, Santo André - SP", -23.663660, -46.507441, "10:00", "22:00", true, true},
		{"São Bernardo Plaza Shopping", "Av. Rotary, 624 - Ferrazópolis, São Bernardo do Campo - SP", -23.723153, -46.543081, "10:00", "22:00", true, true},
		{"Shopping Metrópole", "Praça Samuel Sabatini, 200 - Centro, São Bernardo do Campo - SP", -23.692685, -46.550333, "10:00", "22:00", true, true},
		{"Golden Square Shopping", "Av. Kennedy, 700 - Jardim do Mar, São Bernardo do Campo - SP", -23.683628, -46.557175, "10:00", "22:00", true, true},
		{"ParkShopping São Caetano", "Alameda Terracota, 545 - Cerâmica, São Caetano do Sul - SP", -23.623943, -46.580923, "10:00", "22:00", true, true},
		{"Shopping Praça da Moça", "R. Manoel da Nóbrega, 712 - Centro, Diadema - SP", -23.691477, -46.623373, "10:00", "22:00", true, true},
		{"Mauá Plaza Shopping", "Av. Gov. Mário Covas Júnior, 01 - Centro, Mauá - SP", -23.664456, -46.459498, "10:00", "22:00", true, true},
		{"Shopping Duaik", "Rua do Comércio, 140 - Centro, Ribeirão Pires - SP", -23.712608, -46.414516, "09:00", "20:00", true, true},
	}

	query := `
		INSERT INTO bathrooms (
			name, address, location, is_accessible, has_changing_table, is_free, 
			operating_hours, status, created_at
		) VALUES (
			$1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326), $5, $6, $7, 
			$8::jsonb, $9, NOW()
		)
	`

	deleteQuery := `DELETE FROM bathrooms WHERE name = $1`

	count := 0
	for _, loc := range rawLocations {
		db.Exec(ctx, deleteQuery, loc.Name)

		opHours := fmt.Sprintf(`{"type":"custom", "schedule":{"1":{"open":"%s","close":"%s"}, "2":{"open":"%s","close":"%s"}, "3":{"open":"%s","close":"%s"}, "4":{"open":"%s","close":"%s"}, "5":{"open":"%s","close":"%s"}, "6":{"open":"%s","close":"%s"}, "7":{"open":"%s","close":"%s"}}}`, loc.OpenTime, loc.CloseTime, loc.OpenTime, loc.CloseTime, loc.OpenTime, loc.CloseTime, loc.OpenTime, loc.CloseTime, loc.OpenTime, loc.CloseTime, loc.OpenTime, loc.CloseTime, loc.OpenTime, loc.CloseTime)

		_, err := db.Exec(ctx, query,
			loc.Name, loc.Address, loc.Lng, loc.Lat, loc.IsAccessible, loc.HasChangingTable, true, opHours, "approved",
		)
		if err != nil {
			log.Printf("Failed to insert %s: %v\n", loc.Name, err)
		} else {
			count++
			fmt.Printf("Inserted %s\n", loc.Name)
		}
	}

	fmt.Printf("\nDone! Successfully inserted %d shoppings.\n", count)
}
