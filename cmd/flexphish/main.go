package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"flexphish/internal/api/handlers"
	"flexphish/internal/api/routes"
	"flexphish/internal/auth"
	"flexphish/internal/cli"
	"flexphish/internal/config"
	"flexphish/internal/servers"
	"flexphish/internal/storage"
	"flexphish/pkg/logger"
	"flexphish/pkg/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func main() {
	cli.PrintBanner()
	opts := cli.ParseFlags()

	if err := config.Load(utils.GetBasePath(opts.ConfigFile)); err != nil {
		panic(err)
	}
	logger.Init(opts.DevMode)
	defer logger.Sync()
	config.ApplyCLIOverrides(opts)

	dbPath := opts.DBPath

	testToken := config.GetString("security.test_mode_token")
	if testToken == "" {
		config.SetConfig("security.test_mode_token", uuid.New().String())
	}
	jwtSecret := config.GetString("security.jwt_secret")
	if jwtSecret == "" {
		jwtSecret = utils.GenerateSecret()
		config.SetConfig("security.jwt_secret", jwtSecret)
	}

	if err := storage.RunMigrations(dbPath); err != nil {
		logger.Log.Fatal("Migration failed", zap.Error(err))
	}

	db, err := storage.NewDatabase(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	repo := auth.NewRepository(db)
	service := auth.NewService(repo)

	if cli.HandleUserCommands(opts, service) {
		return
	}

	jwtService := auth.NewJWTService(auth.JWTConfig{
		Secret:     jwtSecret,
		Expiration: time.Hour * 24,
	})

	authHandler := handlers.NewAuthHandler(service, jwtService)

	router := routes.SetupRoutes(authHandler, jwtService, db)

	logger.Log.Info(fmt.Sprintf("[+] api server starting %s:%d",
		config.GetString("server.host"),
		config.GetInt("server.api_port")),
	)

	if opts.RunDashboard {
		go func() {
			err := servers.StartDashboard(config.GetString("server.host"), opts.DashboardPort, "ui")
			if err != nil {
				logger.Log.Error("Failed to start dashboard", zap.Error(err))
			}
		}()
	} else {
		logger.Log.Info("Dashboard disabled via CLI")
	}
	log.Fatal(http.ListenAndServe(
		config.GetString("server.host")+":"+config.GetString("server.api_port"),
		router,
	))
}
