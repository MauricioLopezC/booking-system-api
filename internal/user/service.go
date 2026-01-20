package user

import (
	"errors"
	"fmt"

	"github.com/MauricioLopezC/ReservationSystem/internal/common"
	"github.com/MauricioLopezC/ReservationSystem/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// TODO: this serice implement actions like: CRUD, block user, connect, disconnect user(only employees) with services
type UsersService struct {
	db *gorm.DB
}

func NewUsersService(db *gorm.DB) *UsersService {
	return &UsersService{db: db}
}

func (us *UsersService) FindAll(limit, page int, orderBy string) (*AllUsersResponse, error) {
	offset := (page - 1) * limit
	query := us.db.Model(&model.User{})

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, err
	}

	if orderBy != "" {
		query = query.Order(orderBy)
	}

	users := []model.User{}
	if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, err
	}

	return &AllUsersResponse{
		Data:     users,
		Metadata: metadata{TotalItems: count},
	}, nil

}

func (us *UsersService) FindOne(id uint) (*model.User, error) {
	var user model.User
	if result := us.db.First(&user, id); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "user")
		}
		return nil, result.Error
	}
	return &user, nil
}

// func (s *UsersService) FindByEmail(email string) (*model.User, error) {
// 	var user model.User
// 	if result := s.db.Where("email = ?", email).First(&user); result.Error != nil {
// 		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
// 			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "user")
// 		}
// 		return nil, result.Error
// 	}
// 	return &user, nil
// }

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (us *UsersService) Update(id uint, userDto UpdateUserDto) (*model.User, error) {
	var user model.User
	if restult := us.db.First(&user, id); restult.Error != nil {
		if errors.Is(restult.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "user")
		}
		return nil, restult.Error
	}

	if userDto.Username != nil {
		user.Name = *userDto.Username
	}

	if userDto.Email != nil {
		user.Email = *userDto.Email
	}

	if userDto.Phone != nil {
		user.Phone = *userDto.Phone
	}

	if userDto.Password != nil {
		//hash passwrod first
		hash, err := hashPassword(*userDto.Password)
		if err != nil {
			return nil, err
		}
		user.Password = hash
	}

	if result := us.db.Save(&user); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, fmt.Errorf("%w: %s", common.ErrConflict, "username, email or phone is aready taken!")
		}
		return nil, result.Error
	}

	return &user, nil
}

func (us *UsersService) Delete(id uint) (*model.User, error) {
	var user model.User
	if restult := us.db.First(&user, id); restult.Error != nil {
		if errors.Is(restult.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "user")
		}
		return nil, restult.Error
	}

	if result := us.db.Delete(&user); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "user")
		}
		return nil, result.Error
	}
	return &user, nil
}

func (us *UsersService) AssignServicesToUser(userID uint, ServiceIDs []uint) (*model.User, error) {
	var user model.User
	if result := us.db.First(&user, userID); result.Error != nil {
		return nil, fmt.Errorf("%w: %s", common.ErrRecordNotFound, "user not found")
	}

	if user.Role != model.RoleEmployee.String() {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "Servicen can be assigned only to Employees")
	}

	var services []model.Service
	if result := us.db.Find(&services, ServiceIDs); result.Error != nil {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "invalid services")
	}

	if err := us.db.Model(&user).Association("Services").Replace(&services); err != nil {
		return nil, fmt.Errorf("%w: %s", common.ErrConflict, "invalid services")
	}

	return &user, nil
}
