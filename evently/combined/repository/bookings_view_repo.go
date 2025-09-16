package repository

import (
	"combined/models"
	"errors"

	"gorm.io/gorm"
)

type BookingsViewRepository interface {
	GetByID(id string) (*models.Booking, error)
	GetByEventID(eventID string, limit, page int64, status string) ([]models.Booking, error)
	GetByUserID(userID string, limit, page int64, status string) ([]models.Booking, error)
	GetBookingByRequestID(reqID string) (*models.Booking, error)
	GetTotalBookings() (*models.BookingsCount, error)
	GetDailyBookingStats(eventID, startDate, endDate string) ([]models.DailyBookingStats, error)
}

type bookingsViewRepository struct {
	db *gorm.DB
}

func NewBookingsViewRepository(db *gorm.DB) BookingsViewRepository {
	return &bookingsViewRepository{db}
}

func (r *bookingsViewRepository) GetByID(id string) (*models.Booking, error) {
	var booking models.Booking
	err := r.db.First(&booking, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("booking not found")
		}
		return nil, err
	}

	return &booking, nil
}

func (r *bookingsViewRepository) GetByEventID(eventID string, limit, page int64, status string) ([]models.Booking, error) {
	var bookings []models.Booking

	offset := int((page - 1) * limit)

	query := r.db.Where("event_id = ?", eventID)

	if status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.
		Limit(int(limit)).
		Offset(offset).
		Find(&bookings).Error; err != nil {
		return nil, err
	}
	return bookings, nil
}

func (r *bookingsViewRepository) GetByUserID(userID string, limit, page int64, status string) ([]models.Booking, error) {
	var bookings []models.Booking

	offset := int((page - 1) * limit)

	query := r.db.Where("user_id = ?", userID)

	if status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.
		Limit(int(limit)).
		Offset(offset).
		Find(&bookings).Error; err != nil {
		return nil, err
	}

	return bookings, nil
}

func (r *bookingsViewRepository) GetBookingByRequestID(reqID string) (*models.Booking, error) {
	var booking models.Booking

	err := r.db.Where("request_id = ?", reqID).First(&booking).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *bookingsViewRepository) GetTotalBookings() (*models.BookingsCount, error) {
	var count models.BookingsCount

	err := r.db.Model(&models.Booking{}).
		Select("COUNT(CASE WHEN status = 'confirmed' THEN 1 END) AS confirmed, COUNT(CASE WHEN status = 'cancelled' THEN 1 END) AS cancelled").
		Scan(&count).Error
	if err != nil {
		return nil, err
	}

	count.Total = count.Confirmed + count.Cancelled

	return &count, nil
}

func (r *bookingsViewRepository) GetDailyBookingStats(eventID, startDate, endDate string) ([]models.DailyBookingStats, error) {
	var results []models.DailyBookingStats

	query := r.db.Model(&models.Booking{}).
		Select("DATE(created_at) as date," +
			"COUNT(CASE WHEN status='confirmed' THEN 1 END) as confirmed_count, " +
			"COUNT(CASE WHEN status='cancelled' THEN 1 END) as cancelled_count").
		Group("DATE(created_at)").
		Order("DATE(created_at) ASC")

	if eventID != "" {
		query = query.Where("event_id = ?", eventID)
	}

	query = query.Where("created_at >= ? AND created_at <= ?", startDate, endDate)
	
	err := query.Scan(&results).Error
	return results, err
}

