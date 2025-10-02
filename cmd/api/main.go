package main

import (
	"github.com/MauricioLopezC/ReservationSystem/internal/bookings"
	"github.com/MauricioLopezC/ReservationSystem/internal/database"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/MauricioLopezC/ReservationSystem/internal/schedule"
	"github.com/MauricioLopezC/ReservationSystem/internal/services"
	"github.com/MauricioLopezC/ReservationSystem/internal/slots"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	database.ConnectDb()
	database.Migrate(
		&model.User{},
		&model.Service{},
		&model.Schedule{},
		&model.Booking{},
		&model.ScheduleOverride{},
		&model.TimeSlot{},
	)

	//handles and servicies
	scheduleService := schedule.NewScheduleService(database.DB.Db)
	scheduleHandler := schedule.NewScheduleHandler(scheduleService)

	slotsService := slots.NewSlotsService(database.DB.Db)
	slotsHandler := slots.NewSlotsHandler(slotsService)

	bookingsService := bookings.NewBookingsService(database.DB.Db, slotsService)
	bookingsHandler := bookings.NewBookingsHandler(bookingsService)

	servicesService := services.NewServicesService(database.DB.Db)
	servicesHandler := services.NewServicesHandler(servicesService)

	SetupRoutes(app, scheduleHandler, bookingsHandler, slotsHandler, servicesHandler)

	app.Listen(":3000")
}
