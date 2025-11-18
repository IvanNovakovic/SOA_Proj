package handler

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"

    "follower-service/repository"
)

type followReq struct {
    Follower string `json:"follower"`
    Followee string `json:"followee"`
}

type followHandler struct {
    repo *repository.NeoRepository
}

func RegisterRoutes(r *mux.Router, repo *repository.NeoRepository) {
    h := &followHandler{repo: repo}
    r.HandleFunc("/follow", h.follow).Methods("POST")
    r.HandleFunc("/follow", h.unfollow).Methods("DELETE")
    r.HandleFunc("/followers/{id}", h.getFollowers).Methods("GET")
    r.HandleFunc("/following/{id}", h.getFollowing).Methods("GET")
    r.HandleFunc("/recommendations/{id}", h.recommendations).Methods("GET")
}

func (h *followHandler) follow(w http.ResponseWriter, r *http.Request) {
    var in followReq
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    if in.Follower == "" || in.Followee == "" {
        http.Error(w, "follower and followee required", http.StatusBadRequest)
        return
    }
    if err := h.repo.Follow(r.Context(), in.Follower, in.Followee); err != nil {
        http.Error(w, "failed to follow: "+err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusCreated)
}

func (h *followHandler) unfollow(w http.ResponseWriter, r *http.Request) {
    var in followReq
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    if in.Follower == "" || in.Followee == "" {
        http.Error(w, "follower and followee required", http.StatusBadRequest)
        return
    }
    if _, err := h.repo.Unfollow(r.Context(), in.Follower, in.Followee); err != nil {
        http.Error(w, "failed to unfollow: "+err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
}

func (h *followHandler) getFollowers(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    if id == "" {
        http.Error(w, "id required", http.StatusBadRequest)
        return
    }
    out, err := h.repo.Followers(r.Context(), id)
    if err != nil {
        http.Error(w, "failed to get followers: "+err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(out)
}

func (h *followHandler) getFollowing(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    if id == "" {
        http.Error(w, "id required", http.StatusBadRequest)
        return
    }
    out, err := h.repo.Following(r.Context(), id)
    if err != nil {
        http.Error(w, "failed to get following: "+err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(out)
}

func (h *followHandler) recommendations(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    if id == "" {
        http.Error(w, "id required", http.StatusBadRequest)
        return
    }
    q := r.URL.Query().Get("limit")
    limit := 10
    if q != "" {
        if v, err := strconv.Atoi(q); err == nil && v > 0 {
            limit = v
        }
    }
    out, err := h.repo.Recommendations(r.Context(), id, limit)
    if err != nil {
        http.Error(w, "failed to get recommendations: "+err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(out)
}
