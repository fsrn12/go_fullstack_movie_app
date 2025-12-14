package app

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"multipass/config"
	"multipass/internal/api"
	"multipass/internal/cache"
	"multipass/internal/middleware"
	"multipass/internal/service"
	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/logging"
	"multipass/pkg/response"

	"multipass/internal/auth/tokens"
	"multipass/internal/store"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Application struct {
	Config          *config.Config
	Logger          logging.Logger
	Responder       response.Writer
	DB              *pgxpool.Pool
	Redis           *redis.Client
	MovieHandler    *api.MovieHandler
	AccountHandler  *api.AccountHandler
	WebAuthnHandler *api.WebAuthnHandler
	AuthMiddleware  *middleware.AuthMiddleware
}

func NewApplication() (*Application, error) {
	// ￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
	//           # Load CONFIGURATION
	//￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config", err)
	}

	// ￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
	//           # INITIALIZE LOGGER + JSONWriter
	//￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
	appLogger, err := logging.NewAppLogger(cfg.LogFilePath, slog.LevelInfo)
	if err != nil {
		log.Fatal(err)
	}
	if appLogger == nil {
		fmt.Println("Failed to initialize AppLogger")
		log.Fatal("Failed to initialize AppLogger")
	}

	jsonWriter := response.NewJSONWriter(appLogger)

	if jsonWriter == nil {
		fmt.Println("JSON Writer is nil")
		appLogger.Fatal("Failed to initialize JSONWriter", nil)
	}

	/* ￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
		           # DATABASE SETUP
	￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣*/
	db, err := store.Open(cfg.DatabaseURL, appLogger)
	if err != nil {
		appLogger.Fatalf("Error opening database: %v", err)
	}
	appLogger.Info("✅ Database successfully connected")

	/* ￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
	           # REDIS SETUP
	__________________________________________*/
	redisClient, err := cache.NewRedisClient(cfg.RedisURL, appLogger)
	if err != nil {
		appLogger.Errorf("❌ Failed to connect to redis service: %+w", err)
	}

	/* ￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
	           # TOKENS SETUP
		__________________________________________*/
	// tokenManager := tokens.NewTokenManager(utils.ReadSecret("JWT_ACCESS_SECRET", appLogger), utils.ReadSecret("JWT_REFRESH_SECRET", appLogger), appLogger)
	tokenManager := tokens.NewTokenManager(cfg.JWT.AccessTokenSecret, cfg.JWT.RefreshTokenSecret, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL,
		appLogger)

	tokenStore := store.NewTokenStore(db, appLogger)
	if tokenStore == nil {
		appLogger.Fatal("Failed to initialize token store", nil)
	}

	/* ￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
	           # WEBAUTHEN/PASSKEY SETUP
		__________________________________________*/
	var webAuthnManager *webauthn.WebAuthn
	if webAuthnManager, err = webauthn.New(cfg.WebAuthn); err != nil {
		appLogger.Error("Error creating WebAuthn, Error initialing Passkey engine", err)
	}
	passkeyStore := store.NewPasskeyRepository(db, appLogger)
	passkeyService := service.NewPasskeyService(passkeyStore, webAuthnManager, tokenManager, appLogger)
	webAuthnHandler := api.NewWebAuthnHandler(passkeyService, appLogger, jsonWriter)
	/* ￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
	           # AUTH MIDDLEWARE
		__________________________________________*/
	authMW := middleware.NewAuthMiddleware(tokenManager, redisClient, appLogger, jsonWriter)

	/* ￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
	           # USER ACCOUNT SETUP
		__________________________________________*/
	accountStore := store.NewAccountRepository(db, appLogger)
	if accountStore == nil {
		appLogger.Fatalf("Failed to initialize account store: %+v", err)
	}

	emailSender := service.NewEmailService(cfg, appLogger)

	accountService := service.NewAccountService(
		accountStore,
		tokenStore,
		*tokenManager,
		emailSender,
		appLogger,
		cfg,
	)

	accountHandler := api.NewAccountHandler(
		accountService,
		appLogger,
		jsonWriter,
	)
	if accountHandler == nil {
		appLogger.Fatal("Failed to initialize account handler", nil)
	}
	/* ￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
	           # MOVIES SETUP
		__________________________________________*/
	movieStore := store.NewMovieRepository(db, appLogger)
	if movieStore == nil {
		appLogger.Fatalf("Failed to initialize movie store: %+v", err)
	}

	movieHandler := api.NewMovieHandler(movieStore, appLogger, jsonWriter)
	if movieHandler == nil {
		appLogger.Fatal("Failed to initialize movie handler", nil)
	}
	// accountHandler := api.NewAccountHandler(
	// 	accountStore,
	// 	tokenStore,
	// 	*tokenManager,
	// 	appLogger,
	// 	jsonWriter,
	// )

	/* ￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
	           # RETURN APPLICATION
		__________________________________________*/

	app := &Application{
		Config:          cfg,
		Logger:          appLogger,
		Responder:       jsonWriter,
		Redis:           redisClient,
		DB:              db,
		MovieHandler:    movieHandler,
		AccountHandler:  accountHandler,
		WebAuthnHandler: webAuthnHandler,
		AuthMiddleware:  authMW,
	}
	return app, nil
}

// HealthCheck shows the current status of the application
func (a *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := a.Responder.WriteJSON(w, http.StatusOK, common.Envelop{
		"message": "🚀 App status — Available",
		"method":  r.Method,
		"path":    r.URL.Path,
	}); err != nil {
		apperror.NewAppError(http.StatusNotFound, "💥 App status — Unavailable", "HealthCheck", err, a.Logger, common.Envelop{
			"method": r.Method,
			"path":   r.URL.Path,
		}).WriteJSONError(w, r, a.Responder)
		a.Logger.Error("Application status failure", err)
	}
}

// Custom MIME types
func (a *Application) CustomMIMEServer(publicDir string) http.Handler {
	fs := http.FileServer(http.Dir(publicDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		}
		fs.ServeHTTP(w, r)
	})
}

func (a *Application) CatchAllClientRoutesHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./public/index.html")
}

// func (a *Application) Shutdown(ctx context.Context) error {
// 	if err := a.DB.Close(); err != nil {
// 		a.Logger.Error("Failed to close DB", err)
// 	}
// 	if err := a.Redis.Close(); err != nil {
// 		a.Logger.Error("Failed to close Redis", err)
// 	}
// 	return nil
// }
