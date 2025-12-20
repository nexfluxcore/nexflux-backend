package services

import (
	"errors"
	"nexfi-backend/api/repositories"
	"nexfi-backend/dto"
	"nexfi-backend/models"

	"gorm.io/gorm"
)

// ComponentService handles component business logic
type ComponentService struct {
	repo *repositories.ComponentRepository
}

// NewComponentService creates a new ComponentService
func NewComponentService(db *gorm.DB) *ComponentService {
	return &ComponentService{
		repo: repositories.NewComponentRepository(db),
	}
}

// ListComponents lists components with filters
func (s *ComponentService) ListComponents(req dto.ComponentListRequest, userID string) ([]dto.ComponentResponse, dto.PaginationResponse, error) {
	filter := repositories.ComponentFilter{
		CategoryID: req.Category,
		Search:     req.Search,
		InStock:    req.InStock,
		IsActive:   true,
	}

	components, total, err := s.repo.FindAll(filter, req.Page, req.Limit, req.Sort, req.Order)
	if err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	responses := make([]dto.ComponentResponse, len(components))
	for i, c := range components {
		responses[i] = s.toComponentResponse(&c, userID)
	}

	pagination := dto.PaginationResponse{
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      int(total),
		TotalPages: int((total + int64(req.Limit) - 1) / int64(req.Limit)),
	}

	return responses, pagination, nil
}

// GetComponent gets component by ID
func (s *ComponentService) GetComponent(id, userID string) (*dto.ComponentResponse, error) {
	component, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("component not found")
		}
		return nil, err
	}

	response := s.toComponentResponse(component, userID)
	return &response, nil
}

// SearchComponents searches components
func (s *ComponentService) SearchComponents(req dto.ComponentSearchRequest, userID string) ([]dto.ComponentResponse, error) {
	components, err := s.repo.Search(req.Query, req.Category, req.Limit)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ComponentResponse, len(components))
	for i, c := range components {
		responses[i] = s.toComponentResponse(&c, userID)
	}

	return responses, nil
}

// GetCategories gets all categories
func (s *ComponentService) GetCategories() ([]dto.CategoryResponse, error) {
	categories, err := s.repo.GetCategories()
	if err != nil {
		return nil, err
	}

	responses := make([]dto.CategoryResponse, len(categories))
	for i, c := range categories {
		responses[i] = dto.CategoryResponse{
			ID:          c.ID,
			Name:        c.Name,
			Slug:        c.Slug,
			Icon:        c.Icon,
			Color:       c.Color,
			Description: c.Description,
			Order:       c.Order,
		}
	}

	return responses, nil
}

// GetFavorites gets user's favorite components
func (s *ComponentService) GetFavorites(userID string, page, limit int) ([]dto.ComponentResponse, dto.PaginationResponse, error) {
	components, total, err := s.repo.GetUserFavorites(userID, page, limit)
	if err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	responses := make([]dto.ComponentResponse, len(components))
	for i, c := range components {
		resp := s.toComponentResponse(&c, userID)
		resp.IsFavorite = true
		responses[i] = resp
	}

	pagination := dto.PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
	}

	return responses, pagination, nil
}

// ToggleFavorite toggles component favorite status
func (s *ComponentService) ToggleFavorite(componentID, userID string) (bool, error) {
	// Verify component exists
	_, err := s.repo.FindByID(componentID)
	if err != nil {
		return false, errors.New("component not found")
	}

	return s.repo.ToggleFavorite(userID, componentID)
}

// CreateRequest creates a component request
func (s *ComponentService) CreateRequest(userID string, req dto.ComponentRequestCreateRequest) (*dto.ComponentRequestResponse, error) {
	request := &models.ComponentRequest{
		UserID:        userID,
		ComponentName: req.ComponentName,
		Manufacturer:  req.Manufacturer,
		PartNumber:    req.PartNumber,
		Category:      req.Category,
		Description:   req.Description,
		UseCase:       req.UseCase,
		DatasheetURL:  req.DatasheetURL,
		ProductURL:    req.ProductURL,
		Priority:      models.RequestPriority(req.Priority),
		Status:        models.StatusPending,
	}

	if request.Priority == "" {
		request.Priority = models.PriorityMedium
	}

	if err := s.repo.CreateRequest(request); err != nil {
		return nil, err
	}

	return s.toRequestResponse(request), nil
}

// GetUserRequests gets user's component requests
func (s *ComponentService) GetUserRequests(userID string) ([]dto.ComponentRequestResponse, error) {
	requests, err := s.repo.GetUserRequests(userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ComponentRequestResponse, len(requests))
	for i, r := range requests {
		responses[i] = *s.toRequestResponse(&r)
	}

	return responses, nil
}

// Helper functions

func (s *ComponentService) toComponentResponse(c *models.Component, userID string) dto.ComponentResponse {
	response := dto.ComponentResponse{
		ID:              c.ID,
		CategoryID:      c.CategoryID,
		Name:            c.Name,
		Description:     c.Description,
		Manufacturer:    c.Manufacturer,
		PartNumber:      c.PartNumber,
		Specs:           c.Specs,
		Price:           c.Price,
		Stock:           c.Stock,
		Rating:          c.Rating,
		RatingCount:     c.RatingCount,
		ImageURL:        c.ImageURL,
		DatasheetURL:    c.DatasheetURL,
		SimulationModel: c.SimulationModel,
		IsActive:        c.IsActive,
	}

	if c.Category != nil {
		response.Category = &dto.CategoryResponse{
			ID:   c.Category.ID,
			Name: c.Category.Name,
			Slug: c.Category.Slug,
		}
	}

	if userID != "" {
		response.IsFavorite = s.repo.IsFavorite(userID, c.ID)
	}

	return response
}

func (s *ComponentService) toRequestResponse(r *models.ComponentRequest) *dto.ComponentRequestResponse {
	return &dto.ComponentRequestResponse{
		ID:            r.ID,
		UserID:        r.UserID,
		ComponentName: r.ComponentName,
		Manufacturer:  r.Manufacturer,
		PartNumber:    r.PartNumber,
		Category:      r.Category,
		Description:   r.Description,
		UseCase:       r.UseCase,
		DatasheetURL:  r.DatasheetURL,
		ProductURL:    r.ProductURL,
		Priority:      string(r.Priority),
		Status:        string(r.Status),
		AdminNotes:    r.AdminNotes,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}
