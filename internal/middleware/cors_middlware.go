package middleware

import (
	"net/http"
)

// corsMiddleware is a simple CORS middleware for standard library HTTP.
// It handles preflight OPTIONS requests and sets necessary CORS headers for all responses.
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// IMPORTANT: Replace "http://localhost:5173" with your actual frontend development URL.
		// If you deploy, this should be your production frontend URL (e.g., "https://your-frontend.com").
		// For multiple origins, you can check r.Header.Get("Origin") and set it dynamically,
		// or use a loop if you have a fixed small set of allowed origins.
		// For now, let's assume one specific dev origin.
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")

		// Set allowed HTTP methods for your API (GET, POST, PUT, DELETE are common)
		// Include OPTIONS, as the browser will send preflight requests with this method.
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Set allowed headers. 'Content-Type' is usually needed for JSON requests.
		// 'Authorization' is crucial for sending JWT tokens. Add any other custom headers you use.
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Allow credentials (e.g., cookies, HTTP authentication headers). Set to "true" if your frontend sends them.
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// For preflight requests (OPTIONS method), just send the headers and an OK status.
		// Do not proceed to the actual handler for OPTIONS requests.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return // End the request here for preflights
		}

		// For all other requests, proceed to the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}
