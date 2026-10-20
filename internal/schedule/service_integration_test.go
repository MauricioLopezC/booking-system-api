package schedule

import (
	"testing"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/MauricioLopezC/ReservationSystem/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteReportsNotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	ss := NewScheduleService(db)

	employee := model.User{Name: "emp", Email: "emp@test.com", Phone: "1", Password: "x", Role: model.RoleEmployee.String()}
	require.NoError(t, db.Create(&employee).Error)
	created, err := ss.Create(employee.ID, CreateScheduleDto{WeekDay: 1, StartTime: "08:00", EndTime: "12:00"})
	require.NoError(t, err)

	_, err = ss.Delete(9999)
	assert.ErrorIs(t, err, common.ErrRecordNotFound)

	_, err = ss.DeleteByUserID(created.ID, employee.ID+1)
	assert.ErrorIs(t, err, common.ErrRecordNotFound, "an employee can't delete someone else's schedule")

	deleted, err := ss.DeleteByUserID(created.ID, employee.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, deleted.ID)
}
