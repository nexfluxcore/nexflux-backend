package services

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"nexfi-backend/api/repositories"
	"nexfi-backend/dto"
	"nexfi-backend/models"
	"nexfi-backend/pkg/storage"
	"os"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CircuitService handles circuit business logic
type CircuitService struct {
	repo     *repositories.CircuitRepository
	userRepo *repositories.UserRepository
	db       *gorm.DB
}

// NewCircuitService creates a new CircuitService
func NewCircuitService(db *gorm.DB) *CircuitService {
	return &CircuitService{
		repo:     repositories.NewCircuitRepository(db),
		userRepo: repositories.NewUserRepository(db),
		db:       db,
	}
}

// ListCircuits lists user's circuits
func (s *CircuitService) ListCircuits(userID string, req dto.CircuitListRequest) ([]dto.CircuitResponse, dto.PaginationResponse, error) {
	filter := repositories.CircuitFilter{
		UserID:    userID,
		ProjectID: req.ProjectID,
		Search:    req.Search,
	}

	circuits, total, err := s.repo.FindAll(filter, req.Page, req.Limit, req.Sort, req.Order)
	if err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	responses := make([]dto.CircuitResponse, len(circuits))
	for i, c := range circuits {
		responses[i] = s.toCircuitResponse(&c)
	}

	return responses, dto.PaginationResponse{
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      int(total),
		TotalPages: (int(total) + req.Limit - 1) / req.Limit,
	}, nil
}

// GetCircuit gets circuit by ID
func (s *CircuitService) GetCircuit(circuitID, userID string) (*dto.CircuitDetailResponse, error) {
	circuit, err := s.repo.FindByID(circuitID)
	if err != nil {
		return nil, errors.New("circuit not found")
	}

	// Check access
	if circuit.UserID != userID && !circuit.IsPublic {
		return nil, errors.New("access denied")
	}

	return s.toCircuitDetailResponse(circuit), nil
}

// CreateCircuit creates a new circuit
func (s *CircuitService) CreateCircuit(userID string, req dto.CreateCircuitRequest) (*dto.CircuitResponse, int, error) {
	// Parse schema to count components and wires
	componentsCount, wiresCount := s.countSchemaElements(req.SchemaData)

	circuit := &models.Circuit{
		UserID:          userID,
		ProjectID:       req.ProjectID,
		Name:            req.Name,
		Description:     req.Description,
		SchemaData:      req.SchemaData,
		ComponentsCount: componentsCount,
		WiresCount:      wiresCount,
	}

	if err := s.repo.Create(circuit); err != nil {
		return nil, 0, err
	}

	// Award XP for first circuit
	xpEarned := 10 // Base XP for creating circuit
	s.awardXP(userID, xpEarned, circuit.ID, "circuit_create")

	return s.toCircuitResponsePtr(circuit), xpEarned, nil
}

// UpdateCircuit updates a circuit
func (s *CircuitService) UpdateCircuit(circuitID, userID string, req dto.UpdateCircuitRequest) (*dto.CircuitResponse, error) {
	circuit, err := s.repo.FindByID(circuitID)
	if err != nil {
		return nil, errors.New("circuit not found")
	}

	if circuit.UserID != userID {
		return nil, errors.New("access denied")
	}

	if req.Name != "" {
		circuit.Name = req.Name
	}
	if req.Description != "" {
		circuit.Description = req.Description
	}
	if req.SchemaData != nil {
		circuit.SchemaData = req.SchemaData
		circuit.ComponentsCount, circuit.WiresCount = s.countSchemaElements(req.SchemaData)
	}
	if req.IsPublic != nil {
		circuit.IsPublic = *req.IsPublic
	}

	if err := s.repo.Update(circuit); err != nil {
		return nil, err
	}

	return s.toCircuitResponsePtr(circuit), nil
}

// DeleteCircuit deletes a circuit
func (s *CircuitService) DeleteCircuit(circuitID, userID string) error {
	isOwner, err := s.repo.IsOwner(circuitID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("access denied")
	}

	return s.repo.Delete(circuitID)
}

// DuplicateCircuit duplicates a circuit
func (s *CircuitService) DuplicateCircuit(circuitID, userID string) (*dto.CircuitResponse, error) {
	original, err := s.repo.FindByID(circuitID)
	if err != nil {
		return nil, errors.New("circuit not found")
	}

	// Check access
	if original.UserID != userID && !original.IsPublic {
		return nil, errors.New("access denied")
	}

	newCircuit := &models.Circuit{
		UserID:          userID,
		ProjectID:       original.ProjectID,
		Name:            original.Name + " (Copy)",
		Description:     original.Description,
		SchemaData:      original.SchemaData,
		ComponentsCount: original.ComponentsCount,
		WiresCount:      original.WiresCount,
		IsPublic:        false,
	}

	if err := s.repo.Create(newCircuit); err != nil {
		return nil, err
	}

	return s.toCircuitResponsePtr(newCircuit), nil
}

// UploadThumbnail uploads circuit thumbnail
func (s *CircuitService) UploadThumbnail(circuitID, userID string, base64Data string) (string, error) {
	circuit, err := s.repo.FindByID(circuitID)
	if err != nil {
		return "", errors.New("circuit not found")
	}

	if circuit.UserID != userID {
		return "", errors.New("access denied")
	}

	// Decode base64 image
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", errors.New("invalid image data")
	}

	// Save to storage
	filename := fmt.Sprintf("circuit_%s_thumb.png", circuitID)
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	filepath := fmt.Sprintf("%s/circuits/%s", uploadDir, filename)

	if err := os.MkdirAll(fmt.Sprintf("%s/circuits", uploadDir), 0755); err != nil {
		return "", err
	}

	if err := os.WriteFile(filepath, imageData, 0644); err != nil {
		return "", err
	}

	// Update circuit
	baseURL := os.Getenv("APP_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8005"
	}
	thumbnailURL := fmt.Sprintf("%s/%s", baseURL, filepath)
	circuit.ThumbnailURL = thumbnailURL

	if err := s.repo.Update(circuit); err != nil {
		return "", err
	}

	return thumbnailURL, nil
}

// ListTemplates lists circuit templates
func (s *CircuitService) ListTemplates(req dto.CircuitTemplateListRequest) ([]dto.CircuitTemplateResponse, dto.PaginationResponse, error) {
	templates, total, err := s.repo.FindTemplates(req.Category, req.Difficulty, req.Search, req.Page, req.Limit)
	if err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	responses := make([]dto.CircuitTemplateResponse, len(templates))
	for i, t := range templates {
		responses[i] = dto.CircuitTemplateResponse{
			ID:                   t.ID,
			Name:                 t.Name,
			Description:          t.Description,
			Category:             t.Category,
			Difficulty:           t.Difficulty,
			ThumbnailURL:         t.ThumbnailURL,
			EstimatedTimeMinutes: t.EstimatedTimeMinutes,
			XPReward:             t.XPReward,
			Tags:                 t.Tags,
			UseCount:             t.UseCount,
			CreatedAt:            t.CreatedAt,
		}
	}

	return responses, dto.PaginationResponse{
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      int(total),
		TotalPages: (int(total) + req.Limit - 1) / req.Limit,
	}, nil
}

// UseTemplate clones a template to user's circuits
func (s *CircuitService) UseTemplate(templateID, userID string, req dto.UseTemplateRequest) (*dto.CircuitResponse, int, error) {
	template, err := s.repo.FindTemplateByID(templateID)
	if err != nil {
		return nil, 0, errors.New("template not found")
	}

	name := req.Name
	if name == "" {
		name = template.Name
	}

	circuit := &models.Circuit{
		UserID:          userID,
		ProjectID:       req.ProjectID,
		Name:            name,
		Description:     template.Description,
		SchemaData:      template.SchemaData,
		ComponentsCount: 0, // Will be calculated
		WiresCount:      0,
	}
	circuit.ComponentsCount, circuit.WiresCount = s.countSchemaElements(template.SchemaData)

	if err := s.repo.Create(circuit); err != nil {
		return nil, 0, err
	}

	// Increment template use count
	s.repo.IncrementTemplateUseCount(templateID)

	// Award XP
	xpEarned := 5
	s.awardXP(userID, xpEarned, circuit.ID, "template_use")

	return s.toCircuitResponsePtr(circuit), xpEarned, nil
}

// ExportCircuit exports circuit in different formats
func (s *CircuitService) ExportCircuit(circuitID, userID, format string) (*dto.CircuitExportResponse, error) {
	circuit, err := s.repo.FindByID(circuitID)
	if err != nil {
		return nil, errors.New("circuit not found")
	}

	if circuit.UserID != userID && !circuit.IsPublic {
		return nil, errors.New("access denied")
	}

	filename := strings.ReplaceAll(strings.ToLower(circuit.Name), " ", "_")

	switch format {
	case "json":
		content, _ := json.MarshalIndent(circuit.SchemaData, "", "  ")
		return &dto.CircuitExportResponse{
			Format:   "json",
			Filename: filename + ".json",
			Content:  string(content),
		}, nil

	case "spice":
		spiceContent := s.convertToSPICE(circuit)
		return &dto.CircuitExportResponse{
			Format:   "spice",
			Filename: filename + ".cir",
			Content:  spiceContent,
		}, nil

	default:
		return nil, errors.New("unsupported format")
	}
}

// ============================================
// Helper Functions
// ============================================

func (s *CircuitService) countSchemaElements(schemaData datatypes.JSON) (components, wires int) {
	var schema struct {
		Components []interface{} `json:"components"`
		Wires      []interface{} `json:"wires"`
	}
	json.Unmarshal(schemaData, &schema)
	return len(schema.Components), len(schema.Wires)
}

func (s *CircuitService) toCircuitResponse(c *models.Circuit) dto.CircuitResponse {
	resp := dto.CircuitResponse{
		ID:              c.ID,
		Name:            c.Name,
		Description:     c.Description,
		ThumbnailURL:    c.ThumbnailURL,
		ComponentsCount: c.ComponentsCount,
		WiresCount:      c.WiresCount,
		ProjectID:       c.ProjectID,
		IsTemplate:      c.IsTemplate,
		IsPublic:        c.IsPublic,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}

	if c.Project != nil {
		resp.ProjectName = c.Project.Name
	}

	return resp
}

func (s *CircuitService) toCircuitResponsePtr(c *models.Circuit) *dto.CircuitResponse {
	resp := s.toCircuitResponse(c)
	return &resp
}

func (s *CircuitService) toCircuitDetailResponse(c *models.Circuit) *dto.CircuitDetailResponse {
	return &dto.CircuitDetailResponse{
		CircuitResponse: s.toCircuitResponse(c),
		SchemaData:      c.SchemaData,
	}
}

func (s *CircuitService) awardXP(userID string, xpAmount int, sourceID, description string) {
	tx := &models.UserXPTransaction{
		UserID:      userID,
		XPAmount:    xpAmount,
		XPType:      models.XPSourceProject,
		SourceID:    &sourceID,
		SourceType:  "circuit",
		Description: description,
	}
	s.db.Create(tx)

	s.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"current_xp": gorm.Expr("current_xp + ?", xpAmount),
			"total_xp":   gorm.Expr("total_xp + ?", xpAmount),
		})
}

func (s *CircuitService) convertToSPICE(circuit *models.Circuit) string {
	// Basic SPICE netlist generation
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("* %s\n", circuit.Name))
	sb.WriteString("* Generated by NexFlux Circuit Simulator\n\n")

	var schema struct {
		Components []struct {
			ID         string                 `json:"id"`
			Type       string                 `json:"type"`
			Name       string                 `json:"name"`
			Properties map[string]interface{} `json:"properties"`
		} `json:"components"`
	}
	json.Unmarshal(circuit.SchemaData, &schema)

	nodeCounter := 1
	for _, comp := range schema.Components {
		switch comp.Type {
		case "resistor":
			if r, ok := comp.Properties["resistance"].(float64); ok {
				sb.WriteString(fmt.Sprintf("R%s %d %d %.2f\n", comp.Name, nodeCounter, nodeCounter+1, r))
				nodeCounter++
			}
		case "capacitor":
			if c, ok := comp.Properties["capacitance"].(float64); ok {
				sb.WriteString(fmt.Sprintf("C%s %d %d %.9f\n", comp.Name, nodeCounter, nodeCounter+1, c))
				nodeCounter++
			}
		case "power_source":
			if v, ok := comp.Properties["voltage"].(float64); ok {
				sb.WriteString(fmt.Sprintf("V%s %d 0 DC %.2f\n", comp.Name, nodeCounter, v))
			}
		case "led":
			sb.WriteString(fmt.Sprintf("D%s %d %d LED\n", comp.Name, nodeCounter, nodeCounter+1))
			nodeCounter++
		}
	}

	sb.WriteString("\n.model LED D(Is=1e-20 N=1.5)\n")
	sb.WriteString(".end\n")

	return sb.String()
}

// Placeholder for storage usage
var _ = storage.BucketAvatars
