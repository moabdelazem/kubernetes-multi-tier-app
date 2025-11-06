package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/moabdelazem/k8s-app/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockPollRepository is a mock implementation of PollRepository
type MockPollRepository struct {
	mock.Mock
}

func (m *MockPollRepository) CreatePoll(ctx context.Context, poll *models.Poll, options []models.PollOption) error {
	args := m.Called(ctx, poll, options)
	return args.Error(0)
}

func (m *MockPollRepository) GetPollByID(ctx context.Context, id uuid.UUID) (*models.Poll, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Poll), args.Error(1)
}

func (m *MockPollRepository) GetPollOptions(ctx context.Context, pollID uuid.UUID) ([]models.PollOption, error) {
	args := m.Called(ctx, pollID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PollOption), args.Error(1)
}

func (m *MockPollRepository) ListPolls(ctx context.Context, limit, offset int, activeOnly bool) ([]models.Poll, error) {
	args := m.Called(ctx, limit, offset, activeOnly)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Poll), args.Error(1)
}

func (m *MockPollRepository) CastVote(ctx context.Context, vote *models.Vote) error {
	args := m.Called(ctx, vote)
	return args.Error(0)
}

func (m *MockPollRepository) HasVoted(ctx context.Context, pollID uuid.UUID, voterIdentifier string) (bool, *uuid.UUID, error) {
	args := m.Called(ctx, pollID, voterIdentifier)
	if args.Get(1) == nil {
		return args.Bool(0), nil, args.Error(2)
	}
	return args.Bool(0), args.Get(1).(*uuid.UUID), args.Error(2)
}

func (m *MockPollRepository) DeletePoll(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPollRepository) GetTotalPollsCount(ctx context.Context, activeOnly bool) (int64, error) {
	args := m.Called(ctx, activeOnly)
	return args.Get(0).(int64), args.Error(1)
}
