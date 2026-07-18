package config

import (
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

func NewGormLogger() logger.Interface {
	return logger.New(
		logrus.StandardLogger(),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info, // ALL SQL
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}
