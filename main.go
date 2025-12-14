package main

import (
	"flag"
	"net/http"
	"os"

	"multipass/internal/app"
	"multipass/internal/middleware"
	"multipass/internal/router"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	//￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
	//           # LOAD .env
	//	__________________________________________
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	//￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
	//           # INITIALIZING APPLICATION
	//	__________________________________________
	app, err := app.NewApplication()
	if err != nil {
		panic(err)
	}
	defer app.DB.Close()
	defer app.Logger.Close()

	//￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
	//           # SERVER PORT SETUP
	//____________________________________________
	env := os.Getenv("ENV") // e.g., "development" or "production"
	port := os.Getenv("PORT")
	if port == "" {
		if env == "production" {
			app.Logger.Fatal("PORT must be set in production", nil)
		}

		var portFlag string
		flag.StringVar(&portFlag, "port", "8080", "backend server port")
		flag.Parse()

		port = portFlag
		app.Logger.Info("PORT not set, using default", "default_port", port)
	}

	// Create a new Router instance, passing in the app
	router := router.NewRouter(app)
	mux := http.NewServeMux()
	router.SetupRoutes(mux)
	//￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
	//  # INITIALIZE Middleware with Dependency
	//__________________________________________
	recovery := middleware.NewPanicRecoverMiddleWare(app.Logger, app.Responder)

	// # WRAPPING ROUTES WITH PANIC RECOVERY MIDDLEWARE

	muxWithPanicRecovery := recovery.Middleware(mux)

	//￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
	//         # SERVER STARTUP 🔥
	//__________________________________________*/
	app.Logger.Info("Starting server", "port", port)
	err = http.ListenAndServe(":"+port, muxWithPanicRecovery)
	// err = http.ListenAndServe(":"+port, mux)
	if err != nil {
		app.Logger.Fatal("Server failed to start", err)
	}
}
