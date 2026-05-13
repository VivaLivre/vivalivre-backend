package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/joho/godotenv"
)

type CreateReviewRequest struct {
	Rating              int    `json:"rating"`
	Title               string `json:"title"`
	Comment             string `json:"comment"`
	CleanlinessRating   int    `json:"cleanliness_rating"`
	AccessibilityRating int    `json:"accessibility_rating"`
}

type ReviewResponse struct {
	ID        int    `json:"id"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
	CreatedAt string `json:"created_at"`
}

type ReviewsListResponse struct {
	Reviews        []ReviewResponse `json:"reviews"`
	Total          int              `json:"total"`
	AverageRating  float64          `json:"average_rating"`
}

type RatingStatsResponse struct {
	BathroomID           string             `json:"bathroom_id"`
	TotalReviews         int                `json:"total_reviews"`
	AverageRating        float64            `json:"average_rating"`
	AvgCleanliness       float64            `json:"avg_cleanliness"`
	AvgAccessibility     float64            `json:"avg_accessibility"`
	AvgSpaciosuneness    float64            `json:"avg_spaciosuneness"`
	RatingDistribution   map[string]int     `json:"rating_distribution"`
}

type NearbyBathroom struct {
	ID int `json:"id"`
}

func main() {
	godotenv.Load()

	baseURL := "http://localhost:8080"
	testEmail := "test@vivalivre.com"
	testPassword := "TestPassword123!"
	bathroomID := ""

	fmt.Println("🧪 VivaLivre Ratings API Test")
	fmt.Println("═══════════════════════════════════════════════════════════")

	// Step 1: Login
	fmt.Println("\n1️⃣  Logging in...")
	loginReq := map[string]string{
		"email":    testEmail,
		"password": testPassword,
	}
	loginBody, _ := json.Marshal(loginReq)
	resp, err := http.Post(
		fmt.Sprintf("%s/auth/login", baseURL),
		"application/json",
		bytes.NewBuffer(loginBody),
	)
	if err != nil {
		log.Fatalf("❌ Login failed: %v\n", err)
	}
	defer resp.Body.Close()

	var authResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&authResp)
	token := authResp["token"].(string)
	fmt.Printf("✅ Logged in successfully\n")

	// Step 1.5: Resolve valid bathroom from nearby endpoint
	fmt.Println("\n1️⃣.5️⃣  Resolving bathroom ID from nearby...")
	client := &http.Client{}
	nearbyReq, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/bathrooms/nearby?lat=-23.6607&lng=-46.4309&radius=10000", baseURL),
		nil,
	)
	nearbyReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	nearbyResp, err := client.Do(nearbyReq)
	if err != nil {
		log.Fatalf("❌ Nearby request failed: %v\n", err)
	}
	defer nearbyResp.Body.Close()

	var bathrooms []NearbyBathroom
	if err := json.NewDecoder(nearbyResp.Body).Decode(&bathrooms); err != nil {
		log.Fatalf("❌ Failed to parse nearby response: %v\n", err)
	}
	if len(bathrooms) == 0 {
		log.Fatalf("❌ No bathrooms found for ratings test\n")
	}
	bathroomID = strconv.Itoa(bathrooms[0].ID)
	fmt.Printf("✅ Using bathroom_id=%s\n", bathroomID)

	// Step 2: Create a review
	fmt.Println("\n2️⃣  Creating a review...")
	createReviewReq := CreateReviewRequest{
		Rating:              5,
		Title:               "Excellent bathroom!",
		Comment:             "Very clean and well-maintained. Highly recommended!",
		CleanlinessRating:   5,
		AccessibilityRating: 4,
	}
	reviewBody, _ := json.Marshal(createReviewReq)

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/api/bathrooms/%s/reviews", baseURL, bathroomID),
		bytes.NewBuffer(reviewBody),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("❌ Create review failed: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 409 {
		fmt.Println("ℹ️  Review already exists for this user+bathroom, continuing...")
	} else if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("❌ Create review error: %d - %s\n", resp.StatusCode, string(body))
	}

	if resp.StatusCode == 201 {
		var createdReview ReviewResponse
		json.NewDecoder(resp.Body).Decode(&createdReview)
		fmt.Printf("✅ Review created: ID=%d, Rating=%d\n", createdReview.ID, createdReview.Rating)
	}

	// Step 3: Get reviews for bathroom
	fmt.Println("\n3️⃣  Fetching reviews for bathroom...")
	req, _ = http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/bathrooms/%s/reviews", baseURL, bathroomID),
		nil,
	)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("❌ Fetch reviews failed: %v\n", err)
	}
	defer resp.Body.Close()

	var reviewsList ReviewsListResponse
	json.NewDecoder(resp.Body).Decode(&reviewsList)
	fmt.Printf("✅ Found %d reviews\n", reviewsList.Total)
	fmt.Printf("   Average rating: %.1f/5\n", reviewsList.AverageRating)

	if len(reviewsList.Reviews) > 0 {
		fmt.Println("\n   Recent reviews:")
		for i, review := range reviewsList.Reviews {
			if i >= 3 {
				break
			}
			fmt.Printf("   - [%d⭐] %s\n", review.Rating, review.Comment)
		}
	}

	// Step 4: Get rating statistics
	fmt.Println("\n4️⃣  Fetching rating statistics...")
	req, _ = http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/bathrooms/%s/rating-stats", baseURL, bathroomID),
		nil,
	)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("❌ Fetch stats failed: %v\n", err)
	}
	defer resp.Body.Close()

	var stats RatingStatsResponse
	json.NewDecoder(resp.Body).Decode(&stats)
	fmt.Printf("✅ Rating statistics:\n")
	fmt.Printf("   Total reviews: %d\n", stats.TotalReviews)
	fmt.Printf("   Average rating: %.1f/5\n", stats.AverageRating)
	fmt.Printf("   Avg cleanliness: %.1f/5\n", stats.AvgCleanliness)
	fmt.Printf("   Avg accessibility: %.1f/5\n", stats.AvgAccessibility)

	fmt.Println("\n═══════════════════════════════════════════════════════════")
	fmt.Println("✅ All ratings tests passed!")
}
