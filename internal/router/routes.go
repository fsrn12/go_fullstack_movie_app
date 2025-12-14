package router

import (
	"net/http"

	"multipass/internal/app"
	"multipass/internal/middleware"
)

type Router struct {
	App *app.Application
}

func NewRouter(app *app.Application) *Router {
	return &Router{
		App: app,
	}
}

func (rt *Router) SetupRoutes(mux *http.ServeMux) {
	/*
		 ----------------------------------------------
		* PUBLIC ACCOUNT ENDPOINTS (NO AUTH REQUIRED)
		 ----------------------------------------------
	*/
	mux.Handle("/api/movies/top",
		middleware.CorsMiddleware(
			http.HandlerFunc(rt.App.MovieHandler.HandleGetTopMovies),
		),
	)
	mux.Handle("/api/movies/random",
		middleware.CorsMiddleware(
			http.HandlerFunc(rt.App.MovieHandler.HandleGetRandomMovies),
		),
	)
	mux.Handle("/api/movies/search",
		middleware.CorsMiddleware(
			http.HandlerFunc(rt.App.MovieHandler.HandleSearchMovies),
		),
	)
	mux.Handle("/api/movies/{id}",
		middleware.CorsMiddleware(
			http.HandlerFunc(rt.App.MovieHandler.HandleGetMovieByID),
		),
	)
	mux.Handle("/api/genres",
		middleware.CorsMiddleware(
			http.HandlerFunc(rt.App.MovieHandler.HandleGetAllGenres),
		),
	)

	// POST: REGISTER
	mux.Handle("/api/account/register",
		middleware.CorsMiddleware(
			http.HandlerFunc(rt.App.AccountHandler.SignUp),
		),
	)

	// POST: LOGIN
	mux.Handle("/api/account/login",
		middleware.CorsMiddleware(
			http.HandlerFunc(rt.App.AccountHandler.Login),
		),
	)

	// GET: EMAIL VERIFICATION (Public)
	mux.Handle("/api/account/email/verify", // Cleaner path name
		middleware.CorsMiddleware(
			http.HandlerFunc(rt.App.AccountHandler.HandleVerifyEmail),
		),
	)
	// POST: REQUEST PASSWORD RESET (Public)
	mux.Handle("/api/account/password/reset",
		middleware.CorsMiddleware(
			http.HandlerFunc(rt.App.AccountHandler.HandlePasswordResetRequest),
		),
	)

	// POST: CONFIRM PASSWORD RESET (Public)
	mux.Handle("/api/account/password/confirm",
		middleware.CorsMiddleware(
			http.HandlerFunc(rt.App.AccountHandler.HandleConfirmPasswordReset),
		),
	)

	/*
	 ---------------------------------
	 * AUTHENTICATED ACCOUNT ENDPOINTS
	 ---------------------------------
	*/
	// REFRESH
	mux.Handle("/api/account/refresh",
		rt.App.AuthMiddleware.Authenticate(
			middleware.CorsMiddleware(
				http.HandlerFunc(rt.App.AccountHandler.Refresh),
			),
		),
	)

	// PUT: UPDATE USER
	mux.Handle("/api/account/update-me",
		rt.withAuthAndCORS(
			http.HandlerFunc(rt.App.AccountHandler.HandleUserUpdate),
		),
	)

	// GET: PROFILE
	mux.Handle("/api/account/profile",
		rt.withAuthAndCORS(
			http.HandlerFunc(rt.App.AccountHandler.HandleGetUserProfile),
		),
	)

	// POST: PROFILE PICTURE UPLOAD
	mux.Handle("/api/account/profile-picture",
		rt.withAuthAndCORS(
			http.HandlerFunc(rt.App.AccountHandler.HandleProfilePictureUpload),
		),
	)

	// GET: FAVORITES
	mux.Handle("/api/account/favorites",
		rt.withAuthAndCORS(
			http.HandlerFunc(rt.App.AccountHandler.HandleGetFavorites),
		),
	)

	// GET:  WATCHLIST
	mux.Handle("/api/account/watchlist",
		rt.withAuthAndCORS(
			http.HandlerFunc(rt.App.AccountHandler.HandleGetWatchlist),
		),
	)

	// POST: SAVE MOVIE TO COLLECTION
	mux.Handle("/api/account/save-to-collection",
		rt.withAuthAndCORS(
			http.HandlerFunc(rt.App.AccountHandler.HandleSaveToCollection),
		),
	)

	// POST: REMOVE MOVIE FROM COLLECTION
	mux.Handle("/api/account/remove-from-collection",
		rt.withAuthAndCORS(
			http.HandlerFunc(rt.App.AccountHandler.HandleRemoveMovieFromCollection),
		),
	)

	mux.Handle("/api/passkey/registration-begin",
		rt.withAuthAndCORS(
			http.HandlerFunc(rt.App.WebAuthnHandler.WebAuthnRegistrationBeginHandler),
		),
	)
	mux.Handle("/api/passkey/registration-end",
		rt.withAuthAndCORS(
			http.HandlerFunc(rt.App.WebAuthnHandler.WebAuthRegistrationEndHandler),
		),
	)
	mux.Handle("/api/passkey/authentication-begin",
		middleware.CorsMiddleware(
			http.HandlerFunc(rt.App.WebAuthnHandler.WebAuthnAuthenticationBeginHandler),
		),
	)
	mux.Handle("/api/passkey/authentication-end",
		middleware.CorsMiddleware(
			http.HandlerFunc(rt.App.WebAuthnHandler.WebAuthnAuthenticationEndHandler),
		),
	)

	//￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
	//         # CATCH ALL CLIENT ROUTES
	//__________________________________________
	mux.HandleFunc("/movies", rt.App.CatchAllClientRoutesHandler)
	mux.HandleFunc("/movies/", rt.App.CatchAllClientRoutesHandler)
	mux.HandleFunc("/account/", rt.App.CatchAllClientRoutesHandler)
	//￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
	//         # STATIC ROUTE
	//__________________________________________

	mux.Handle("/", rt.App.CustomMIMEServer(rt.App.Config.STATIC))
}

// Utility to apply Auth and CORS to routes that need both
func (rt *Router) withAuthAndCORS(next http.Handler) http.Handler {
	return rt.App.AuthMiddleware.Authenticate(middleware.CorsMiddleware(next))
}

// And ensure your main function uses this mux for ListenAndServe:
// func main() {
//     // ... your app setup ...
//     router := setupRoutes(app)
//     log.Fatal(http.ListenAndServe(":4000", router)) // Or whatever port you use
// }

// func SetupRoutes(app *app.Application) *http.ServeMux {
// 	mux := http.NewServeMux()

// 	mux.Handle("/api/movies/top",
// 		middleware.CorsMiddleware(
// 			http.HandlerFunc(app.MovieHandler.HandleGetTopMovies),
// 		),
// 	)
// 	mux.Handle("/api/movies/random",
// 		middleware.CorsMiddleware(
// 			http.HandlerFunc(app.MovieHandler.HandleGetRandomMovies),
// 		),
// 	)
// 	mux.Handle("/api/movies/search",
// 		middleware.CorsMiddleware(
// 			http.HandlerFunc(app.MovieHandler.HandleSearchMovies),
// 		),
// 	)
// 	mux.Handle("/api/movies/{id}",
// 		middleware.CorsMiddleware(
// 			http.HandlerFunc(app.MovieHandler.HandleGetMovieByID),
// 		),
// 	)
// 	mux.Handle("/api/genres",
// 		middleware.CorsMiddleware(
// 			http.HandlerFunc(app.MovieHandler.HandleGetAllGenres),
// 		),
// 	)

// 	// POST: REGISTER
// 	mux.Handle("/api/account/register",
// 		middleware.CorsMiddleware(
// 			http.HandlerFunc(app.AccountHandler.SignUp),
// 		),
// 	)

// 	// POST: LOGIN
// 	mux.Handle("/api/account/login",
// 		middleware.CorsMiddleware(
// 			http.HandlerFunc(app.AccountHandler.Login),
// 		),
// 	)

// 	// REFRESH
// 	mux.Handle("/api/account/refresh",
// 		app.AuthMiddleware.Authenticate(
// 			middleware.CorsMiddleware(
// 				http.HandlerFunc(app.AccountHandler.Refresh),
// 			),
// 		),
// 	)

// 	// PUT: UPDATE USER
// 	mux.Handle("/api/account/updateMe",
// 		app.AuthMiddleware.Authenticate(
// 			middleware.CorsMiddleware(
// 				http.HandlerFunc(app.AccountHandler.HandleUserUpdate),
// 			),
// 		),
// 	)

// 	// GET: PROFILE
// 	mux.Handle("/api/account/profile",
// 		app.AuthMiddleware.Authenticate(
// 			middleware.CorsMiddleware(
// 				http.HandlerFunc(app.AccountHandler.HandleGetUserProfile),
// 			),
// 		),
// 	)

// 	// POST: PROFILE PICTURE UPLOAD
// 	mux.Handle("/api/account/profile-picture",
// 		app.AuthMiddleware.Authenticate(
// 			middleware.CorsMiddleware(
// 				http.HandlerFunc(app.AccountHandler.HandleProfilePictureUpload),
// 			),
// 		),
// 	)

// 	// GET: FAVORITES
// 	mux.Handle("/api/account/favorite",
// 		app.AuthMiddleware.Authenticate(
// 			middleware.CorsMiddleware(
// 				http.HandlerFunc(app.AccountHandler.HandleGetFavorites),
// 			),
// 		),
// 	)

// 	// GET:  WATCHLIST
// 	mux.Handle("/api/account/watchlist",
// 		app.AuthMiddleware.Authenticate(
// 			middleware.CorsMiddleware(
// 				http.HandlerFunc(app.AccountHandler.HandleGetWatchlist),
// 			),
// 		),
// 	)

// 	// POST: SAVE MOVIE TO COLLECTION
// 	mux.Handle("/api/account/save-to-collection",
// 		app.AuthMiddleware.Authenticate(
// 			middleware.CorsMiddleware(
// 				http.HandlerFunc(app.AccountHandler.HandleSaveToCollection),
// 			),
// 		),
// 	)

// 	//￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
// 	//         # CATCH ALL CLIENT ROUTES
// 	//__________________________________________
// 	mux.HandleFunc("/movies", app.CatchAllClientRoutesHandler)
// 	mux.HandleFunc("/movies/", app.CatchAllClientRoutesHandler)
// 	mux.HandleFunc("/account/", app.CatchAllClientRoutesHandler)
// 	//￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣
// 	//         # STATIC ROUTE
// 	//__________________________________________
// 	static := os.Getenv("STATIC")
// 	if static == "" {
// 		app.Logger.Error("STATIC is not set in environment", nil)
// 	}
// 	// mux.Handle("/", http.FileServer(http.Dir(static)))
// 	mux.Handle("/", app.CustomMIMEServer(static))

// 	return mux
// }
