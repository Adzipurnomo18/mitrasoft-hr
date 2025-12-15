package attendance

import (
	"context"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) Checkin(ctx context.Context, userID int64, now time.Time) (*Record, error) {
	if _, err := s.repo.EnsureToday(ctx, userID, now); err != nil {
		return nil, err
	}
	return s.repo.SetCheckin(ctx, userID, now)
}

func (s *Service) Checkout(ctx context.Context, userID int64, now time.Time) (*Record, error) {
	if _, err := s.repo.EnsureToday(ctx, userID, now); err != nil {
		return nil, err
	}
	return s.repo.SetCheckout(ctx, userID, now)
}

func (s *Service) Summary(ctx context.Context, userID int64, from, to time.Time) (*Summary, error) {
	summ, err := s.repo.GetSummary(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}
	// Compute working days (Mon-Fri) between from (inclusive) and to (exclusive)
	wd := 0
	for d := from; d.Before(to); d = d.AddDate(0, 0, 1) {
		switch d.Weekday() {
		case time.Saturday, time.Sunday:
			// skip weekends
		default:
			wd++
		}
	}
	summ.WorkingDays = wd
	// Derive absent from working days - present to include non-recorded days
	if wd > 0 {
		if summ.Present < wd {
			summ.Absent = wd - summ.Present
		} else {
			summ.Absent = 0
		}
	}
	return summ, nil
}

func (s *Service) List(ctx context.Context, userID int64, from, to time.Time) ([]*Record, error) {
	return s.repo.ListBetween(ctx, userID, from, to)
}
