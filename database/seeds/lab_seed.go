package seeds

import (
	"encoding/json"
	"log"
	"nexfi-backend/models"

	"gorm.io/gorm"
)

// LabHardwareSpecs represents hardware specifications for a lab
type LabHardwareSpecs struct {
	Board     string         `json:"board"`
	Camera    string         `json:"camera"`
	Sensors   []SensorSpec   `json:"sensors"`
	Actuators []ActuatorSpec `json:"actuators"`
}

// SensorSpec for sensor specification
type SensorSpec struct {
	Name string            `json:"name"`
	Type string            `json:"type"`
	Pin  string            `json:"pin,omitempty"`
	Pins map[string]string `json:"pins,omitempty"`
}

// ActuatorSpec for actuator specification
type ActuatorSpec struct {
	Name string            `json:"name"`
	Type string            `json:"type"`
	Pin  string            `json:"pin,omitempty"`
	Pins map[string]string `json:"pins,omitempty"`
}

// SeedLabs seeds initial lab data
func SeedLabs(db *gorm.DB) error {
	// Check if labs already exist
	var count int64
	db.Model(&models.Lab{}).Count(&count)
	if count > 0 {
		log.Println("✅ Labs data already exists (skipping seed). Proceeding...")
		return nil
	}

	log.Println("🔬 Seeding Virtual Labs...")

	labs := []struct {
		Name          string
		Slug          string
		Description   string
		Platform      models.LabPlatform
		HardwareSpecs LabHardwareSpecs
		ThumbnailURL  string
		MQTTTopic     string
		AgentID       string
	}{
		{
			Name:        "Arduino Lab #1",
			Slug:        "arduino-lab-1",
			Description: "Complete Arduino Uno setup with various sensors and actuators. Perfect for beginners to learn the basics of electronics and programming.",
			Platform:    models.PlatformArduinoUno,
			HardwareSpecs: LabHardwareSpecs{
				Board:  "Arduino Uno R3",
				Camera: "Logitech C920 1080p",
				Sensors: []SensorSpec{
					{Name: "DHT22", Type: "temperature_humidity", Pin: "D2"},
					{Name: "LDR", Type: "light", Pin: "A0"},
					{Name: "HC-SR04", Type: "ultrasonic", Pins: map[string]string{"trig": "D3", "echo": "D4"}},
				},
				Actuators: []ActuatorSpec{
					{Name: "RGB LED", Type: "led", Pins: map[string]string{"r": "D9", "g": "D10", "b": "D11"}},
					{Name: "Servo Motor", Type: "servo", Pin: "D5"},
					{Name: "Buzzer", Type: "buzzer", Pin: "D6"},
				},
			},
			ThumbnailURL: "https://images.unsplash.com/photo-1553406830-ef2513450d76?w=400",
			MQTTTopic:    "lab/arduino-lab-1",
			AgentID:      "agent-arduino-01",
		},
		{
			Name:        "Arduino Lab #2",
			Slug:        "arduino-lab-2",
			Description: "Advanced Arduino Uno lab with motor control and additional sensors for robotics projects.",
			Platform:    models.PlatformArduinoUno,
			HardwareSpecs: LabHardwareSpecs{
				Board:  "Arduino Uno R3",
				Camera: "Logitech C920 1080p",
				Sensors: []SensorSpec{
					{Name: "MPU6050", Type: "accelerometer_gyroscope", Pins: map[string]string{"sda": "A4", "scl": "A5"}},
					{Name: "Potentiometer", Type: "analog", Pin: "A1"},
					{Name: "IR Sensor", Type: "infrared", Pin: "D7"},
				},
				Actuators: []ActuatorSpec{
					{Name: "DC Motor", Type: "motor", Pins: map[string]string{"en": "D3", "in1": "D4", "in2": "D5"}},
					{Name: "Stepper Motor", Type: "stepper", Pins: map[string]string{"step": "D8", "dir": "D9"}},
					{Name: "LCD Display", Type: "display", Pins: map[string]string{"rs": "D12", "en": "D11", "d4": "D5", "d5": "D4", "d6": "D3", "d7": "D2"}},
				},
			},
			ThumbnailURL: "https://images.unsplash.com/photo-1518770660439-4636190af475?w=400",
			MQTTTopic:    "lab/arduino-lab-2",
			AgentID:      "agent-arduino-02",
		},
		{
			Name:        "ESP32 Lab #1",
			Slug:        "esp32-lab-1",
			Description: "WiFi-enabled ESP32 lab with IoT capabilities. Learn to build connected devices with real-time data monitoring.",
			Platform:    models.PlatformESP32,
			HardwareSpecs: LabHardwareSpecs{
				Board:  "ESP32 DevKit V1",
				Camera: "Logitech C922 Pro",
				Sensors: []SensorSpec{
					{Name: "BME280", Type: "temp_humidity_pressure", Pins: map[string]string{"sda": "GPIO21", "scl": "GPIO22"}},
					{Name: "Soil Moisture", Type: "moisture", Pin: "GPIO34"},
					{Name: "PIR Motion", Type: "motion", Pin: "GPIO27"},
				},
				Actuators: []ActuatorSpec{
					{Name: "Relay Module", Type: "relay", Pin: "GPIO26"},
					{Name: "OLED Display", Type: "display", Pins: map[string]string{"sda": "GPIO21", "scl": "GPIO22"}},
					{Name: "NeoPixel Strip", Type: "led_strip", Pin: "GPIO13"},
				},
			},
			ThumbnailURL: "https://images.unsplash.com/photo-1601749607628-2d9e1e1bcd45?w=400",
			MQTTTopic:    "lab/esp32-lab-1",
			AgentID:      "agent-esp32-01",
		},
		{
			Name:        "ESP32 Lab #2",
			Slug:        "esp32-lab-2",
			Description: "Advanced ESP32 lab with camera module and machine learning capabilities. Build smart IoT applications.",
			Platform:    models.PlatformESP32,
			HardwareSpecs: LabHardwareSpecs{
				Board:  "ESP32-CAM",
				Camera: "OV2640 Camera Module + External USB Camera",
				Sensors: []SensorSpec{
					{Name: "MAX30102", Type: "heart_rate", Pins: map[string]string{"sda": "GPIO21", "scl": "GPIO22"}},
					{Name: "MQ-2", Type: "gas", Pin: "GPIO35"},
					{Name: "SW-420", Type: "vibration", Pin: "GPIO33"},
				},
				Actuators: []ActuatorSpec{
					{Name: "Servo Pan-Tilt", Type: "servo", Pins: map[string]string{"pan": "GPIO12", "tilt": "GPIO13"}},
					{Name: "LED Matrix", Type: "led_matrix", Pins: map[string]string{"din": "GPIO14", "clk": "GPIO15", "cs": "GPIO16"}},
					{Name: "Speaker", Type: "audio", Pin: "GPIO25"},
				},
			},
			ThumbnailURL: "https://images.unsplash.com/photo-1558618666-fcd25c85cd64?w=400",
			MQTTTopic:    "lab/esp32-lab-2",
			AgentID:      "agent-esp32-02",
		},
		{
			Name:        "Raspberry Pi Lab #1",
			Slug:        "raspi-lab-1",
			Description: "Full Linux-powered Raspberry Pi lab for advanced projects. Run Python scripts and control GPIO pins.",
			Platform:    models.PlatformRaspberryPi,
			HardwareSpecs: LabHardwareSpecs{
				Board:  "Raspberry Pi 4 Model B 4GB",
				Camera: "Raspberry Pi Camera Module V2",
				Sensors: []SensorSpec{
					{Name: "DS18B20", Type: "temperature", Pin: "GPIO4"},
					{Name: "ADS1115", Type: "adc", Pins: map[string]string{"sda": "GPIO2", "scl": "GPIO3"}},
					{Name: "GPS Module", Type: "gps", Pins: map[string]string{"tx": "GPIO14", "rx": "GPIO15"}},
				},
				Actuators: []ActuatorSpec{
					{Name: "LED Strip (WS2812)", Type: "led_strip", Pin: "GPIO18"},
					{Name: "Servo Array", Type: "servo", Pins: map[string]string{"s1": "GPIO17", "s2": "GPIO27", "s3": "GPIO22"}},
					{Name: "7-Segment Display", Type: "display", Pins: map[string]string{"clk": "GPIO5", "dio": "GPIO6"}},
				},
			},
			ThumbnailURL: "https://images.unsplash.com/photo-1629654297299-c8506221ca97?w=400",
			MQTTTopic:    "lab/raspi-lab-1",
			AgentID:      "agent-raspi-01",
		},
		{
			Name:        "STM32 Lab #1",
			Slug:        "stm32-lab-1",
			Description: "Professional-grade STM32 microcontroller lab for embedded systems development. High-performance ARM Cortex-M processor.",
			Platform:    models.PlatformSTM32,
			HardwareSpecs: LabHardwareSpecs{
				Board:  "STM32F4 Discovery",
				Camera: "Logitech C930e",
				Sensors: []SensorSpec{
					{Name: "LIS3DSH", Type: "accelerometer", Pins: map[string]string{"cs": "PE3", "sck": "PA5", "miso": "PA6", "mosi": "PA7"}},
					{Name: "MP45DT02", Type: "microphone", Pins: map[string]string{"clk": "PC3", "dout": "PB10"}},
					{Name: "Encoder", Type: "rotary", Pins: map[string]string{"a": "PA0", "b": "PA1", "btn": "PA2"}},
				},
				Actuators: []ActuatorSpec{
					{Name: "User LEDs", Type: "led", Pins: map[string]string{"green": "PD12", "orange": "PD13", "red": "PD14", "blue": "PD15"}},
					{Name: "CS43L22", Type: "audio_codec", Pins: map[string]string{"i2c_sda": "PB9", "i2c_scl": "PB6"}},
					{Name: "PWM Output", Type: "pwm", Pins: map[string]string{"ch1": "PB4", "ch2": "PB5", "ch3": "PB0", "ch4": "PB1"}},
				},
			},
			ThumbnailURL: "https://images.unsplash.com/photo-1518770660439-4636190af475?w=400",
			MQTTTopic:    "lab/stm32-lab-1",
			AgentID:      "agent-stm32-01",
		},
	}

	for _, labData := range labs {
		specs, _ := json.Marshal(labData.HardwareSpecs)

		lab := &models.Lab{
			Name:               labData.Name,
			Slug:               labData.Slug,
			Description:        labData.Description,
			Platform:           labData.Platform,
			Status:             models.LabStatusAvailable,
			HardwareSpecs:      specs,
			ThumbnailURL:       labData.ThumbnailURL,
			LivekitRoomName:    labData.Slug,
			MaxSessionDuration: 30,
			QueueCount:         0,
			TotalSessions:      0,
			MQTTTopic:          labData.MQTTTopic,
			AgentID:            labData.AgentID,
			IsOnline:           false, // Will be set to true when agent connects
		}

		if err := db.Create(lab).Error; err != nil {
			log.Printf("Warning: Failed to seed lab %s: %v", labData.Name, err)
			continue
		}
		log.Printf("  → Created lab: %s (%s)", lab.Name, lab.Platform)
	}

	log.Println("✅ Virtual Labs seeded successfully")
	return nil
}
