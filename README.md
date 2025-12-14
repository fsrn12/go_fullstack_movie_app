# 🎬 Fullstack Movie Application

A production-ready, full-stack movie discovery platform built with modern web technologies, featuring secure authentication, real-time caching, and a custom component architecture.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-18+-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## 🌟 Overview

This application demonstrates enterprise-grade full-stack development practices with a focus on performance, security, and maintainability. Built entirely from scratch without relying on heavy frameworks, it showcases deep understanding of web fundamentals and systems design.

**Live Features:**

- Browse trending and top-rated movies with rich metadata
- Secure user authentication with JWT + Refresh Token rotation
- WebAuthn/Passkey integration for passwordless authentication
- Personalized watchlists and favorites management
- Profile customization with image uploads
- Responsive, component-based vanilla JavaScript frontend
- Service Worker caching for offline-first experience

---

## 🏗️ Architecture

### Backend Stack

- **Language:** Go (Golang) - leveraging goroutines, channels, and Go's concurrency model
- **Database:** PostgreSQL with PGX driver for optimal performance
- **Query Layer:** Raw SQL for precise control and query optimization
- **Authentication:** JWT access tokens with secure refresh token rotation
- **Caching:** Redis for session management and performance optimization
- **Security:** Bcrypt password hashing, CORS middleware, rate limiting

### Frontend Stack

- **Core:** Vanilla JavaScript (ES6+) - no frameworks, pure web standards
- **Architecture:** Custom Web Components for reusable, encapsulated UI
- **Routing:** Custom SPA router with history API integration
- **State Management:** Custom authentication store with reactivity
- **Caching:** Service Workers for offline-first PWA capabilities
- **Styling:** Modern CSS with custom properties and responsive design

---

## 🚀 Key Technical Features

### Backend Engineering

#### Custom Middleware Pipeline

```
Request → Panic Recovery → CORS → Rate Limiting → Auth Validation → Handler
```

- **Panic Middleware:** Graceful error recovery with detailed logging
- **Auth Middleware:** JWT validation with token refresh flow
- **Rate Limiting:** Token bucket algorithm for API protection
- **CORS:** Configurable cross-origin resource sharing

#### Security Implementation

- Password hashing with Bcrypt (cost factor 12)
- JWT access tokens (short-lived) + Refresh tokens (HTTP-only, secure cookies)
- Token rotation on refresh to prevent replay attacks
- WebAuthn/Passkey support for FIDO2 passwordless authentication
- SQL injection prevention through parameterized queries
- XSS protection through proper content-type headers

#### Custom Packages

- **`pkg/logging`**: Structured logging with multiple output handlers and log levels
- **`pkg/apperror`**: Type-safe error handling with HTTP status code mapping
- **`pkg/validator`**: Request validation with custom rules
- **`pkg/response`**: Standardized JSON response writer

#### Database Design

- Normalized schema with foreign key constraints
- Migration-based version control (golang-migrate)
- Connection pooling for optimal resource usage
- Transaction support for data consistency

### Frontend Engineering

#### Custom Web Components

```javascript
// Self-contained, reusable components
<movie-item>, <genre-list>, <youtube-embed>, <alert-modal>
```

- Shadow DOM for style encapsulation
- Custom event system for inter-component communication
- Lifecycle hooks (connectedCallback, disconnectedCallback)

#### Client-Side Routing

- History API-based navigation (no page reloads)
- Route guards for protected pages
- Dynamic route parameters
- Browser back/forward support

#### State Management

- Centralized authentication store
- Observer pattern for reactive updates
- LocalStorage persistence for user sessions

#### Progressive Web App (PWA)

- Service Worker for offline functionality
- Cache-first strategy for static assets
- Background sync for failed requests
- Web App Manifest for installability

---

## 📂 Project Structure

```
root/
├── cmd/                        # Application entrypoint
│   └── main.go
├── internal/                   # Private application code
│   ├── api/                   # HTTP handlers (controllers)
│   ├── service/               # Business logic layer
│   ├── store/                 # Data access layer (repositories)
│   ├── model/                 # Domain models
│   ├── middleware/            # HTTP middleware chain
│   ├── router/                # Route definitions
│   └── auth/                  # Authentication utilities
├── pkg/                       # Reusable packages
│   ├── logging/              # Structured logger
│   ├── apperror/             # Error handling
│   ├── validator/            # Request validation
│   └── utils/                # Helper functions
├── public/                    # Frontend application
│   ├── components/           # Web Components
│   ├── services/             # API clients, Router, State
│   └── utils/                # Frontend utilities
├── migrations/               # SQL migrations
└── config/                   # Configuration management
```

---

## 🛠️ Installation & Setup

### Prerequisites

- Go 1.24 or higher
- PostgreSQL 17+
- Redis 9+
- Node.js (optional, for development tools)

### Quick Start

1. **Clone the repository**

   ```bash
   git clone https://github.com/yourusername/movie-app.git
   cd movie-app
   ```

2. **Environment Configuration**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Database Setup**

   ```bash
   # Create database
   Use the provided 'import.go' script in /import to seed your movie data to your postgres database or use TMDB API.

   # Run migrations
   Use goose from 'https://github.com/pressly/goose' to run migrations
   ```

4. **Install Dependencies**

   ```bash
   go mod download
   ```

5. **Run the Application**

   ```bash
   # Development with hot reload
   air

   # Production
   go run main.go
   ```

6. **Access the Application**
   ```
   http://localhost:8080
   ```

---

## 🔧 Configuration

Configuration is managed through environment variables:

```env
# Server
PORT=? #SERVER PORT
LOG_FILE_PATH=? #LOG FiLE
STATIC=? #FRONTEND DIRECTORY PATH
PROFILE_PICTURE_PATH=?  #FILESYSTEM PATH FOR STORING IMAGES
PROFILE_PICTURE_BASE=?  #URL BASE FOR SERVING IMAGES

DATABASE_URL=? #POSTGRES DATABASE STRING
REDIS_URL=? #REDIS INSTANCE STRING

JWT_ACCESS_SECRET=? #SECRET STRING
REFRESH_SECRET=? #SECRET STRING
ACCESS_TOKEN_TTL=? #10m
REFRESH_TOKEN_TTL=?  #48h for 7 days

WEBAUTHN_RP_DISPLAY_NAME=? #App
WEBAUTHN_RP_ID=? #LOCALHOST
WEBAUTHN_RP_ORIGINS=? #http://HOSTNAME:PORTNUMBER

# EMAIL CONFIG
FRONTEND_URL=? #http://HOSTNAME:PORTNUMBER/
EMAIL_FROM=? #ADMIN@APP.COM
SMTP_HOST=? #SMTP HOST
SMTP_PORT=? #EMAIL SERVER PORT
SMTP_USER=? #EMAIL USERNAME
SMTP_PASS=? #EMAIL PASSWORD

# Movie API
TMDB_API_KEY=your-tmdb-api-key
```

---

## 🎯 API Endpoints

### Authentication

```
POST   /api/account/register          # User registration
POST   /api/account/login             # User login
POST   /api/account/refresh           # Refresh access token
POST   /api/account/logout            # User logout
POST   /api/account/email/verify      # verify email
POST   /api/account/password/reset    # forgot password
POST   /api/account/password/confirm  # confirm password
```

### Passkey Authentication

```
POST   /api/passkey/registration-begin    # Initiate passkey registration
POST   /api/passkey/registration-end      # Complete passkey registration
POST   /api/passkey/authentication-begin  # Initiate passkey login
POST   /api/passkey/authentication-end    # Complete passkey login
```

### User Account

```
GET    /api/account/profile              # Get user profile
PUT    /api/account/update-me            # Update user profile
POST   /api/account/profile-picture      # Upload profile picture

```

### Movies

```
GET    /api/movies/random             # Get random movies
GET    /api/movies/top                # Get top-rated movies
GET    /api/movies/:id                # Get movie details
GET    /api/v1/movies/:id/cast        # Get movie cast
POST   /api/movies/search             # Search Movie
```

### Watchlist & Favorites

```
GET    /api/account/favorites            # Get Favorites List
GET    /api/account/watchlist            # Get Watchlist
POST   /api/account/collection/add       # Add to Favorites / Watchlist List
POST   /api/account/collection/remove    # Remove from Favorites / Watchlist List
```

---

## 🧪 Development

### Hot Reload

```bash
air  # Uses .air.toml configuration
```

### Linting

```bash
golangci-lint run
```

### Database Migrations

```bash
# Create new migration
goose create create_user sql

# Run migrations
goose -dir migrations postgres "postgres://postgres:PASSWORD@localhost:5432/postgres?sslmode=disable" up

# Rollback
goose -dir migrations postgres "postgres://postgres:PASSWORD@localhost:5432/postgres?sslmode=disable" down
```

---

## 🏆 Technical Highlights

### Performance Optimizations

- **Connection Pooling:** Configured PGX pool for optimal database connections
- **Redis Caching:** Frequently accessed data cached to reduce database load
- **Service Worker:** Aggressive caching strategy for static assets
- **Lazy Loading:** Frontend components loaded on-demand
- **SQL Query Optimization:** Indexed columns and optimized joins

### Code Quality

- **Clean Architecture:** Separation of concerns (handlers → services → repositories)
- **Error Handling:** Centralized error types with proper HTTP status mapping
- **Type Safety:** Strict Go typing with interface contracts
- **Validation:** Request validation at API boundaries
- **Logging:** Structured logging with different severity levels

### Security Best Practices

- OWASP Top 10 mitigation strategies
- Rate limiting to prevent abuse
- Secure cookie flags (HttpOnly, Secure, SameSite)
- Input sanitization and validation
- Prepared statements for SQL injection prevention

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Feel free to check the [issues page](../../issues).

---

## 📝 License

This project is [MIT](LICENSE) licensed.

---

## 👨‍💻 Author

**Faisal Rehman**

- GitHub: [@yourusername](https://github.com/fsrn12)
- LinkedIn: [Your LinkedIn](https://www.linkedin.com/in/faisal-rehman-a58032177/)

---

## 🙏 Acknowledgments

- The Movie Database (TMDB) for movie data API
- Go community for excellent packages and documentation
- WebAuthn specification for passwordless authentication standards

---

**Built with passion and attention to detail** 🚀
