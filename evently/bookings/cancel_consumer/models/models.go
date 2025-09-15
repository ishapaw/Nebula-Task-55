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

type KafkaCancelEvent struct {
	BookingRequestId string `json:"booking_request_id"`
	EventId   string `json:"event_id"`
	BookingId string `json:"booking_id"`
	Seats     int64  `json:"seats"`
}

type KafkaUpdateEvent struct {
	EventId   string `json:"event_id"`
	Seats     int64  `json:"seats"`
	Operation string `json:"operation"`
}
