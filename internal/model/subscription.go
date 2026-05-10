package model

import (
	"time"

	"github.com/google/uuid"
)

type Month time.Time

const MonthLayout = "01-2006"

func ParseMonth(value string) (time.Time, error) {
	return time.Parse(MonthLayout, value)
}

func FormatMonth(value time.Time) string {
	return value.Format(MonthLayout)
}

type Subscription struct {
	ID          int64      `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   time.Time  `json:"-"`
	EndDate     *time.Time `json:"-"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type SubscriptionResponse struct {
	ID          int64   `json:"id"`
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"12-2025"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func ToResponse(s Subscription) SubscriptionResponse {
	var end *string
	if s.EndDate != nil {
		v := FormatMonth(*s.EndDate)
		end = &v
	}
	return SubscriptionResponse{
		ID:          s.ID,
		ServiceName: s.ServiceName,
		Price:       s.Price,
		UserID:      s.UserID.String(),
		StartDate:   FormatMonth(s.StartDate),
		EndDate:     end,
		CreatedAt:   s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   s.UpdatedAt.Format(time.RFC3339),
	}
}

type CreateSubscriptionRequest struct {
	ServiceName string `json:"service_name" example:"Yandex Plus"`
	Price       int    `json:"price" example:"400"`
	UserID      string `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string `json:"start_date" example:"07-2025"`
	EndDate     string `json:"end_date,omitempty" example:"12-2025"`
}

type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name,omitempty"`
	Price       *int    `json:"price,omitempty"`
	UserID      *string `json:"user_id,omitempty"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
}

type ListFilter struct {
	Limit       int
	Offset      int
	UserID      *uuid.UUID
	ServiceName *string
}

type SumFilter struct {
	From        time.Time
	To          time.Time
	UserID      *uuid.UUID
	ServiceName *string
}

type SumResponse struct {
	Total       int    `json:"total" example:"1200"`
	From        string `json:"from" example:"07-2025"`
	To          string `json:"to" example:"09-2025"`
	UserID      string `json:"user_id,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"invalid request"`
}
