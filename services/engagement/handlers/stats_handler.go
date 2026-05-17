package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"microblog/services/engagement/logger"
	"microblog/services/engagement/service"
)

const (
	urlPartsStats = 4 //кол-во частей после разделения url
)

type StatsHandler struct {
	engService *service.EngService
	log        *logger.Logger
}

func NewStatsHandler(engService *service.EngService, log *logger.Logger) *StatsHandler {
	if engService == nil {
		panic("StatsHandlers: statsService cannot be nil")
	}
	if log == nil {
		panic("StatsHandlers: log cannot be nil")
	}
	return &StatsHandler{
		engService: engService,
		log:        log,
	}
}

func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.log.Log("http_method_not_allowed", map[string]any{
			"method": r.Method,
		})
		http.Error(w, "Error! Only GET", http.StatusMethodNotAllowed)
		return
	}

	// /stats/posts/{id}
	path := r.URL.Path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < urlPartsStats {
		h.log.Log("http_invalid_stats_path", map[string]any{
			"path": path,
		})
		http.Error(w, "Invalid path. Use /stats/posts/{id}", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(parts[3])
	if err != nil || postID <= 0 {
		h.log.Log("http_invalid_post_id", map[string]any{
			"path": path,
		})
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	likes, err := h.engService.GetPostStats(postID)
	if err != nil {
		h.log.Log("get_post_failed", map[string]any{
			"like_id": postID,
			"error":   err.Error(),
		})
		http.Error(w, "Failed to get post data", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"post_id":    postID,
		"like_count": likes,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.log.Log("http_stats_response_sent", map[string]any{
		"post_id": postID,
		"likes":   likes,
	})
}
