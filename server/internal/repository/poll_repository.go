package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/moabdelazem/k8s-app/internal/models"
)

// PollRepositoryInterface defines the contract for poll data access
type PollRepositoryInterface interface {
	CreatePoll(ctx context.Context, poll *models.Poll, options []models.PollOption) error
	GetPollByID(ctx context.Context, id uuid.UUID) (*models.Poll, error)
	GetPollOptions(ctx context.Context, pollID uuid.UUID) ([]models.PollOption, error)
	ListPolls(ctx context.Context, limit, offset int, activeOnly bool) ([]models.Poll, error)
	CastVote(ctx context.Context, vote *models.Vote) error
	HasVoted(ctx context.Context, pollID uuid.UUID, voterIdentifier string) (bool, *uuid.UUID, error)
	DeletePoll(ctx context.Context, id uuid.UUID) error
	GetTotalPollsCount(ctx context.Context, activeOnly bool) (int64, error)
}

type PollRepository struct {
	db *sql.DB
}

func NewPollRepository(db *sql.DB) *PollRepository {
	return &PollRepository{db: db}
}

// CreatePoll creates a new poll with options
func (r *PollRepository) CreatePoll(ctx context.Context, poll *models.Poll, options []models.PollOption) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert poll
	query := `
		INSERT INTO polls (question, description, expires_at, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, total_votes`

	err = tx.QueryRowContext(ctx, query,
		poll.Question,
		poll.Description,
		poll.ExpiresAt,
		poll.IsActive,
	).Scan(&poll.ID, &poll.CreatedAt, &poll.TotalVotes)

	if err != nil {
		return fmt.Errorf("failed to insert poll: %w", err)
	}

	// Insert options
	optionQuery := `
		INSERT INTO poll_options (poll_id, option_text, position)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, vote_count`

	for i := range options {
		options[i].PollID = poll.ID
		options[i].Position = i

		err = tx.QueryRowContext(ctx, optionQuery,
			options[i].PollID,
			options[i].OptionText,
			options[i].Position,
		).Scan(&options[i].ID, &options[i].CreatedAt, &options[i].VoteCount)

		if err != nil {
			return fmt.Errorf("failed to insert option: %w", err)
		}
	}

	return tx.Commit()
}

// GetPollByID retrieves a poll by ID
func (r *PollRepository) GetPollByID(ctx context.Context, id uuid.UUID) (*models.Poll, error) {
	query := `
		SELECT id, question, description, created_at, expires_at, is_active, total_votes
		FROM polls
		WHERE id = $1`

	poll := &models.Poll{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&poll.ID,
		&poll.Question,
		&poll.Description,
		&poll.CreatedAt,
		&poll.ExpiresAt,
		&poll.IsActive,
		&poll.TotalVotes,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get poll: %w", err)
	}

	return poll, nil
}

// GetPollOptions retrieves all options for a poll
func (r *PollRepository) GetPollOptions(ctx context.Context, pollID uuid.UUID) ([]models.PollOption, error) {
	query := `
		SELECT id, poll_id, option_text, vote_count, position, created_at
		FROM poll_options
		WHERE poll_id = $1
		ORDER BY position ASC`

	rows, err := r.db.QueryContext(ctx, query, pollID)
	if err != nil {
		return nil, fmt.Errorf("failed to query options: %w", err)
	}
	defer rows.Close()

	var options []models.PollOption
	for rows.Next() {
		var opt models.PollOption
		err := rows.Scan(
			&opt.ID,
			&opt.PollID,
			&opt.OptionText,
			&opt.VoteCount,
			&opt.Position,
			&opt.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan option: %w", err)
		}
		options = append(options, opt)
	}

	return options, rows.Err()
}

// ListPolls retrieves polls with pagination
func (r *PollRepository) ListPolls(ctx context.Context, limit, offset int, activeOnly bool) ([]models.Poll, error) {
	query := `
		SELECT id, question, description, created_at, expires_at, is_active, total_votes
		FROM polls
		WHERE ($1 = false OR (is_active = true AND (expires_at IS NULL OR expires_at > NOW())))
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, activeOnly, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query polls: %w", err)
	}
	defer rows.Close()

	var polls []models.Poll
	for rows.Next() {
		var poll models.Poll
		err := rows.Scan(
			&poll.ID,
			&poll.Question,
			&poll.Description,
			&poll.CreatedAt,
			&poll.ExpiresAt,
			&poll.IsActive,
			&poll.TotalVotes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan poll: %w", err)
		}
		polls = append(polls, poll)
	}

	return polls, rows.Err()
}

// CastVote records a vote for an option
func (r *PollRepository) CastVote(ctx context.Context, vote *models.Vote) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert vote (will fail if voter already voted due to unique constraint)
	voteQuery := `
		INSERT INTO votes (poll_id, option_id, voter_identifier)
		VALUES ($1, $2, $3)
		RETURNING id, voted_at`

	err = tx.QueryRowContext(ctx, voteQuery,
		vote.PollID,
		vote.OptionID,
		vote.VoterIdentifier,
	).Scan(&vote.ID, &vote.VotedAt)

	if err != nil {
		return fmt.Errorf("failed to cast vote: %w", err)
	}

	// Increment option vote count
	updateQuery := `
		UPDATE poll_options
		SET vote_count = vote_count + 1
		WHERE id = $1`

	_, err = tx.ExecContext(ctx, updateQuery, vote.OptionID)
	if err != nil {
		return fmt.Errorf("failed to update vote count: %w", err)
	}

	return tx.Commit()
}

// HasVoted checks if a voter has already voted on a poll
func (r *PollRepository) HasVoted(ctx context.Context, pollID uuid.UUID, voterIdentifier string) (bool, *uuid.UUID, error) {
	query := `
		SELECT option_id
		FROM votes
		WHERE poll_id = $1 AND voter_identifier = $2`

	var optionID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, pollID, voterIdentifier).Scan(&optionID)

	if err == sql.ErrNoRows {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, fmt.Errorf("failed to check vote: %w", err)
	}

	return true, &optionID, nil
}

// DeletePoll soft deletes a poll
func (r *PollRepository) DeletePoll(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE polls
		SET is_active = false
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete poll: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("poll not found")
	}

	return nil
}

// GetTotalPollsCount returns the total number of polls
func (r *PollRepository) GetTotalPollsCount(ctx context.Context, activeOnly bool) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM polls
		WHERE ($1 = false OR (is_active = true AND (expires_at IS NULL OR expires_at > NOW())))`

	var count int64
	err := r.db.QueryRowContext(ctx, query, activeOnly).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count polls: %w", err)
	}

	return count, nil
}
