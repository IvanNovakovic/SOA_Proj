package handler

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "time"

    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "tour-service/model"
)

type kpRepo interface {
    CreateKeyPoint(ctx context.Context, kp *model.KeyPoint) (*model.KeyPoint, error)
    GetKeyPointsByTour(ctx context.Context, tourId primitive.ObjectID) ([]model.KeyPoint, error)
}

func RegisterKeyPointRoutes(public *mux.Router, authRouter *mux.Router, repo kpRepo) {
    // protected routes
    if authRouter != nil {
        authRouter.HandleFunc("/tours/{tourId}/keypoints", createKeyPoint(repo)).Methods("POST")
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
        kps, err := repo.GetKeyPointsByTour(ctx, tourID)
        if err != nil {
            log.Println("list keypoints error:", err)
            http.Error(w, "failed to list keypoints", http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(kps)
    }
}
