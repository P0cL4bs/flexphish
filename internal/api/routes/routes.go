package routes

import (
	"log"
	"net/http"

	"flexphish/internal/api/handlers"
	"flexphish/internal/api/middleware"
	"flexphish/internal/application/controller"
	"flexphish/internal/application/runtime"
	"flexphish/internal/auth"
	"flexphish/internal/config"
	"flexphish/internal/infrastructure/repository"
	"flexphish/internal/servers"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func SetupRoutes(
	authHandler *handlers.AuthHandler,
	jwtService *auth.JWTService,
	db *gorm.DB,
) http.Handler {

	cfg := config.Get().All()
	router := mux.NewRouter()
	router.Use(commonMiddleware)
	router.Use(middleware.RecoveryMiddleware)
	router.Use(middleware.LoggingMiddleware)

	public := router.PathPrefix("/api").Subrouter()
	public.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost, http.MethodOptions)

	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware(jwtService))

	templateRepo := repository.NewTemplateRepository(cfg)
	campaignRepo := repository.NewCampaignRepository(db, templateRepo)
	htmlFileRepo := repository.NewHtmlFileRepository(config.GetString("template_assets_dir"), templateRepo)
	staticFileRepo := repository.NewStaticFileRepository(config.GetString("template_assets_dir"), templateRepo)
	groupRepo := repository.NewGroupRepository(db)
	smtpRepo := repository.NewSMTPRepository(db)
	emailTemplateRepo := repository.NewEmailTemplateRepository(db)

	templateHandler := handlers.NewTemplateHandler(templateRepo, campaignRepo)
	configHandler := handlers.NewConfigHandler()

	RegisterTemplateRoutes(protected, templateHandler)
	RegisterHtmlfilesRoutes(protected, handlers.NewHtmlFileHandler(htmlFileRepo, templateRepo))
	RegisterStaticFilesRoutes(protected, handlers.NewStaticFileHandler(staticFileRepo, templateRepo))
	RegisterCampaignRoutes(
		protected,
		campaignRepo,
		templateRepo,
		groupRepo,
		smtpRepo,
		emailTemplateRepo,
		middleware.AuthMiddleware(jwtService),
	)
	RegisterConfigRoutes(protected, configHandler)
	RegisterGroupRoutes(protected, groupRepo)
	RegisterSMTPRoutes(protected, smtpRepo)
	RegisterEmailTemplateRoutes(protected, emailTemplateRepo)

	emailScheduler := handlers.NewCampaignEmailScheduler(
		campaignRepo,
		handlers.NewCampaignHandler(
			campaignRepo,
			templateRepo,
			groupRepo,
			smtpRepo,
			emailTemplateRepo,
		),
		config.EmailSchedulerConfig{
			Enabled:              config.GetBool("email_scheduler.enabled"),
			PollIntervalSeconds:  config.GetInt("email_scheduler.poll_interval_seconds"),
			EmailsPerMinute:      config.GetInt("email_scheduler.emails_per_minute"),
			BatchSize:            config.GetInt("email_scheduler.batch_size"),
			BatchPauseMS:         config.GetInt("email_scheduler.batch_pause_ms"),
			MaxParallelCampaigns: config.GetInt("email_scheduler.max_parallel_campaigns"),
		},
	)
	emailScheduler.Start()

	protected.HandleFunc("/auth/validate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"valid": true}`))
	}).Methods(http.MethodGet, http.MethodOptions)

	mux := http.NewServeMux()

	resultRepo := repository.NewResultRepository(db)
	eventRepo := repository.NewEventRepository(db)

	geoService, err := runtime.NewGeoService("configs/GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}

	defer geoService.Close()

	sessionService := runtime.NewSessionService(
		resultRepo,
		geoService,
		campaignRepo,
	)
	eventService := runtime.NewEventService(eventRepo, resultRepo, campaignRepo)

	runtimeController := controller.NewCampaignRuntimeController(
		campaignRepo,
		templateRepo,
		sessionService,
		eventService,
		runtime.NewMemoryStateStore(),
	)

	RegisterRuntimeRoutes(mux, runtimeController)

	servers.StartCampaignServer(mux)

	return router
}
