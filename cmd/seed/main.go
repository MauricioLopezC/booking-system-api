package main

import (
	"flag"
	"log"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/config"
	"github.com/MauricioLopezC/ReservationSystem/internal/database"
	"github.com/MauricioLopezC/ReservationSystem/internal/database/seed"
)

func main() {
	reset := flag.Bool("reset", false, "truncate all tables before seeding")
	flag.Parse()

	time.Local = time.UTC

	config.Load()
	database.ConnectDB()

	db := database.DB.Db

	if *reset {
		if err := seed.Reset(db); err != nil {
			log.Fatalf("seed: %v", err)
		}
	}

	services, err := seed.SeedServices(db)
	if err != nil {
		log.Fatalf("seed: %v", err)
	}

	users, err := seed.SeedUsers(db, services)
	if err != nil {
		log.Fatalf("seed: %v", err)
	}

	// users[0] = admin, users[1] = emp1, users[2] = emp2, users[3..4] = customers
	employees := users[1:3]
	customers := users[3:5]

	if err := seed.SeedSchedules(db, employees); err != nil {
		log.Fatalf("seed: %v", err)
	}

	businessLoc, err := config.BusinessLocation()
	if err != nil {
		log.Fatalf("seed: invalid BUSINESS_TIMEZONE: %v", err)
	}

	if err := seed.SeedBookings(db, customers, employees, services, businessLoc); err != nil {
		log.Fatalf("seed: %v", err)
	}

	log.Println("seed: done!")
}
