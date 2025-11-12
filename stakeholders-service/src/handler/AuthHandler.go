package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"stakeholders-service/auth"
	"stakeholders-service/model"
	"stakeholders-service/repository"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

type AuthHandler struct {
	repo *repository.UserRepository
}

func RegisterAuthRoutes(r *mux.Router, repo *repository.UserRepository) {
	h := &AuthHandler{repo: repo}
	r.HandleFunc("/auth/register", h.Register).Methods("POST")
	r.HandleFunc("/auth/login", h.Login).Methods("POST")
}

type registerReq struct {
	Username string        `json:"username"`
	Password string        `json:"password"`
	Email    string        `json:"email"`
	Name     string        `json:"name"`
	Surname  string        `json:"surname"`
	Address  model.Address `json:"address"`
	Roles    []string      `json:"roles"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var in registerReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	in.Username = strings.TrimSpace(in.Username)
	in.Email = strings.TrimSpace(in.Email)
	if in.Username == "" || len(in.Password) < 6 || in.Email == "" {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		http.Error(w, "hash error", http.StatusInternalServerError)
		return
	}
	u := model.User{
		Username: in.Username,
		Password: hash,
		Email:    in.Email,
		Name:     strings.TrimSpace(in.Name),
		Surname:  strings.TrimSpace(in.Surname),
		Roles:    append([]string{}, in.Roles...),
		Address:  in.Address,
	}
	id, err := h.repo.Create(r.Context(), &u)
	if err != nil {
		// probably duplicate username/email
		http.Error(w, "user exists: "+err.Error(), http.StatusConflict)
		return
	}
	u.ID = id
	u.Password = "" // never return password hash
	writeJSON(w, http.StatusCreated, u)
}

type loginReq struct {
	Username string `json:"username"` // or email
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var in loginReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	in.Username = strings.TrimSpace(in.Username)

	// find by username or email
	filter := bson.M{"$or": []bson.M{{"username": in.Username}, {"email": in.Username}}}
	users, err := h.repo.List(r.Context(), filter, 0, 1)
	if err != nil || len(users) == 0 {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	u := users[0]
	if !auth.CheckPassword(u.Password, in.Password) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.IssueToken(u.ID.Hex(), u.Username, u.Roles, 60*time.Minute)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}
	resp := map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   int(60 * 60),
		"user": map[string]any{
			"id":       u.ID.Hex(),
			"username": u.Username,
			"email":    u.Email,
			"name":     u.Name,
			"surname":  u.Surname,
			"roles":    u.Roles,
		},
	}
	writeJSON(w, http.StatusOK, resp)
}

// MIDDLEWARE ----------------------------------------------------------------

type authContextKey struct{}

type AuthContext struct {
	UserID   string
	Username string
	Roles    []string
}

func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(h, "Bearer ")
		claims, err := auth.ParseToken(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), authContextKey{}, &AuthContext{
			UserID:   claims.UserID,
			Username: claims.Username,
			Roles:    claims.Roles,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetAuth(r *http.Request) *AuthContext {
	v := r.Context().Value(authContextKey{})
	if v == nil {
		return nil
	}
	return v.(*AuthContext)
}
