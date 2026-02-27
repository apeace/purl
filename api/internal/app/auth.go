package app

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type googleTokenInfo struct {
	Aud           string `json:"aud"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Sub           string `json:"sub"`
}

type googleAuthRequest struct {
	IDToken string `json:"id_token"`
}

type authUserResponse struct {
	ID    string       `json:"id"`
	Name  string       `json:"name"`
	Email string       `json:"email"`
	Org   orgResponse  `json:"org"`
}

type authResponse struct {
	Token string           `json:"token"`
	User  authUserResponse `json:"user"`
}

func (a *App) googleAuth(w http.ResponseWriter, r *http.Request) {
	var req googleAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IDToken == "" {
		http.Error(w, "missing id_token", http.StatusBadRequest)
		return
	}

	// Verify token with Google
	resp, err := http.Get(fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", req.IDToken))
	if err != nil {
		http.Error(w, "failed to verify token", http.StatusInternalServerError)
		log.Printf("googleAuth verify: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		log.Printf("googleAuth token rejected: %s", body)
		return
	}

	var tokenInfo googleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		http.Error(w, "failed to parse token info", http.StatusInternalServerError)
		log.Printf("googleAuth decode: %v", err)
		return
	}

	if tokenInfo.Aud != a.googleClientID {
		http.Error(w, "invalid token audience", http.StatusUnauthorized)
		return
	}
	if tokenInfo.EmailVerified != "true" {
		http.Error(w, "email not verified", http.StatusUnauthorized)
		return
	}

	// Look up user by email
	var userID, userName, userEmail, orgID, orgName string
	var googleID sql.NullString
	err = a.db.QueryRowContext(r.Context(), `
		SELECT u.id, u.name, u.email, u.google_id, u.org_id, o.name
		FROM users u
		JOIN organizations o ON o.id = u.org_id
		WHERE u.email = $1
	`, tokenInfo.Email).Scan(&userID, &userName, &userEmail, &googleID, &orgID, &orgName)
	if err == sql.ErrNoRows {
		http.Error(w, "no account found for this email", http.StatusForbidden)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("googleAuth user lookup: %v", err)
		return
	}

	// Link Google ID on first sign-in, or verify it matches
	if !googleID.Valid {
		_, err = a.db.ExecContext(r.Context(),
			`UPDATE users SET google_id = $1 WHERE id = $2`, tokenInfo.Sub, userID)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Printf("googleAuth link google_id: %v", err)
			return
		}
	} else if googleID.String != tokenInfo.Sub {
		http.Error(w, "google account mismatch", http.StatusForbidden)
		return
	}

	// Create session
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("googleAuth rand: %v", err)
		return
	}
	sessionToken := hex.EncodeToString(tokenBytes)

	err = a.redis.Set(r.Context(), "session:"+sessionToken, userID, 7*24*time.Hour).Err()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("googleAuth redis set: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authResponse{
		Token: sessionToken,
		User: authUserResponse{
			ID:    userID,
			Name:  userName,
			Email: userEmail,
			Org:   orgResponse{ID: orgID, Name: orgName},
		},
	})
}

func (a *App) authMe(w http.ResponseWriter, r *http.Request) {
	u := userFromContext(r.Context())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authUserResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		Org:   orgResponse{ID: u.OrgID, Name: u.OrgName},
	})
}

func (a *App) authLogout(w http.ResponseWriter, r *http.Request) {
	token := sessionTokenFromContext(r.Context())
	if token != "" {
		a.redis.Del(r.Context(), "session:"+token)
	}
	w.WriteHeader(http.StatusNoContent)
}
