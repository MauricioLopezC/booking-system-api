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

// UserHasBooking reports whether the user can act on the booking: admins on any booking,
// employees on the ones they attend and customers on the ones they made.
func UserHasBooking(db *gorm.DB, userID, bookingID uint, userRole string) (bool, error) {
	query := db.Model(&model.Booking{}).Where("id = ?", bookingID)

	switch userRole {
	case model.RoleAdmin.String():
	case model.RoleEmployee.String():
		query = query.Where("employee_id = ?", userID)
	case model.RoleCustomer.String():
		query = query.Where("customer_id = ?", userID)
	default:
		return false, nil
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// lockNamespaceEmployeeAgenda namespaces the advisory locks taken on an employee's agenda.
const lockNamespaceEmployeeAgenda = 1

// LockEmployeeAgenda serializes, until tx commits or rolls back, the operations that check
// an employee's agenda and then change it: creating a booking and creating a blocking
// schedule exception. Without it a booking could slip in between an exception's
// conflict check and its insert (or vice versa).
func LockEmployeeAgenda(tx *gorm.DB, employeeID uint) error {
	return tx.Exec("SELECT pg_advisory_xact_lock(?::int, ?::int)", lockNamespaceEmployeeAgenda, int32(employeeID)).Error
}
