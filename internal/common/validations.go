package common

import (
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"gorm.io/gorm"
)

func EmployeeHasService(db *gorm.DB, employeeID, serviceID uint) (bool, error) {
	var count int64
	result := db.Table("user_services").Where("user_id = ? AND service_id = ?", employeeID, serviceID).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

func UserHasBooking(db *gorm.DB, userID, bookingID uint, userRole string) (bool, error) {
	var count int64

	if userRole == model.RoleAdmin.String() || userRole == model.RoleEmployee.String() {
		result := db.Table("bookings").Where("id = ? AND employee_id = ?", bookingID, userID).Count(&count)
		if result.Error != nil {
			return false, result.Error
		}
	}

	if userRole == model.RoleCustomer.String() {
		result := db.Table("bookings").Where("id = ? AND customer_id = ?", bookingID, userID).Count(&count)
		if result.Error != nil {
			return false, result.Error
		}
	}

	return count > 0, nil
}
