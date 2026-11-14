package handlers

import (
	"context"
	"crypto/rand"
	"log"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/gabrieljose2004/vivalivre-backend/internal/auth"
	"github.com/gabrieljose2004/vivalivre-backend/internal/database"
	"github.com/gin-gonic/gin"
)

const (
	resetCodeLength   = 8
	resetCodeTTL      = 15 * time.Minute
	resetCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // sem 0/O/1/I para evitar confusão
)

type resetEntry struct {
	email     string
	expiresAt time.Time
}

// passwordResetStore guarda os códigos de recuperação em memória.
// NOTA: num deploy com várias réplicas isto deve migrar para a BD/Redis.
var passwordResetStore = struct {
	sync.Mutex
	codes map[string]resetEntry
}{codes: make(map[string]resetEntry)}

// generateResetCode devolve um código aleatório criptograficamente seguro.
func generateResetCode() (string, error) {
	code := make([]byte, resetCodeLength)
	max := big.NewInt(int64(len(resetCodeAlphabet)))
	for i := range code {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		code[i] = resetCodeAlphabet[n.Int64()]
	}
	return string(code), nil
}

func storeResetCode(code, email string) {
	passwordResetStore.Lock()
	defer passwordResetStore.Unlock()
	now := time.Now()
	// limpeza oportunista de códigos expirados
	for k, v := range passwordResetStore.codes {
		if now.After(v.expiresAt) {
			delete(passwordResetStore.codes, k)
		}
	}
	passwordResetStore.codes[code] = resetEntry{email: email, expiresAt: now.Add(resetCodeTTL)}
}

// consumeResetCode devolve o email associado ao código e invalida-o.
func consumeResetCode(code string) (string, bool) {
	passwordResetStore.Lock()
	defer passwordResetStore.Unlock()
	entry, ok := passwordResetStore.codes[code]
	if !ok {
		return "", false
	}
	delete(passwordResetStore.codes, code)
	if time.Now().After(entry.expiresAt) {
		return "", false
	}
	return entry.email, true
}

// ForgotPassword gera um código de recuperação para o email indicado.
// TODO: enviar o código por email. Enquanto não existe provider, é escrito no log do servidor.
func ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email,max=255"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email inválido."})
		return
	}

	const genericMsg = "Se o email existir, receberá um código de recuperação."

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var id int
	if err := db.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", req.Email).Scan(&id); err != nil {
		// Resposta idêntica para não revelar quais emails existem.
		c.JSON(http.StatusOK, gin.H{"message": genericMsg})
		return
	}

	code, err := generateResetCode()
	if err != nil {
		log.Printf("ForgotPassword: failed to generate code: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Não foi possível gerar o código de recuperação."})
		return
	}
	storeResetCode(code, req.Email)

	log.Printf("PASSWORD RESET user_id=%d code=%s (expira em %s)", id, code, resetCodeTTL)

	c.JSON(http.StatusOK, gin.H{"message": genericMsg})
}

// ResetPassword valida o código e define a nova palavra-passe.
func ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required,max=64"`
		NewPassword string `json:"new_password" binding:"required,min=8,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Código ou palavra-passe inválidos (mínimo 8 caracteres)."})
		return
	}

	email, ok := consumeResetCode(req.Token)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Código de recuperação inválido ou expirado."})
		return
	}

	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao encriptar a nova palavra-passe."})
		return
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if _, err := db.Exec(ctx, "UPDATE users SET password_hash = $1 WHERE email = $2", newHash, email); err != nil {
		log.Printf("ResetPassword db update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar a palavra-passe."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Palavra-passe alterada com sucesso."})
}
