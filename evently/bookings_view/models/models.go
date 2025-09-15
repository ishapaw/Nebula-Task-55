package models

import (
	"time"
)

type Booking struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	RequestID string    `gorm:"type:varchar(255);not null" json:"requestId"`
	UserID    string    `gorm:"type:varchar(255);not null" json:"userId"`
	EventID   string    `gorm:"type:varchar(255);not null" json:"eventId"`
	Price     float64   `gorm:"type:numeric;not null" json:"price"`
	Seats     int64     `gorm:"not null" json:"seats"`
	Status    string    `gorm:"type:varchar(50);not null" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

type BookingsCount struct {
	Total     int64 `json:"total"`
	Confirmed int64 `json:"confirmed"`
	Cancelled int64 `json:"cancelled"`
}

type DailyBookingStats struct {
	Date           string `json:"date"` 
	EventID        string `json:"event_id,omitempty"`
	ConfirmedCount int64  `json:"confirmed_count"`
	CancelledCount int64  `json:"cancelled_count"`
}
