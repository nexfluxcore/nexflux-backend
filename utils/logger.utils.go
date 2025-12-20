// utils/logger.go
package utils

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log = logrus.New()

func InitLogger() {
	logFile := &lumberjack.Logger{
		Filename:   "app.log",
		MaxSize:    10,   // Ukuran maksimal file (MB)
		MaxBackups: 3,    // Jumlah backup file
		MaxAge:     30,   // Umur maksimal file (hari)
		Compress:   true, // Kompres file lama
	}
	Log.SetOutput(logFile)

	// Set format log
	Log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Set level log (misalnya: Info, Debug, Error)
	Log.SetLevel(logrus.InfoLevel)
}
