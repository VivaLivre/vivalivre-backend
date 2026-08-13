# 🤝 Guia de Contribuição — VivaLivre Backend

Este documento define o fluxo de trabalho Git obrigatório para todos que desenvolvem na API do VivaLivre. **Nenhum código vai direto para a `main`.**

---

## 🌿 Estrutura de Branches

```
main          ← Produção. Somente Pull Requests vindos de develop.
│
└── develop   ← Área de testes/integração. Branch padrão de trabalho.
      │
      ├── feature/login-jwt
      ├── feature/postgis-radius
      ├── fix/auth-middleware
      └── chore/atualizar-go
```

| Branch | Finalidade | Quem pode publicar |
|---|---|---|
| `main` | Versão estável de produção | Somente via PR aprovado de `develop` |
| `develop` | Integração e testes | Somente via PR aprovado de `feature/*` ou `fix/*` |
| `feature/*` | Nova funcionalidade | Desenvolvedor, a partir de `develop` |
| `fix/*` | Correção de bug | Desenvolvedor, a partir de `develop` |
| `chore/*` | Tarefas técnicas (deps, config) | Desenvolvedor, a partir de `develop` |
| `hotfix/*` | Correção urgente em produção | A partir de `main`, merge em `main` E `develop` |

---

## 🚀 Fluxo de Trabalho Diário

### 1. Criar uma branch para a sua tarefa

Sempre a partir de `develop` (nunca de `main`):

```bash
git checkout develop
git pull origin develop          # garante que está atualizado

# Escolha o prefixo correto:
git checkout -b feature/nome-da-feature
git checkout -b fix/nome-do-bug
git checkout -b chore/nome-da-tarefa
```

### 2. Desenvolver e commitar

Use mensagens de commit no padrão **Conventional Commits**:

```bash
git add .
git commit -m "feat: endpoint de raio de proximidade com PostGIS"
git commit -m "fix: validação de token expirado no middleware"
git commit -m "chore: atualiza gin-gonic para v1.9.1"
git commit -m "docs: adiciona guia de contribuição"
git commit -m "test: adiciona testes de integração no login"
```

| Prefixo | Quando usar |
|---|---|
| `feat:` | Nova funcionalidade |
| `fix:` | Correção de bug |
| `chore:` | Atualização de dependências, configurações |
| `docs:` | Mudanças na documentação |
| `style:` | Formatação, sem mudança de lógica |
| `refactor:` | Refatoração de código |
| `test:` | Adição ou correção de testes |
| `ci:` | Configuração de CI/CD |

### 3. Publicar a branch e abrir PR para `develop`

```bash
git push origin feature/nome-da-feature
```

Em seguida, acesse o GitHub e abra um **Pull Request**:
- **De:** `feature/nome-da-feature`
- **Para:** `develop`
- Descreva o que foi feito no PR

### 4. Revisão e merge em `develop`

Após revisão, faça o merge em `develop`. Isso disponibiliza o código para testes de integração.

### 5. Promover `develop` → `main` (Release)

Quando `develop` estiver estável e testado:

```bash
git checkout main
git pull origin main
git merge develop
git push origin main
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin --tags
```

---

## 🚑 Hotfix (Correção Urgente em Produção)

Quando há um bug crítico em produção que não pode esperar o ciclo normal:

```bash
git checkout main
git pull origin main
git checkout -b hotfix/descricao-do-bug

# ... corrigir o bug ...

git commit -m "fix: corrige panic no handler de login"

# Merge em MAIN
git checkout main
git merge hotfix/descricao-do-bug
git push origin main
git tag -a v1.0.1 -m "Hotfix v1.0.1"
git push origin --tags

# Merge também em DEVELOP para não perder a correção
git checkout develop
git merge hotfix/descricao-do-bug
git push origin develop

# Deletar branch de hotfix
git branch -d hotfix/descricao-do-bug
git push origin --delete hotfix/descricao-do-bug
```

---

## 📋 Regras Obrigatórias

> [!CAUTION]
> **Proibido** fazer `git push` diretamente para `main`. Sempre use Pull Request.

> [!IMPORTANT]
> Toda branch deve sair de `develop`, não de `main`.

> [!WARNING]
> Antes de criar uma branch, sempre faça `git pull origin develop` para evitar conflitos.

> [!TIP]
> Delete branches locais após o merge: `git branch -d feature/nome-da-feature`

---

## 🏷️ Versionamento (SemVer)

O projeto segue **Semantic Versioning**: `MAJOR.MINOR.PATCH`

---

## 💻 Setup Inicial para Novos Devs

```bash
# 1. Clonar
git clone https://github.com/VivaLivre/vivalivre-backend.git
cd vivalivre-backend

# 2. Configurar Variáveis de Ambiente
cp .env.example .env
# Edite o .env com as suas credenciais locais do PostgreSQL (PostGIS)

# 3. Baixar dependências Go
go mod download

# 4. Confirmar que está na branch develop
git checkout develop

# 5. Rodar o servidor
go run cmd/api/main.go
```
