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

	//Bookings
	mybookings := api.Group("/me/bookings")
	mybookings.Post("/", auth.Protected(), bookingsH.CreateMyBooking)
	mybookings.Get("/", auth.Protected(), bookingsH.FindMyBookings)
	mybookings.Post("/:id/cancel", auth.Protected(), bookingsH.CancelBooking)
	//Bookings - ADMIN
	bookings := api.Group("/bookings")
	bookings.Get("/", auth.Protected(), auth.RoleMiddleware("admin"), bookingsH.FindAll)
	bookings.Get("/:id", auth.Protected(), auth.RoleMiddleware("admin"), bookingsH.FindOne)

	//Shedules - ADMIN
	schedules := api.Group("/schedules")
	schedules.Get("/", auth.Protected(), auth.RoleMiddleware("admin"), schedulesH.FindAll) //Admin get all employee schedules
	schedules.Get("/:id", auth.Protected(), auth.RoleMiddleware("admin"), schedulesH.FindOne)

	//Shedules - user
	mySchedules := api.Group("/me/schedules")
	mySchedules.Post("/", auth.Protected(), auth.RoleMiddleware("employee, admin"), schedulesH.CreateMySchedule)
	mySchedules.Get("/", auth.Protected(), auth.RoleMiddleware("admin", "employee"), schedulesH.FindMySchedules)
	mySchedules.Delete("/:id", auth.Protected(), auth.RoleMiddleware("admin", "employee"), schedulesH.DeleteMySchedules)

	//available slots
	api.Get("/available-slots", slotsH.FindAvailableSlots) //Customer find all available slots for employee and service

	//Services
	services := api.Group("/services")
	services.Post("/", auth.Protected(), auth.RoleMiddleware("admin", "employee"), servicesH.Create)
	services.Get("/", servicesH.FindAll)
	services.Patch("/:id", auth.Protected(), auth.RoleMiddleware("admin", "employee"), servicesH.Update)
	services.Delete("/:id", auth.Protected(), auth.RoleMiddleware("admin", "employee"), servicesH.Delete)

	//SchedulesEx - USER
	myScheduleEx := api.Group("me/scheduleex")
	myScheduleEx.Post("/", auth.Protected(), auth.RoleMiddleware("employee", "admin"), schedulesExH.CreateMyScheduleEx)
	myScheduleEx.Delete("/:id", auth.Protected(), schedulesExH.DeleteMyScheduleEx)

}
