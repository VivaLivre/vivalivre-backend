# VivaLivre Backend - Guia de Configuração

Este é o backend do VivaLivre, desenvolvido em Go para alta performance e autonomia total do Firebase.

## Requisitos
- Go 1.21+
- PostgreSQL com extensão PostGIS

## Instalação e Execução Local

1. **Configurar Variáveis de Ambiente:**
   Crie um arquivo `.env` na raiz da pasta `vivalivre-backend/` com as seguintes variáveis:
   ```env
   DB_URL=postgres://usuario:senha@host:porta/database?sslmode=disable
   JWT_SECRET=sua_chave_secreta_aqui
   PORT=8080
   ```

2. **Preparar o Banco de Dados:**
   Execute o script `setup_database.sql` no seu banco de dados PostgreSQL para criar as tabelas e funções necessárias.

3. **Rodar o Servidor:**
   ```bash
   go mod tidy
   go run cmd/api/main.go
   ```

## Endpoints REST

### Públicos
- `GET /health`: Verifica se o servidor está online.
- `POST /auth/register`: Registro de novo usuário.
  - Body: `{"name": "...", "email": "...", "password": "..."}`
- `POST /auth/login`: Autenticação e geração de token.
  - Body: `{"email": "...", "password": "..."}`
  - Retorno: `{"token": "JWT_TOKEN", "user": {...}}`

### Protegidos (Requer header `Authorization: Bearer <TOKEN>`)
- `GET /api/bathrooms/nearby`: Busca banheiros próximos via PostGIS.
- `GET /api/health/entries`: Recupera entradas de saúde do usuário logado.

## Estrutura do Projeto
- `cmd/api/`: Ponto de entrada do servidor.
- `internal/auth/`: Lógica de JWT, Bcrypt e Middleware.
- `internal/database/`: Singleton de conexão com PostgreSQL.
- `internal/handlers/`: Controladores das rotas.
- `internal/models/`: Estruturas de dados (User, Bathroom, HealthEntry).
