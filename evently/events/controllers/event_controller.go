package controllers

import (
	"events/models"
	"events/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EventController struct {
	service service.EventService
}

func NewEventController(s service.EventService) *EventController {
	return &EventController{service: s}
}

func (ec *EventController) CreateEvent(c *gin.Context) {
	ctx := c.Request.Context()

	var event models.Event

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdEvent, err := ec.service.CreateEvent(ctx, &event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "event created successfully", "data": createdEvent})
}

func (ec *EventController) GetEventByID(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event id is missing"})
		return
	}

	event, err := ec.service.GetEventByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

func (ec *EventController) GetAllEvents(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	events, err := ec.service.GetAllEvents(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

func (ec *EventController) GetAllUpcomingEvents(c *gin.Context) {
	ctx := c.Request.Context()

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)

	events, err := ec.service.GetAllUpcomingEvents(ctx, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch upcoming events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"upcoming_events": events})
}

func (ec *EventController) UpdateEvent(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event id is missing"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	updatedEvent, err := ec.service.UpdateEvent(ctx, id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "event updated successfully", "updatedEvent": updatedEvent})
}

func (ec *EventController) DeleteEvent(c *gin.Context) {

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event id is missing"})
		return
	}

	err := ec.service.DeleteEvent(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "event deleted successfully"})
}

func (ec *EventController) GetCapacityUtilization(c *gin.Context) {
	ctx := c.Request.Context()

	eventId := c.Query("event_id")
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "100"), 10, 64)

	analytics, err := ec.service.GetCapacityUtilization(ctx, eventId, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)

}

func (ec *EventController) GetMostBookedEvents(c *gin.Context) {
	ctx := c.Request.Context()

	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)


	analytics, err := ec.service.GetMostBookedEvents(ctx, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)

}

func (ec *EventController) GetMostPopularEvents(c *gin.Context) {
	ctx := c.Request.Context()

	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)


	analytics, err := ec.service.GetMostPopularEvents(ctx, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)

}