package requests

import (
	"context"
	"errors"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateRequest(ctx context.Context, userID int64, reqType string, startDate, endDate time.Time, reason string) (*Request, error) {
	if startDate.After(endDate) {
		return nil, errors.New("start date must be before end date")
	}
	if reason == "" {
		return nil, errors.New("reason is required")
	}

	req := &Request{
		UserID:    userID,
		Type:      reqType,
		StartDate: startDate,
		EndDate:   endDate,
		Reason:    reason,
	}
	err := s.repo.Create(ctx, req)
	return req, err
}

func (s *Service) GetMyRequests(ctx context.Context, userID int64) ([]*Request, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *Service) GetPendingRequests(ctx context.Context) ([]*Request, error) {
	return s.repo.FindPending(ctx)
}

func (s *Service) ApproveRequest(ctx context.Context, id int64, approverID int64) error {
	// Check if already processed?
	req, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if req.Status != "PENDING" {
		return errors.New("request is not pending")
	}

	return s.repo.UpdateStatus(ctx, id, "APPROVED", approverID, nil)
}

func (s *Service) RejectRequest(ctx context.Context, id int64, approverID int64, reason string) error {
	req, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if req.Status != "PENDING" {
		return errors.New("request is not pending")
	}
	if reason == "" {
		return errors.New("rejection reason is required")
	}

	return s.repo.UpdateStatus(ctx, id, "REJECTED", approverID, &reason)
}
