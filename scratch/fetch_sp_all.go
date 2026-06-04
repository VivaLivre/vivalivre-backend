package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
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

type Place struct {
	Category     string
	Name         string
	Address      string
	OpeningHours string
	Lat          float64
	Lon          float64
}

func fetchCategory(categoryName string, tagQuery string) ([]Place, error) {
	query := fmt.Sprintf(`
		[out:json][timeout:90];
		area["name"="São Paulo"]["admin_level"="8"]->.sp;
		(
			node%s(area.sp);
			way%s(area.sp);
		);
		out center;
	`, tagQuery, tagQuery)

	encodedQuery := url.QueryEscape(query)
	req, err := http.NewRequest("GET", "https://lz4.overpass-api.de/api/interpreter?data="+encodedQuery, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "VivaLivreApp/1.0 (fetch_all script)")

	client := &http.Client{Timeout: 100 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data OverpassResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	var places []Place
	for _, el := range data.Elements {
		lat := el.Lat
		lon := el.Lon
		if el.Type == "way" {
			lat = el.Center.Lat
			lon = el.Center.Lon
		}

		// Filter for São Paulo city roughly
		if lat > -23.0 || lat < -24.0 || lon > -46.0 || lon < -47.0 {
			continue
		}

		name := el.Tags["name"]
		if name == "" {
			if categoryName == "Banheiro Público" {
				name = "Banheiro Público"
			} else if categoryName == "Parque" {
				name = "Parque (Sem Nome)"
			} else {
				continue // For malls and supermarkets, name is strictly required
			}
		}

		brand := el.Tags["brand"]
		if brand != "" && !strings.Contains(name, brand) {
			name = brand + " - " + name
		}

		if categoryName == "Banheiro Público" {
			if el.Tags["description"] != "" {
				name += " (" + el.Tags["description"] + ")"
			}
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
			address = "Endereço não especificado no mapa"
		}

		// For amenity=toilets, we can accept "Endereço não especificado" because they are often just coordinate points inside a park.
		// For others, we might want to be strict, but let's allow it and the user can review.
		// Wait, the user said "ao adicionar o endereço, coloque o endereço completo, inclusive cidade e estado, se tiver, coloque até o número"
		// This implies we extract whatever we have.

		places = append(places, Place{
			Category:     categoryName,
			Name:         name,
			Address:      address,
			OpeningHours: openingHours,
			Lat:          lat,
			Lon:          lon,
		})
	}

	return places, nil
}

func main() {
	categories := []struct {
		Name  string
		Query string
	}{
		{"Parque", `["leisure"="park"]`},
		{"Banheiro Público", `["amenity"="toilets"]`},
	}

	var allPlaces []Place

	for _, cat := range categories {
		fmt.Printf("Buscando %s...\n", cat.Name)
		places, err := fetchCategory(cat.Name, cat.Query)
		if err != nil {
			log.Printf("Erro ao buscar %s: %v", cat.Name, err)
			continue
		}
		fmt.Printf("Encontrados %d %s válidos\n", len(places), cat.Name)
		allPlaces = append(allPlaces, places...)
		time.Sleep(5 * time.Second) // Be nice to Overpass
	}

	mdFile, err := os.OpenFile("../sp_places_list.md", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Erro ao abrir arquivo md: %v", err)
	}
	defer mdFile.Close()

	// Group by category
	grouped := make(map[string][]Place)
	for _, p := range allPlaces {
		grouped[p.Category] = append(grouped[p.Category], p)
	}

	for _, cat := range categories {
		places := grouped[cat.Name]
		if len(places) == 0 {
			continue
		}
		fmt.Fprintf(mdFile, "## %s (%d encontrados)\n\n", cat.Name, len(places))
		fmt.Fprintf(mdFile, "| Nome | Endereço | Horário de Funcionamento | Coordenadas (Lat, Lng) |\n")
		fmt.Fprintf(mdFile, "|---|---|---|---|\n")
		for _, p := range places {
			fmt.Fprintf(mdFile, "| %s | %s | %s | %f, %f |\n", p.Name, p.Address, p.OpeningHours, p.Lat, p.Lon)
		}
		fmt.Fprintf(mdFile, "\n")
	}

	fmt.Println("Processo concluído. Arquivo atualizado.")
}
