package models

import (
	"time"

	"github.com/google/uuid"
)

// Poll represents a poll question
type Poll struct {
	ID          uuid.UUID  `json:"id"`
	Question    string     `json:"question"`
	Description *string    `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active"`
	TotalVotes  int64      `json:"total_votes"`
}

// PollOption represents a poll option/choice
type PollOption struct {
	ID         uuid.UUID `json:"id"`
	PollID     uuid.UUID `json:"poll_id"`
	OptionText string    `json:"option_text"`
	VoteCount  int64     `json:"vote_count"`
	Position   int       `json:"position"`
	CreatedAt  time.Time `json:"created_at"`
}

// Vote represents a user's vote
type Vote struct {
	ID              uuid.UUID `json:"id"`
	PollID          uuid.UUID `json:"poll_id"`
	OptionID        uuid.UUID `json:"option_id"`
	VoterIdentifier string    `json:"-"` // Hidden from JSON response
	VotedAt         time.Time `json:"voted_at"`
}

// PollWithOptions combines poll with its options
type PollWithOptions struct {
	Poll
	Options []PollOption `json:"options"`
}

// PollResults represents poll results with percentages
type PollResults struct {
	Poll
	Options     []OptionResult `json:"options"`
	TotalVotes  int64          `json:"total_votes"`
	HasVoted    bool           `json:"has_voted"`
	VotedOption *uuid.UUID     `json:"voted_option,omitempty"`
}

// OptionResult represents an option with calculated percentage
type OptionResult struct {
	PollOption
	Percentage float64 `json:"percentage"`
}

// CreatePollRequest represents the request to create a poll
type CreatePollRequest struct {
	Question    string     `json:"question"`
	Description *string    `json:"description,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Options     []string   `json:"options"`
}

// VoteRequest represents the request to vote on a poll
type VoteRequest struct {
	OptionID uuid.UUID `json:"option_id"`
}
