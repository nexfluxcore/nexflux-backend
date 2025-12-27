package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"nexfi-backend/api/repositories"
	"nexfi-backend/dto"
	"nexfi-backend/models"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mssola/user_agent"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SecurityService handles security-related business logic
type SecurityService struct {
	repo     *repositories.SecurityRepository
	userRepo *repositories.UserRepository
	db       *gorm.DB
}

// NewSecurityService creates a new SecurityService
func NewSecurityService(db *gorm.DB) *SecurityService {
	return &SecurityService{
		repo:     repositories.NewSecurityRepository(db),
		userRepo: repositories.NewUserRepository(db),
		db:       db,
	}
}

// ============================================
// Password Management
// ============================================

// ChangePassword changes user's password
func (s *SecurityService) ChangePassword(userID string, req dto.ChangePasswordRequest, ipAddress, userAgent string) error {
	// Get user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	// Check if new password is same as current
	if req.CurrentPassword == req.NewPassword {
		return errors.New("new password cannot be the same as current password")
	}

	// Check password history (prevent reuse of last N passwords)
	historyCount := getEnvInt("PASSWORD_HISTORY_COUNT", 5)
	history, _ := s.repo.GetPasswordHistory(userID, historyCount)
	for _, h := range history {
		if err := bcrypt.CompareHashAndPassword([]byte(h.PasswordHash), []byte(req.NewPassword)); err == nil {
			return errors.New("password has been used recently, please choose a different one")
		}
	}

	// Hash new password
	bcryptCost := getEnvInt("PASSWORD_BCRYPT_COST", 12)
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcryptCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Update password
	now := time.Now()
	user.PasswordHash = string(newHash)
	user.PasswordChangedAt = &now
	if err := s.userRepo.Update(user); err != nil {
		return errors.New("failed to update password")
	}

	// Add old password to history
	s.repo.CreatePasswordHistory(&models.PasswordHistory{
		UserID:       userID,
		PasswordHash: string(newHash),
	})

	// Log security event
	s.logSecurityEvent(userID, models.SecurityEventPasswordChange, ipAddress, userAgent, nil)

	return nil
}

// ============================================
// Two-Factor Authentication
// ============================================

// Enable2FA initiates 2FA setup
func (s *SecurityService) Enable2FA(userID, email string) (*dto.Enable2FAResponse, error) {
	// Check if already enabled
	existing, _ := s.repo.GetUser2FA(userID)
	if existing != nil && existing.IsEnabled {
		return nil, errors.New("two-factor authentication is already enabled")
	}

	// Generate TOTP secret
	issuer := getEnvStr("TOTP_ISSUER", "NexFlux")
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: email,
	})
	if err != nil {
		return nil, errors.New("failed to generate 2FA secret")
	}

	// Generate backup codes
	backupCodes := generateBackupCodes(getEnvInt("BACKUP_CODES_COUNT", 5))

	// Save or update 2FA settings (not enabled yet - waiting for verification)
	backupCodesJSON, _ := json.Marshal(backupCodes)
	twoFA := &models.User2FA{
		UserID:      userID,
		SecretKey:   key.Secret(),
		IsEnabled:   false,
		BackupCodes: datatypes.JSON(backupCodesJSON),
	}

	if existing != nil {
		twoFA.ID = existing.ID
		s.repo.UpdateUser2FA(twoFA)
	} else {
		s.repo.CreateUser2FA(twoFA)
	}

	return &dto.Enable2FAResponse{
		Secret:      key.Secret(),
		QRCodeURL:   key.URL(),
		BackupCodes: backupCodes,
	}, nil
}

// Verify2FA verifies and activates 2FA
func (s *SecurityService) Verify2FA(userID, code, ipAddress, userAgent string) error {
	twoFA, err := s.repo.GetUser2FA(userID)
	if err != nil {
		return errors.New("2FA setup not found, please enable 2FA first")
	}

	if twoFA.IsEnabled {
		return errors.New("2FA is already verified and active")
	}

	// Verify TOTP code
	if !totp.Validate(code, twoFA.SecretKey) {
		return errors.New("invalid verification code")
	}

	// Activate 2FA
	now := time.Now()
	twoFA.IsEnabled = true
	twoFA.EnabledAt = &now
	if err := s.repo.UpdateUser2FA(twoFA); err != nil {
		return errors.New("failed to activate 2FA")
	}

	// Update user
	s.db.Model(&models.User{}).Where("id = ?", userID).Update("two_factor_enabled", true)

	// Log security event
	s.logSecurityEvent(userID, models.SecurityEvent2FAEnable, ipAddress, userAgent, nil)

	return nil
}

// Disable2FA disables 2FA
func (s *SecurityService) Disable2FA(userID string, req dto.Disable2FARequest, ipAddress, userAgent string) error {
	// Get user and verify password
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return errors.New("password is incorrect")
	}

	// Get 2FA settings
	twoFA, err := s.repo.GetUser2FA(userID)
	if err != nil || !twoFA.IsEnabled {
		return errors.New("2FA is not enabled")
	}

	// Verify code (TOTP or backup code)
	if !s.verify2FACode(twoFA, req.Code) {
		return errors.New("invalid code")
	}

	// Disable 2FA
	if err := s.repo.DeleteUser2FA(userID); err != nil {
		return errors.New("failed to disable 2FA")
	}

	// Update user
	s.db.Model(&models.User{}).Where("id = ?", userID).Update("two_factor_enabled", false)

	// Log security event
	s.logSecurityEvent(userID, models.SecurityEvent2FADisable, ipAddress, userAgent, nil)

	return nil
}

// Get2FAStatus gets 2FA status for a user
func (s *SecurityService) Get2FAStatus(userID string) (*dto.TwoFAStatusResponse, error) {
	twoFA, err := s.repo.GetUser2FA(userID)
	if err != nil {
		// 2FA not set up
		return &dto.TwoFAStatusResponse{
			Enabled:              false,
			BackupCodesRemaining: 0,
		}, nil
	}

	// Count remaining backup codes
	backupCodesRemaining := 0
	if twoFA.BackupCodes != nil {
		var codes []string
		json.Unmarshal(twoFA.BackupCodes, &codes)
		backupCodesRemaining = len(codes)
	}

	return &dto.TwoFAStatusResponse{
		Enabled:              twoFA.IsEnabled,
		EnabledAt:            twoFA.EnabledAt,
		BackupCodesRemaining: backupCodesRemaining,
	}, nil
}

// Validate2FACode validates a 2FA code (for login)
func (s *SecurityService) Validate2FACode(userID, code string) (bool, error) {
	twoFA, err := s.repo.GetUser2FA(userID)
	if err != nil || !twoFA.IsEnabled {
		return false, errors.New("2FA is not enabled")
	}

	return s.verify2FACode(twoFA, code), nil
}

// ============================================
// Session Management
// ============================================

// CreateSession creates a new session
func (s *SecurityService) CreateSession(userID, token, ipAddress, userAgentStr string) (*models.UserSession, error) {
	// Parse user agent
	device, browser := parseUserAgent(userAgentStr)

	// Hash token
	tokenHash := hashToken(token)

	// Get location from IP (simplified - in production use GeoIP)
	location := getLocationFromIP(ipAddress)

	// Calculate expiration
	sessionMaxAge := getEnvDuration("SESSION_MAX_AGE", 7*24*time.Hour)

	session := &models.UserSession{
		UserID:       userID,
		TokenHash:    tokenHash,
		Device:       device,
		Browser:      browser,
		IPAddress:    ipAddress,
		Location:     location,
		UserAgent:    userAgentStr,
		LastActiveAt: time.Now(),
		ExpiresAt:    time.Now().Add(sessionMaxAge),
	}

	if err := s.repo.CreateSession(session); err != nil {
		return nil, err
	}

	// Check max sessions limit
	s.enforceMaxSessions(userID)

	return session, nil
}

// GetSessions gets all sessions for a user
func (s *SecurityService) GetSessions(userID, currentSessionID string) ([]dto.SecuritySessionResponse, error) {
	sessions, err := s.repo.GetUserSessions(userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.SecuritySessionResponse, len(sessions))
	for i, session := range sessions {
		responses[i] = dto.SecuritySessionResponse{
			ID:         session.ID,
			Device:     session.Device,
			Browser:    session.Browser,
			IPAddress:  session.IPAddress,
			Location:   session.Location,
			LastActive: session.LastActiveAt,
			IsCurrent:  session.ID == currentSessionID,
			CreatedAt:  session.CreatedAt,
		}
	}

	return responses, nil
}

// RevokeSession revokes a specific session
func (s *SecurityService) RevokeSession(userID, sessionID, ipAddress, userAgent string) error {
	if err := s.repo.RevokeSession(sessionID, userID); err != nil {
		return errors.New("failed to revoke session")
	}

	s.logSecurityEvent(userID, models.SecurityEventSessionRevoke, ipAddress, userAgent, map[string]interface{}{
		"session_id": sessionID,
	})

	return nil
}

// RevokeAllSessions revokes all sessions except current
func (s *SecurityService) RevokeAllSessions(userID, exceptSessionID, ipAddress, userAgent string) (int, error) {
	count, err := s.repo.RevokeAllSessionsExcept(userID, exceptSessionID)
	if err != nil {
		return 0, errors.New("failed to revoke sessions")
	}

	s.logSecurityEvent(userID, models.SecurityEventSessionRevokeAll, ipAddress, userAgent, map[string]interface{}{
		"revoked_count": count,
	})

	return int(count), nil
}

// UpdateSessionActivity updates session last active time
func (s *SecurityService) UpdateSessionActivity(sessionID string) {
	s.repo.UpdateSessionActivity(sessionID)
}

// ============================================
// Login History
// ============================================

// RecordLoginAttempt records a login attempt
func (s *SecurityService) RecordLoginAttempt(userID *string, email, ipAddress, userAgentStr, status, failureReason string) {
	device, browser := parseUserAgent(userAgentStr)
	location := getLocationFromIP(ipAddress)

	history := &models.LoginHistory{
		UserID:    userID,
		Email:     email,
		Device:    device,
		Browser:   browser,
		IPAddress: ipAddress,
		Location:  location,
		UserAgent: userAgentStr,
		Status:    status,
	}

	if failureReason != "" {
		history.FailureReason = &failureReason
	}

	s.repo.CreateLoginHistory(history)
}

// GetLoginHistory gets login history for a user
func (s *SecurityService) GetLoginHistory(userID string, req dto.LoginHistoryRequest) ([]dto.LoginHistoryResponse, dto.PaginationResponse, error) {
	history, total, err := s.repo.GetLoginHistory(userID, req.Page, req.Limit)
	if err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	responses := make([]dto.LoginHistoryResponse, len(history))
	for i, h := range history {
		responses[i] = dto.LoginHistoryResponse{
			ID:            h.ID,
			Device:        h.Device,
			Browser:       h.Browser,
			IPAddress:     h.IPAddress,
			Location:      h.Location,
			Status:        h.Status,
			CreatedAt:     h.CreatedAt,
			FailureReason: h.FailureReason,
		}
	}

	return responses, dto.PaginationResponse{
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      int(total),
		TotalPages: (int(total) + req.Limit - 1) / req.Limit,
	}, nil
}

// CheckAccountLockout checks if account should be locked
func (s *SecurityService) CheckAccountLockout(email string) (bool, *time.Time, error) {
	// Check recent failed attempts
	since := time.Now().Add(-15 * time.Minute)
	failedAttempts, _ := s.repo.GetRecentFailedAttempts(email, since)

	maxAttempts := int64(getEnvInt("LOGIN_MAX_ATTEMPTS", 5))
	if failedAttempts >= maxAttempts {
		lockoutDuration := 30 * time.Minute
		lockedUntil := time.Now().Add(lockoutDuration)
		return true, &lockedUntil, nil
	}

	return false, nil, nil
}

// GetSecuritySettings gets overall security settings for a user
func (s *SecurityService) GetSecuritySettings(userID string) (*dto.SecuritySettingsResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Count active sessions
	sessionCount, _ := s.repo.CountUserSessions(userID)

	// Count recent login attempts (last 24 hours)
	since := time.Now().Add(-24 * time.Hour)
	recentAttempts, _ := s.repo.GetRecentFailedAttempts(user.Email, since)

	// Get last successful login
	var lastLoginAt *time.Time
	var lastLoginIP, lastLoginLocation string
	lastLogin, err := s.repo.GetLastSuccessfulLogin(userID)
	if err == nil {
		lastLoginAt = &lastLogin.CreatedAt
		lastLoginIP = lastLogin.IPAddress
		lastLoginLocation = lastLogin.Location
	}

	return &dto.SecuritySettingsResponse{
		TwoFactorEnabled:    user.TwoFactorEnabled,
		PasswordChangedAt:   user.PasswordChangedAt,
		ActiveSessionsCount: int(sessionCount),
		RecentLoginAttempts: int(recentAttempts),
		LastLoginAt:         lastLoginAt,
		LastLoginIP:         lastLoginIP,
		LastLoginLocation:   lastLoginLocation,
		AccountLocked:       user.LockedUntil != nil && user.LockedUntil.After(time.Now()),
		LockedUntil:         user.LockedUntil,
	}, nil
}

// ============================================
// Helper Functions
// ============================================

func (s *SecurityService) verify2FACode(twoFA *models.User2FA, code string) bool {
	// Try TOTP first
	if totp.Validate(code, twoFA.SecretKey) {
		return true
	}

	// Try backup codes
	if twoFA.BackupCodes != nil {
		var codes []string
		json.Unmarshal(twoFA.BackupCodes, &codes)

		for i, c := range codes {
			if c == code {
				// Remove used backup code
				codes = append(codes[:i], codes[i+1:]...)
				newCodes, _ := json.Marshal(codes)
				twoFA.BackupCodes = datatypes.JSON(newCodes)
				s.repo.UpdateUser2FA(twoFA)
				return true
			}
		}
	}

	return false
}

func (s *SecurityService) enforceMaxSessions(userID string) {
	maxSessions := getEnvInt("SESSION_MAX_DEVICES", 5)
	sessions, _ := s.repo.GetUserSessions(userID)

	if len(sessions) > maxSessions {
		// Remove oldest sessions
		for i := maxSessions; i < len(sessions); i++ {
			s.repo.RevokeSession(sessions[i].ID, userID)
		}
	}
}

func (s *SecurityService) logSecurityEvent(userID, eventType, ipAddress, userAgent string, details map[string]interface{}) {
	detailsJSON, _ := json.Marshal(details)
	s.repo.CreateSecurityLog(&models.SecurityLog{
		UserID:    userID,
		EventType: eventType,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Details:   datatypes.JSON(detailsJSON),
	})
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func parseUserAgent(ua string) (device, browser string) {
	parser := user_agent.New(ua)

	device = "Unknown Device"
	if parser.Mobile() {
		device = "Mobile Device"
	} else if parser.Bot() {
		device = "Bot"
	} else {
		os := parser.OS()
		if os != "" {
			device = os
		}
	}

	browserName, browserVersion := parser.Browser()
	if browserName != "" {
		browser = fmt.Sprintf("%s %s", browserName, browserVersion)
	} else {
		browser = "Unknown Browser"
	}

	return
}

func getLocationFromIP(ip string) string {
	// Simplified - in production use MaxMind GeoIP2 or similar service
	// For now, return a placeholder
	if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") || ip == "127.0.0.1" || ip == "::1" {
		return "Local Network"
	}
	return "Unknown Location"
}

func generateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		bytes := make([]byte, 5)
		rand.Read(bytes)
		codes[i] = strings.ToUpper(base32.StdEncoding.EncodeToString(bytes)[:8])
	}
	return codes
}

func getEnvStr(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}
