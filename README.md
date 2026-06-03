# 💙 VivaLivre — Backend API

> API REST robusta, performática e escalável para suporte a pacientes com DII.

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![Gin](https://img.shields.io/badge/Gin-Gonic-00ADD8?style=flat-square&logo=go&logoColor=white)](https://gin-gonic.com)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-PostGIS-336791?style=flat-square&logo=postgresql&logoColor=white)](https://www.postgresql.org)
[![JWT](https://img.shields.io/badge/JWT-Auth-blueviolet?style=flat-square)](https://jwt.io)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow?style=flat-square)](LICENSE)

---

## 📖 Sobre o Backend

**VivaLivre Backend** é uma API REST desenvolvida em **Go (Golang)** com **Gin Gonic**, fornecendo uma infraestrutura robusta, performática e escalável para suporte a pacientes com **Doenças Inflamatórias Intestinais (DII)**.

O backend implementa:
- **Autenticação segura** com JWT e Bcrypt
- **Geolocalização avançada** com PostGIS para queries geoespaciais
- **Arquitetura escalável** desacoplada do Firebase
- **API REST RESTful** com validação de entrada
- **Persistência de dados** em PostgreSQL
- **Middleware de segurança** (CORS, rate limiting, validação)

---

## 🎯 Responsabilidades Principais

- Autenticação e autorização de utilizadores
- Gestão de registos de saúde (CRUD)
- Localização de banheiros adaptados via PostGIS
- Sistema de avaliações de banheiros (reviews + estatísticas + votos úteis)
- Validação e persistência de dados
- Segurança e proteção de dados sensíveis
- Performance e escalabilidade

### Endpoints de Ratings (protegidos)
- `POST /api/bathrooms/:bathroom_id/reviews`
- `GET /api/bathrooms/:bathroom_id/reviews`
- `GET /api/bathrooms/:bathroom_id/rating-stats`
- `GET /api/reviews/:review_id`
- `PUT /api/reviews/:review_id`
- `DELETE /api/reviews/:review_id`
- `POST /api/reviews/:review_id/helpful`

Regra de produto: foto pertence ao banheiro, não à avaliação.

---

## 🛠️ Stack Tecnológico

| Camada | Tecnologia | Propósito |
|---|---|---|
| **Linguagem** | Go 1.21+ | Linguagem compilada, rápida e eficiente |
| **Framework Web** | Gin Gonic | Roteamento HTTP e middleware |
| **Banco de Dados** | PostgreSQL 14+ | Dados relacionais e ACID |
| **Geolocalização** | PostGIS | Queries geoespaciais avançadas |
| **Autenticação** | JWT + Bcrypt | Tokens seguros e hashing de senhas |
| **Validação** | Validator | Validação de structs e inputs |
| **Logging** | Logrus | Logs estruturados |
| **Testes** | Testing (stdlib) | Testes unitários e integração |
| **Containerização** | Docker | Deployment consistente |
| **CI/CD** | GitHub Actions | Automação de testes e deploy |

### Dependências Principais

```go
// Framework Web
github.com/gin-gonic/gin v1.9.1

// Banco de Dados
github.com/lib/pq v1.10.9              // Driver PostgreSQL
github.com/jmoiron/sqlx v1.3.5         // SQL utilities

// Autenticação
github.com/golang-jwt/jwt/v5 v5.0.0    // JWT
golang.org/x/crypto v0.14.0            // Bcrypt

// Validação
github.com/go-playground/validator/v10 v10.15.0

// Logging
github.com/sirupsen/logrus v1.9.3

// Utilitários
github.com/joho/godotenv v1.5.1        // .env files
github.com/google/uuid v1.3.1          // UUIDs
```

---

## ⚙️ Configuração Local

### Pré-requisitos

- [Go 1.21+](https://go.dev/dl) — Verificar com `go version`
- [PostgreSQL 14+](https://www.postgresql.org/download/) — Banco de dados
- [PostGIS](https://postgis.net/install/) — Extensão geoespacial
- [Git](https://git-scm.com/) — Controle de versão
- [Docker](https://www.docker.com/) (opcional) — Para containerização

### Passo a Passo

**1. Clone o repositório**
```bash
git clone https://github.com/VivaLivre/vivalivre-backend.git
cd vivalivre-backend
```

**2. Configure variáveis de ambiente**

Crie um arquivo `.env` baseado em `.env.example`:

```bash
cp .env.example .env
```

Edite `.env` com suas configurações:

```env
# Banco de Dados
DB_HOST=localhost
DB_PORT=5432
DB_USER=vivalivre
DB_PASSWORD=sua_senha_segura
DB_NAME=vivalivre_db
DB_SSLMODE=disable

# Autenticação
JWT_SECRET=sua_chave_secreta_muito_longa_e_segura_aqui
JWT_EXPIRATION=24h

# Servidor
PORT=8080
ENV=development

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8081

# Logging
LOG_LEVEL=debug
```

**3. Configure o Banco de Dados**

```bash
# Crie o banco de dados
createdb vivalivre_db

# Execute as migrations
psql -U vivalivre -d vivalivre_db -f scripts/setup_database.sql

# Verifique se PostGIS foi instalado
psql -U vivalivre -d vivalivre_db -c "CREATE EXTENSION IF NOT EXISTS postgis;"
```

**4. Instale as dependências Go**

```bash
go mod download
go mod tidy
```

**5. Execute o servidor**

```bash
# Modo desenvolvimento
go run cmd/api/main.go

# Modo release (otimizado)
go build -o vivalivre-api cmd/api/main.go
./vivalivre-api
```

O servidor estará disponível em `http://localhost:8080`

**6. Verifique a saúde da API**

```bash
curl http://localhost:8080/health
```

Resposta esperada:
```json
{
  "status": "ok",
  "timestamp": "2026-05-12T21:24:06Z"
}
```

### Troubleshooting

| Problema | Solução |
|---|---|
| `go: command not found` | Instale Go 1.21+ e adicione ao PATH |
| Erro de conexão PostgreSQL | Verifique credenciais em `.env` e se PostgreSQL está rodando |
| PostGIS não encontrado | Execute `CREATE EXTENSION postgis;` no banco |
| Porta 8080 já em uso | Altere `PORT` em `.env` ou libere a porta |
| Erro de JWT_SECRET | Gere uma chave segura: `openssl rand -base64 32` |

---

## 🏗️ Arquitetura

O backend segue princípios de **Clean Architecture** com separação clara entre camadas:

```
cmd/
├── api/
│   └── main.go                  # Ponto de entrada

internal/
├── handlers/                    # HTTP handlers (Controllers)
│   ├── auth.go                  # Autenticação
│   ├── health.go                # Registos de saúde
│   ├── bathrooms.go             # Banheiros
│   └── users.go                 # Perfil de utilizador
│
├── services/                    # Lógica de negócio
│   ├── auth_service.go
│   ├── health_service.go
│   └── bathroom_service.go
│
├── repositories/                # Acesso a dados
│   ├── user_repository.go
│   ├── health_repository.go
│   └── bathroom_repository.go
│
├── models/                      # Structs de dados
│   ├── user.go
│   ├── health_entry.go
│   └── bathroom.go
│
├── middleware/                  # Middleware HTTP
│   ├── auth.go                  # JWT validation
│   ├── cors.go                  # CORS
│   └── logging.go               # Request logging
│
├── database/                    # Conexão e migrations
│   ├── connection.go
│   └── migrations.go
│
└── config/                      # Configuração
    └── config.go
```

### Fluxo de Requisição

```
HTTP Request
    ↓
Middleware (CORS, Logging, Auth)
    ↓
Router (Gin)
    ↓
Handler (Validação de input)
    ↓
Service (Lógica de negócio)
    ↓
Repository (Acesso a dados)
    ↓
Database (PostgreSQL + PostGIS)
    ↓
Response JSON
```

---

## 📚 API Endpoints

### 🔐 Autenticação (Públicos)

#### Registar novo utilizador
```http
POST /auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "senha_segura_123",
  "name": "João Silva"
}
```

Resposta (201 Created):
```json
{
  "id": "uuid-123",
  "email": "user@example.com",
  "name": "João Silva",
  "created_at": "2026-05-12T21:24:06Z"
}
```

#### Login
```http
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "senha_segura_123"
}
```

Resposta (200 OK):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 86400,
  "user": {
    "id": "uuid-123",
    "email": "user@example.com",
    "name": "João Silva"
  }
}
```

### 👤 Utilizador (Protegidos)

#### Obter perfil do utilizador logado
```http
GET /api/users/me
Authorization: Bearer {token}
```

Resposta (200 OK):
```json
{
  "id": "uuid-123",
  "email": "user@example.com",
  "name": "João Silva",
  "created_at": "2026-05-12T21:24:06Z"
}
```

### 🩺 Registos de Saúde (Protegidos)

#### Listar registos de saúde
```http
GET /api/health/entries
Authorization: Bearer {token}
```

Resposta (200 OK):
```json
[
  {
    "id": "uuid-456",
    "user_id": "uuid-123",
    "type": "symptom",
    "description": "Dor abdominal",
    "severity": 7,
    "symptoms": ["dor", "inchaço"],
    "created_at": "2026-05-12T21:24:06Z"
  }
]
```

#### Criar novo registo de saúde
```http
POST /api/health/entries
Authorization: Bearer {token}
Content-Type: application/json

{
  "type": "symptom",
  "description": "Dor abdominal",
  "severity": 7,
  "symptoms": ["dor", "inchaço"]
}
```

Resposta (201 Created):
```json
{
  "id": "uuid-456",
  "user_id": "uuid-123",
  "type": "symptom",
  "description": "Dor abdominal",
  "severity": 7,
  "symptoms": ["dor", "inchaço"],
  "created_at": "2026-05-12T21:24:06Z"
}
```

#### Eliminar registo de saúde
```http
DELETE /api/health/entries/{id}
Authorization: Bearer {token}
```

Resposta (204 No Content)

### 🗺️ Banheiros (Protegidos)

#### Localizar banheiros próximos (PostGIS)
```http
GET /api/bathrooms/nearby?lat=-23.5505&lng=-46.6333&radius=1000
Authorization: Bearer {token}
```

Parâmetros:
- `lat` — Latitude do utilizador
- `lng` — Longitude do utilizador
- `radius` — Raio de busca em metros (padrão: 1000)

Resposta (200 OK):
```json
[
  {
    "id": "uuid-789",
    "name": "Banheiro Adaptado Centro",
    "latitude": -23.5505,
    "longitude": -46.6333,
    "distance_meters": 250,
    "rating": 4.5,
    "accessible": true,
    "spacious": true,
    "clean": true,
    "created_at": "2026-05-12T21:24:06Z"
  }
]
```

#### Criar novo banheiro (Contribuição)
```http
POST /api/bathrooms
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "Banheiro Adaptado Centro",
  "latitude": -23.5505,
  "longitude": -46.6333,
  "accessible": true,
  "spacious": true,
  "clean": true
}
```

Resposta (201 Created):
```json
{
  "id": "uuid-789",
  "name": "Banheiro Adaptado Centro",
  "latitude": -23.5505,
  "longitude": -46.6333,
  "distance_meters": 0,
  "rating": 0,
  "accessible": true,
  "spacious": true,
  "clean": true,
  "created_at": "2026-05-12T21:24:06Z"
}
```

### ❌ Códigos de Erro

| Código | Descrição |
|---|---|
| `400 Bad Request` | Validação de input falhou |
| `401 Unauthorized` | Token JWT inválido ou expirado |
| `403 Forbidden` | Sem permissão para acessar recurso |
| `404 Not Found` | Recurso não encontrado |
| `409 Conflict` | Email já registado |
| `500 Internal Server Error` | Erro no servidor |

---

## 🤝 Contribuindo

Contribuições são muito bem-vindas! Siga os passos:

1. **Faça um fork** do projeto
2. **Crie uma branch** a partir de `develop`:
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feat/sua-feature
   ```

3. **Faça suas alterações** seguindo os padrões:
   - Código limpo e bem documentado
   - Testes unitários para novas funcionalidades
   - Sem hardcoded secrets
   - Validação de input robusta

4. **Commit com mensagens semânticas**:
   ```bash
   git commit -m "feat(health): add symptom tracking endpoint"
   git commit -m "fix(auth): handle token expiration correctly"
   ```

5. **Verifique a qualidade**:
   ```bash
   go fmt ./...
   go vet ./...
   go test ./...
   ```

6. **Envie para seu fork**:
   ```bash
   git push origin feat/sua-feature
   ```

7. **Abra um Pull Request** para `develop`

### Padrões de Código

#### Handlers
```go
// ✅ CORRETO - Handler bem estruturado
func (h *HealthHandler) CreateEntry(c *gin.Context) {
    var req CreateHealthEntryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    entry, err := h.service.CreateEntry(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create entry"})
        return
    }
    
    c.JSON(http.StatusCreated, entry)
}

// ❌ ERRADO - Sem validação
func (h *HealthHandler) CreateEntry(c *gin.Context) {
    var req CreateHealthEntryRequest
    c.BindJSON(&req)  // Sem verificar erro
    entry := h.service.CreateEntry(req)
    c.JSON(200, entry)
}
```

#### Repositories
```go
// ✅ CORRETO - Query com prepared statement
func (r *HealthRepository) GetEntries(ctx context.Context, userID string) ([]HealthEntry, error) {
    query := `SELECT id, user_id, type, description, severity FROM health_entries WHERE user_id = $1`
    rows, err := r.db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var entries []HealthEntry
    for rows.Next() {
        var entry HealthEntry
        if err := rows.Scan(&entry.ID, &entry.UserID, &entry.Type, &entry.Description, &entry.Severity); err != nil {
            return nil, err
        }
        entries = append(entries, entry)
    }
    return entries, nil
}

// ❌ ERRADO - SQL injection vulnerability
func (r *HealthRepository) GetEntries(userID string) []HealthEntry {
    query := fmt.Sprintf("SELECT * FROM health_entries WHERE user_id = '%s'", userID)
    rows := r.db.Query(query)
    // ...
}
```

---

## 📚 Documentação Adicional

- **[AGENTS.md](../AGENTS.md)** — Definição de papéis e responsabilidades
- **[docs/](./docs/)** — Documentação técnica e ADRs
- **[Frontend VivaLivre](https://github.com/VivaLivre/vivalivre-app)** — Aplicativo Mobile Flutter
- **[Admin Portal VivaLivre](https://github.com/VivaLivre/vivalivre-admin)** — Painel Administrativo Web em Flutter

---

## 📞 Suporte

- **Issues**: [GitHub Issues](https://github.com/VivaLivre/vivalivre-backend/issues)
- **Discussões**: [GitHub Discussions](https://github.com/VivaLivre/vivalivre-backend/discussions)

---

## 📄 Licença

Este projeto está sob a licença **MIT**. Veja o arquivo [LICENSE](LICENSE) para detalhes.

---

<div align="center">

Feito com 💙 por **Gabriel José de Souza** para a comunidade DII brasileira.

*"Toda pessoa com DII merece viver com liberdade e dignidade."*

[⬆ Voltar ao topo](#-vivalivre--backend-api)

</div>
