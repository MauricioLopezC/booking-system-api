package common

import (
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"gorm.io/gorm"
)

func EmployeeHasService(db *gorm.DB, employeeId, serviceId uint) (bool, error) {
	var count int64
	result := db.Table("user_services").Where("user_id = ? AND service_id = ?", employeeId, serviceId).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

func EmployeeHasSchedule(db *gorm.DB, employeeId, scheduleId uint) (bool, error) {
	var count int64
	result := db.Table("schedules").Where("id = ? AND employee_id = ?", scheduleId, employeeId).Count(&count)

	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}

func UserHasBooking(db *gorm.DB, userId, bookingId uint, userRole string) (bool, error) {
	var count int64

	if userRole == model.RoleAdmin.String() || userRole == model.RoleEmployee.String() {
		result := db.Table("bookings").Where("id = ? AND employee_id = ?", bookingId, userId).Count(&count)
		if result.Error != nil {
			return false, result.Error
		}
	}

	if userRole == model.RoleCustomer.String() {
		result := db.Table("bookings").Where("id = ? AND customer_id = ?", bookingId, userId).Count(&count)
		if result.Error != nil {
			return false, result.Error
		}
	}

	return count > 0, nil
}
