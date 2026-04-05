// Package bookings
package bookings

import (
	"errors"
	"fmt"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"github.com/gofiber/fiber/v3"
)

type BookingsHandler struct {
	bookingsService *BookingsService
}

func NewBookingsHandler(bookingsSerice *BookingsService) *BookingsHandler {
	return &BookingsHandler{bookingsService: bookingsSerice}
}

// CreateMyBooking godoc
// @Summary      Create a booking
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      BookingDto            true  "Booking data"
// @Success      201   {object}  model.BookingResponse
// @Failure      400   {object}  common.ErrorResponse
// @Failure      401   {object}  common.ErrorResponse
// @Failure      404   {object}  common.ErrorResponse
// @Failure      409   {object}  common.ErrorResponse
// @Router       /me/bookings [post]
func (h *BookingsHandler) CreateMyBooking(c fiber.Ctx) error {
	bookingDto := new(BookingDto)

	if err := c.Bind().Body(bookingDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{StatusCode: fiber.StatusBadRequest, Message: err.Error()})
	}

	if err := utils.Validate(bookingDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    err.Error(),
		})
	}

	authUser, err := utils.GetAuthUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusUnauthorized,
			Message:    "Invalid auth token",
		})
	}

	newBooking, err := h.bookingsService.CreateBooking(*bookingDto, authUser.ID)
	if err != nil {
		if errors.Is(err, common.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusNotFound,
				Message:    err.Error(),
			})
		}
		if errors.Is(err, common.ErrConflict) {
			return c.Status(fiber.StatusConflict).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusConflict,
				Message:    err.Error(),
			})
		}
		if errors.Is(err, common.ErrValidation) {
			return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusBadRequest,
				Message:    err.Error(),
			})
		}
		// unexpected error
		fmt.Print(err.Error())
		return common.NewInternalServerErrorResponse(c)
	}
	return c.Status(fiber.StatusCreated).JSON(newBooking)
}

// FindAll godoc
// @Summary      Get all bookings (admin)
// @Tags         bookings
// @Produce      json
// @Security     BearerAuth
// @Param        limit    query     int     true   "Items per page"
// @Param        page     query     int     true   "Page number"
// @Param        orderBy  query     string  false  "Sort field: booking_date | created_at"
// @Success      200  {object}  model.AllBookingsSwaggerResponse
// @Failure      400  {object}  common.ErrorResponse
// @Failure      401  {object}  common.ErrorResponse
// @Router       /bookings [get]
func (h *BookingsHandler) FindAll(c fiber.Ctx) error {
	limit := fiber.Query[int](c, "limit")
	if limit <= 0 {
		return c.Status(fiber.ErrBadRequest.Code).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message:    "Limit muste be an integer number and gt 0",
		})
	}

	page := fiber.Query[int](c, "page")
	if page <= 0 {
		return c.Status(fiber.ErrBadRequest.Code).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message:    "page must be an integer number and gt 0",
		})
	}

	orderBy := c.Query("orderBy")
	allowedFields := map[string]bool{
		"booking_date": true,
		"created_at":   true,
	}
	orderBy, _ = utils.ParseSortParam(orderBy, allowedFields)

	allBookingsResponse, err := h.bookingsService.FindAll(limit, page, orderBy)
	if err != nil {
		// unexpected error
		return common.NewInternalServerErrorResponse(c)
	}

	return c.Status(200).JSON(allBookingsResponse)
}

// FindOne godoc
// @Summary      Get booking by ID (admin)
// @Tags         bookings
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Booking ID"
// @Success      200  {object}  model.BookingResponse
// @Failure      400  {object}  common.ErrorResponse
// @Failure      401  {object}  common.ErrorResponse
// @Failure      404  {object}  common.ErrorResponse
// @Router       /bookings/{id} [get]
// FindOne ADMIN ONLY
func (h *BookingsHandler) FindOne(c fiber.Ctx) error {
	id := fiber.Params[int](c, "id")
	if id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Numeric string greater than 0 is expected",
		})
	}

	booking, err := h.bookingsService.FindByID(uint(id))
	if err != nil {
		if errors.Is(err, common.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusNotFound,
				Message:    err.Error(),
			})
		}
		// unexpected error
		return common.NewInternalServerErrorResponse(c)
	}
	return c.Status(fiber.StatusOK).JSON(booking)
}

// FindMyBookings godoc
// @Summary      Get current user bookings
// @Tags         bookings
// @Produce      json
// @Security     BearerAuth
// @Param        limit    query     int     false  "Items per page (default 10)"
// @Param        page     query     int     false  "Page number (default 1)"
// @Param        orderBy  query     string  false  "Sort field: booking_date | created_at"
// @Success      200  {object}  model.AllBookingsSwaggerResponse
// @Failure      400  {object}  common.ErrorResponse
// @Failure      401  {object}  common.ErrorResponse
// @Router       /me/bookings [get]
func (h *BookingsHandler) FindMyBookings(c fiber.Ctx) error {
	limit := fiber.Query(c, "limit", 10)
	if limit <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Limit muste be an integer number and gt 0",
		})
	}

	page := fiber.Query(c, "page", 1)
	if page <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "page must be an integer number and gt 0",
		})
	}

	orderBy := c.Query("orderBy")
	allowedFields := map[string]bool{
		"booking_date": true,
		"created_at":   true,
	}
	orderBy, _ = utils.ParseSortParam(orderBy, allowedFields)

	authUser, err := utils.GetAuthUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusUnauthorized,
			Message:    "Invalid auth token",
		})
	}

	allBookingsResponse, err := h.bookingsService.FindByUserID(limit, page, orderBy, authUser.ID, authUser.Role)
	if err != nil {
		if errors.Is(err, common.ErrUnauthorized) {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusUnauthorized,
				Message:    err.Error(),
			})
		}
		return common.NewInternalServerErrorResponse(c)
	}

	return c.Status(fiber.StatusOK).JSON(allBookingsResponse)
}

// CancelBooking godoc
// @Summary      Cancel a booking
// @Tags         bookings
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Booking ID"
// @Success      200  {object}  model.BookingResponse
// @Failure      400  {object}  common.ErrorResponse
// @Failure      401  {object}  common.ErrorResponse
// @Failure      404  {object}  common.ErrorResponse
// @Failure      409  {object}  common.ErrorResponse
// @Router       /me/bookings/{id}/cancel [post]
func (h *BookingsHandler) CancelBooking(c fiber.Ctx) error {
	// Get booking id
	id := fiber.Params[int](c, "id")
	if id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Numeric string greater than 0 is expected",
		})
	}
	// Get user id
	authUser, err := utils.GetAuthUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusUnauthorized,
			Message:    "Invalid auth token",
		})
	}

	booking, err := h.bookingsService.CancelBooking(id, authUser.ID, authUser.Role)
	if err != nil {
		if errors.Is(err, common.ErrUnauthorized) {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusUnauthorized,
				Message:    err.Error(),
			})
		}
		if errors.Is(err, common.ErrConflict) {
			return c.Status(fiber.StatusConflict).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusConflict,
				Message:    err.Error(),
			})
		}
		if errors.Is(err, common.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusNotFound,
				Message:    err.Error(),
			})
		}
		return common.NewInternalServerErrorResponse(c)
	}

	return c.Status(fiber.StatusOK).JSON(booking)
}
