package bookings

import (
	"errors"
	"fmt"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type BookingsHandler struct {
	bookingsService *BookingsService
}

func NewBookingsHandler(bookingsSerice *BookingsService) *BookingsHandler {
	return &BookingsHandler{bookingsService: bookingsSerice}
}

func (h *BookingsHandler) Create(c *fiber.Ctx) error {
	bookingDto := new(BookingDto)

	if err := c.BodyParser(bookingDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{StatusCode: fiber.StatusBadRequest, Message: err.Error()})
	}

	if err := utils.Validate(bookingDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    err.Error(),
		})
	}

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userId := uint(claims["user_id"].(float64))

	newBooking, err := h.bookingsService.CreateBooking(*bookingDto, userId)
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

func (h *BookingsHandler) FindAll(c *fiber.Ctx) error {
	limit := c.QueryInt("limit")
	if limit <= 0 {
		return c.Status(fiber.ErrBadRequest.Code).JSON(common.ErrorResponse{
			StatusCode: 400,
			Message:    "Limit muste be an integer number and gt 0",
		})
	}

	page := c.QueryInt("page")
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

	allBookingsResponse, err := h.bookingsService.FindAllBookings(limit, page, orderBy)
	if err != nil {
		// unexpected error
		return common.NewInternalServerErrorResponse(c)
	}

	return c.Status(200).JSON(allBookingsResponse)
}

func (h *BookingsHandler) FindOne(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Numeric string greater than 0 is expected",
		})
	}

	createdBooking, err := h.bookingsService.FindOneBooking(uint(id))
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
	return c.Status(fiber.StatusCreated).JSON(createdBooking)
}

// WARNING: check this handler
func (h *BookingsHandler) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Numeric string greater than 0 is expected",
		})
	}

	bookingDto := new(BookingDto)
	if err := c.BodyParser(bookingDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{StatusCode: fiber.StatusBadRequest, Message: err.Error()})
	}

	if err := utils.Validate(bookingDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    err.Error(),
		})
	}

	updatedBooking, err := h.bookingsService.Update(uint(id))

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
	}
	return c.Status(fiber.StatusOK).JSON(updatedBooking)
}

func (h *BookingsHandler) FindBookingsByUser(c *fiber.Ctx) error {
	limit := c.QueryInt("limit")
	if limit <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Limit muste be an integer number and gt 0",
		})
	}

	page := c.QueryInt("page")
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
	orderBy, err := utils.ParseSortParam(orderBy, allowedFields)

	//get user data from context
	userToken, ok1 := c.Locals("user").(*jwt.Token)
	claims, ok2 := userToken.Claims.(jwt.MapClaims)
	userId, ok3 := claims["user_id"].(float64)
	role, ok4 := claims["role"].(string)
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return common.NewInternalServerErrorResponse(c)
	}

	allBookingsResponse, err := h.bookingsService.FindBookingsByUser(limit, page, orderBy, uint(userId), role)
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

func (h *BookingsHandler) CancelBooking(c *fiber.Ctx) error {
	//Get booking id
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(common.ErrorResponse{
			StatusCode: fiber.StatusBadRequest,
			Message:    "Numeric string greater than 0 is expected",
		})
	}
	//Get user id
	userToken, ok1 := c.Locals("user").(*jwt.Token)
	claims, ok2 := userToken.Claims.(jwt.MapClaims)
	userId, ok3 := claims["user_id"].(float64)
	role, ok4 := claims["role"].(string)
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return common.NewInternalServerErrorResponse(c)
	}

	booking, err := h.bookingsService.CancelBooking(id, uint(userId), role)
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
