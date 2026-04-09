package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MauricioLopezC/ReservationSystem/config"
	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// type UsersService interface {
// 	FindByEmail(email string) (*model.User, error)
// 	Create(userDto CreateUserDto) (*model.User, error)
// }

type AuthService struct {
	db *gorm.DB
	// usersService UsersService
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

func checkPasswordHash(password string, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (as *AuthService) Login(loginDto LoginDto) (*LoginResponse, error) {
	// get user by email
	var user model.User
	if result := as.db.Where("email = ?", loginDto.Email).First(&user); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "Wrong email!")
		}
		return nil, result.Error
	}

	// check password
	if !checkPasswordHash(loginDto.Password, user.Password) {
		return nil, fmt.Errorf("%w: %s", common.ErrUnauthorized, "Wrong password!")
	}

	token := jwt.New(jwt.SigningMethodHS256)

	// claims are token payload
	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = user.Email
	claims["user_id"] = user.ID
	claims["role"] = user.Role
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte(config.Config("SECRET")))
	if err != nil {
		return nil, errors.New("error: try again later")
	}

	userResponse := UserResponse{
		ID:       user.ID,
		Username: user.Name,
		Email:    user.Email,
		Role:     user.Role,
	}

	return &LoginResponse{
		UserResponse: userResponse,
		Token:        t,
	}, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (as *AuthService) Register(userDto CreateUserDto) (*model.User, error) {
	// check if user already exists
	var existingUser model.User
	userResult := as.db.Where("email = ?", userDto.Email).First(&existingUser)
	if userResult.Error == nil {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "username, email or phone number already exists!")
	} else if !errors.Is(userResult.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("database error")
	}

	hash, err := hashPassword(userDto.Password)
	if err != nil {
		return nil, errors.New("couldn't hash password")
	}

	newUser := model.User{
		Name:     strings.ToLower(userDto.Username),
		Email:    userDto.Email,
		Password: hash,
		Phone:    userDto.Phone,
	}

	if result := as.db.Create(&newUser); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, fmt.Errorf("%w: %s", common.ErrConflict, "email, username or phone number already exists")
		}
		return nil, errors.New("error creating user")
	}

	return &newUser, nil
}
