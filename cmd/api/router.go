package main

import (
	"github.com/MauricioLopezC/ReservationSystem/internal/auth"
	"github.com/MauricioLopezC/ReservationSystem/internal/bookings"
	"github.com/MauricioLopezC/ReservationSystem/internal/schedule"
	scheduleexceptions "github.com/MauricioLopezC/ReservationSystem/internal/schedule_exceptions"
	"github.com/MauricioLopezC/ReservationSystem/internal/services"
	"github.com/MauricioLopezC/ReservationSystem/internal/slots"
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
) {
	//middleware
	api := app.Group("/api", logger.New())

	//auth
	authGroup := api.Group("/auth")
	authGroup.Post("/login", authH.Login)
	authGroup.Post("/register", authH.Register)

	//

	//Bookings
	bookings := api.Group("/bookings")
	bookings.Post("/", auth.Protected(), auth.RoleMiddleware("employee", "admin"), bookingsH.Create)
	bookings.Get("/", auth.Protected(), auth.RoleMiddleware("admin"), bookingsH.FindAll)
	bookings.Post("/:id/cancel", auth.Protected(), bookingsH.CancelBooking)
	//cancel, reasign
	//Bookings - User
	userBookings := api.Group("/me/bookings")
	userBookings.Get("/", auth.Protected(), bookingsH.FindBookingsByUser)

	//Shedules - ADMIN
	schedules := api.Group("/schedules")
	schedules.Get("/", auth.Protected(), auth.RoleMiddleware("admin"), schedulesH.FindAll) //Admin get all employee schedules
	schedules.Get("/:id", auth.Protected(), auth.RoleMiddleware("admin"), schedulesH.FindOne)

	//Shedules - user
	userSchedules := api.Group("/me/schedules")
	userSchedules.Post("/", auth.Protected(), auth.RoleMiddleware("employee, admin"), schedulesH.CreateCurrentUserSchedule)
	userSchedules.Get("/", auth.Protected(), auth.RoleMiddleware("admin", "employee"), schedulesH.CurrentUserFindAll)
	userSchedules.Delete("/:id", auth.Protected(), auth.RoleMiddleware("admin", "employee"), schedulesH.CurrentUserDelete)

	//available slots
	api.Get("/available-slots", slotsH.FindAvailableSlots) //Customer find all available slots for employee and service

	//Services
	services := api.Group("/services")
	services.Post("/", auth.Protected(), auth.RoleMiddleware("admin", "employee"), servicesH.Create)
	services.Get("/", servicesH.FindAll)
	services.Patch("/:id", auth.Protected(), auth.RoleMiddleware("admin", "employee"), servicesH.Update)
	services.Delete("/:id", auth.Protected(), auth.RoleMiddleware("admin", "employee"), servicesH.Delete)

	//SchedulesEx - ADMIN
	//SchedulesEx - USER
	userScheduleEx := api.Group("me/scheduleex")
	userScheduleEx.Post("/", auth.Protected(), auth.RoleMiddleware("employee", "admin"), schedulesExH.CurrentUserCreate)
	userScheduleEx.Delete("/:id", auth.Protected(), schedulesExH.CurrentUserDelete)

}
