package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/moabdelazem/k8s-app/internal/models"
	"github.com/moabdelazem/k8s-app/internal/service"
	"github.com/moabdelazem/k8s-app/pkg/logger"
	"github.com/moabdelazem/k8s-app/pkg/response"
	"go.uber.org/zap"
)

type PollHandler struct {
	service *service.PollService
}

func NewPollHandler(service *service.PollService) *PollHandler {
	return &PollHandler{service: service}
}

// getVoterIdentifier generates a voter identifier from request
// In production, this could be user ID from auth token
// For now, we use IP address as identifier
func (h *PollHandler) getVoterIdentifier(r *http.Request) string {
	// Try to get real IP from headers (for load balancer/proxy scenarios)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	return r.RemoteAddr
}

// CreatePoll creates a new poll
func (h *PollHandler) CreatePoll(w http.ResponseWriter, r *http.Request) {
	logger.Info("Creating new poll", zap.String("handler", "CreatePoll"))

	var req models.CreatePollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request", zap.Error(err))
		response.BadRequest(w, "Invalid request body")
		return
	}

	poll, err := h.service.CreatePoll(r.Context(), &req)
	if err != nil {
		logger.Error("Failed to create poll", zap.Error(err))
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, "Poll created successfully", poll)
}

// GetPoll retrieves a poll with results
func (h *PollHandler) GetPoll(w http.ResponseWriter, r *http.Request) {
	pollIDStr := chi.URLParam(r, "id")
	pollID, err := uuid.Parse(pollIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid poll ID")
		return
	}

	voterIdentifier := h.getVoterIdentifier(r)
	results, err := h.service.GetPollResults(r.Context(), pollID, voterIdentifier)
	if err != nil {
		logger.Error("Failed to get poll results",
			zap.Error(err),
			zap.String("poll_id", pollIDStr),
		)
		response.NotFound(w, err.Error())
		return
	}

	response.Success(w, "", results)
}

// ListPolls lists all polls with pagination
func (h *PollHandler) ListPolls(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	activeOnlyStr := r.URL.Query().Get("active")

	limit, _ := strconv.Atoi(limitStr)
	if limit == 0 {
		limit = 20
	}

	offset, _ := strconv.Atoi(offsetStr)

	activeOnly := activeOnlyStr == "true"

	polls, total, err := h.service.ListPolls(r.Context(), limit, offset, activeOnly)
	if err != nil {
		logger.Error("Failed to list polls", zap.Error(err))
		response.InternalServerError(w, "Failed to retrieve polls")
		return
	}

	response.Success(w, "", map[string]interface{}{
		"polls":  polls,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// VoteOnPoll casts a vote on a poll
func (h *PollHandler) VoteOnPoll(w http.ResponseWriter, r *http.Request) {
	pollIDStr := chi.URLParam(r, "id")
	pollID, err := uuid.Parse(pollIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid poll ID")
		return
	}

	var req models.VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode vote request", zap.Error(err))
		response.BadRequest(w, "Invalid request body")
		return
	}

	voterIdentifier := h.getVoterIdentifier(r)

	err = h.service.CastVote(r.Context(), pollID, req.OptionID, voterIdentifier)
	if err != nil {
		logger.Error("Failed to cast vote",
			zap.Error(err),
			zap.String("poll_id", pollIDStr),
			zap.String("option_id", req.OptionID.String()),
		)
		response.BadRequest(w, err.Error())
		return
	}

	// Get updated results
	results, err := h.service.GetPollResults(r.Context(), pollID, voterIdentifier)
	if err != nil {
		logger.Warn("Failed to get updated results after vote", zap.Error(err))
		response.Success(w, "Vote cast successfully", nil)
		return
	}

	response.Success(w, "Vote cast successfully", results)
}

// DeletePoll soft deletes a poll
func (h *PollHandler) DeletePoll(w http.ResponseWriter, r *http.Request) {
	pollIDStr := chi.URLParam(r, "id")
	pollID, err := uuid.Parse(pollIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid poll ID")
		return
	}

	err = h.service.DeletePoll(r.Context(), pollID)
	if err != nil {
		logger.Error("Failed to delete poll",
			zap.Error(err),
			zap.String("poll_id", pollIDStr),
		)
		response.InternalServerError(w, err.Error())
		return
	}

	response.Success(w, "Poll deleted successfully", nil)
}
