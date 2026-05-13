package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Bathroom struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Address      string  `json:"address"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	IsAccessible bool    `json:"is_accessible"`
	Distance     float64 `json:"distance"`
	CreatedAt    string  `json:"created_at"`
}

func main() {
	// Load .env
	godotenv.Load()

	baseURL := "http://localhost:8080"
	testEmail := "test@vivalivre.com"
	testPassword := "TestPassword123!"
	testName := "Test User"

	fmt.Println("🚀 VivaLivre API Test Script")
	fmt.Println("═══════════════════════════════════════════════════════════")

	// Step 1: Register user
	fmt.Println("\n1️⃣  Registering test user...")
	registerReq := RegisterRequest{
		Name:     testName,
		Email:    testEmail,
		Password: testPassword,
	}

	registerBody, _ := json.Marshal(registerReq)
	resp, err := http.Post(
		fmt.Sprintf("%s/auth/register", baseURL),
		"application/json",
		bytes.NewBuffer(registerBody),
	)
	if err != nil {
		log.Fatalf("❌ Registration failed: %v\n", err)
	}
	defer resp.Body.Close()

	var authResp AuthResponse
	json.NewDecoder(resp.Body).Decode(&authResp)

	if resp.StatusCode != 201 && resp.StatusCode != 409 {
		log.Fatalf("❌ Registration error: %d\n", resp.StatusCode)
	}

	if resp.StatusCode == 201 {
		fmt.Printf("✅ User registered: %s (%s)\n", authResp.User.Name, authResp.User.Email)
	} else {
		fmt.Println("ℹ️  User already exists, proceeding to login...")
	}

	// Step 2: Login
	fmt.Println("\n2️⃣  Logging in...")
	loginReq := LoginRequest{
		Email:    testEmail,
		Password: testPassword,
	}

	loginBody, _ := json.Marshal(loginReq)
	resp, err = http.Post(
		fmt.Sprintf("%s/auth/login", baseURL),
		"application/json",
		bytes.NewBuffer(loginBody),
	)
	if err != nil {
		log.Fatalf("❌ Login failed: %v\n", err)
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&authResp)
	if resp.StatusCode != 200 {
		log.Fatalf("❌ Login error: %d\n", resp.StatusCode)
	}

	token := authResp.Token
	fmt.Printf("✅ Login successful\n")
	fmt.Printf("🔑 Token: %s...\n", token[:20])

	// Step 3: Get nearby bathrooms
	fmt.Println("\n3️⃣  Fetching nearby bathrooms...")
	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/bathrooms/nearby?lat=-23.5615&lng=-46.6558&radius=5000", baseURL),
		nil,
	)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("❌ Request failed: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("❌ Error: %d - %s\n", resp.StatusCode, string(body))
	}

	var bathrooms []Bathroom
	json.NewDecoder(resp.Body).Decode(&bathrooms)

	fmt.Printf("✅ Found %d bathrooms\n", len(bathrooms))
	fmt.Println("\n📍 Bathrooms near São Paulo:")
	fmt.Println("─────────────────────────────────────────────────────────────")

	if len(bathrooms) == 0 {
		fmt.Println("⚠️  No bathrooms found. Run insert_test_bathrooms.go first!")
	} else {
		for i, b := range bathrooms {
			accessible := "❌"
			if b.IsAccessible {
				accessible = "✓"
			}
			fmt.Printf("%d. %s\n", i+1, b.Name)
			fmt.Printf("   📍 %s\n", b.Address)
			fmt.Printf("   📌 Lat: %.4f, Lng: %.4f | Distance: %.0fm | Accessible: %s\n",
				b.Latitude, b.Longitude, b.Distance, accessible)
		}
	}

	fmt.Println("\n═══════════════════════════════════════════════════════════")
	fmt.Println("✅ Test completed successfully!")
}
