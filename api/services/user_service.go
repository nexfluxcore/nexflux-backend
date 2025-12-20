package services

import (
	"nexfi-backend/models"
	"nexfi-backend/utils"

	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{DB: db}
}

func (us *UserService) GetUserByID(userID string) (*models.User, error) {
	var user models.User
	err := us.DB.First(&user, "id = ?", userID).Error
	return &user, err
}

func (us *UserService) UpdateUserProfile(userID, username string) error {
	return us.DB.Model(&models.User{}).Where("id = ?", userID).Update("username", username).Error
}

func (us *UserService) GetAllUsers() ([]models.User, error) {
	var users []models.User
	err := us.DB.Find(&users).Error
	return users, err
}

func (us *UserService) CreateUser(email, password, username string) error {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	user := models.User{
		Email:         email,
		PasswordHash:  hashedPassword,
		Username:      username,
		EmailVerified: false,
	}

	return us.DB.Create(&user).Error
}

func (us *UserService) DeleteUser(userID string) error {
	return us.DB.Delete(&models.User{}, "id = ?", userID).Error
}
