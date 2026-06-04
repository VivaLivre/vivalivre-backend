package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

func main() {
	query := `
		[out:json][timeout:90];
		area["name"="São Paulo"]["admin_level"="8"]->.sp;
		(
			node["leisure"="park"](area.sp);
			way["leisure"="park"](area.sp);
			node["shop"="mall"](area.sp);
			way["shop"="mall"](area.sp);
			node["amenity"="toilets"](area.sp);
			way["amenity"="toilets"](area.sp);
			node["shop"="supermarket"](area.sp);
			way["shop"="supermarket"](area.sp);
		);
		out count;
	`

	encodedQuery := url.QueryEscape(query)
	req, err := http.NewRequest("GET", "https://lz4.overpass-api.de/api/interpreter?data="+encodedQuery, nil)
	if err != nil {
		log.Fatalf("Erro ao criar request: %v", err)
	}
	req.Header.Set("User-Agent", "VivaLivreApp/1.0 (test script)")

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

	fmt.Println(string(body))
}
