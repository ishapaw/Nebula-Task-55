package controllers

import (
	"bookings_view/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BookingsViewController struct {
	bookingsViewService service.BookingsViewService
}

func NewBookingsViewController(bookingsViewService service.BookingsViewService) *BookingsViewController {
	return &BookingsViewController{bookingsViewService: bookingsViewService}
}

func (c *BookingsViewController) GetBookingByID(ctx *gin.Context) {
	id := ctx.Param("id")

	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "booking id is required"})
		return
	}

	booking, err := c.bookingsViewService.GetBookingByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, booking)
}

func (c *BookingsViewController) GetBookingsByEventID(ctx *gin.Context) {
	eventID := ctx.Param("event_id")
	status := ctx.DefaultQuery("status", "all")

	if eventID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "eventID is required"})
		return
	}

	page, _ := strconv.ParseInt(ctx.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)

	bookings, err := c.bookingsViewService.GetBookingsByEventID(eventID, limit, page, status)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, bookings)
}

func (c *BookingsViewController) GetBookingsByUserID(ctx *gin.Context) {
	userID := ctx.Param("user_id")
	status := ctx.DefaultQuery("status", "all")

	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	page, _ := strconv.ParseInt(ctx.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)

	bookings, err := c.bookingsViewService.GetBookingsByUserID(userID, limit, page, status)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, bookings)
}

func (c *BookingsViewController) GetBookingByRequestID(ctx *gin.Context) {
	reqID := ctx.Param("request_id")

	if reqID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "reqID is required"})
		return
	}

	booking, err := c.bookingsViewService.GetBookingByRequestID(reqID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Booking not found"})
		return
	}

	ctx.JSON(http.StatusOK, booking)
}

func (c *BookingsViewController) GetTotalBookings(ctx *gin.Context) {

	bookingsCount, err := c.bookingsViewService.GetTotalBookings()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, bookingsCount)
}

func (c *BookingsViewController) GetDailyBookingStats(ctx *gin.Context) {
	eventID := ctx.Query("event_id")
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")

	if startDate == "" || endDate == ""{
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "startDate and endDate both required"})
		return
	}

	stats, err := c.bookingsViewService.GetDailyBookingStats(eventID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}
