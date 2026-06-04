package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/gabrieljose2004/vivalivre-backend/internal/database"
	"github.com/joho/godotenv"
)

type RawLocation struct {
	Name           string  `json:"name"`
	Address        string  `json:"address"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	Category       string  `json:"category"`
	OperatingHours string  `json:"operating_hours"`
	IsAccessible   bool    `json:"is_accessible"`
	IsFree         bool    `json:"is_free"`
}

func parseOperatingHours(raw string) string {
	if raw == "" || raw == "Desconhecido / Não informado" {
		return `{"type":"unknown"}`
	}
	if raw == "24/7" || raw == "24 hours" {
		return `{"type":"24h"}`
	}

	re := regexp.MustCompile(`(\d{2}:\d{2})[^\d]*(\d{2}:\d{2})`)
	matches := re.FindStringSubmatch(raw)
	if len(matches) == 3 {
		openTime := matches[1]
		closeTime := matches[2]
		return fmt.Sprintf(`{"type":"custom", "schedule":{"1":{"open":"%s","close":"%s"}, "2":{"open":"%s","close":"%s"}, "3":{"open":"%s","close":"%s"}, "4":{"open":"%s","close":"%s"}, "5":{"open":"%s","close":"%s"}, "6":{"open":"%s","close":"%s"}, "7":{"open":"%s","close":"%s"}}}`,
			openTime, closeTime, openTime, closeTime, openTime, closeTime, openTime, closeTime, openTime, closeTime, openTime, closeTime, openTime, closeTime)
	}

	// If we can't parse it but it's not empty, fallback to unknown
	return `{"type":"unknown"}`
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		err = godotenv.Load("../../.env")
	}

	db := database.GetDB()
	ctx := context.Background()

	file, err := os.Open("../scripts/banheiros_sp_regioes_raw.json")
	if err != nil {
		log.Fatalf("Erro ao abrir json: %v", err)
	}
	defer file.Close()

	var locations []RawLocation
	if err := json.NewDecoder(file).Decode(&locations); err != nil {
		log.Fatalf("Erro ao decodificar json: %v", err)
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

	count := 0
	for _, loc := range locations {
		// Only parse what is needed
		opHours := parseOperatingHours(loc.OperatingHours)

		hasChangingTable := false
		if loc.Category == "mall" {
			hasChangingTable = true
		}

		_, err := db.Exec(ctx, query,
			loc.Name, loc.Address, loc.Longitude, loc.Latitude, loc.IsAccessible, hasChangingTable, loc.IsFree, opHours, "approved",
		)
		if err != nil {
			log.Printf("Erro ao inserir %s: %v", loc.Name, err)
		} else {
			count++
		}
	}

	fmt.Printf("\nDone! Successfully inserted %d locations from json.\n", count)
}
