package repositories

import (
	"nexfi-backend/models"

	"gorm.io/gorm"
)

// ComponentRepository handles component database operations
type ComponentRepository struct {
	*BaseRepository
}

// NewComponentRepository creates a new ComponentRepository
func NewComponentRepository(db *gorm.DB) *ComponentRepository {
	return &ComponentRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// ComponentFilter defines filter options
type ComponentFilter struct {
	CategoryID string
	Search     string
	InStock    *bool
	IsActive   bool
}

// FindAll finds all components with filters
func (r *ComponentRepository) FindAll(filter ComponentFilter, page, limit int, sort, order string) ([]models.Component, int64, error) {
	var components []models.Component
	var total int64

	query := r.DB.Model(&models.Component{})

	if filter.IsActive {
		query = query.Where("is_active = ?", true)
	}

	if filter.CategoryID != "" {
		query = query.Where("category_id = ?", filter.CategoryID)
	}

	if filter.Search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ? OR manufacturer ILIKE ?",
			"%"+filter.Search+"%", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	if filter.InStock != nil && *filter.InStock {
		query = query.Where("stock > 0")
	}

	// Count
	query.Count(&total)

	// Sort
	if sort == "" {
		sort = "name"
	}
	if order == "" {
		order = "asc"
	}
	query = query.Order(sort + " " + order)

	// Paginate and fetch
	err := query.Scopes(Paginate(page, limit)).
		Preload("Category").
		Find(&components).Error

	return components, total, err
}

// FindByID finds component by ID
func (r *ComponentRepository) FindByID(id string) (*models.Component, error) {
	var component models.Component
	err := r.DB.Preload("Category").First(&component, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

// Search searches components by name
func (r *ComponentRepository) Search(query string, categoryID string, limit int) ([]models.Component, error) {
	var components []models.Component

	q := r.DB.Model(&models.Component{}).
		Where("is_active = ?", true).
		Where("name ILIKE ? OR part_number ILIKE ?", "%"+query+"%", "%"+query+"%")

	if categoryID != "" {
		q = q.Where("category_id = ?", categoryID)
	}

	err := q.Limit(limit).Preload("Category").Find(&components).Error
	return components, err
}

// GetCategories gets all component categories
func (r *ComponentRepository) GetCategories() ([]models.ComponentCategory, error) {
	var categories []models.ComponentCategory
	err := r.DB.Order("\"order\" ASC").Find(&categories).Error
	return categories, err
}

// GetCategoryBySlug gets category by slug
func (r *ComponentRepository) GetCategoryBySlug(slug string) (*models.ComponentCategory, error) {
	var category models.ComponentCategory
	err := r.DB.First(&category, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetUserFavorites gets user's favorite components
func (r *ComponentRepository) GetUserFavorites(userID string, page, limit int) ([]models.Component, int64, error) {
	var total int64
	var favorites []models.UserFavoriteComponent

	query := r.DB.Model(&models.UserFavoriteComponent{}).Where("user_id = ?", userID)
	query.Count(&total)

	err := query.Scopes(Paginate(page, limit)).
		Preload("Component").
		Preload("Component.Category").
		Find(&favorites).Error

	if err != nil {
		return nil, 0, err
	}

	components := make([]models.Component, len(favorites))
	for i, fav := range favorites {
		if fav.Component != nil {
			components[i] = *fav.Component
		}
	}

	return components, total, nil
}

// ToggleFavorite toggles component favorite status
func (r *ComponentRepository) ToggleFavorite(userID, componentID string) (bool, error) {
	var existing models.UserFavoriteComponent
	err := r.DB.First(&existing, "user_id = ? AND component_id = ?", userID, componentID).Error

	if err == gorm.ErrRecordNotFound {
		// Add favorite
		fav := models.UserFavoriteComponent{
			UserID:      userID,
			ComponentID: componentID,
		}
		return true, r.DB.Create(&fav).Error
	}

	// Remove favorite
	return false, r.DB.Delete(&existing).Error
}

// IsFavorite checks if component is favorited by user
func (r *ComponentRepository) IsFavorite(userID, componentID string) bool {
	var count int64
	r.DB.Model(&models.UserFavoriteComponent{}).Where("user_id = ? AND component_id = ?", userID, componentID).Count(&count)
	return count > 0
}

// CreateRequest creates a component request
func (r *ComponentRepository) CreateRequest(request *models.ComponentRequest) error {
	return r.DB.Create(request).Error
}

// GetUserRequests gets user's component requests
func (r *ComponentRepository) GetUserRequests(userID string) ([]models.ComponentRequest, error) {
	var requests []models.ComponentRequest
	err := r.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&requests).Error
	return requests, err
}
