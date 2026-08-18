package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func main() {
	url := "https://vivalivre-backend-production.up.railway.app/auth/register"
	payload := []byte(`{"name":"Test","email":"test998@test.com","password":"Password123!","cpf":"12345678902","date_of_birth":"1990-01-01","gender":"Masculino","clinical_condition":"Crohn","comorbidities":[],"weight":70,"height":1.7}`)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\nResponse: %s\n", resp.StatusCode, string(body))
}
