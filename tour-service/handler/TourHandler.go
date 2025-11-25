package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"tour-service/auth"
	"tour-service/model"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type tourRepo interface {
	CreateTour(ctx context.Context, t *model.Tour) (*model.Tour, error)
	GetTourByID(ctx context.Context, tourId string) (*model.Tour, error)
	GetToursByAuthor(ctx context.Context, authorId string) ([]model.Tour, error)
	UpdateTour(ctx context.Context, tourId string, authorId string, updates map[string]interface{}) (*model.Tour, error)
	GetKeyPointsByTour(ctx context.Context, tourId primitive.ObjectID) ([]model.KeyPoint, error)
	PublishTour(ctx context.Context, tourId string, authorId string) (*model.Tour, error)
	ArchiveTour(ctx context.Context, tourId string, authorId string) (*model.Tour, error)
	ActivateTour(ctx context.Context, tourId string, authorId string) (*model.Tour, error)
}

func RegisterRoutes(public *mux.Router, authRouter *mux.Router, repo tourRepo) {
	// protected routes
	if authRouter != nil {
		authRouter.HandleFunc("/tours", createTour(repo)).Methods("POST")
		authRouter.HandleFunc("/tours/{id}", updateTour(repo)).Methods("PUT")
		authRouter.HandleFunc("/tours/{id}/publish", publishTour(repo)).Methods("POST")
		authRouter.HandleFunc("/tours/{id}/archive", archiveTour(repo)).Methods("POST")
		authRouter.HandleFunc("/tours/{id}/activate", activateTour(repo)).Methods("POST")
	}
	// public routes
	public.HandleFunc("/tours/{id}", getTourByID(repo)).Methods("GET")
	public.HandleFunc("/tours/author/{authorId}", listToursByAuthor(repo)).Methods("GET")
}

type createTourRequest struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Difficulty  string                   `json:"difficulty"`
	Tags        []string                 `json:"tags"`
	Status      string                   `json:"status"`
	Durations   *model.TransportDuration `json:"durations,omitempty"`
}

func createTour(repo tourRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract authenticated user ID from JWT
		a := auth.GetAuth(r)
		if a == nil || a.UserID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req createTourRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		t := &model.Tour{
			AuthorID:    a.UserID,
			Name:        req.Name,
			Description: req.Description,
			Difficulty:  req.Difficulty,
			Tags:        req.Tags,
			Status:      req.Status,
		}
		if req.Durations != nil {
			t.Durations = *req.Durations
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		created, err := repo.CreateTour(ctx, t)
		if err != nil {
			log.Println("create tour error:", err)
			http.Error(w, "failed to create tour", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
	}
}

type updateTourRequest struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Difficulty  string                   `json:"difficulty"`
	Tags        []string                 `json:"tags"`
	Status      string                   `json:"status"`
	Price       float64                  `json:"price"`
	Distance    float64                  `json:"distance"`
	Durations   *model.TransportDuration `json:"durations,omitempty"`
}

func updateTour(repo tourRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract authenticated user ID from JWT
		a := auth.GetAuth(r)
		if a == nil || a.UserID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		tourId := vars["id"]
		if tourId == "" {
			http.Error(w, "tour id required", http.StatusBadRequest)
			return
		}

		var req updateTourRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		// Build update map
		updates := make(map[string]interface{})
		if req.Name != "" {
			updates["name"] = req.Name
		}
		if req.Description != "" {
			updates["description"] = req.Description
		}
		if req.Difficulty != "" {
			updates["difficulty"] = req.Difficulty
		}
		if req.Tags != nil {
			updates["tags"] = req.Tags
		}
		if req.Status != "" {
			updates["status"] = req.Status
		}
		if req.Price >= 0 {
			updates["price"] = req.Price
		}
		if req.Distance >= 0 {
			updates["distance"] = req.Distance
		}
		if req.Durations != nil {
			updates["durations"] = req.Durations
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		updated, err := repo.UpdateTour(ctx, tourId, a.UserID, updates)
		if err != nil {
			log.Println("update tour error:", err)
			http.Error(w, "failed to update tour", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updated)
	}
}

func getTourByID(repo tourRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tourId := vars["id"]
		if tourId == "" {
			http.Error(w, "tour id required", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		tour, err := repo.GetTourByID(ctx, tourId)
		if err != nil {
			log.Println("get tour error:", err)
			http.Error(w, "tour not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tour)
	}
}

func listToursByAuthor(repo tourRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		authorId := vars["authorId"]
		if authorId == "" {
			http.Error(w, "authorId required", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		tours, err := repo.GetToursByAuthor(ctx, authorId)
		if err != nil {
			log.Println("list tours error:", err)
			http.Error(w, "failed to list tours", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tours)
	}
}

func publishTour(repo tourRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a := auth.GetAuth(r)
		if a == nil || a.UserID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		tourId := vars["id"]
		if tourId == "" {
			http.Error(w, "tour id required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Validate tour can be published
		tour, err := repo.GetTourByID(ctx, tourId)
		if err != nil {
			http.Error(w, "tour not found", http.StatusNotFound)
			return
		}

		if tour.AuthorID != a.UserID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		// Check requirements
		if tour.Name == "" || tour.Description == "" || tour.Difficulty == "" || len(tour.Tags) == 0 {
			http.Error(w, "tour missing basic information", http.StatusBadRequest)
			return
		}

		// Convert tourId to ObjectID for keypoint query
		tourObjID, err := primitive.ObjectIDFromHex(tourId)
		if err != nil {
			http.Error(w, "invalid tour id", http.StatusBadRequest)
			return
		}

		keypoints, err := repo.GetKeyPointsByTour(ctx, tourObjID)
		if err != nil || len(keypoints) < 2 {
			http.Error(w, "tour must have at least 2 key points", http.StatusBadRequest)
			return
		}

		if tour.Durations.Walking == 0 && tour.Durations.Biking == 0 && tour.Durations.Driving == 0 {
			http.Error(w, "tour must have at least one duration defined", http.StatusBadRequest)
			return
		}

		published, err := repo.PublishTour(ctx, tourId, a.UserID)
		if err != nil {
			log.Println("publish tour error:", err)
			http.Error(w, "failed to publish tour", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(published)
	}
}

func archiveTour(repo tourRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a := auth.GetAuth(r)
		if a == nil || a.UserID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		tourId := vars["id"]
		if tourId == "" {
			http.Error(w, "tour id required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		archived, err := repo.ArchiveTour(ctx, tourId, a.UserID)
		if err != nil {
			log.Println("archive tour error:", err)
			http.Error(w, "failed to archive tour", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(archived)
	}
}

func activateTour(repo tourRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a := auth.GetAuth(r)
		if a == nil || a.UserID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		tourId := vars["id"]
		if tourId == "" {
			http.Error(w, "tour id required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		activated, err := repo.ActivateTour(ctx, tourId, a.UserID)
		if err != nil {
			log.Println("activate tour error:", err)
			http.Error(w, "failed to activate tour", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(activated)
	}
}
