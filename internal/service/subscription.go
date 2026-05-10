package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"subscription-service/internal/model"
	"subscription-service/internal/repository"

	"github.com/google/uuid"
)

type SubscriptionRepo interface {
	Create(context.Context, model.Subscription) (model.Subscription, error)
	Get(context.Context, int64) (model.Subscription, error)
	List(context.Context, model.ListFilter) ([]model.Subscription, error)
	Update(context.Context, model.Subscription) (model.Subscription, error)
	Delete(context.Context, int64) error
	Sum(context.Context, model.SumFilter) (int, error)
}

type SubscriptionService struct{ repo SubscriptionRepo }

func NewSubscriptionService(repo SubscriptionRepo) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) Create(ctx context.Context, req model.CreateSubscriptionRequest) (model.Subscription, error) {
	sub, err := validateCreate(req)
	if err != nil {
		return model.Subscription{}, err
	}
	return s.repo.Create(ctx, sub)
}

func (s *SubscriptionService) Get(ctx context.Context, id int64) (model.Subscription, error) {
	return s.repo.Get(ctx, id)
}

func (s *SubscriptionService) List(ctx context.Context, f model.ListFilter) ([]model.Subscription, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return s.repo.List(ctx, f)
}

func (s *SubscriptionService) Update(ctx context.Context, id int64, req model.UpdateSubscriptionRequest) (model.Subscription, error) {
	current, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.Subscription{}, err
	}
	if req.ServiceName != nil {
		current.ServiceName = strings.TrimSpace(*req.ServiceName)
	}
	if req.Price != nil {
		current.Price = *req.Price
	}
	if req.UserID != nil {
		u, err := uuid.Parse(*req.UserID)
		if err != nil {
			return model.Subscription{}, fmt.Errorf("invalid user_id")
		}
		current.UserID = u
	}
	if req.StartDate != nil {
		d, err := model.ParseMonth(*req.StartDate)
		if err != nil {
			return model.Subscription{}, fmt.Errorf("invalid start_date, expected MM-YYYY")
		}
		current.StartDate = repository.MonthStart(d)
	}
	if req.EndDate != nil {
		if *req.EndDate == "" {
			current.EndDate = nil
		} else {
			d, err := model.ParseMonth(*req.EndDate)
			if err != nil {
				return model.Subscription{}, fmt.Errorf("invalid end_date, expected MM-YYYY")
			}
			v := repository.MonthStart(d)
			current.EndDate = &v
		}
	}
	if err := validateSubscription(current); err != nil {
		return model.Subscription{}, err
	}
	return s.repo.Update(ctx, current)
}

func (s *SubscriptionService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *SubscriptionService) Sum(ctx context.Context, f model.SumFilter) (int, error) {
	if f.From.After(f.To) {
		return 0, errors.New("from must be before or equal to to")
	}
	return s.repo.Sum(ctx, f)
}

func validateCreate(req model.CreateSubscriptionRequest) (model.Subscription, error) {
	u, err := uuid.Parse(req.UserID)
	if err != nil {
		return model.Subscription{}, fmt.Errorf("invalid user_id")
	}
	start, err := model.ParseMonth(req.StartDate)
	if err != nil {
		return model.Subscription{}, fmt.Errorf("invalid start_date, expected MM-YYYY")
	}
	var end *time.Time
	if req.EndDate != "" {
		d, err := model.ParseMonth(req.EndDate)
		if err != nil {
			return model.Subscription{}, fmt.Errorf("invalid end_date, expected MM-YYYY")
		}
		v := repository.MonthStart(d)
		end = &v
	}
	sub := model.Subscription{ServiceName: strings.TrimSpace(req.ServiceName), Price: req.Price, UserID: u, StartDate: repository.MonthStart(start), EndDate: end}
	if err := validateSubscription(sub); err != nil {
		return model.Subscription{}, err
	}
	return sub, nil
}

func validateSubscription(s model.Subscription) error {
	if strings.TrimSpace(s.ServiceName) == "" {
		return errors.New("service_name is required")
	}
	if s.Price <= 0 {
		return errors.New("price must be positive")
	}
	if s.EndDate != nil && s.EndDate.Before(s.StartDate) {
		return errors.New("end_date must be after or equal to start_date")
	}
	return nil
}
