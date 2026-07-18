package main

import (
	"interviewexcel-backend-go/config"
	"interviewexcel-backend-go/routes"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func buildRouter() *gin.Engine {
	r := gin.Default()
	runtimeConfig := config.RuntimeConfig()

	// Apply CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     runtimeConfig.CorsAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Cookie"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/healthz", healthHandler)

	// Register routes
	routes.RegisterExpertRoutes(r)
	routes.RegisterStudentRoutes(r)
	routes.AuthRoutes(r)

	// Banner
	banner := `


  __ __ _ ____ ____ ____ _  _ __ ____ _  _    ____ _  _ ___ ____ __      ____  __   ___ __ _ ____ __ _ ____ 
 (  (  ( (_  _(  __(  _ / )( (  (  __/ )( \  (  __( \/ / __(  __(  )    (  _ \/ _\ / __(  / (  __(  ( (    \
  )(/    / )(  ) _) )   \ \/ /)( ) _)\ /\ /   ) _) )  ( (__ ) _)/ (_/\   ) _ /    ( (__ )  ( ) _)/    /) D (
 (__\_)__)(__)(____(__\_)\__/(__(____(_/\_)  (____(_/\_\___(____\____/  (____\_/\_/\___(__\_(____\_)__(____/


	`
	color.Green(banner)

	return r
}

func healthHandler(c *gin.Context) {
	statusCode := http.StatusOK
	status := gin.H{
		"status": "ok",
		"db":     "ok",
		"redis":  "disabled",
	}

	if config.DB == nil {
		statusCode = http.StatusServiceUnavailable
		status["status"] = "degraded"
		status["db"] = "not_initialized"
	} else if sqlDB, err := config.DB.DB(); err != nil || sqlDB.Ping() != nil {
		statusCode = http.StatusServiceUnavailable
		status["status"] = "degraded"
		status["db"] = "unhealthy"
	}

	if config.RedisClient != nil {
		if _, err := config.RedisClient.Ping(config.Ctx).Result(); err != nil {
			statusCode = http.StatusServiceUnavailable
			status["status"] = "degraded"
			status["redis"] = "unhealthy"
		} else {
			status["redis"] = "ok"
		}
	}

	c.JSON(statusCode, status)
}

func runServe() error {
	if err := config.InitDB(); err != nil {
		return err
	}

	if err := config.InitRedis(); err != nil {
		return err
	}

	if err := config.InitRazorpay(); err != nil {
		return err
	}

	router := buildRouter()
	port := config.RuntimeConfig().Port
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	return router.Run(port)
}

func runMigrate() error {
	if err := config.InitDB(); err != nil {
		return err
	}

	return config.RunMigrations()
}

func main() {
	command := "serve"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	var err error

	switch command {
	case "serve":
		err = runServe()
	case "migrate":
		err = runMigrate()
	default:
		log.Fatalf("unknown command %q, expected serve or migrate", command)
	}

	if err != nil {
		log.Fatal(err)
	}
}
