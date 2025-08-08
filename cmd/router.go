package main

import (
	"github.com/MauricioLopezC/ReservationSystem/internal/auth"
	"github.com/MauricioLopezC/ReservationSystem/internal/bookings"
	"github.com/MauricioLopezC/ReservationSystem/internal/schedule"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func SetupRoutes(app *fiber.App) {
	//middleware
	api := app.Group("/api", logger.New())

	//auth
	authGroup := api.Group("/auth")
	authGroup.Post("/login", auth.Login)
	authGroup.Post("/register", auth.Register)

	//Bookings
	booking := api.Group("/booking")
	booking.Post("/", auth.Protected(), auth.RoleMiddleware("employee", "admin"), bookings.CreateBooking)
	booking.Get("/", auth.Protected(), auth.RoleMiddleware("admin"), bookings.FindAllBookings)
	//cancel, reasign

	//Shedules
	scheduleGroup := api.Group("/schedule")
	scheduleGroup.Post("/", auth.Protected(), auth.RoleMiddleware("employee", "admin"), schedule.CreateSchedule)
	scheduleGroup.Get("/", auth.Protected(), auth.RoleMiddleware("admin"), schedule.FindAllSchedules) //Admin get all employee schedules
	scheduleGroup.Get("/:id", schedule.FindOneSchedule)
	scheduleGroup.Delete("/:id", auth.Protected(), auth.RoleMiddleware("employee", "admin"), schedule.DeleteSchedule)

	//available slots
	api.Get("/available-slots", schedule.FindAvailableSlots) //Customer find all available slots for employee and service

	//current user specific data
	meGroup := api.Group("/me")
	meGroup.Get("/bookings", auth.Protected(), bookings.FindBookingsByUser)
	meGroup.Get("/schedules", auth.Protected(), schedule.FindSchedulesByEmployee)
}
