package main

import (
	"github.com/MauricioLopezC/ReservationSystem/internal/auth"
	"github.com/MauricioLopezC/ReservationSystem/internal/bookings"
	"github.com/MauricioLopezC/ReservationSystem/internal/schedule"
	scheduleexceptions "github.com/MauricioLopezC/ReservationSystem/internal/schedule_exceptions"
	"github.com/MauricioLopezC/ReservationSystem/internal/services"
	"github.com/MauricioLopezC/ReservationSystem/internal/slots"
	"github.com/MauricioLopezC/ReservationSystem/internal/user"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
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
	// Middleware
	api := app.Group("/api", logger.New())

	// Auth
	authGroup := api.Group("/auth")
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
	// Services - Protected
	servicesProtected := services.Group("/", auth.Protected(), auth.RoleMiddleware("admin", "employee"))
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
}
