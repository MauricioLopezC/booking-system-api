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
	booking.Post("/", auth.Protected(), bookings.CreateBooking)

	//Shedules
	scheduleGroup := api.Group("/schedule")
	scheduleGroup.Post("/", auth.Protected(), schedule.CreateSchedule)
	scheduleGroup.Get("/", schedule.FindAllSchedules)
}
