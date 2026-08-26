package config

import (
	"fmt"
	"interviewexcel-backend-go/models"
	"log"

	"context"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

// R2 client and derived settings, populated by InitR2.
var (
	R2Client        *s3.Client
	R2Bucket        string
	R2PublicBaseURL string
)

// InitR2 initializes the Cloudflare R2 (S3-compatible) client from runtime config.
func InitR2() error {
	rc := RuntimeConfig()

	if rc.R2AccountID == "" || rc.R2AccessKeyID == "" || rc.R2SecretAccessKey == "" || rc.R2Bucket == "" {
		return fmt.Errorf("missing R2 credentials (need R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, R2_BUCKET)")
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", rc.R2AccountID)

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		// R2 ignores region but the SDK requires a non-empty value.
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(rc.R2AccessKeyID, rc.R2SecretAccessKey, ""),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to load R2 config: %w", err)
	}

	R2Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = &endpoint
	})
	R2Bucket = rc.R2Bucket
	R2PublicBaseURL = rc.R2PublicBaseURL

	log.Println("Cloudflare R2 client initialized")
	return nil
}
