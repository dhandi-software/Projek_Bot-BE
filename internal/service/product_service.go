package service

import (
	"strconv"
	"time"

	"bot_be/internal/model"

	"gorm.io/gorm"
)

type ProductFilter struct {
	Category  string
	Query     string
	Featured  bool
	Active    bool
	Status    string
	BestDeals bool
	Now       time.Time
}

type ProductService interface {
	GetProducts(filter ProductFilter) ([]model.Product, error)
	GetProductByID(idOrSKU string) (*model.Product, error)
	CreateProduct(product *model.Product) error
	UpdateProduct(product *model.Product) error
	DeleteProduct(id uint) error
	BulkCreateProducts(products []model.Product) error
}

type productService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) ProductService {
	return &productService{db: db}
}

func (s *productService) GetProducts(filter ProductFilter) ([]model.Product, error) {
	var products []model.Product
	q := s.db.Model(&model.Product{})

	if filter.Category != "" && filter.Category != "ALL" {
		q = q.Where("LOWER(category) = LOWER(?)", filter.Category)
	}
	if filter.Query != "" {
		searchTerm := "%" + filter.Query + "%"
		q = q.Where("LOWER(title) LIKE LOWER(?) OR LOWER(sku) LIKE LOWER(?) OR LOWER(brand) LIKE LOWER(?) OR LOWER(materials) LIKE LOWER(?)", searchTerm, searchTerm, searchTerm, searchTerm)
	}
	if filter.Featured {
		q = q.Where("is_featured = ?", true)
	}
	if filter.Active {
		q = q.Where("is_active = ?", true)
	}
	if filter.Status != "" {
		q = q.Where("LOWER(status) = LOWER(?)", filter.Status)
	}
	if filter.BestDeals {
		q = q.Where("is_best_deal = ? AND best_deal_expires_at IS NOT NULL AND best_deal_expires_at > ?", true, filter.Now)
	}

	err := q.Order("id desc").Find(&products).Error
	return products, err
}

func (s *productService) GetProductByID(idOrSKU string) (*model.Product, error) {
	var product model.Product
	if num, err := strconv.ParseUint(idOrSKU, 10, 64); err == nil {
		if err := s.db.Where("id = ? OR LOWER(sku) = LOWER(?)", num, idOrSKU).First(&product).Error; err == nil {
			return &product, nil
		}
	}
	if err := s.db.Where("LOWER(sku) = LOWER(?)", idOrSKU).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (s *productService) CreateProduct(product *model.Product) error {
	return s.db.Create(product).Error
}

func (s *productService) UpdateProduct(product *model.Product) error {
	return s.db.Select("*").Save(product).Error
}

func (s *productService) DeleteProduct(id uint) error {
	return s.db.Delete(&model.Product{}, id).Error
}

func (s *productService) BulkCreateProducts(products []model.Product) error {
	return s.db.Create(&products).Error
}
