package models

import(
	"time"
	"bookings_consumer/kafka" 
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
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

type KafkaEvent struct {
	RequestID string `json:"request_id"`
	EventID   string `json:"event_id"`
	Seats     int64  `json:"seats"`
	UserID    string `json:"user_id"`
	Price float64 `json:"price"`
	State     string `json:"state"`
}

type KafkaUpdateEvent struct {
	EventId   string `json:"event_id"`
	Seats     int64  `json:"seats"`
	Operation string `json:"operation"`
}

type ProcessorDeps struct {
    RedisReq   *redis.Client
    RedisSeats *redis.Client
    RedisPrice *redis.Client
    DB         *gorm.DB
    Producer   *kafka.Producer
}
