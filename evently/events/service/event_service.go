package service

import (
	"context"
	"encoding/json"
	"errors"
	"events/models"
	"events/repository"
	"strconv"

	"github.com/redis/go-redis/v9"

	"fmt"
	"strings"
	"time"
)

type EventService interface {
	CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error)
	GetEventByID(ctx context.Context, id string) (*models.Event, error)
	GetAllEvents(page, limit int64) ([]models.Event, error)
	GetAllUpcomingEvents(ctx context.Context, page, limit int64) ([]models.UpcomingEvent, error)
	UpdateEvent(ctx context.Context, id string, updates map[string]interface{}) (*models.Event, error)
	GetCapacityUtilization(ctx context.Context, eventID string, page, limit int64) ([]models.CapacityUtilization, error)
	GetMostBookedEvents(ctx context.Context, limit int64) ([]models.MostBookedEvent, error)
	GetMostPopularEvents(ctx context.Context, limit int64) ([]models.MostPopularEvent, error)
	DeleteEvent(id string) error
}

type eventService struct {
	repo       repository.EventRepository
	redis      *redis.Client
	redisSeats *redis.Client
	redisPrice *redis.Client
}

func NewEventService(r repository.EventRepository, redisClient *redis.Client, redisSeats *redis.Client, redisPrice *redis.Client) EventService {
	return &eventService{
		repo:       r,
		redis:      redisClient,
		redisSeats: redisSeats,
		redisPrice: redisPrice,
	}
}

func (s *eventService) CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error) {

	err := validate(event)
	if err != nil {
		return nil, err
	}

	createdEvent, err1 := s.repo.Create(event)
	if err1 != nil {
		return nil, err1
	}

	keys, _ := s.redis.Keys(ctx, "events:upcoming:*").Result()
	if len(keys) > 0 {
		s.redis.Del(ctx, keys...)
	}

	seatsKey := "seatsLeft:" + createdEvent.ID.Hex()
	priceKey := "price:" + createdEvent.ID.Hex()
	s.redisSeats.Set(ctx, seatsKey, createdEvent.AvailableSeats, 0)
	s.redisPrice.Set(ctx, priceKey, createdEvent.Price, 0)

	return createdEvent, nil
}

func (s *eventService) getEventFromCache(ctx context.Context, id string) (*models.Event, error) {
	cacheKey := "event:" + id

	val, err := s.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}

	var ev models.Event

	if jsonErr := json.Unmarshal([]byte(val), &ev); jsonErr != nil {
		return nil, jsonErr
	}

	seatKey := "seatsLeft:" + id

	availableSeatsStr, err := s.redisSeats.Get(ctx, seatKey).Result()
	if err != nil {
		seats, err1 := s.repo.FindAvailableSeatsForIds([]string{id})
		if err1 != nil {
			return nil, err1
		} else {
			ev.AvailableSeats = seats[id]
		}
	} else {
		availableSeats, _ := strconv.Atoi(availableSeatsStr)
		ev.AvailableSeats= int64(availableSeats)
	}

	s.redis.Expire(ctx, cacheKey, 10*time.Minute)
	return &ev, nil

}

func (s *eventService) GetEventByID(ctx context.Context, id string) (*models.Event, error) {
	cacheKey := "event:" + id

	cachedEvent, err := s.getEventFromCache(ctx, id)
	if err == nil {
		return cachedEvent, nil
	}

	event, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(event); err == nil {
		s.redis.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return event, nil
}

func (s *eventService) GetAllEvents(page, limit int64) ([]models.Event, error) {
	return s.repo.FindAll(page, limit)
}

func (s *eventService) getUpcomingEventsFromCache(ctx context.Context, cacheKey string) ([]models.UpcomingEvent, error) {

	val, err := s.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}

	var events []models.UpcomingEvent
	if jsonErr := json.Unmarshal([]byte(val), &events); jsonErr != nil {
		return nil, jsonErr
	}

	ids := make([]string, len(events))
	for i, ev := range events {
		ids[i] = "seatsLeft:" + ev.ID.Hex()
	}

	vals, err1 := s.redisSeats.MGet(ctx, ids...).Result()
	if err1 == nil {

		seatMap := make(map[string]int)
		for i, val := range vals {
			if val != nil {
				seatMap[ids[i]], _ = strconv.Atoi(val.(string))
			}
		}

		for i, ev := range events {
			key := "seatsLeft:" + ev.ID.Hex()
			if seats, ok := seatMap[key]; ok {
				events[i].AvailableSeats = int64(seats)
			}
		}

	} else {

		availMap, err := s.repo.FindAvailableSeatsForIds(ids)
		if err != nil {
			return nil, err
		}

		for i, ev := range events {
			key := "seatsLeft:" + ev.ID.Hex()
			if avail, ok := availMap[key]; ok {
				events[i].AvailableSeats = avail
			}
		}

	}

	s.redis.Expire(ctx, cacheKey, 5*time.Minute)
	return events, nil
}

func (s *eventService) GetAllUpcomingEvents(ctx context.Context, page, limit int64) ([]models.UpcomingEvent, error) {
	today := time.Now().Format("2006-01-02")
	cacheKey := fmt.Sprintf("events:upcoming:%s:page=%d:limit=%d", today, page, limit)

	cachedEvents, err := s.getUpcomingEventsFromCache(ctx, cacheKey)
	if err == nil {
		return cachedEvents, nil
	}

	events, err := s.repo.FindAllUpcomingEvents(page, limit)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(events)
	s.redis.Set(ctx, cacheKey, data, 5*time.Minute)

	return events, nil
}

func (s *eventService) updateCache(ctx context.Context, id string, updates map[string]interface{}) {
	_, ok := updates["available_seats"]

	if len(updates) > 1 || !ok {
		s.redis.Del(ctx, "event:"+id)
	}

	upcomingFields := map[string]bool{
		"title":         true,
		"venue":         true,
		"date":          true,
		"total_seats": true,
		"price":         true,
	}

	if _, exists := updates["price"]; exists {
		priceKey := "price:" + id
        s.redisPrice.Set(ctx, priceKey, updates["price"], 0).Err()
    }

	shouldInvalidateUpcoming := false
	for field := range updates {
		if upcomingFields[field] {
			shouldInvalidateUpcoming = true
			break
		}
	}

	if shouldInvalidateUpcoming {
		keys, _ := s.redis.Keys(ctx, "events:upcoming:*").Result()
		if len(keys) > 0 {
			s.redis.Del(ctx, keys...)
		}
	}
}

func (s *eventService) UpdateEvent(ctx context.Context, id string, updates map[string]interface{}) (*models.Event, error) {

	if err := validateUpdates(updates); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateFields(id, updates); err != nil {
		return nil, err
	}

	updatedEvent, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	s.updateCache(ctx, id, updates)

	return updatedEvent, nil
}

func (s *eventService) GetCapacityUtilization(ctx context.Context, eventID string, page, limit int64) ([]models.CapacityUtilization, error) {
	return s.repo.GetCapacityUtilization(ctx, eventID, page, limit)
}

func (s *eventService) GetMostBookedEvents(ctx context.Context, limit int64) ([]models.MostBookedEvent, error) {
	return s.repo.GetMostBookedEvents(ctx, limit)
}

func (s *eventService) GetMostPopularEvents(ctx context.Context, limit int64) ([]models.MostPopularEvent, error) {
	return s.repo.GetMostPopularEvents(ctx, limit)
}

func (s *eventService) DeleteEvent(id string) error {
	return s.repo.Delete(id)
}

func validate(e *models.Event) error {

	if strings.TrimSpace(e.Title) == "" {
		return errors.New("title is required")
	}

	if strings.TrimSpace(e.Venue) == "" {
		return errors.New("venue is required")
	}

	if e.Date.IsZero() || e.Date.Before(time.Now()) {
		return errors.New("date is required and must be in the future")
	}

	if e.Price <= 0 {
		return errors.New("price must be greater than 0")
	}

	if e.TotalSeats == 0 || e.AvailableSeats == 0{
		return errors.New("total and available seats must be greater than 0")
	}

	if e.TotalSeats < e.AvailableSeats {
		return errors.New("total seats must be >= available seats")
	}

	return nil
}

func validateUpdates(updates map[string]interface{}) error {

	for key, value := range updates {

		switch key {

		case "title", "venue":
			strVal, ok := value.(string)
			if !ok || strings.TrimSpace(strVal) == "" {
				return fmt.Errorf("%s must be a non-empty string", key)
			}

		case "date":
			dateStr, ok := value.(string)
			if !ok {
				return fmt.Errorf("date must be a string (ISO format)")
			}

			parsed, err := time.Parse(time.RFC3339, dateStr)
			if err != nil {
				return fmt.Errorf("invalid date format, must be RFC3339")
			}

			if parsed.Before(time.Now()) {
				return fmt.Errorf("date must be in the future")
			}

			updates[key] = parsed

		case "price":
			price, ok := value.(float64)
			if !ok || price <= 0 {
				return fmt.Errorf("price must be a positive number")
			}

		case "total_seats":
			seats, ok := value.(float64)
			if !ok || seats <= 0 {
				return fmt.Errorf("total_seats must be greater than 0")
			}

			updates[key] = int(seats)
		}
	}
	return nil
}

