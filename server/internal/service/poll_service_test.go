package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moabdelazem/k8s-app/internal/mocks"
	"github.com/moabdelazem/k8s-app/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreatePoll_Success(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPollRepository)
	service := NewPollService(mockRepo)

	req := &models.CreatePollRequest{
		Question:    "What's your favorite language?",
		Description: strPtr("Programming languages poll"),
		Options:     []string{"Go", "Python", "JavaScript"},
	}

	mockRepo.On("CreatePoll", mock.Anything, mock.AnythingOfType("*models.Poll"), mock.AnythingOfType("[]models.PollOption")).
		Return(nil).
		Run(func(args mock.Arguments) {
			poll := args.Get(1).(*models.Poll)
			poll.ID = uuid.New()
			poll.CreatedAt = time.Now()
			poll.TotalVotes = 0
			poll.IsActive = true

			options := args.Get(2).([]models.PollOption)
			for i := range options {
				options[i].ID = uuid.New()
				options[i].PollID = poll.ID
				options[i].CreatedAt = time.Now()
				options[i].VoteCount = 0
			}
		})

	// Act
	result, err := service.CreatePoll(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.Question, result.Question)
	assert.Equal(t, 3, len(result.Options))
	mockRepo.AssertExpectations(t)
}

func TestCreatePoll_ValidationError_QuestionTooShort(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPollRepository)
	service := NewPollService(mockRepo)

	req := &models.CreatePollRequest{
		Question: "Hi?", // Too short (< 5 characters)
		Options:  []string{"Option 1", "Option 2"},
	}

	// Act
	result, err := service.CreatePoll(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "question must be between 5 and 500 characters")
	mockRepo.AssertNotCalled(t, "CreatePoll")
}

func TestCreatePoll_ValidationError_NotEnoughOptions(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPollRepository)
	service := NewPollService(mockRepo)

	req := &models.CreatePollRequest{
		Question: "Valid question?",
		Options:  []string{"Only one option"}, // Need at least 2
	}

	// Act
	result, err := service.CreatePoll(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "poll must have at least 2 options")
	mockRepo.AssertNotCalled(t, "CreatePoll")
}

func TestCreatePoll_ValidationError_TooManyOptions(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPollRepository)
	service := NewPollService(mockRepo)

	req := &models.CreatePollRequest{
		Question: "Valid question?",
		Options:  []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}, // More than 10
	}

	// Act
	result, err := service.CreatePoll(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "poll can have at most 10 options")
	mockRepo.AssertNotCalled(t, "CreatePoll")
}

func TestCreatePoll_ValidationError_ExpirationInPast(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPollRepository)
	service := NewPollService(mockRepo)

	pastTime := time.Now().Add(-1 * time.Hour)
	req := &models.CreatePollRequest{
		Question:  "Valid question?",
		Options:   []string{"Option 1", "Option 2"},
		ExpiresAt: &pastTime,
	}

	// Act
	result, err := service.CreatePoll(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "expiration date must be in the future")
	mockRepo.AssertNotCalled(t, "CreatePoll")
}

func TestCastVote_Success(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPollRepository)
	service := NewPollService(mockRepo)

	pollID := uuid.New()
	optionID := uuid.New()
	voterIdentifier := "127.0.0.1"

	poll := &models.Poll{
		ID:         pollID,
		Question:   "Test poll?",
		IsActive:   true,
		ExpiresAt:  nil,
		TotalVotes: 0,
	}

	options := []models.PollOption{
		{ID: optionID, PollID: pollID, OptionText: "Option 1"},
		{ID: uuid.New(), PollID: pollID, OptionText: "Option 2"},
	}

	mockRepo.On("GetPollByID", mock.Anything, pollID).Return(poll, nil)
	mockRepo.On("HasVoted", mock.Anything, pollID, voterIdentifier).Return(false, nil, nil)
	mockRepo.On("GetPollOptions", mock.Anything, pollID).Return(options, nil)
	mockRepo.On("CastVote", mock.Anything, mock.AnythingOfType("*models.Vote")).Return(nil)

	// Act
	err := service.CastVote(context.Background(), pollID, optionID, voterIdentifier)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCastVote_Error_PollNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPollRepository)
	service := NewPollService(mockRepo)

	pollID := uuid.New()
	optionID := uuid.New()
	voterIdentifier := "127.0.0.1"

	mockRepo.On("GetPollByID", mock.Anything, pollID).Return(nil, nil)

	// Act
	err := service.CastVote(context.Background(), pollID, optionID, voterIdentifier)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "poll not found")
	mockRepo.AssertExpectations(t)
}

func TestCastVote_Error_PollNotActive(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPollRepository)
	service := NewPollService(mockRepo)

	pollID := uuid.New()
	optionID := uuid.New()
	voterIdentifier := "127.0.0.1"

	poll := &models.Poll{
		ID:       pollID,
		Question: "Test poll?",
		IsActive: false, // Not active
	}

	mockRepo.On("GetPollByID", mock.Anything, pollID).Return(poll, nil)

	// Act
	err := service.CastVote(context.Background(), pollID, optionID, voterIdentifier)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "poll is not active")
	mockRepo.AssertExpectations(t)
}

func TestCastVote_Error_PollExpired(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPollRepository)
	service := NewPollService(mockRepo)

	pollID := uuid.New()
	optionID := uuid.New()
	voterIdentifier := "127.0.0.1"

	expiredTime := time.Now().Add(-1 * time.Hour)
	poll := &models.Poll{
		ID:        pollID,
		Question:  "Test poll?",
		IsActive:  true,
		ExpiresAt: &expiredTime, // Expired
	}

	mockRepo.On("GetPollByID", mock.Anything, pollID).Return(poll, nil)

	// Act
	err := service.CastVote(context.Background(), pollID, optionID, voterIdentifier)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "poll has expired")
	mockRepo.AssertExpectations(t)
}

func TestCastVote_Error_AlreadyVoted(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPollRepository)
	service := NewPollService(mockRepo)

	pollID := uuid.New()
	optionID := uuid.New()
	voterIdentifier := "127.0.0.1"

	poll := &models.Poll{
		ID:        pollID,
		Question:  "Test poll?",
		IsActive:  true,
		ExpiresAt: nil,
	}

	mockRepo.On("GetPollByID", mock.Anything, pollID).Return(poll, nil)
	mockRepo.On("HasVoted", mock.Anything, pollID, voterIdentifier).Return(true, &optionID, nil)

	// Act
	err := service.CastVote(context.Background(), pollID, optionID, voterIdentifier)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you have already voted on this poll")
	mockRepo.AssertExpectations(t)
}

func TestCastVote_Error_InvalidOption(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPollRepository)
	service := NewPollService(mockRepo)

	pollID := uuid.New()
	invalidOptionID := uuid.New()
	voterIdentifier := "127.0.0.1"

	poll := &models.Poll{
		ID:        pollID,
		Question:  "Test poll?",
		IsActive:  true,
		ExpiresAt: nil,
	}

	options := []models.PollOption{
		{ID: uuid.New(), PollID: pollID, OptionText: "Option 1"},
		{ID: uuid.New(), PollID: pollID, OptionText: "Option 2"},
	}

	mockRepo.On("GetPollByID", mock.Anything, pollID).Return(poll, nil)
	mockRepo.On("HasVoted", mock.Anything, pollID, voterIdentifier).Return(false, nil, nil)
	mockRepo.On("GetPollOptions", mock.Anything, pollID).Return(options, nil)

	// Act
	err := service.CastVote(context.Background(), pollID, invalidOptionID, voterIdentifier)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid option for this poll")
	mockRepo.AssertExpectations(t)
}

func TestListPolls_Success(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPollRepository)
	service := NewPollService(mockRepo)

	expectedPolls := []models.Poll{
		{ID: uuid.New(), Question: "Poll 1", IsActive: true},
		{ID: uuid.New(), Question: "Poll 2", IsActive: true},
	}

	mockRepo.On("ListPolls", mock.Anything, 20, 0, true).Return(expectedPolls, nil)
	mockRepo.On("GetTotalPollsCount", mock.Anything, true).Return(int64(2), nil)

	// Act
	polls, total, err := service.ListPolls(context.Background(), 20, 0, true)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, len(polls))
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestListPolls_DefaultLimit(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPollRepository)
	service := NewPollService(mockRepo)

	expectedPolls := []models.Poll{}

	// When limit is 0 or negative, should default to 20
	mockRepo.On("ListPolls", mock.Anything, 20, 0, false).Return(expectedPolls, nil)
	mockRepo.On("GetTotalPollsCount", mock.Anything, false).Return(int64(0), nil)

	// Act
	polls, total, err := service.ListPolls(context.Background(), 0, 0, false)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 0, len(polls))
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}

func TestDeletePoll_Success(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockPollRepository)
	service := NewPollService(mockRepo)

	pollID := uuid.New()
	mockRepo.On("DeletePoll", mock.Anything, pollID).Return(nil)

	// Act
	err := service.DeletePoll(context.Background(), pollID)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Helper function
func strPtr(s string) *string {
	return &s
}
