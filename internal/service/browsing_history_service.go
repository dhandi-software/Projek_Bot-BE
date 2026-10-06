package service

import (
	"errors"
	"strings"
	"time"

	"bot_be/internal/model"

	"gorm.io/gorm"
)

type BrowsingHistoryFilter struct {
	UserEmail string
	SessionID string
	Query     string
	Date      string
	Page      int
	Limit     int
}

type BrowsingHistoryService interface {
	RecordView(userID uint, userEmail string, sessionID string, productID uint) (*model.BrowsingHistory, error)
	GetHistory(filter BrowsingHistoryFilter) ([]model.BrowsingHistory, int64, error)
	DeleteHistoryItem(id uint, userEmail string, sessionID string) error
	ClearHistory(userEmail string, sessionID string) error
}

type browsingHistoryService struct {
	db *gorm.DB
}

func NewBrowsingHistoryService(db *gorm.DB) BrowsingHistoryService {
	return &browsingHistoryService{db: db}
}

func (s *browsingHistoryService) RecordView(userID uint, userEmail string, sessionID string, productID uint) (*model.BrowsingHistory, error) {
	if productID == 0 {
		return nil, errors.New("product_id is required")
	}

	var product model.Product
	if err := s.db.First(&product, productID).Error; err != nil {
		return nil, errors.New("product not found")
	}

	now := time.Now()
	var existing model.BrowsingHistory

	q := s.db.Where("product_id = ?", productID)
	if userEmail != "" {
		q = q.Where("user_email = ?", userEmail)
	} else if sessionID != "" {
		q = q.Where("session_id = ?", sessionID)
	}

	// Update timestamp if viewed earlier today, otherwise create fresh entry
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	err := q.Where("viewed_at >= ?", todayStart).First(&existing).Error
	if err == nil {
		existing.ViewedAt = now
		if userID > 0 && existing.UserID == 0 {
			existing.UserID = userID
		}
		if userEmail != "" && existing.UserEmail == "" {
			existing.UserEmail = userEmail
		}
		if err := s.db.Save(&existing).Error; err != nil {
			return nil, err
		}
		existing.Product = product
		return &existing, nil
	}

	history := model.BrowsingHistory{
		UserID:    userID,
		UserEmail: userEmail,
		SessionID: sessionID,
		ProductID: productID,
		Product:   product,
		ViewedAt:  now,
	}

	if err := s.db.Create(&history).Error; err != nil {
		return nil, err
	}

	return &history, nil
}

func (s *browsingHistoryService) GetHistory(filter BrowsingHistoryFilter) ([]model.BrowsingHistory, int64, error) {
	var histories []model.BrowsingHistory
	var total int64

	q := s.db.Model(&model.BrowsingHistory{}).Preload("Product")

	// Match by user email or session identifier, with fallback to initial sample session if empty
	var userCount int64
	checkQ := s.db.Model(&model.BrowsingHistory{})
	if filter.UserEmail != "" {
		checkQ = checkQ.Where("user_email = ?", filter.UserEmail)
	} else if filter.SessionID != "" {
		checkQ = checkQ.Where("session_id = ?", filter.SessionID)
	}
	_ = checkQ.Count(&userCount).Error

	if userCount > 0 {
		if filter.UserEmail != "" {
			q = q.Where("user_email = ?", filter.UserEmail)
		} else if filter.SessionID != "" {
			q = q.Where("session_id = ?", filter.SessionID)
		}
	} else {
		if filter.UserEmail != "" {
			q = q.Where("user_email = ? OR session_id = 'guest-session'", filter.UserEmail)
		} else if filter.SessionID != "" {
			q = q.Where("session_id = ? OR session_id = 'guest-session'", filter.SessionID)
		}
	}

	// Optional text search through joined product details
	if filter.Query != "" {
		searchTerm := "%" + strings.ToLower(filter.Query) + "%"
		q = q.Joins("JOIN products ON products.id = browsing_histories.product_id").
			Where("LOWER(products.title) LIKE ? OR LOWER(products.sku) LIKE ? OR LOWER(products.brand) LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	// Optional date filter (supports YYYY-MM-DD or DD/MM/YYYY)
	if filter.Date != "" {
		var targetDate time.Time
		var parseErr error

		cleanDate := strings.TrimSpace(filter.Date)
		if strings.Contains(cleanDate, "/") {
			targetDate, parseErr = time.Parse("02/01/2006", cleanDate)
		} else {
			targetDate, parseErr = time.Parse("2006-01-02", cleanDate)
		}

		if parseErr == nil {
			dateStr := targetDate.Format("2006-01-02")
			startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.Local)
			endOfDay := startOfDay.Add(24 * time.Hour)
			q = q.Where("TO_CHAR(browsing_histories.viewed_at, 'YYYY-MM-DD') = ? OR DATE(browsing_histories.viewed_at) = ? OR (browsing_histories.viewed_at >= ? AND browsing_histories.viewed_at < ?)", dateStr, dateStr, startOfDay, endOfDay)
		}
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}

	limit := filter.Limit
	if limit < 1 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	err := q.Order("browsing_histories.viewed_at DESC").Offset(offset).Limit(limit).Find(&histories).Error
	return histories, total, err
}

func (s *browsingHistoryService) DeleteHistoryItem(id uint, userEmail string, sessionID string) error {
	q := s.db.Where("id = ?", id)
	if userEmail != "" {
		q = q.Where("user_email = ?", userEmail)
	} else if sessionID != "" {
		q = q.Where("session_id = ?", sessionID)
	}
	return q.Delete(&model.BrowsingHistory{}).Error
}

func (s *browsingHistoryService) ClearHistory(userEmail string, sessionID string) error {
	q := s.db
	if userEmail != "" {
		q = q.Where("user_email = ?", userEmail)
	} else if sessionID != "" {
		q = q.Where("session_id = ?", sessionID)
	} else {
		return errors.New("user_email or session_id is required to clear history")
	}
	return q.Delete(&model.BrowsingHistory{}).Error
}
