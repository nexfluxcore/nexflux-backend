package database

import (
	"fmt"
	"log"
	"nexfi-backend/models"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Construct DSN from environment variables
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	config := &gorm.Config{}

	// Enable GORM debug mode if DB_DEBUG is true
	if os.Getenv("DB_DEBUG") == "true" {
		config.Logger = logger.Default.LogMode(logger.Info)
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	log.Println("✅ Database connected successfully")

	// Setup database extensions
	setupExtensions()

	// Auto migrate models if DB_AUTO_MIGRATE is true
	if os.Getenv("DB_AUTO_MIGRATE") == "true" {
		RunMigrations()
	}
}

// setupExtensions ensures required PostgreSQL extensions are enabled
func setupExtensions() {
	var hasUUID bool
	err := DB.Raw("SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'uuid-ossp')").Scan(&hasUUID).Error
	if err != nil {
		panic("Failed to check uuid-ossp extension: " + err.Error())
	}

	if !hasUUID {
		log.Println("⚠️ WARNING: uuid-ossp extension is not installed.")
		log.Println("Run: CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
		panic("Database needs uuid-ossp extension to be installed by a superuser")
	}
}

// RunMigrations runs all database migrations in order
func RunMigrations() {
	log.Println("🔄 Running database migrations...")

	// Migration order is important due to foreign key relationships
	migrations := []struct {
		Name   string
		Models []interface{}
	}{
		{
			Name: "Users Module",
			Models: []interface{}{
				&models.User{},
				&models.UserSettings{},
				&models.UserSession{},
				&models.UserStreak{},
			},
		},
		{
			Name: "Components Module",
			Models: []interface{}{
				&models.ComponentCategory{},
				&models.Component{},
				&models.ComponentRequest{},
				&models.UserFavoriteComponent{},
			},
		},
		{
			Name: "Projects Module",
			Models: []interface{}{
				&models.Project{},
				&models.ProjectCollaborator{},
				&models.ProjectComponent{},
			},
		},
		{
			Name: "Challenges Module",
			Models: []interface{}{
				&models.Challenge{},
				&models.ChallengeProgress{},
				&models.DailyChallenge{},
			},
		},
		{
			Name: "Notifications Module",
			Models: []interface{}{
				&models.Notification{},
			},
		},
		{
			Name: "Gamification Module",
			Models: []interface{}{
				&models.Achievement{},
				&models.UserAchievement{},
				&models.Leaderboard{},
				&models.LeaderboardEntry{},
			},
		},
		{
			Name: "Documentation Module",
			Models: []interface{}{
				&models.DocCategory{},
				&models.DocArticle{},
				&models.DocVideo{},
			},
		},
		{
			Name: "Auth Tokens",
			Models: []interface{}{
				&models.PasswordResetToken{},
				&models.EmailVerificationToken{},
			},
		},
		{
			Name: "Virtual Lab Module",
			Models: []interface{}{
				&models.Lab{},
				&models.LabSession{},
				&models.LabBooking{},
				&models.LabQueue{},
				&models.LabHardwareLog{},
				&models.CodeCompilation{},
			},
		},
	}

	for _, migration := range migrations {
		log.Printf("  → Migrating %s...", migration.Name)
		if err := DB.AutoMigrate(migration.Models...); err != nil {
			panic(fmt.Sprintf("Failed to migrate %s: %v", migration.Name, err))
		}
	}

	// Create indexes after migration
	createIndexes()

	log.Println("✅ Database migrations completed")
}

// createIndexes creates additional indexes for performance
func createIndexes() {
	log.Println("  → Creating indexes...")

	indexes := []string{
		// User indexes
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)",
		"CREATE INDEX IF NOT EXISTS idx_users_provider ON users(provider, provider_id)",

		// Project indexes
		"CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_projects_is_public ON projects(is_public)",
		"CREATE INDEX IF NOT EXISTS idx_projects_difficulty ON projects(difficulty)",

		// Component indexes
		"CREATE INDEX IF NOT EXISTS idx_components_category ON components(category_id)",
		"CREATE INDEX IF NOT EXISTS idx_components_is_active ON components(is_active)",

		// Challenge indexes
		"CREATE INDEX IF NOT EXISTS idx_challenges_type ON challenges(type)",
		"CREATE INDEX IF NOT EXISTS idx_challenges_difficulty ON challenges(difficulty)",
		"CREATE INDEX IF NOT EXISTS idx_challenges_is_active ON challenges(is_active)",

		// Challenge progress indexes
		"CREATE INDEX IF NOT EXISTS idx_challenge_progress_user ON challenge_progress(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_challenge_progress_status ON challenge_progress(status)",

		// Notification indexes
		"CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications(user_id, is_read)",

		// Leaderboard indexes
		"CREATE INDEX IF NOT EXISTS idx_leaderboard_entries_rank ON leaderboard_entries(leaderboard_id, rank)",

		// Lab indexes
		"CREATE INDEX IF NOT EXISTS idx_labs_status ON labs(status)",
		"CREATE INDEX IF NOT EXISTS idx_labs_platform ON labs(platform)",
		"CREATE INDEX IF NOT EXISTS idx_labs_slug ON labs(slug)",
		"CREATE INDEX IF NOT EXISTS idx_labs_current_user ON labs(current_user_id)",

		// Lab session indexes
		"CREATE INDEX IF NOT EXISTS idx_lab_sessions_lab ON lab_sessions(lab_id)",
		"CREATE INDEX IF NOT EXISTS idx_lab_sessions_user ON lab_sessions(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_lab_sessions_status ON lab_sessions(status)",
		"CREATE INDEX IF NOT EXISTS idx_lab_sessions_active ON lab_sessions(user_id, status) WHERE status = 'active'",

		// Lab queue indexes
		"CREATE INDEX IF NOT EXISTS idx_lab_queue_lab ON lab_queue(lab_id)",
		"CREATE INDEX IF NOT EXISTS idx_lab_queue_user ON lab_queue(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_lab_queue_position ON lab_queue(lab_id, position)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_lab_queue_unique ON lab_queue(lab_id, user_id)",

		// Lab booking indexes
		"CREATE INDEX IF NOT EXISTS idx_lab_bookings_lab ON lab_bookings(lab_id)",
		"CREATE INDEX IF NOT EXISTS idx_lab_bookings_user ON lab_bookings(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_lab_bookings_scheduled ON lab_bookings(scheduled_at)",

		// Lab hardware log indexes
		"CREATE INDEX IF NOT EXISTS idx_lab_hardware_logs_lab ON lab_hardware_logs(lab_id)",
		"CREATE INDEX IF NOT EXISTS idx_lab_hardware_logs_session ON lab_hardware_logs(session_id)",
		"CREATE INDEX IF NOT EXISTS idx_lab_hardware_logs_type ON lab_hardware_logs(event_type)",

		// Code compilation indexes
		"CREATE INDEX IF NOT EXISTS idx_code_compilations_session ON code_compilations(session_id)",
		"CREATE INDEX IF NOT EXISTS idx_code_compilations_status ON code_compilations(status)",
	}

	for _, idx := range indexes {
		if err := DB.Exec(idx).Error; err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
