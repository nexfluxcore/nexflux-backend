package services

import (
	"errors"
	"nexfi-backend/api/repositories"
	"nexfi-backend/dto"
	"nexfi-backend/models"

	"gorm.io/gorm"
)

// ProjectService handles project business logic
type ProjectService struct {
	repo     *repositories.ProjectRepository
	userRepo *repositories.UserRepository
}

// NewProjectService creates a new ProjectService
func NewProjectService(db *gorm.DB) *ProjectService {
	return &ProjectService{
		repo:     repositories.NewProjectRepository(db),
		userRepo: repositories.NewUserRepository(db),
	}
}

// ListProjects lists user's projects with filters
func (s *ProjectService) ListProjects(userID string, req dto.ProjectListRequest) ([]dto.ProjectResponse, dto.PaginationResponse, error) {
	filter := repositories.ProjectFilter{
		UserID:     userID,
		Search:     req.Search,
		Difficulty: req.Difficulty,
		IsFavorite: req.IsFavorite,
		IsPublic:   req.IsPublic,
	}

	projects, total, err := s.repo.FindAll(filter, req.Page, req.Limit)
	if err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	responses := make([]dto.ProjectResponse, len(projects))
	for i, p := range projects {
		responses[i] = s.toProjectResponse(&p)
	}

	pagination := dto.PaginationResponse{
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      int(total),
		TotalPages: int((total + int64(req.Limit) - 1) / int64(req.Limit)),
	}

	return responses, pagination, nil
}

// GetProject gets project by ID
func (s *ProjectService) GetProject(projectID, userID string) (*dto.ProjectDetailResponse, error) {
	project, err := s.repo.FindByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		return nil, err
	}

	// Check access
	hasAccess, err := s.repo.HasAccess(projectID, userID)
	if err != nil {
		return nil, err
	}
	if !hasAccess && !project.IsPublic {
		return nil, errors.New("access denied")
	}

	return s.toProjectDetailResponse(project), nil
}

// CreateProject creates a new project
func (s *ProjectService) CreateProject(userID string, req dto.ProjectCreateRequest) (*dto.ProjectResponse, error) {
	project := &models.Project{
		UserID:           userID,
		Name:             req.Name,
		Description:      req.Description,
		Difficulty:       models.Difficulty(req.Difficulty),
		HardwarePlatform: req.HardwarePlatform,
		Tags:             req.Tags,
		IsPublic:         req.IsPublic,
		IsTemplate:       req.IsTemplate,
		XPReward:         100, // Default XP reward
	}

	if project.Difficulty == "" {
		project.Difficulty = models.DifficultyBeginner
	}

	if err := s.repo.Create(project); err != nil {
		return nil, err
	}

	response := s.toProjectResponse(project)
	return &response, nil
}

// UpdateProject updates a project
func (s *ProjectService) UpdateProject(projectID, userID string, req dto.ProjectUpdateRequest) (*dto.ProjectResponse, error) {
	project, err := s.repo.FindByIDSimple(projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	// Check ownership
	if project.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Update fields
	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.ThumbnailURL != "" {
		project.ThumbnailURL = req.ThumbnailURL
	}
	if req.Difficulty != "" {
		project.Difficulty = models.Difficulty(req.Difficulty)
	}
	if req.Progress != nil {
		project.Progress = *req.Progress
	}
	if req.HardwarePlatform != "" {
		project.HardwarePlatform = req.HardwarePlatform
	}
	if req.Tags != nil {
		project.Tags = req.Tags
	}
	if req.IsPublic != nil {
		project.IsPublic = *req.IsPublic
	}
	if req.IsFavorite != nil {
		project.IsFavorite = *req.IsFavorite
	}
	if req.SchemaData != nil {
		project.SchemaData = req.SchemaData
	}
	if req.CodeData != nil {
		project.CodeData = req.CodeData
	}
	if req.SimulationData != nil {
		project.SimulationData = req.SimulationData
	}

	if err := s.repo.Update(project); err != nil {
		return nil, err
	}

	response := s.toProjectResponse(project)
	return &response, nil
}

// DeleteProject deletes a project
func (s *ProjectService) DeleteProject(projectID, userID string) error {
	project, err := s.repo.FindByIDSimple(projectID)
	if err != nil {
		return errors.New("project not found")
	}

	if project.UserID != userID {
		return errors.New("access denied")
	}

	return s.repo.Delete(projectID)
}

// DuplicateProject duplicates a project
func (s *ProjectService) DuplicateProject(projectID, userID string) (*dto.ProjectResponse, error) {
	project, err := s.repo.DuplicateProject(projectID, userID)
	if err != nil {
		return nil, err
	}

	response := s.toProjectResponse(project)
	return &response, nil
}

// ToggleFavorite toggles project favorite status
func (s *ProjectService) ToggleFavorite(projectID, userID string) (bool, error) {
	project, err := s.repo.FindByIDSimple(projectID)
	if err != nil {
		return false, errors.New("project not found")
	}

	if project.UserID != userID {
		return false, errors.New("access denied")
	}

	return s.repo.ToggleFavorite(projectID)
}

// GetTemplates gets template projects
func (s *ProjectService) GetTemplates(page, limit int) ([]dto.ProjectResponse, dto.PaginationResponse, error) {
	projects, total, err := s.repo.GetTemplates(page, limit)
	if err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	responses := make([]dto.ProjectResponse, len(projects))
	for i, p := range projects {
		responses[i] = s.toProjectResponse(&p)
	}

	pagination := dto.PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
	}

	return responses, pagination, nil
}

// GetCollaborators gets collaborators for a project
func (s *ProjectService) GetCollaborators(projectID, userID string) ([]dto.CollaboratorResponse, error) {
	project, err := s.repo.FindByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		return nil, err
	}

	// Check access
	hasAccess, err := s.repo.HasAccess(projectID, userID)
	if err != nil {
		return nil, err
	}
	if !hasAccess && !project.IsPublic {
		return nil, errors.New("access denied")
	}

	collaborators := make([]dto.CollaboratorResponse, len(project.Collaborators))
	for i, c := range project.Collaborators {
		collaborators[i] = dto.CollaboratorResponse{
			ID:         c.ID,
			UserID:     c.UserID,
			Role:       string(c.Role),
			InvitedAt:  c.InvitedAt,
			AcceptedAt: c.AcceptedAt,
		}
		if c.User != nil {
			collaborators[i].User = &dto.UserResponse{
				ID:        c.User.ID,
				Name:      c.User.Name,
				Username:  c.User.Username,
				AvatarURL: c.User.AvatarURL,
			}
		}
	}

	return collaborators, nil
}

// Helper functions

func (s *ProjectService) toProjectResponse(p *models.Project) dto.ProjectResponse {
	response := dto.ProjectResponse{
		ID:                 p.ID,
		UserID:             p.UserID,
		Name:               p.Name,
		Description:        p.Description,
		ThumbnailURL:       p.ThumbnailURL,
		Difficulty:         string(p.Difficulty),
		Progress:           p.Progress,
		XPReward:           p.XPReward,
		IsPublic:           p.IsPublic,
		IsFavorite:         p.IsFavorite,
		IsTemplate:         p.IsTemplate,
		HardwarePlatform:   p.HardwarePlatform,
		Tags:               p.Tags,
		ComponentsCount:    len(p.Components),
		CollaboratorsCount: len(p.Collaborators),
		CompletedAt:        p.CompletedAt,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}

	if p.User != nil {
		response.User = &dto.UserResponse{
			ID:       p.User.ID,
			Name:     p.User.Name,
			Username: p.User.Username,
		}
	}

	return response
}

func (s *ProjectService) toProjectDetailResponse(p *models.Project) *dto.ProjectDetailResponse {
	base := s.toProjectResponse(p)

	components := make([]dto.ProjectComponentResponse, len(p.Components))
	for i, c := range p.Components {
		components[i] = dto.ProjectComponentResponse{
			ID:          c.ID,
			ComponentID: c.ComponentID,
			Quantity:    c.Quantity,
			PositionX:   c.PositionX,
			PositionY:   c.PositionY,
			Rotation:    c.Rotation,
			ConfigData:  c.ConfigData,
		}
		if c.Component != nil {
			components[i].Component = &dto.ComponentResponse{
				ID:   c.Component.ID,
				Name: c.Component.Name,
			}
		}
	}

	collaborators := make([]dto.CollaboratorResponse, len(p.Collaborators))
	for i, c := range p.Collaborators {
		collaborators[i] = dto.CollaboratorResponse{
			ID:         c.ID,
			UserID:     c.UserID,
			Role:       string(c.Role),
			InvitedAt:  c.InvitedAt,
			AcceptedAt: c.AcceptedAt,
		}
		if c.User != nil {
			collaborators[i].User = &dto.UserResponse{
				ID:       c.User.ID,
				Name:     c.User.Name,
				Username: c.User.Username,
			}
		}
	}

	return &dto.ProjectDetailResponse{
		ProjectResponse: base,
		SchemaData:      p.SchemaData,
		CodeData:        p.CodeData,
		SimulationData:  p.SimulationData,
		Components:      components,
		Collaborators:   collaborators,
	}
}
