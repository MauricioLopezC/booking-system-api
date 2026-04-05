package main

import (
	"github.com/MauricioLopezC/ReservationSystem/config"
	"github.com/MauricioLopezC/ReservationSystem/internal/auth"
	"github.com/MauricioLopezC/ReservationSystem/internal/bookings"
	"github.com/MauricioLopezC/ReservationSystem/internal/database"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/MauricioLopezC/ReservationSystem/internal/schedule"
	scheduleexceptions "github.com/MauricioLopezC/ReservationSystem/internal/schedule_exceptions"
	"github.com/MauricioLopezC/ReservationSystem/internal/services"
	"github.com/MauricioLopezC/ReservationSystem/internal/slots"
	"github.com/MauricioLopezC/ReservationSystem/internal/user"
	"github.com/gofiber/fiber/v3"

	_ "github.com/MauricioLopezC/ReservationSystem/docs"
)

// @title           Reservation System API
// @version         1.0
// @description     API for managing reservations, schedules, and services.

// @host      localhost:3000
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	config.Load()
	app := fiber.New()
	database.ConnectDB()
	database.Migrate(
		&model.User{},
		&model.Service{},
		&model.Schedule{},
		&model.Booking{},
		&model.ScheduleOverride{},
		&model.TimeSlot{},
	)

	// handles and servicies
	scheduleService := schedule.NewScheduleService(database.DB.Db)
	scheduleHandler := schedule.NewScheduleHandler(scheduleService)

	slotsService := slots.NewSlotsService(database.DB.Db)
	slotsHandler := slots.NewSlotsHandler(slotsService)

	bookingsService := bookings.NewBookingsService(database.DB.Db, slotsService)
	bookingsHandler := bookings.NewBookingsHandler(bookingsService)

	servicesService := services.NewServicesService(database.DB.Db)
	servicesHandler := services.NewServicesHandler(servicesService)

	authService := auth.NewAuthService(database.DB.Db)
	authHandler := auth.NewAuthHandler(authService)

	scheduleExService := scheduleexceptions.NewScheduleExService(database.DB.Db)
	scheduleExHandler := scheduleexceptions.NewScheduleExHandler(scheduleExService)

	usersService := user.NewUsersService(database.DB.Db)
	usersHandler := user.NewUsersHandler(usersService)

	SetupRoutes(
		app,
		scheduleHandler,
		bookingsHandler,
		slotsHandler,
		servicesHandler,
		authHandler,
		scheduleExHandler,
		usersHandler,
	)

	app.Listen(":3000")
}
