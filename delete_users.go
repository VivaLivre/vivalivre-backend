package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gabrieljose2004/vivalivre-backend/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		err = godotenv.Load("../../.env")
	}

	db := database.GetDB()
	ctx := context.Background()

	// Target email
	targetEmail := "exemplo@gmail.com"

	// Get user ID for exemplo@gmail.com
	var keepId int
	err = db.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", targetEmail).Scan(&keepId)
	if err != nil {
		log.Printf("Aviso: Utilizador %s não encontrado no banco de dados. Cuidado: isto irá apagar TODOS os utilizadores.", targetEmail)
		keepId = -1
	} else {
		fmt.Printf("Mantendo o utilizador com ID %d (%s)\n", keepId, targetEmail)
	}

	// Tentar deletar dependências
	tables := []string{"health_entries", "bathroom_reviews", "bathroom_reports", "bathroom_suggestions"}
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE user_id != $1", table)
		tag, err := db.Exec(ctx, query, keepId)
		if err != nil {
			log.Printf("Aviso: Falha ao limpar tabela %s: %v", table, err)
		} else {
			fmt.Printf("Limpou %d registros de %s\n", tag.RowsAffected(), table)
		}
	}

	// Deletar os usuários
	deleteQuery := `DELETE FROM users WHERE email != $1`
	tag, err := db.Exec(ctx, deleteQuery, targetEmail)
	if err != nil {
		log.Fatalf("Erro ao deletar usuários: %v", err)
	}
	
	fmt.Printf("Sucesso. Foram deletados %d utilizadores do banco de dados.\n", tag.RowsAffected())
}
