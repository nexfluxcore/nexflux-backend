package livekit

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

var (
	apiKey    string
	apiSecret string
	host      string
	roomSvc   *lksdk.RoomServiceClient
)

// Config holds LiveKit configuration
type Config struct {
	URL       string
	APIKey    string
	APISecret string
}

// GetConfig returns LiveKit configuration from environment
func GetConfig() Config {
	return Config{
		URL:       getEnv("LIVEKIT_URL", "wss://livekit.nexflux.io"),
		APIKey:    os.Getenv("LIVEKIT_API_KEY"),
		APISecret: os.Getenv("LIVEKIT_API_SECRET"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// InitLiveKit initializes the LiveKit client
func InitLiveKit() error {
	config := GetConfig()

	if config.APIKey == "" || config.APISecret == "" {
		return fmt.Errorf("LIVEKIT_API_KEY and LIVEKIT_API_SECRET must be set")
	}

	apiKey = config.APIKey
	apiSecret = config.APISecret
	host = config.URL

	// Initialize room service client
	roomSvc = lksdk.NewRoomServiceClient(host, apiKey, apiSecret)

	log.Printf("✅ LiveKit client initialized with host: %s", host)
	return nil
}

// GetHost returns the LiveKit host URL
func GetHost() string {
	return host
}

// ============================================
// Token Generation
// ============================================

// TokenOptions for generating access tokens
type TokenOptions struct {
	Identity string
	Name     string
	RoomName string
	// Permissions
	CanPublish     bool
	CanSubscribe   bool
	CanPublishData bool
	// Duration
	ValidFor time.Duration
	// Metadata
	Metadata string
}

// GenerateToken generates a LiveKit access token
func GenerateToken(opts TokenOptions) (string, error) {
	if apiKey == "" || apiSecret == "" {
		return "", fmt.Errorf("LiveKit not initialized")
	}

	at := auth.NewAccessToken(apiKey, apiSecret)

	grant := &auth.VideoGrant{
		Room:           opts.RoomName,
		RoomJoin:       true,
		CanPublish:     &opts.CanPublish,
		CanSubscribe:   &opts.CanSubscribe,
		CanPublishData: &opts.CanPublishData,
	}

	at.AddGrant(grant).
		SetIdentity(opts.Identity).
		SetName(opts.Name).
		SetValidFor(opts.ValidFor)

	if opts.Metadata != "" {
		at.SetMetadata(opts.Metadata)
	}

	return at.ToJWT()
}

// GenerateLabViewerToken generates a token for viewing a lab stream
func GenerateLabViewerToken(userID, userName, roomName string) (string, error) {
	return GenerateToken(TokenOptions{
		Identity:       userID,
		Name:           userName,
		RoomName:       roomName,
		CanPublish:     false, // Viewers can't publish
		CanSubscribe:   true,  // Can subscribe to streams
		CanPublishData: true,  // Can send data messages (for actuator control)
		ValidFor:       time.Hour,
	})
}

// GenerateLabAgentToken generates a token for a lab agent (hardware streamer)
func GenerateLabAgentToken(labID, roomName string) (string, error) {
	return GenerateToken(TokenOptions{
		Identity:       fmt.Sprintf("lab-agent-%s", labID),
		Name:           fmt.Sprintf("Lab Agent %s", labID),
		RoomName:       roomName,
		CanPublish:     true, // Agent publishes video/audio
		CanSubscribe:   true, // Agent subscribes to data messages
		CanPublishData: true, // Agent sends sensor data
		ValidFor:       24 * time.Hour,
	})
}

// ============================================
// Room Management
// ============================================

// CreateRoom creates a new LiveKit room
func CreateRoom(roomName string, maxParticipants uint32) (*livekit.Room, error) {
	if roomSvc == nil {
		return nil, fmt.Errorf("LiveKit not initialized")
	}

	room, err := roomSvc.CreateRoom(context.Background(), &livekit.CreateRoomRequest{
		Name:            roomName,
		EmptyTimeout:    300, // 5 minutes timeout when empty
		MaxParticipants: maxParticipants,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %v", err)
	}

	log.Printf("🎬 Created LiveKit room: %s", roomName)
	return room, nil
}

// DeleteRoom deletes a LiveKit room
func DeleteRoom(roomName string) error {
	if roomSvc == nil {
		return fmt.Errorf("LiveKit not initialized")
	}

	_, err := roomSvc.DeleteRoom(context.Background(), &livekit.DeleteRoomRequest{
		Room: roomName,
	})
	if err != nil {
		return fmt.Errorf("failed to delete room: %v", err)
	}

	log.Printf("🗑️ Deleted LiveKit room: %s", roomName)
	return nil
}

// ListRooms lists all LiveKit rooms
func ListRooms() ([]*livekit.Room, error) {
	if roomSvc == nil {
		return nil, fmt.Errorf("LiveKit not initialized")
	}

	resp, err := roomSvc.ListRooms(context.Background(), &livekit.ListRoomsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list rooms: %v", err)
	}

	return resp.Rooms, nil
}

// GetRoom gets a LiveKit room
func GetRoom(roomName string) (*livekit.Room, error) {
	if roomSvc == nil {
		return nil, fmt.Errorf("LiveKit not initialized")
	}

	rooms, err := roomSvc.ListRooms(context.Background(), &livekit.ListRoomsRequest{
		Names: []string{roomName},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %v", err)
	}

	if len(rooms.Rooms) == 0 {
		return nil, fmt.Errorf("room not found: %s", roomName)
	}

	return rooms.Rooms[0], nil
}

// ============================================
// Participant Management
// ============================================

// ListParticipants lists participants in a room
func ListParticipants(roomName string) ([]*livekit.ParticipantInfo, error) {
	if roomSvc == nil {
		return nil, fmt.Errorf("LiveKit not initialized")
	}

	resp, err := roomSvc.ListParticipants(context.Background(), &livekit.ListParticipantsRequest{
		Room: roomName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list participants: %v", err)
	}

	return resp.Participants, nil
}

// RemoveParticipant removes a participant from a room
func RemoveParticipant(roomName, identity string) error {
	if roomSvc == nil {
		return fmt.Errorf("LiveKit not initialized")
	}

	_, err := roomSvc.RemoveParticipant(context.Background(), &livekit.RoomParticipantIdentity{
		Room:     roomName,
		Identity: identity,
	})
	if err != nil {
		return fmt.Errorf("failed to remove participant: %v", err)
	}

	return nil
}

// ============================================
// Data Broadcasting
// ============================================

// SendData sends data to participants in a room
func SendData(roomName string, data []byte, destinationIdentities []string) error {
	if roomSvc == nil {
		return fmt.Errorf("LiveKit not initialized")
	}

	req := &livekit.SendDataRequest{
		Room: roomName,
		Data: data,
		Kind: livekit.DataPacket_RELIABLE,
	}

	if len(destinationIdentities) > 0 {
		req.DestinationIdentities = destinationIdentities
	}

	_, err := roomSvc.SendData(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to send data: %v", err)
	}

	return nil
}

// ============================================
// Lab Room Helpers
// ============================================

// BuildLabRoomName builds a standardized room name for a lab session
func BuildLabRoomName(labSlug, sessionID string) string {
	return fmt.Sprintf("lab-%s-session-%s", labSlug, sessionID)
}

// GetOrCreateLabRoom ensures a lab room exists
func GetOrCreateLabRoom(roomName string) (*livekit.Room, error) {
	room, err := GetRoom(roomName)
	if err == nil {
		return room, nil
	}

	// Room doesn't exist, create it
	return CreateRoom(roomName, 10) // Max 10 participants
}

// LabSessionInfo contains info for a lab session connection
type LabSessionInfo struct {
	RoomName   string `json:"room_name"`
	Token      string `json:"token"`
	LivekitURL string `json:"livekit_url"`
	AgentToken string `json:"agent_token,omitempty"` // Only for internal use
}

// CreateLabSession creates a new lab session with room and tokens
func CreateLabSession(labSlug, sessionID, userID, userName string) (*LabSessionInfo, error) {
	roomName := BuildLabRoomName(labSlug, sessionID)

	// Create or get the room
	_, err := GetOrCreateLabRoom(roomName)
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %v", err)
	}

	// Generate user token
	userToken, err := GenerateLabViewerToken(userID, userName, roomName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate user token: %v", err)
	}

	return &LabSessionInfo{
		RoomName:   roomName,
		Token:      userToken,
		LivekitURL: host,
	}, nil
}

// EndLabSession ends a lab session and cleans up
func EndLabSession(roomName string) error {
	// Remove all participants first
	participants, err := ListParticipants(roomName)
	if err == nil {
		for _, p := range participants {
			RemoveParticipant(roomName, p.Identity)
		}
	}

	// Delete the room
	return DeleteRoom(roomName)
}
