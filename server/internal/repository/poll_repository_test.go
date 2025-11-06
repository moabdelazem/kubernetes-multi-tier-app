package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/moabdelazem/k8s-app/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: These are integration tests that require a running database
// They should be run with build tag: go test -tags=integration

// setupTestDB creates a test database connection
// This is a placeholder - in real scenarios, use testcontainers or similar
func setupTestDB(t *testing.T) *sql.DB {
	// This would connect to a test database
	// For now, we'll skip these tests if no DB is available
	t.Skip("Integration test - requires database")
	return nil
}

func TestCreatePoll_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPollRepository(db)
	ctx := context.Background()

	poll := &models.Poll{
		Question:  "Test poll question?",
		IsActive:  true,
		ExpiresAt: nil,
	}

	options := []models.PollOption{
		{OptionText: "Option 1", Position: 0},
		{OptionText: "Option 2", Position: 1},
	}

	// Act
	err := repo.CreatePoll(ctx, poll, options)

	// Assert
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, poll.ID)
	assert.False(t, poll.CreatedAt.IsZero())

	for i, opt := range options {
		assert.NotEqual(t, uuid.Nil, opt.ID)
		assert.Equal(t, poll.ID, opt.PollID)
		assert.Equal(t, i, opt.Position)
	}
}

func TestGetPollByID_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPollRepository(db)
	ctx := context.Background()

	// Create a poll first
	poll := &models.Poll{
		Question: "Test poll?",
		IsActive: true,
	}
	options := []models.PollOption{
		{OptionText: "Yes", Position: 0},
		{OptionText: "No", Position: 1},
	}
	err := repo.CreatePoll(ctx, poll, options)
	require.NoError(t, err)

	// Act
	retrieved, err := repo.GetPollByID(ctx, poll.ID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, poll.ID, retrieved.ID)
	assert.Equal(t, poll.Question, retrieved.Question)
}

func TestCastVote_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPollRepository(db)
	ctx := context.Background()

	// Create a poll
	poll := &models.Poll{
		Question: "Test poll?",
		IsActive: true,
	}
	options := []models.PollOption{
		{OptionText: "Yes", Position: 0},
		{OptionText: "No", Position: 1},
	}
	err := repo.CreatePoll(ctx, poll, options)
	require.NoError(t, err)

	// Cast vote
	vote := &models.Vote{
		PollID:          poll.ID,
		OptionID:        options[0].ID,
		VoterIdentifier: "test-voter-1",
	}

	err = repo.CastVote(ctx, vote)

	// Assert
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, vote.ID)
	assert.False(t, vote.VotedAt.IsZero())
}

func TestHasVoted_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPollRepository(db)
	ctx := context.Background()

	// Create poll and cast vote
	poll := &models.Poll{
		Question: "Test poll?",
		IsActive: true,
	}
	options := []models.PollOption{
		{OptionText: "Yes", Position: 0},
	}
	err := repo.CreatePoll(ctx, poll, options)
	require.NoError(t, err)

	vote := &models.Vote{
		PollID:          poll.ID,
		OptionID:        options[0].ID,
		VoterIdentifier: "test-voter-1",
	}
	err = repo.CastVote(ctx, vote)
	require.NoError(t, err)

	// Check if voted
	hasVoted, optionID, err := repo.HasVoted(ctx, poll.ID, "test-voter-1")

	// Assert
	require.NoError(t, err)
	assert.True(t, hasVoted)
	assert.NotNil(t, optionID)
	assert.Equal(t, options[0].ID, *optionID)

	// Check for non-voter
	hasVoted2, optionID2, err := repo.HasVoted(ctx, poll.ID, "non-existent-voter")
	require.NoError(t, err)
	assert.False(t, hasVoted2)
	assert.Nil(t, optionID2)
}
