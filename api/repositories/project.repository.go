package repositories

import (
	"nexfi-backend/models"

	"gorm.io/gorm"
)

// ProjectRepository handles project database operations
type ProjectRepository struct {
	*BaseRepository
}

// NewProjectRepository creates a new ProjectRepository
func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// ProjectFilter defines filter options for projects
type ProjectFilter struct {
	UserID     string
	Search     string
	Difficulty string
	IsFavorite *bool
	IsPublic   *bool
	IsTemplate bool
}

// FindAll finds all projects with filters and pagination
func (r *ProjectRepository) FindAll(filter ProjectFilter, page, limit int) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	query := r.DB.Model(&models.Project{})

	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if filter.Search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	if filter.Difficulty != "" {
		query = query.Where("difficulty = ?", filter.Difficulty)
	}

	if filter.IsFavorite != nil {
		query = query.Where("is_favorite = ?", *filter.IsFavorite)
	}

	if filter.IsPublic != nil {
		query = query.Where("is_public = ?", *filter.IsPublic)
	}

	if filter.IsTemplate {
		query = query.Where("is_template = ?", true)
	}

	// Count total
	query.Count(&total)

	// Apply pagination and get results
	err := query.Scopes(Paginate(page, limit)).
		Order("created_at DESC").
		Preload("User").
		Find(&projects).Error

	return projects, total, err
}

// FindByID finds project by ID
func (r *ProjectRepository) FindByID(id string) (*models.Project, error) {
	var project models.Project
	err := r.DB.Preload("User").
		Preload("Components").
		Preload("Components.Component").
		Preload("Collaborators").
		Preload("Collaborators.User").
		First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// FindByIDSimple finds project by ID without preloading
func (r *ProjectRepository) FindByIDSimple(id string) (*models.Project, error) {
	var project models.Project
	err := r.DB.First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// Create creates a new project
func (r *ProjectRepository) Create(project *models.Project) error {
	return r.DB.Create(project).Error
}

// Update updates a project
func (r *ProjectRepository) Update(project *models.Project) error {
	return r.DB.Save(project).Error
}

// Delete deletes a project
func (r *ProjectRepository) Delete(id string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// Delete project components
		if err := tx.Delete(&models.ProjectComponent{}, "project_id = ?", id).Error; err != nil {
			return err
		}
		// Delete collaborators
		if err := tx.Delete(&models.ProjectCollaborator{}, "project_id = ?", id).Error; err != nil {
			return err
		}
		// Delete project
		return tx.Delete(&models.Project{}, "id = ?", id).Error
	})
}

// ToggleFavorite toggles project favorite status
func (r *ProjectRepository) ToggleFavorite(id string) (bool, error) {
	var project models.Project
	if err := r.DB.First(&project, "id = ?", id).Error; err != nil {
		return false, err
	}

	project.IsFavorite = !project.IsFavorite
	err := r.DB.Save(&project).Error
	return project.IsFavorite, err
}

// DuplicateProject creates a copy of a project
func (r *ProjectRepository) DuplicateProject(id, newUserID string) (*models.Project, error) {
	var original models.Project
	if err := r.DB.Preload("Components").First(&original, "id = ?", id).Error; err != nil {
		return nil, err
	}

	// Create new project
	newProject := models.Project{
		UserID:           newUserID,
		Name:             original.Name + " (Copy)",
		Description:      original.Description,
		ThumbnailURL:     original.ThumbnailURL,
		Difficulty:       original.Difficulty,
		Progress:         0,
		XPReward:         original.XPReward,
		IsPublic:         false,
		IsFavorite:       false,
		IsTemplate:       false,
		HardwarePlatform: original.HardwarePlatform,
		Tags:             original.Tags,
		SchemaData:       original.SchemaData,
		CodeData:         original.CodeData,
		SimulationData:   original.SimulationData,
	}

	if err := r.DB.Create(&newProject).Error; err != nil {
		return nil, err
	}

	// Copy components
	for _, comp := range original.Components {
		newComp := models.ProjectComponent{
			ProjectID:   newProject.ID,
			ComponentID: comp.ComponentID,
			Quantity:    comp.Quantity,
			PositionX:   comp.PositionX,
			PositionY:   comp.PositionY,
			Rotation:    comp.Rotation,
			ConfigData:  comp.ConfigData,
		}
		r.DB.Create(&newComp)
	}

	return &newProject, nil
}

// GetCollaborators gets project collaborators
func (r *ProjectRepository) GetCollaborators(projectID string) ([]models.ProjectCollaborator, error) {
	var collaborators []models.ProjectCollaborator
	err := r.DB.Where("project_id = ?", projectID).
		Preload("User").
		Find(&collaborators).Error
	return collaborators, err
}

// AddCollaborator adds a collaborator to project
func (r *ProjectRepository) AddCollaborator(collab *models.ProjectCollaborator) error {
	return r.DB.Create(collab).Error
}

// RemoveCollaborator removes a collaborator from project
func (r *ProjectRepository) RemoveCollaborator(projectID, userID string) error {
	return r.DB.Delete(&models.ProjectCollaborator{}, "project_id = ? AND user_id = ?", projectID, userID).Error
}

// IsOwner checks if user is project owner
func (r *ProjectRepository) IsOwner(projectID, userID string) (bool, error) {
	var count int64
	err := r.DB.Model(&models.Project{}).Where("id = ? AND user_id = ?", projectID, userID).Count(&count).Error
	return count > 0, err
}

// HasAccess checks if user has access to project
func (r *ProjectRepository) HasAccess(projectID, userID string) (bool, error) {
	// Check if owner
	isOwner, err := r.IsOwner(projectID, userID)
	if err != nil {
		return false, err
	}
	if isOwner {
		return true, nil
	}

	// Check if collaborator
	var count int64
	err = r.DB.Model(&models.ProjectCollaborator{}).
		Where("project_id = ? AND user_id = ? AND accepted_at IS NOT NULL", projectID, userID).
		Count(&count).Error
	return count > 0, err
}

// GetTemplates gets template projects
func (r *ProjectRepository) GetTemplates(page, limit int) ([]models.Project, int64, error) {
	return r.FindAll(ProjectFilter{IsTemplate: true, IsPublic: func() *bool { b := true; return &b }()}, page, limit)
}
