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

type reviewRepo interface {
	CreateReview(ctx context.Context, rev *model.Review) (*model.Review, error)
	GetReviewsByTour(ctx context.Context, tourId primitive.ObjectID) ([]model.Review, error)
	HasUserReviewedTour(ctx context.Context, tourId primitive.ObjectID, authorId string) (bool, error)
}

func RegisterReviewRoutes(public *mux.Router, authRouter *mux.Router, repo reviewRepo) {
	// protected routes
	if authRouter != nil {
		authRouter.HandleFunc("/tours/{tourId}/reviews", createReview(repo)).Methods("POST")
	}
	// public routes
	public.HandleFunc("/tours/{tourId}/reviews", listReviews(repo)).Methods("GET")
}

type createReviewRequest struct {
	Rating    int        `json:"rating"`
	Comment   string     `json:"comment,omitempty"`
	Images    []string   `json:"images,omitempty"`
	VisitedAt *time.Time `json:"visitedAt,omitempty"`
}

func createReview(repo reviewRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract authenticated user ID from JWT
		a := auth.GetAuth(r)
		if a == nil || a.UserID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		tourIdStr := vars["tourId"]
		if tourIdStr == "" {
			http.Error(w, "tourId required", http.StatusBadRequest)
			return
		}
		tourID, err := primitive.ObjectIDFromHex(tourIdStr)
		if err != nil {
			http.Error(w, "invalid tourId", http.StatusBadRequest)
			return
		}

		// Check if user has already reviewed this tour
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		hasReviewed, err := repo.HasUserReviewedTour(ctx, tourID, a.UserID)
		if err != nil {
			log.Println("error checking existing review:", err)
			http.Error(w, "failed to check existing review", http.StatusInternalServerError)
			return
		}
		if hasReviewed {
			http.Error(w, "you have already reviewed this tour", http.StatusConflict)
			return
		}

		var req createReviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		if req.Rating < 1 || req.Rating > 5 {
			http.Error(w, "rating(1-5) required", http.StatusBadRequest)
			return
		}
		rev := &model.Review{
			TourID:     tourID,
			AuthorID:   a.UserID,
			AuthorName: a.Username,
			Rating:     req.Rating,
			Comment:    req.Comment,
			Images:     req.Images,
			VisitedAt:  req.VisitedAt,
			CreatedAt:  time.Now().UTC(),
		}

		created, err := repo.CreateReview(ctx, rev)
		if err != nil {
			log.Println("create review error:", err)
			http.Error(w, "failed to create review", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
	}
}

func listReviews(repo reviewRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tourIdStr := vars["tourId"]
		if tourIdStr == "" {
			http.Error(w, "tourId required", http.StatusBadRequest)
			return
		}
		tourID, err := primitive.ObjectIDFromHex(tourIdStr)
		if err != nil {
			http.Error(w, "invalid tourId", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		revs, err := repo.GetReviewsByTour(ctx, tourID)
		if err != nil {
			log.Println("list reviews error:", err)
			http.Error(w, "failed to list reviews", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(revs)
	}
}
