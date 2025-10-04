package main

import (
	"github.com/MauricioLopezC/ReservationSystem/internal/auth"
	"github.com/MauricioLopezC/ReservationSystem/internal/bookings"
	"github.com/MauricioLopezC/ReservationSystem/internal/schedule"
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
) {
	//middleware
	api := app.Group("/api", logger.New())

	//auth
	authGroup := api.Group("/auth")
	authGroup.Post("/login", authH.Login)
	authGroup.Post("/register", authH.Register)

	//Bookings
	booking := api.Group("/bookings")
	booking.Post("/", auth.Protected(), auth.RoleMiddleware("employee", "admin"), bookingsH.Create)
	booking.Get("/", auth.Protected(), auth.RoleMiddleware("admin"), bookingsH.FindAll)
	booking.Post("/:id/cancel", auth.Protected(), bookingsH.CancelBooking)
	//cancel, reasign

	//Shedules
	scheduleGroup := api.Group("/schedules")
	scheduleGroup.Post("/", auth.Protected(), auth.RoleMiddleware("employee", "admin"), schedulesH.CreateSchedule)
	scheduleGroup.Get("/", auth.Protected(), auth.RoleMiddleware("admin"), schedulesH.FindAllSchedules) //Admin get all employee schedules
	scheduleGroup.Get("/:id", schedulesH.FindOneSchedule)
	scheduleGroup.Delete("/:id", auth.Protected(), auth.RoleMiddleware("employee", "admin"), schedulesH.DeleteSchedule)

	//available slots
	api.Get("/available-slots", slotsH.FindAvailableSlots) //Customer find all available slots for employee and service

	//current user specific data
	meGroup := api.Group("/me")
	meGroup.Get("/bookings", auth.Protected(), bookingsH.FindBookingsByUser)
	meGroup.Get("/schedules", auth.Protected(), schedulesH.FindSchedulesByEmployee)

	//Services
	serviceGroup := api.Group("/services")
	serviceGroup.Post("/", auth.Protected(), auth.RoleMiddleware("admin", "employee"), servicesH.Create)
	serviceGroup.Get("/", servicesH.FindAll)
	serviceGroup.Patch("/:id", servicesH.Update)
	serviceGroup.Delete("/:id", servicesH.Delete)

}
