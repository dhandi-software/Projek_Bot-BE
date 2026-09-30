package service

import (
	"bot_be/internal/model"

	"gorm.io/gorm"
)

type BannerService interface {
	GetBanners(activeOnly bool) ([]model.Banner, error)
	GetBannerByID(id string) (*model.Banner, error)
	CreateBanner(b *model.Banner) error
	UpdateBanner(b *model.Banner) error
	DeleteBanner(id string) error
}

type bannerService struct {
	db *gorm.DB
}

func NewBannerService(db *gorm.DB) BannerService {
	return &bannerService{db: db}
}

func (s *bannerService) GetBanners(activeOnly bool) ([]model.Banner, error) {
	var banners []model.Banner
	q := s.db.Model(&model.Banner{})
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	err := q.Order("sort_order asc, id desc").Find(&banners).Error
	return banners, err
}

func (s *bannerService) GetBannerByID(id string) (*model.Banner, error) {
	var b model.Banner
	if err := s.db.First(&b, id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *bannerService) CreateBanner(b *model.Banner) error {
	return s.db.Create(b).Error
}

func (s *bannerService) UpdateBanner(b *model.Banner) error {
	return s.db.Save(b).Error
}

func (s *bannerService) DeleteBanner(id string) error {
	return s.db.Delete(&model.Banner{}, id).Error
}
