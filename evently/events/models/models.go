package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Event struct {
	ID             primitive.ObjectID     `bson:"_id,omitempty" json:"_id"`
	Title          string                 `bson:"title" json:"title"`
	Description    string                 `bson:"description,omitempty" json:"description"`
	Venue          string                 `bson:"venue" json:"venue"`
	Date           time.Time              `bson:"date" json:"date"`
	Price          float64                `bson:"price" json:"price"`
	AvailableSeats int64                  `bson:"available_seats" json:"available_seats"`
	TotalSeats     int64                  `bson:"total_seats" json:"total_seats"`
	Metadata       map[string]interface{} `bson:",inline"`
	CreatedAt      time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time              `bson:"updated_at" json:"updated_at"`
}

type UpcomingEvent struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Title          string             `bson:"title" json:"title"`
	Venue          string             `bson:"venue" json:"venue"`
	Date           time.Time          `bson:"date" json:"date"`
	AvailableSeats int64              `bson:"available_seats" json:"available_seats"`
	TotalSeats     int64              `bson:"total_seats" json:"total_seats"`
}

type MostBookedEvent struct {
    EventID     string `bson:"event_id" json:"event_id"`
    Name        string `bson:"title" json:"title"`
    BookedSeats int64  `bson:"booked_seats" json:"booked_seats"`
}

type MostPopularEvent struct {
    EventID       string  `bson:"event_id" json:"event_id"`
    Name          string  `bson:"title" json:"title"`
    BookedSeats   int64   `bson:"booked_seats" json:"booked_seats"`
    TotalSeats    int64   `bson:"total_seats" json:"total_seats"`
    OccupancyRate float64 `bson:"occupancy_rate" json:"occupancy_rate"`
}

type CapacityUtilization struct {
    EventID         string  `bson:"event_id" json:"event_id"`
    Title           string  `bson:"title" json:"title"`
    CapacityUtilization float64 `bson:"capacity_utilisation" json:"capacity_utilisation"`
}

