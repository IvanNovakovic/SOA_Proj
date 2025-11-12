package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	// "time"

	"stakeholders-service/model"
	"stakeholders-service/repository"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserHandler struct {
	repo *repository.UserRepository
}

func RegisterRoutes(router *mux.Router, userRepo *repository.UserRepository) {
	h := &UserHandler{repo: userRepo}

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	router.HandleFunc("/users", h.ListUsers).Methods("GET")
	router.HandleFunc("/users", h.CreateUser).Methods("POST")
	router.HandleFunc("/users/{id}", h.GetUserByID).Methods("GET")
	router.HandleFunc("/users/{id}", h.UpdateUserFields).Methods("PUT", "PATCH")
	router.HandleFunc("/users/{id}/password", h.UpdatePassword).Methods("PATCH")
	router.HandleFunc("/users/{id}", h.DeleteUser).Methods("DELETE")
}

// HANDLERS -----------------------------------------------------------------

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Simple filtering + pagination
	q := r.URL.Query()
	filter := bson.M{}
	if v := strings.TrimSpace(q.Get("username")); v != "" {
		filter["username"] = v
	}
	if v := strings.TrimSpace(q.Get("email")); v != "" {
		filter["email"] = v
	}
	skip := parseInt64(q.Get("skip"), 0)
	limit := parseInt64(q.Get("limit"), 0)

	var users []model.User
	var err error
	if len(filter) == 0 && skip == 0 && limit == 0 {
		users, err = h.repo.GetAll(ctx)
	} else {
		users, err = h.repo.List(ctx, filter, skip, limit)
	}
	if err != nil {
		http.Error(w, "failed to list users", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, http.StatusOK, users)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := parseObjectID(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	user, err := h.repo.GetByID(ctx, id)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var in model.User
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := validateNewUser(in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.repo.Create(ctx, &in)
	if err != nil {
		// likely duplicate username/email -> 409
		http.Error(w, "failed to create user", http.StatusConflict)
		return
	}
	in.ID = id
	w.Header().Set("Location", "/users/"+id.Hex())
	writeJSON(w, http.StatusCreated, in)
}

func (h *UserHandler) UpdateUserFields(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := parseObjectID(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// Accept partial fields
	var fields map[string]any
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	// Prevent changing password here (use /password) and _id
	delete(fields, "_id")
	delete(fields, "password")

	if err := h.repo.UpdateFields(ctx, id, bson.M(fields)); err != nil {
		http.Error(w, "failed to update user", http.StatusInternalServerError)
		return
	}
	updated, err := h.repo.GetByID(ctx, id)
	if err != nil {
		http.Error(w, "failed to load updated user", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := parseObjectID(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var in struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(strings.TrimSpace(in.Password)) < 6 {
		http.Error(w, "password too short", http.StatusBadRequest)
		return
	}

	// NOTE: hash the password here in real apps
	if err := h.repo.UpdatePassword(ctx, id, in.Password); err != nil {
		http.Error(w, "failed to update password", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := parseObjectID(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.repo.DeleteByID(ctx, id); err != nil {
		http.Error(w, "failed to delete user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UTILITIES ----------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func parseObjectID(s string) (primitive.ObjectID, error) {
	if !primitive.IsValidObjectID(s) {
		return primitive.NilObjectID, errors.New("invalid")
	}
	return primitive.ObjectIDFromHex(s)
}

func parseInt64(s string, def int64) int64 {
	if s == "" {
		return def
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return n
}

var emailRegex = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)

func validateNewUser(u model.User) error {
	if strings.TrimSpace(u.Username) == "" {
		return errors.New("username required")
	}
	if len(u.Username) < 3 {
		return errors.New("username too short")
	}
	if strings.TrimSpace(u.Email) == "" {
		return errors.New("email required")
	}
	if !emailRegex.MatchString(u.Email) {
		return errors.New("invalid email")
	}
	if strings.TrimSpace(u.Name) == "" {
		return errors.New("name required")
	}
	if strings.TrimSpace(u.Surname) == "" {
		return errors.New("surname required")
	}
	// password cannot be validated here because json:"-" prevents input
	return nil
}
