package repositories

import (
	"nexfi-backend/models"
	"time"

	"gorm.io/gorm"
)

// SimulationRepository handles simulation database operations
type SimulationRepository struct {
	*BaseRepository
}

// NewSimulationRepository creates a new SimulationRepository
func NewSimulationRepository(db *gorm.DB) *SimulationRepository {
	return &SimulationRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// SimulationFilter defines filter options
type SimulationFilter struct {
	UserID    string
	ProjectID string
	Status    string
	Type      string
	Search    string
}

// FindAll finds all simulations with filters
func (r *SimulationRepository) FindAll(filter SimulationFilter, page, limit int, sort, order string) ([]models.Simulation, int64, error) {
	var simulations []models.Simulation
	var total int64

	query := r.DB.Model(&models.Simulation{})

	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if filter.ProjectID != "" {
		query = query.Where("project_id = ?", filter.ProjectID)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}

	if filter.Search != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Search+"%")
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
		Find(&simulations).Error

	return simulations, total, err
}

// FindByID finds simulation by ID
func (r *SimulationRepository) FindByID(id string) (*models.Simulation, error) {
	var simulation models.Simulation
	err := r.DB.Preload("Project").Preload("User").First(&simulation, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &simulation, nil
}

// Create creates a new simulation
func (r *SimulationRepository) Create(simulation *models.Simulation) error {
	return r.DB.Create(simulation).Error
}

// Update updates a simulation
func (r *SimulationRepository) Update(simulation *models.Simulation) error {
	return r.DB.Save(simulation).Error
}

// Delete deletes a simulation
func (r *SimulationRepository) Delete(id string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// Delete runs first
		if err := tx.Delete(&models.SimulationRun{}, "simulation_id = ?", id).Error; err != nil {
			return err
		}
		// Delete simulation
		return tx.Delete(&models.Simulation{}, "id = ?", id).Error
	})
}

// IsOwner checks if user is owner
func (r *SimulationRepository) IsOwner(simulationID, userID string) (bool, error) {
	var count int64
	err := r.DB.Model(&models.Simulation{}).Where("id = ? AND user_id = ?", simulationID, userID).Count(&count).Error
	return count > 0, err
}

// GetStats gets simulation statistics for a user
func (r *SimulationRepository) GetStats(userID string) (*models.SimulationStats, error) {
	stats := &models.SimulationStats{
		ByType: make(map[string]int),
	}

	// Total simulations
	var totalSim int64
	r.DB.Model(&models.Simulation{}).Where("user_id = ?", userID).Count(&totalSim)
	stats.TotalSimulations = int(totalSim)

	// By status
	var statusCounts []struct {
		Status string
		Count  int
	}
	r.DB.Model(&models.Simulation{}).
		Select("status, count(*) as count").
		Where("user_id = ?", userID).
		Group("status").
		Scan(&statusCounts)

	for _, sc := range statusCounts {
		switch sc.Status {
		case string(models.SimStatusRunning):
			stats.RunningNow = sc.Count
		case string(models.SimStatusCompleted):
			stats.Completed = sc.Count
		case string(models.SimStatusPaused):
			stats.Paused = sc.Count
		case string(models.SimStatusError):
			stats.Error = sc.Count
		}
	}

	// Total runtime
	var totalMs int64
	r.DB.Model(&models.Simulation{}).Where("user_id = ?", userID).Select("COALESCE(SUM(total_runtime_ms), 0)").Scan(&totalMs)
	stats.TotalRuntimeHours = float64(totalMs) / 3600000.0

	// Success rate
	if stats.TotalSimulations > 0 {
		stats.SuccessRate = float64(stats.Completed) / float64(stats.TotalSimulations) * 100
	}

	// Simulations this week
	weekAgo := time.Now().AddDate(0, 0, -7)
	var weekCount int64
	r.DB.Model(&models.Simulation{}).Where("user_id = ? AND created_at >= ?", userID, weekAgo).Count(&weekCount)
	stats.SimulationsThisWeek = int(weekCount)

	// By type
	var typeCounts []struct {
		Type  string
		Count int
	}
	r.DB.Model(&models.Simulation{}).
		Select("type, count(*) as count").
		Where("user_id = ?", userID).
		Group("type").
		Scan(&typeCounts)

	for _, tc := range typeCounts {
		stats.ByType[tc.Type] = tc.Count
	}

	return stats, nil
}

// ============================================
// Simulation Run Repository
// ============================================

// CreateRun creates a new simulation run
func (r *SimulationRepository) CreateRun(run *models.SimulationRun) error {
	return r.DB.Create(run).Error
}

// UpdateRun updates a simulation run
func (r *SimulationRepository) UpdateRun(run *models.SimulationRun) error {
	return r.DB.Save(run).Error
}

// GetRuns gets run history for a simulation
func (r *SimulationRepository) GetRuns(simulationID string, limit int) ([]models.SimulationRun, error) {
	var runs []models.SimulationRun
	err := r.DB.Where("simulation_id = ?", simulationID).
		Order("started_at DESC").
		Limit(limit).
		Find(&runs).Error
	return runs, err
}

// GetLatestRun gets the latest run for a simulation
func (r *SimulationRepository) GetLatestRun(simulationID string) (*models.SimulationRun, error) {
	var run models.SimulationRun
	err := r.DB.Where("simulation_id = ?", simulationID).
		Order("started_at DESC").
		First(&run).Error
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// IncrementRunCount increments simulation run count
func (r *SimulationRepository) IncrementRunCount(simulationID string, durationMs int) error {
	now := time.Now()
	return r.DB.Model(&models.Simulation{}).Where("id = ?", simulationID).Updates(map[string]interface{}{
		"run_count":        gorm.Expr("run_count + 1"),
		"total_runtime_ms": gorm.Expr("total_runtime_ms + ?", durationMs),
		"last_run_at":      now,
	}).Error
}
