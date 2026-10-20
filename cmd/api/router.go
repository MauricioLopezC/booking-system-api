package main

import (
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/auth"
	"github.com/MauricioLopezC/ReservationSystem/internal/bookings"
	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/schedule"
	scheduleexceptions "github.com/MauricioLopezC/ReservationSystem/internal/schedule_exceptions"
	"github.com/MauricioLopezC/ReservationSystem/internal/services"
	"github.com/MauricioLopezC/ReservationSystem/internal/slots"
	"github.com/MauricioLopezC/ReservationSystem/internal/user"
	swaggo "github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func SetupRoutes(
	app *fiber.App,
	schedulesH *schedule.ScheduleHandler,
	bookingsH *bookings.BookingsHandler,
	slotsH *slots.SlotsHandler,
	servicesH *services.ServicesHandler,
	authH *auth.AuthHandler,
	schedulesExH *scheduleexceptions.ScheduleExHandler,
	usersH *user.UsersHandler,
) {
	generalLimiter := limiter.New(limiter.Config{
		Max:        60,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusTooManyRequests,
				Message:    "Too many requests, please slow down.",
			})
		},
	})

	authLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(common.ErrorResponse{
				StatusCode: fiber.StatusTooManyRequests,
				Message:    "Too many login attempts. Please wait before trying again.",
			})
		},
	})

	// Swagger UI (no auth, no rate limit)
	app.Get("/swagger/*", swaggo.HandlerDefault)

	// Middleware
	api := app.Group("/api", logger.New(), generalLimiter)

	// Auth
	authGroup := api.Group("/auth", authLimiter)
	authGroup.Post("/login", authH.Login)
	authGroup.Post("/register", authH.Register)

	// --- ME (User Context) ---
	me := api.Group("/me", auth.Protected())

	// Me - Bookings
	me.Post("/bookings", bookingsH.CreateMyBooking)
	me.Get("/bookings", bookingsH.FindMyBookings)
	me.Post("/bookings/:id/cancel", bookingsH.CancelBooking)

	// Me - Schedules
	me.Post("/schedules", auth.RoleMiddleware("employee", "admin"), schedulesH.CreateMySchedule)
	me.Get("/schedules", auth.RoleMiddleware("admin", "employee"), schedulesH.FindMySchedules)
	me.Delete("/schedules/:id", auth.RoleMiddleware("admin", "employee"), schedulesH.DeleteMySchedules)

	// Me - Schedule Exceptions
	me.Post("/schedule-exceptions", auth.RoleMiddleware("employee", "admin"), schedulesExH.CreateMyScheduleEx)
	me.Delete("/schedule-exceptions/:id", auth.RoleMiddleware("employee", "admin"), schedulesExH.DeleteMyScheduleEx)

	// Me - Services
	me.Post("/services", auth.RoleMiddleware("admin", "employee"), usersH.AssignServiceToMe)

	// --- PUBLIC / SHARED ---

	// Available Slots
	api.Get("/available-slots", slotsH.FindAvailableSlots)

	// Services
	services := api.Group("/services")
	services.Get("/", servicesH.FindAll)
	// Services - Professionals
	services.Get("/:id/professionals", usersH.FindEmployeesByService)

	// Services - Protected
	servicesProtected := services.Group("/", auth.Protected(), auth.RoleMiddleware("admin"))
	servicesProtected.Post("/", servicesH.Create)
	servicesProtected.Patch("/:id", servicesH.Update)
	servicesProtected.Delete("/:id", servicesH.Delete)

	// --- ADMIN ---

	// Bookings - Admin
	bookings := api.Group("/bookings", auth.Protected(), auth.RoleMiddleware("admin"))
	bookings.Get("/", bookingsH.FindAll)
	bookings.Get("/:id", bookingsH.FindOne)

	// Schedules - Admin
	schedules := api.Group("/schedules", auth.Protected(), auth.RoleMiddleware("admin"))
	schedules.Get("/", schedulesH.FindAll)
	schedules.Get("/:id", schedulesH.FindOne)

	// Users - Admin
	users := api.Group("/users", auth.Protected(), auth.RoleMiddleware("admin"))
	users.Get("/", usersH.FindAll)
	users.Get("/:id", usersH.FindOne)
	users.Patch("/:id", usersH.Update)
	users.Delete("/:id", usersH.Delete)
}
