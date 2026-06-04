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

	// The user requested to delete ALL bathrooms
	deleteQuery := `DELETE FROM bathrooms`
	tag, err := db.Exec(ctx, deleteQuery)
	if err != nil {
		log.Fatalf("Erro ao deletar todos os banheiros: %v", err)
	}
	
	fmt.Printf("Sucesso. Foram deletados %d registros de banheiros do banco de dados.\n", tag.RowsAffected())
}
