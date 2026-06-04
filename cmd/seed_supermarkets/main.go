package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gabrieljose2004/vivalivre-backend/internal/database"
	"github.com/joho/godotenv"
)

type OverpassResponse struct {
	Elements []struct {
		Type string `json:"type"`
		ID   int64  `json:"id"`
		Lat  float64 `json:"lat"`
		Lon  float64 `json:"lon"`
		Center struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"center"`
		Tags map[string]string `json:"tags"`
	} `json:"elements"`
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	db := database.GetDB()
	defer database.CloseDB()

	// Clean up previous insertions to avoid duplicates and update the addresses
	_, err := db.Exec(context.Background(), "DELETE FROM bathrooms WHERE observations = 'Banheiro localizado no interior do supermercado'")
	if err != nil {
		log.Printf("Aviso: Falha ao limpar supermercados anteriores: %v", err)
	}

	query := `
		[out:json][timeout:25];
		area["name"="São Paulo"]->.state;
		(
			area["name"="Santo André"](area.state)->.sa;
			area["name"="São Bernardo do Campo"](area.state)->.sbc;
			area["name"="São Caetano do Sul"](area.state)->.scs;
		);
		(
			node["shop"="supermarket"](area.sa);
			way["shop"="supermarket"](area.sa);
			node["shop"="supermarket"](area.sbc);
			way["shop"="supermarket"](area.sbc);
			node["shop"="supermarket"](area.scs);
			way["shop"="supermarket"](area.scs);
		);
		out center;
	`

	encodedQuery := url.QueryEscape(query)
	req, err := http.NewRequest("GET", "https://lz4.overpass-api.de/api/interpreter?data="+encodedQuery, nil)
	if err != nil {
		log.Fatalf("Erro ao criar request: %v", err)
	}
	req.Header.Set("User-Agent", "VivaLivreApp/1.0 (seed script)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Erro na requisição: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Erro ao ler body: %v", err)
	}

	var data OverpassResponse
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatalf("Erro ao parsear JSON: %v", err)
	}

	insertQuery := `
		INSERT INTO bathrooms (
			name, address, location, is_accessible, has_changing_table, is_free, status, operating_hours, observations
		) VALUES (
			$1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326), false, false, true, 'approved', $5::jsonb, 'Banheiro localizado no interior do supermercado'
		)
	`

	insertedCount := 0
	for _, el := range data.Elements {
		name := el.Tags["name"]
		if name == "" {
			name = "Supermercado (Sem Nome)"
		}

		brand := el.Tags["brand"]
		if brand != "" && !strings.Contains(name, brand) {
			name = brand + " - " + name
		}

		openingHours := el.Tags["opening_hours"]
		if openingHours == "" {
			openingHours = "Não informado"
		}

		street := el.Tags["addr:street"]
		housenumber := el.Tags["addr:housenumber"]
		suburb := el.Tags["addr:suburb"]
		city := el.Tags["addr:city"]
		state := el.Tags["addr:state"]
		postcode := el.Tags["addr:postcode"]

		var parts []string
		if street != "" {
			if housenumber != "" {
				parts = append(parts, street+", "+housenumber)
			} else {
				parts = append(parts, street)
			}
		}
		if suburb != "" {
			parts = append(parts, suburb)
		}
		if city != "" {
			parts = append(parts, city)
		}
		if state != "" {
			parts = append(parts, state)
		}
		if postcode != "" {
			parts = append(parts, "CEP "+postcode)
		}

		address := strings.Join(parts, " - ")
		if address == "" {
			address = "Não informado"
		}

		lat := el.Lat
		lon := el.Lon
		if el.Type == "way" {
			lat = el.Center.Lat
			lon = el.Center.Lon
		}

		if name != "Supermercado (Sem Nome)" && address != "Não informado" {
			if lat < -23.0 && lat > -24.5 && lon < -46.0 && lon > -47.0 {
				
				var opHoursMap map[string]interface{}
				if openingHours == "Não informado" {
					opHoursMap = map[string]interface{}{"type": "unknown"}
				} else if openingHours == "24/7" {
					opHoursMap = map[string]interface{}{"type": "24h"}
				} else {
					// We'll store it as unknown with the raw string in observations if we wanted, 
					// but for now, we'll store type: unknown since we can't easily parse OSM string into our custom format.
					// Actually, wait, let's just use unknown for everything we can't parse easily.
					opHoursMap = map[string]interface{}{"type": "unknown", "raw_osm": openingHours}
				}
				
				hoursJSON, _ := json.Marshal(opHoursMap)
				
				// PostGIS MakePoint expects (longitude, latitude)
				_, err := db.Exec(context.Background(), insertQuery, name, address, lon, lat, string(hoursJSON))
				if err != nil {
					log.Printf("Erro ao inserir %s: %v", name, err)
				} else {
					insertedCount++
				}
			}
		}
	}

	fmt.Printf("Seed concluído. %d supermercados inseridos.\n", insertedCount)
}
