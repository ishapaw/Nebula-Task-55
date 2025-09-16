package service

import (
	"combined/models"
	"combined/repository"
	"errors"
)

type BookingsViewService interface {
	GetAllBookings(page,limit int64) ([]models.Booking, error)
	GetBookingByID(id string) (*models.Booking, error)
	GetBookingsByEventID(eventID string, limit, page int64, status string) ([]models.Booking, error)
	GetBookingByRequestID(reqID string) (*models.Booking, error)
	GetBookingsByUserID(userID string, limit, page int64, status string) ([]models.Booking, error)
	GetTotalBookings() (*models.BookingsCount, error) 
	GetDailyBookingStats(eventID, startDate, endDate string) ([]models.DailyBookingStats, error)
}

type bookingsViewService struct {
	repo repository.BookingsViewRepository
}

func NewBookingsViewService(repo repository.BookingsViewRepository) BookingsViewService {
	return &bookingsViewService{repo: repo}
}

func (s *bookingsViewService) GetAllBookings(page, limit int64) ([]models.Booking, error) {
	return s.repo.GetAllBookings(page, limit)
}


func (s *bookingsViewService) GetBookingByID(id string) (*models.Booking, error) {
	booking, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if booking == nil {
		return nil, errors.New("booking not found")
	}

	return booking, nil
}

func (s *bookingsViewService) GetBookingByRequestID(reqID string) (*models.Booking, error) {
	return s.repo.GetBookingByRequestID(reqID)
}

func (s *bookingsViewService) GetBookingsByEventID(eventID string, limit, page int64, status string) ([]models.Booking, error) {
	return s.repo.GetByEventID(eventID, limit, page, status)
}

func (s *bookingsViewService) GetBookingsByUserID(userID string, limit, page int64, status string) ([]models.Booking, error) {
	return s.repo.GetByUserID(userID, limit, page, status)
}

func (s *bookingsViewService) GetTotalBookings() (*models.BookingsCount, error) {
	return s.repo.GetTotalBookings()
}

func (s *bookingsViewService) GetDailyBookingStats(eventID, startDate, endDate string) ([]models.DailyBookingStats, error) {
	return s.repo.GetDailyBookingStats(eventID, startDate, endDate)
}

