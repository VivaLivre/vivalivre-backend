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

	// 2. Deletar todos os registros de health_entries do usuário
	queryDelete := `DELETE FROM health_entries WHERE user_id = $1`
	tag, err := db.Exec(ctx, queryDelete, userID)
	if err != nil {
		log.Fatalf("Erro ao limpar registros de sintomas: %v", err)
	}

	fmt.Printf("Sucesso! Foram eliminados %d registros de sintomas da tabela health_entries para este utilizador.\n", tag.RowsAffected())
}
