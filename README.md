# VivaLivre Backend

O backend oficial do projeto **VivaLivre**, desenvolvido em Go para fornecer uma infraestrutura robusta, performática e escalável para suporte a pacientes com DII (Doenças Inflamatórias Intestinais). 🚀

Inclui geolocalização avançada com PostGIS, autenticação proprietária com JWT e arquitetura escalável desacoplada do Firebase.

## 🚀 Tecnologias
- **Linguagem:** Go (Golang) 1.21+
- **Framework Web:** Gin Gonic
- **Banco de Dados:** PostgreSQL + PostGIS (Geolocalização)
- **Autenticação:** JWT (JSON Web Tokens) & Bcrypt

## ⚙️ Configuração Local

1. **Clonar o repositório:**
   ```bash
   git clone https://github.com/VivaLivre/vivalivre-backend.git
   cd vivalivre-backend
   ```

2. **Configurar variáveis de ambiente:**
   Crie um arquivo `.env` baseado no `.env.example`:
   ```env
   DB_URL=postgres://usuario:senha@host:porta/database?sslmode=disable
   JWT_SECRET=sua_chave_secreta_aqui
   PORT=8080
   ```

3. **Banco de Dados:**
   Execute o script `setup_database.sql` no seu PostgreSQL para inicializar as tabelas e extensões do PostGIS.

4. **Executar:**
   ```bash
   go mod tidy
   go run cmd/api/main.go
   ```

## 🛠 Endpoints Principais

### Autenticação
- `POST /auth/register` - Criar nova conta
- `POST /auth/login` - Obter token de acesso

### Funcionalidades (Protegidas)
- `GET /api/users/me` - Perfil do usuário logado
- `GET /api/bathrooms/nearby` - Localizar banheiros adaptados próximos
- `GET /api/health/entries` - Histórico de saúde e sintomas

## 📂 Estrutura
- `/cmd/api`: Ponto de entrada da aplicação.
- `/internal/handlers`: Lógica de processamento das requisições.
- `/internal/database`: Gerenciamento da conexão com o banco.
- `/internal/auth`: Middleware e segurança.

## 📄 Licença
Este projeto está sob a licença MIT. Veja o arquivo [LICENSE](LICENSE) para detalhes.
