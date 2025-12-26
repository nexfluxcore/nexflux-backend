package repositories

import (
	"nexfi-backend/models"

	"gorm.io/gorm"
)

// CircuitRepository handles circuit database operations
type CircuitRepository struct {
	*BaseRepository
}

// NewCircuitRepository creates a new CircuitRepository
func NewCircuitRepository(db *gorm.DB) *CircuitRepository {
	return &CircuitRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// CircuitFilter defines filter options
type CircuitFilter struct {
	UserID    string
	ProjectID string
	Search    string
	IsPublic  *bool
}

// FindAll finds all circuits with filters
func (r *CircuitRepository) FindAll(filter CircuitFilter, page, limit int, sort, order string) ([]models.Circuit, int64, error) {
	var circuits []models.Circuit
	var total int64

	query := r.DB.Model(&models.Circuit{})

	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if filter.ProjectID != "" {
		query = query.Where("project_id = ?", filter.ProjectID)
	}

	if filter.Search != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Search+"%")
	}

	if filter.IsPublic != nil {
		query = query.Where("is_public = ?", *filter.IsPublic)
	}

	query.Count(&total)

	// Sort handling
	orderClause := "created_at DESC"
	if sort != "" {
		if order == "asc" {
			orderClause = sort + " ASC"
		} else {
			orderClause = sort + " DESC"
		}
	}

	err := query.Scopes(Paginate(page, limit)).
		Order(orderClause).
		Preload("Project").
		Find(&circuits).Error

	return circuits, total, err
}

// FindByID finds circuit by ID
func (r *CircuitRepository) FindByID(id string) (*models.Circuit, error) {
	var circuit models.Circuit
	err := r.DB.Preload("Project").Preload("User").First(&circuit, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &circuit, nil
}

// Create creates a new circuit
func (r *CircuitRepository) Create(circuit *models.Circuit) error {
	return r.DB.Create(circuit).Error
}

// Update updates a circuit
func (r *CircuitRepository) Update(circuit *models.Circuit) error {
	return r.DB.Save(circuit).Error
}

// Delete deletes a circuit
func (r *CircuitRepository) Delete(id string) error {
	return r.DB.Delete(&models.Circuit{}, "id = ?", id).Error
}

// IsOwner checks if user is owner
func (r *CircuitRepository) IsOwner(circuitID, userID string) (bool, error) {
	var count int64
	err := r.DB.Model(&models.Circuit{}).Where("id = ? AND user_id = ?", circuitID, userID).Count(&count).Error
	return count > 0, err
}

// FindTemplates finds all circuit templates
func (r *CircuitRepository) FindTemplates(category, difficulty, search string, page, limit int) ([]models.CircuitTemplate, int64, error) {
	var templates []models.CircuitTemplate
	var total int64

	query := r.DB.Model(&models.CircuitTemplate{})

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}

	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	query.Count(&total)

	err := query.Scopes(Paginate(page, limit)).
		Order("use_count DESC").
		Find(&templates).Error

	return templates, total, err
}

// FindTemplateByID finds template by ID
func (r *CircuitRepository) FindTemplateByID(id string) (*models.CircuitTemplate, error) {
	var template models.CircuitTemplate
	err := r.DB.First(&template, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// IncrementTemplateUseCount increments template use count
func (r *CircuitRepository) IncrementTemplateUseCount(id string) error {
	return r.DB.Model(&models.CircuitTemplate{}).Where("id = ?", id).
		Update("use_count", gorm.Expr("use_count + 1")).Error
}
