package config

import (
	"fmt"
	"interviewexcel-backend-go/models"
	"log"

	_ "github.com/lib/pq"
	"github.com/razorpay/razorpay-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var RazorpayClient *razorpay.Client

type Config struct {
	GoogleLoginConfig oauth2.Config
}

var AppConfig Config

func GoogleConfig() *oauth2.Config {
	runtimeConfig := RuntimeConfig()
	AppConfig.GoogleLoginConfig = oauth2.Config{
		RedirectURL:  runtimeConfig.GoogleRedirectURL,
		ClientID:     runtimeConfig.GoogleClientID,
		ClientSecret: runtimeConfig.GoogleClientSecret,
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint: google.Endpoint,
	}

	return &AppConfig.GoogleLoginConfig
}

var DB *gorm.DB

func OpenDB() (*gorm.DB, error) {
	// Ensure logrus level allows SQL logs
	logrus.SetLevel(logrus.InfoLevel)

	db, err := gorm.Open(postgres.Open(DatabaseDSN()), &gorm.Config{
		Logger: NewGormLogger(),
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting to the DB: %w", err)
	}

	return db, nil
}

func InitDB() error {
	db, err := OpenDB()
	if err != nil {
		return err
	}

	DB = db
	log.Println("Database connected successfully")
	return nil
}

func RunMigrations() error {
	if DB == nil {
		if err := InitDB(); err != nil {
			return err
		}
	}

	if err := DB.AutoMigrate(models.GetMigrationModel()...); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func InitRazorpay() error {
	runtimeConfig := RuntimeConfig()

	if runtimeConfig.RazorpayKey == "" || runtimeConfig.RazorpaySecret == "" {
		return fmt.Errorf("missing Razorpay credentials")
	}

	RazorpayClient = razorpay.NewClient(runtimeConfig.RazorpayKey, runtimeConfig.RazorpaySecret)
	log.Println("Razorpay client initialized")
	return nil
}
