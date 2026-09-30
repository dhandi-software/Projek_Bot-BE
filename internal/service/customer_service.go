package service

import (
	"errors"

	"bot_be/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CustomerService interface {
	RegisterCustomer(cust *model.Customer, rawPassword string) error
	GetCustomerByEmail(email string) (*model.Customer, error)
	GetCustomerByID(id uint) (*model.Customer, error)
	UpdateCustomer(cust *model.Customer) error
	GetAllCustomers() ([]model.Customer, error)
}

type customerService struct {
	db *gorm.DB
}

func NewCustomerService(db *gorm.DB) CustomerService {
	return &customerService{db: db}
}

func (s *customerService) RegisterCustomer(cust *model.Customer, rawPassword string) error {
	var existing model.Customer
	if err := s.db.Where("email = ?", cust.Email).First(&existing).Error; err == nil {
		return errors.New("Email sudah terdaftar. Silakan login.")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("Gagal memproses password")
	}

	cust.Password = string(hashedPassword)
	return s.db.Create(cust).Error
}

func (s *customerService) GetCustomerByEmail(email string) (*model.Customer, error) {
	var cust model.Customer
	if err := s.db.Where("email = ?", email).First(&cust).Error; err != nil {
		return nil, err
	}
	return &cust, nil
}

func (s *customerService) GetCustomerByID(id uint) (*model.Customer, error) {
	var cust model.Customer
	if err := s.db.First(&cust, id).Error; err != nil {
		return nil, err
	}
	return &cust, nil
}

func (s *customerService) UpdateCustomer(cust *model.Customer) error {
	return s.db.Save(cust).Error
}

func (s *customerService) GetAllCustomers() ([]model.Customer, error) {
	var customers []model.Customer
	err := s.db.Order("id desc").Find(&customers).Error
	return customers, err
}
