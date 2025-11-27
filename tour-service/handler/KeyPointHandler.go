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

type kpRepo interface {
	CreateKeyPoint(ctx context.Context, kp *model.KeyPoint) (*model.KeyPoint, error)
	GetKeyPointsByTour(ctx context.Context, tourId primitive.ObjectID) ([]model.KeyPoint, error)
	UpdateKeyPoint(ctx context.Context, keypointId string, updates map[string]interface{}) (*model.KeyPoint, error)
	DeleteKeyPoint(ctx context.Context, keypointId string) error
	UpdateKeyPointsOrder(ctx context.Context, tourId primitive.ObjectID, orderedIds []string) error
	GetTourByID(ctx context.Context, tourId string) (*model.Tour, error)
	HasUserPurchasedTour(ctx context.Context, userId string, tourId string) (bool, error)
}

func RegisterKeyPointRoutes(public *mux.Router, authRouter *mux.Router, repo kpRepo) {
	// protected routes
	if authRouter != nil {
		authRouter.HandleFunc("/tours/{tourId}/keypoints", createKeyPoint(repo)).Methods("POST")
		authRouter.HandleFunc("/tours/{tourId}/keypoints/reorder", reorderKeyPoints(repo)).Methods("PUT")
		authRouter.HandleFunc("/keypoints/{keypointId}", updateKeyPoint(repo)).Methods("PUT")
		authRouter.HandleFunc("/keypoints/{keypointId}", deleteKeyPoint(repo)).Methods("DELETE")
	}
	// public routes
	public.HandleFunc("/tours/{tourId}/keypoints", listKeyPoints(repo)).Methods("GET")
}

type createKeyPointRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	ImageURL    string  `json:"imageUrl,omitempty"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

func createKeyPoint(repo kpRepo) http.HandlerFunc {
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
		var req createKeyPointRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}
		kp := &model.KeyPoint{
			TourID:      tourID,
			Name:        req.Name,
			Description: req.Description,
			ImageURL:    req.ImageURL,
			Latitude:    req.Latitude,
			Longitude:   req.Longitude,
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		created, err := repo.CreateKeyPoint(ctx, kp)
		if err != nil {
			log.Println("create keypoint error:", err)
			http.Error(w, "failed to create keypoint", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
	}
}

func listKeyPoints(repo kpRepo) http.HandlerFunc {
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

		// Get the tour to check status and author
		tour, err := repo.GetTourByID(ctx, tourIdStr)
		if err != nil {
			log.Println("get tour error:", err)
			http.Error(w, "failed to get tour", http.StatusInternalServerError)
			return
		}

		kps, err := repo.GetKeyPointsByTour(ctx, tourID)
		if err != nil {
			log.Println("list keypoints error:", err)
			http.Error(w, "failed to list keypoints", http.StatusInternalServerError)
			return
		}

		// Check if we should restrict keypoints for tourists
		// Get user from JWT (optional - middleware sets it if token is present)
		authCtx := auth.GetAuth(r)

		// If tour is published and user is not the author, check if they purchased the tour
		if tour.Status == "published" {
			if authCtx == nil || authCtx.UserID != tour.AuthorID {
				// User is not the author - check if they purchased the tour
				hasPurchased := false
				if authCtx != nil {
					// Check if user has purchased this tour
					purchased, err := repo.HasUserPurchasedTour(ctx, authCtx.UserID, tourIdStr)
					if err != nil {
						log.Println("error checking purchase status:", err)
					} else {
						hasPurchased = purchased
					}
				}

				// If user hasn't purchased, return only first keypoint
				if !hasPurchased && len(kps) > 0 {
					kps = kps[:1] // Return only first keypoint
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(kps)
	}
}

func updateKeyPoint(repo kpRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		keypointId := vars["keypointId"]
		if keypointId == "" {
			http.Error(w, "keypointId required", http.StatusBadRequest)
			return
		}

		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		updated, err := repo.UpdateKeyPoint(ctx, keypointId, updates)
		if err != nil {
			log.Println("update keypoint error:", err)
			http.Error(w, "failed to update keypoint", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updated)
	}
}

func deleteKeyPoint(repo kpRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		keypointId := vars["keypointId"]
		if keypointId == "" {
			http.Error(w, "keypointId required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err := repo.DeleteKeyPoint(ctx, keypointId)
		if err != nil {
			log.Println("delete keypoint error:", err)
			http.Error(w, "failed to delete keypoint", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

type reorderRequest struct {
	KeyPointIds []string `json:"keypointIds"`
}

func reorderKeyPoints(repo kpRepo) http.HandlerFunc {
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

		var req reorderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		if len(req.KeyPointIds) == 0 {
			http.Error(w, "keypointIds required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err = repo.UpdateKeyPointsOrder(ctx, tourID, req.KeyPointIds)
		if err != nil {
			log.Println("reorder keypoints error:", err)
			http.Error(w, "failed to reorder keypoints", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "keypoints reordered successfully"})
	}
}
