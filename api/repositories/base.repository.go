package repositories

import (
	"gorm.io/gorm"
)

// BaseRepository provides common database operations
type BaseRepository struct {
	DB *gorm.DB
}

// NewBaseRepository creates a new BaseRepository
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{DB: db}
}

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Page  int
	Limit int
}

// PaginationResult holds pagination result
type PaginationResult struct {
	Page       int
	Limit      int
	Total      int64
	TotalPages int
}

// GetOffset calculates offset for pagination
func (p *PaginationParams) GetOffset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 20
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
	return (p.Page - 1) * p.Limit
}

// CalculateTotalPages calculates total pages
func (p *PaginationResult) CalculateTotalPages() {
	if p.Limit <= 0 {
		p.TotalPages = 0
		return
	}
	p.TotalPages = int((p.Total + int64(p.Limit) - 1) / int64(p.Limit))
}

// Paginate applies pagination to a query
func Paginate(page, limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page < 1 {
			page = 1
		}
		if limit < 1 {
			limit = 20
		}
		if limit > 100 {
			limit = 100
		}
		offset := (page - 1) * limit
		return db.Offset(offset).Limit(limit)
	}
}
