package auth

import (
	"testing"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/internal/utils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// The dummy hash must cost as much as a real one, otherwise login timing reveals unknown emails.
func TestDummyHashMatchesPasswordCost(t *testing.T) {
	cost, err := bcrypt.Cost([]byte(dummyHash))
	assert.NoError(t, err)
	assert.Equal(t, utils.PasswordCost, cost)

	start := time.Now()
	assert.False(t, utils.CheckPasswordHash("whatever", dummyHash))
	assert.Greater(t, time.Since(start), 50*time.Millisecond, "comparison must do real bcrypt work")
}
