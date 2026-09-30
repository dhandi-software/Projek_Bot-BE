package service

import (
	"bot_be/internal/model"

	"gorm.io/gorm"
)

type CategoryService interface {
	GetCategories() ([]model.Category, error)
	CreateCategory(cat *model.Category) error
	GetCategoryByID(id string) (*model.Category, error)
	GetCategoryByName(name string) (*model.Category, error)
	UpdateCategory(cat *model.Category) error
	DeleteCategory(id string) error
}

type categoryService struct {
	db *gorm.DB
}

func NewCategoryService(db *gorm.DB) CategoryService {
	return &categoryService{db: db}
}

func (s *categoryService) GetCategories() ([]model.Category, error) {
	var categories []model.Category
	if err := s.db.Order("name asc").Find(&categories).Error; err != nil {
		return nil, err
	}

	if len(categories) == 0 {
		defaultCategories := []model.Category{
			{Name: "Computer & Laptop", Slug: "computer-laptop", Description: "Perangkat Komputer dan Laptop", Icon: "Laptop", IsActive: true},
			{Name: "Gaming Console", Slug: "gaming-console", Description: "Konsol Game dan Aksesoris", Icon: "Gaming", IsActive: true},
			{Name: "Smartphone", Slug: "smartphone", Description: "Ponsel Pintar dan Tablet", Icon: "Smartphone", IsActive: true},
			{Name: "Headphone", Slug: "headphone", Description: "Headphone, Earphone, Audio", Icon: "Headphone", IsActive: true},
			{Name: "Computer Accessories", Slug: "computer-accessories", Description: "Aksesoris Komputer", Icon: "Plug", IsActive: true},
			{Name: "Umum", Slug: "umum", Description: "Kategori Umum", Icon: "Package", IsActive: true},
		}
		for _, cat := range defaultCategories {
			s.db.Create(&cat)
		}
		s.db.Order("name asc").Find(&categories)
	}

	return categories, nil
}

func (s *categoryService) GetCategoryByID(id string) (*model.Category, error) {
	var cat model.Category
	if err := s.db.First(&cat, id).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

func (s *categoryService) GetCategoryByName(name string) (*model.Category, error) {
	var cat model.Category
	if err := s.db.Where("LOWER(name) = LOWER(?)", name).First(&cat).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

func (s *categoryService) CreateCategory(cat *model.Category) error {
	return s.db.Create(cat).Error
}

func (s *categoryService) UpdateCategory(cat *model.Category) error {
	return s.db.Save(cat).Error
}

func (s *categoryService) DeleteCategory(id string) error {
	return s.db.Delete(&model.Category{}, id).Error
}
