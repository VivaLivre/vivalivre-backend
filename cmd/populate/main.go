package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gabrieljose2004/vivalivre-backend/internal/database"
)

type RawLocation struct {
	Name             string
	Address          string
	Lat              float64
	Lng              float64
	OpenTime         string
	CloseTime        string
	IsAccessible     bool
	HasChangingTable bool
}

func main() {
	os.Setenv("DB_URL", "postgresql://postgres.leteadhfnmefdmpivcll:ynwhwFQbLLZY0dwJ@aws-1-sa-east-1.pooler.supabase.com:6543/postgres?pgbouncer=true&default_query_exec_mode=exec")
	db := database.GetDB()
	ctx := context.Background()

	rawLocations := []RawLocation{
		// Santo André
		{"Parque Celso Daniel", "Av. Dom Pedro II, 940 - Bairro Jardim, Santo André - SP", -23.6475245, -46.5359018, "06:00", "22:00", true, true},
		{"Parque Central (Dep. José Cicote)", "Rua José Bonifácio, s/n - Vila Assunção, Santo André - SP", -23.6456329, -46.5250579, "06:00", "20:00", true, true},
		{"Parque Antônio Fláquer (Ipiranguinha)", "Rua Coronel Seabra, 210 - Vila Alzira, Santo André - SP", -23.6684178, -46.5214352, "06:00", "22:00", false, false},
		{"Parque Escola", "Rua Anacleto Popote, 46 - Vila Valparaíso, Santo André - SP", -23.6646653, -46.5509253, "06:00", "20:00", true, false},
		{"Parque Regional da Criança", "Av. Itamarati, 536 - Parque Jaçatuba, Santo André - SP", -23.6492141, -46.5193094, "06:00", "20:00", true, true},
		{"Parque da Juventude Ana Brandão", "Av. Capitão Mario de Toledo, s/n - Vila Humaitá, Santo André - SP", -23.6826315, -46.5045461, "06:00", "18:00", false, false},

		// SBC
		{"Esplanada do Paço Municipal", "Praça Samuel Sabatini - Centro, São Bernardo do Campo - SP", -23.69663, -46.5517015, "08:00", "21:00", true, true},
		{"Praça Giovanni Breda", "Área Verde do Assunção, São Bernardo do Campo - SP", -23.724167, -46.574444, "07:00", "21:00", true, true},
		{"Parque da Juventude (Città di Maróstica)", "Av. Armando Ítalo Setti, 65 - Centro, São Bernardo do Campo - SP", -23.6997747, -46.5425972, "06:00", "22:00", true, true},
		{"Parque Eng. Salvador Arena", "Av. Caminho do Mar, 2980 - Rudge Ramos, São Bernardo do Campo - SP", -23.6604118, -46.5701324, "06:00", "22:00", true, true},
		{"Parque Raphael Lazzuri", "Av. Kennedy, 1111 - Anchieta, São Bernardo do Campo - SP", -23.6799706, -46.5586099, "06:00", "22:00", true, true},
		{"Parque Chácara Silvestre", "Av. Wallace Simonsen, 1800 - Nova Petrópolis, São Bernardo do Campo - SP", -23.7108345, -46.5331599, "06:00", "20:00", true, false},

		// SCS
		{"Espaço Verde Chico Mendes", "Av. Fernando Simonsen, 566 - Bairro Cerâmica, São Caetano do Sul - SP", -23.6325253, -46.5730651, "06:00", "22:00", true, true},
		{"Parque Linear Kennedy", "Av. Presidente Kennedy, 2300 - Bairro Olímpico, São Caetano do Sul - SP", -23.6326125, -46.5576656, "06:00", "22:00", true, false},
		{"Parque Tom Jobim", "Blvd. São Caetano - Bairro Cerâmica, São Caetano do Sul - SP", -23.6227531, -46.579668, "06:00", "22:00", true, false},
		{"Parque Bosque do Povo", "Estrada das Lágrimas, 320 - Bairro São José, São Caetano do Sul - SP", -23.63435, -46.580082, "06:00", "18:00", false, false},
		{"Cidade das Crianças", "Alameda Conde de Porto Alegre, 840 - Bairro Santa Maria, São Caetano do Sul - SP", -23.6324584, -46.5559946, "06:00", "18:00", true, true},

		// Diadema
		{"Parque do Paço", "Avenida Antônio Piranga, 1380 - Centro, Diadema - SP", -23.6861754, -46.6101267, "06:00", "22:00", true, true},
		{"Parque Pousada dos Jesuítas", "Rua Prof. Vitalina Caiaffa Esquivel, s/n - Centro, Diadema - SP", -23.6878897, -46.6188302, "07:00", "18:00", true, false},
		{"Parque Ecológico Fernando Vítor", "Av. Nossa Senhora dos Navegantes, 145 - Eldorado, Diadema - SP", -23.7167943, -46.6250696, "05:00", "22:00", false, false},

		// Mauá
		{"Parque Natural Municipal Guapituba", "Av. Capitão João, 3220 - Jardim Guapituba, Mauá - SP", -23.6706318, -46.4597627, "07:00", "17:00", true, false},
		{"Parque Ecológico da Gruta Santa Luzia", "Rua Luzia da Silva Itabaiana, 101 - Jardim Itapeva, Mauá - SP", -23.6762105, -46.409765, "07:00", "17:00", true, false},
		{"Parque da Juventude (Francisco de Carvalho Filho)", "Paço Municipal, Centro, Mauá - SP", -23.6679655, -46.4676535, "06:00", "22:00", true, true},
		{"Estação Mauá (CPTM)", "Praça Vinte e Dois de Novembro, Mauá - SP", -23.6682599, -46.4615541, "04:00", "23:59", true, true},

		// Ribeirão Pires
		{"Parque Municipal Prof. Luiz Carlos Grecco", "Rua Diamantino de Oliveira, 220 - Jardim Pastoril, Ribeirão Pires - SP", -23.7113089, -46.4082089, "08:00", "17:00", true, false},
		{"Parque Oriental (Milton Marinho de Moraes)", "Rua Major Cardim, 3100 - Estância Noblesse, Ribeirão Pires - SP", -23.7209369, -46.4350823, "08:00", "17:00", true, true},
		{"Terminal Rodoviário de Ribeirão Pires", "Rua Capitão José Gallo, Centro, Ribeirão Pires - SP", -23.7122709, -46.4158247, "05:00", "23:00", true, true},

		// Rio Grande da Serra
		{"Estação Rio Grande da Serra (CPTM)", "Praça da Matriz - Centro, Rio Grande da Serra - SP", -23.7434354, -46.391837, "04:00", "23:59", true, true},
	}

	query := `
		INSERT INTO bathrooms (
			name, address, location, is_accessible, has_changing_table, is_free, 
			operating_hours, status, created_at
		) VALUES (
			$1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326), $5, $6, $7, 
			$8::jsonb, $9, NOW()
		)
	`

	deleteQuery := `DELETE FROM bathrooms WHERE name = $1`

	count := 0
	for _, loc := range rawLocations {
		db.Exec(ctx, deleteQuery, loc.Name)

		opHours := fmt.Sprintf(`{"type":"custom", "schedule":{"1":{"open":"%s","close":"%s"}, "2":{"open":"%s","close":"%s"}, "3":{"open":"%s","close":"%s"}, "4":{"open":"%s","close":"%s"}, "5":{"open":"%s","close":"%s"}, "6":{"open":"%s","close":"%s"}, "7":{"open":"%s","close":"%s"}}}`, loc.OpenTime, loc.CloseTime, loc.OpenTime, loc.CloseTime, loc.OpenTime, loc.CloseTime, loc.OpenTime, loc.CloseTime, loc.OpenTime, loc.CloseTime, loc.OpenTime, loc.CloseTime, loc.OpenTime, loc.CloseTime)

		_, err := db.Exec(ctx, query,
			loc.Name, loc.Address, loc.Lng, loc.Lat, loc.IsAccessible, loc.HasChangingTable, true, opHours, "approved",
		)
		if err != nil {
			log.Printf("Failed to insert %s: %v\n", loc.Name, err)
		} else {
			count++
			fmt.Printf("Inserted %s\n", loc.Name)
		}
	}

	fmt.Printf("\nDone! Successfully inserted %d bathrooms.\n", count)
}
