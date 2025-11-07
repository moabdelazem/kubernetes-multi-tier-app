package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/moabdelazem/k8s-app/internal/models"
	"github.com/moabdelazem/k8s-app/internal/repository"
	"github.com/moabdelazem/k8s-app/pkg/logger"
	"go.uber.org/zap"
)

type PollService struct {
	repo repository.PollRepositoryInterface
}

func NewPollService(repo repository.PollRepositoryInterface) *PollService {
	return &PollService{repo: repo}
}

// CreatePoll creates a new poll with validation
func (s *PollService) CreatePoll(ctx context.Context, req *models.CreatePollRequest) (*models.PollWithOptions, error) {
	// Validate request
	if len(req.Question) < 5 || len(req.Question) > 500 {
		return nil, fmt.Errorf("question must be between 5 and 500 characters")
	}

	if len(req.Options) < 2 {
		return nil, fmt.Errorf("poll must have at least 2 options")
	}

	if len(req.Options) > 10 {
		return nil, fmt.Errorf("poll can have at most 10 options")
	}

	// Validate each option
	for i, opt := range req.Options {
		if len(opt) < 1 || len(opt) > 200 {
			return nil, fmt.Errorf("option %d must be between 1 and 200 characters", i+1)
		}
	}

	// Check expiration date
	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("expiration date must be in the future")
	}

	// Create poll
	poll := &models.Poll{
		Question:    req.Question,
		Description: req.Description,
		ExpiresAt:   req.ExpiresAt,
		IsActive:    true,
	}

	// Create options
	options := make([]models.PollOption, len(req.Options))
	for i, optText := range req.Options {
		options[i] = models.PollOption{
			OptionText: optText,
			Position:   i,
		}
	}

	// Save to database
	err := s.repo.CreatePoll(ctx, poll, options)
	if err != nil {
		logger.Error("Failed to create poll", zap.Error(err))
		return nil, fmt.Errorf("failed to create poll: %w", err)
	}

	logger.Info("Poll created successfully",
		zap.String("poll_id", poll.ID.String()),
		zap.String("question", poll.Question),
		zap.Int("options_count", len(options)),
	)

	return &models.PollWithOptions{
		Poll:    *poll,
		Options: options,
	}, nil
}

// GetPollResults retrieves poll with results and checks if voter has voted
func (s *PollService) GetPollResults(ctx context.Context, pollID uuid.UUID, voterIdentifier string) (*models.PollResults, error) {
	// Get poll
	poll, err := s.repo.GetPollByID(ctx, pollID)
	if err != nil {
		return nil, fmt.Errorf("failed to get poll: %w", err)
	}
	if poll == nil {
		return nil, fmt.Errorf("poll not found")
	}

	// Get options
	options, err := s.repo.GetPollOptions(ctx, pollID)
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}

	// Check if voter has voted
	hasVoted, votedOptionID, err := s.repo.HasVoted(ctx, pollID, voterIdentifier)
	if err != nil {
		logger.Warn("Failed to check vote status", zap.Error(err))
	}

	// Calculate percentages
	results := make([]models.OptionResult, len(options))
	for i, opt := range options {
		percentage := 0.0
		if poll.TotalVotes > 0 {
			percentage = float64(opt.VoteCount) / float64(poll.TotalVotes) * 100
		}
		results[i] = models.OptionResult{
			PollOption: opt,
			Percentage: percentage,
		}
	}

	return &models.PollResults{
		Poll:        *poll,
		Options:     results,
		TotalVotes:  poll.TotalVotes,
		HasVoted:    hasVoted,
		VotedOption: votedOptionID,
	}, nil
}

// CastVote casts a vote on a poll
func (s *PollService) CastVote(ctx context.Context, pollID uuid.UUID, optionID uuid.UUID, voterIdentifier string) error {
	// Get poll
	poll, err := s.repo.GetPollByID(ctx, pollID)
	if err != nil {
		return fmt.Errorf("failed to get poll: %w", err)
	}
	if poll == nil {
		return fmt.Errorf("poll not found")
	}

	// Check if poll is active
	if !poll.IsActive {
		return fmt.Errorf("poll is not active")
	}

	// Check if poll is expired
	if poll.ExpiresAt != nil && poll.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("poll has expired")
	}

	// Check if voter has already voted
	hasVoted, _, err := s.repo.HasVoted(ctx, pollID, voterIdentifier)
	if err != nil {
		return fmt.Errorf("failed to check vote status: %w", err)
	}
	if hasVoted {
		return fmt.Errorf("you have already voted on this poll")
	}

	// Verify option belongs to this poll
	options, err := s.repo.GetPollOptions(ctx, pollID)
	if err != nil {
		return fmt.Errorf("failed to get poll options: %w", err)
	}

	validOption := false
	for _, opt := range options {
		if opt.ID == optionID {
			validOption = true
			break
		}
	}
	if !validOption {
		return fmt.Errorf("invalid option for this poll")
	}

	// Cast vote
	vote := &models.Vote{
		PollID:          pollID,
		OptionID:        optionID,
		VoterIdentifier: voterIdentifier,
	}

	err = s.repo.CastVote(ctx, vote)
	if err != nil {
		logger.Error("Failed to cast vote",
			zap.Error(err),
			zap.String("poll_id", pollID.String()),
			zap.String("option_id", optionID.String()),
		)
		return fmt.Errorf("failed to cast vote: %w", err)
	}

	logger.Info("Vote cast successfully",
		zap.String("poll_id", pollID.String()),
		zap.String("option_id", optionID.String()),
		zap.String("voter", voterIdentifier),
	)

	return nil
}

// ListPolls lists polls with pagination and includes options
func (s *PollService) ListPolls(ctx context.Context, limit, offset int, activeOnly bool) ([]models.PollWithOptions, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	polls, err := s.repo.ListPollsWithOptions(ctx, limit, offset, activeOnly)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list polls: %w", err)
	}

	total, err := s.repo.GetTotalPollsCount(ctx, activeOnly)
	if err != nil {
		logger.Warn("Failed to get total count", zap.Error(err))
		total = 0
	}

	return polls, total, nil
}

// DeletePoll soft deletes a poll
func (s *PollService) DeletePoll(ctx context.Context, pollID uuid.UUID) error {
	err := s.repo.DeletePoll(ctx, pollID)
	if err != nil {
		logger.Error("Failed to delete poll",
			zap.Error(err),
			zap.String("poll_id", pollID.String()),
		)
		return fmt.Errorf("failed to delete poll: %w", err)
	}

	logger.Info("Poll deleted successfully",
		zap.String("poll_id", pollID.String()),
	)

	return nil
}
