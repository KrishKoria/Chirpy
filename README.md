 # Chirpy API

 ## Overview
 Chirpy is a social media backend API service built with Go.
 It allows users to post short messages called "chirps" (similar to tweets),
 follow other users, and interact with the platform.

 ## Features
 - User registration and authentication with JWT tokens
 - Secure password handling with bcrypt
 - Refresh token management
 - Posting, retrieving and deleting chirps
 - Automatic profanity filtering
 - Premium subscription (Chirpy Red)
 - Webhook integration
 - API key authentication for third-party services

 ## Installation
 1. Clone the repository
 2. Create a .env file with the following variables:
    - DB_URL=postgresql:username:password@localhost:5432/chirpy
    - JWT_SECRET=your_jwt_secret_key
    - POLKA_KEY=your_polka_api_key
    - PLATFORM=dev (or "prod" for production)
 3. Run database migrations: `goose postgres $DB_URL up`
 4. Start the server: `go run .`

 ## API Endpoints

 ### Authentication
 - `POST /api/users` - Register a new user
 - `POST /api/login` - Login and get access/refresh tokens
 - `POST /api/refresh` - Get a new access token using refresh token
 - `POST /api/revoke` - Revoke a refresh token

 ### User Management
 - `PUT /api/users` - Update user email or password

 ### Chirps
 - `POST /api/chirps` - Create a new chirp
 - `GET /api/chirps` - Get all chirps (supports filtering and sorting)
 - `GET /api/chirps/{chirpID}` - Get a specific chirp
 - `DELETE /api/chirps/{chirpID}` - Delete a chirp (user must be author)

 ### Webhooks
 - `POST /api/polka/webhooks` - Process webhook events from Polka

 ### Admin/System
 - `GET /api/healthz` - Health check endpoint
 - `GET /admin/metrics` - View metrics (hits counter)
 - `POST /admin/reset` - Reset system (dev mode only)

 ## Authentication
 The API uses JWT for authentication with a two-token system:
 - Access tokens: Short-lived (1 hour) for API access
 - Refresh tokens: Long-lived (60 days) for obtaining new access tokens

 Include the access token in requests with the header:
 `Authorization: Bearer {token}`

 ## Database Structure
 - `users`: User accounts including hashed passwords
 - `chirps`: Short messages with author references
 - `refresh_tokens`: Token storage with expiration and revocation support

 ## Contributing
 Pull requests are welcome. For major changes, please open an issue first
 to discuss what you would like to change.

 ## License
 [MIT](https:choosealicense.com/licenses/mit/)