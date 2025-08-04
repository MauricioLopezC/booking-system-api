package main

import (
	"github.com/MauricioLopezC/ReservationSystem/internal/database"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
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
	SetupRoutes(app)

	app.Listen(":3000")
}
