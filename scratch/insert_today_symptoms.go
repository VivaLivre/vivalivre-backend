package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gabrieljose2004/vivalivre-backend/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	// Carregar variáveis de ambiente do .env na raiz do projeto
	err := godotenv.Load(".env")
	if err != nil {
		err = godotenv.Load("../.env")
		if err != nil {
			log.Fatalf("Erro ao carregar o arquivo .env: %v", err)
		}
	}

	db := database.GetDB()
	ctx := context.Background()

	// 1. Obter o utilizador específico pelo email
	targetEmail := "josegabriel13112004@gmail.com"
	var userID int
	var email string
	var name string
	queryUser := "SELECT id, name, email FROM users WHERE email = $1"
	err = db.QueryRow(ctx, queryUser, targetEmail).Scan(&userID, &name, &email)
	if err != nil {
		log.Fatalf("Erro ao buscar o utilizador com email %s: %v", targetEmail, err)
	}

	fmt.Printf("Utilizador alvo:\nID: %d | Nome: %s | Email: %s\n\n", userID, name, email)

	// 2. Inserir sintomas com a data de hoje (NOW())
	symptomsList := []struct {
		entryType   string
		severity    string
		description string
		symptoms    []string
	}{
		{
			entryType:   "symptom",
			severity:    "moderate",
			description: "Sensibilidade à luz e indisposição (Hoje)",
			symptoms:    []string{"photophobia", "fatigue"},
		},
		{
			entryType:   "symptom",
			severity:    "mild",
			description: "Pequena tontura pela manhã (Hoje)",
			symptoms:    []string{"dizziness"},
		},
	}

	queryInsert := `
		INSERT INTO health_entries (user_id, type, severity, description, symptoms, entry_date)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, entry_date
	`

	fmt.Println("A inserir registos clínicos de hoje no banco de dados...")
	for _, s := range symptomsList {
		var entryID int
		var entryDate interface{}
		err = db.QueryRow(ctx, queryInsert, userID, s.entryType, s.severity, s.description, s.symptoms).Scan(&entryID, &entryDate)
		if err != nil {
			log.Fatalf("Erro ao inserir sintoma: %v", err)
		}
		fmt.Printf("✅ Registro inserido - ID: %d | Descrição: %s | Data Gravada: %v\n", entryID, s.description, entryDate)
	}

	fmt.Println("\nSucesso! Os sintomas de hoje foram cadastrados para o teste de filtro.")
}
