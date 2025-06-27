# ChirpStack API Documentation

ChirpStack provides a RESTful API for managing users, chirps (posts), and administrative functions. All endpoints return JSON unless otherwise specified.

## Base URL

```
http://localhost:8080
```

## Authentication

ChirpStack uses JWT (JSON Web Tokens) for authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

Refresh tokens are also supported for token renewal.

## API Endpoints

### Health Check

#### GET /api/health
Returns the health status of the API.

**Response:**
- **200 OK**: `"OK"` (text/plain)

### User Management

#### POST /api/users
Create a new user account.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response:**
- **201 Created**: User object (password field excluded)
- **400 Bad Request**: Invalid email or password too long
- **500 Internal Server Error**: Server error

**Example Response:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z",
  "is_chirpy_red": false
}
```

#### PUT /api/users
Update user information. Requires authentication.

**Request Body:**
```json
{
  "email": "newemail@example.com",
  "password": "newpassword123"
}
```

**Response:**
- **200 OK**: Updated user object
- **400 Bad Request**: Invalid email or password too long
- **401 Unauthorized**: Missing or invalid token
- **500 Internal Server Error**: Server error

### Authentication

#### POST /api/login
Authenticate a user and receive access and refresh tokens.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
- **200 OK**: User object with tokens
- **400 Bad Request**: Invalid email format
- **401 Unauthorized**: Invalid credentials
- **500 Internal Server Error**: Server error

**Example Response:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z",
  "is_chirpy_red": false,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "refresh_token_string"
}
```

#### POST /api/refresh
Refresh an access token using a refresh token.

**Headers:**
```
Authorization: Bearer <refresh-token>
```

**Response:**
- **200 OK**: New access token
- **401 Unauthorized**: Invalid refresh token

**Example Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### POST /api/revoke
Revoke a refresh token. Requires refresh token in Authorization header.

**Headers:**
```
Authorization: Bearer <refresh-token>
```

**Response:**
- **204 No Content**: Token revoked successfully
- **401 Unauthorized**: Invalid refresh token
- **500 Internal Server Error**: Server error

### Chirp Management

#### POST /api/chirps
Create a new chirp. Requires authentication.

**Request Body:**
```json
{
  "body": "This is my chirp content!"
}
```

**Response:**
- **201 Created**: Chirp object
- **400 Bad Request**: Chirp too long (>140 characters)
- **401 Unauthorized**: Missing or invalid token
- **500 Internal Server Error**: Server error

**Example Response:**
```json
{
  "id": "456e7891-e89b-12d3-a456-426614174001",
  "body": "This is my chirp content!",
  "created_at": "2024-01-01T12:30:00Z",
  "updated_at": "2024-01-01T12:30:00Z",
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Note:** Profanity is automatically filtered. The words "kerfuffle", "sharbert", and "fornax" are replaced with "****".

#### GET /api/chirps
Retrieve all chirps with optional filtering and sorting.

**Query Parameters:**
- `author_id` (optional): Filter chirps by author UUID
- `sort` (optional): Sort order ("desc" for newest first)

**Response:**
- **200 OK**: Array of chirp objects
- **204 No Content**: Invalid author_id format
- **500 Internal Server Error**: Server error

**Example Response:**
```json
[
  {
    "id": "456e7891-e89b-12d3-a456-426614174001",
    "body": "This is my chirp content!",
    "created_at": "2024-01-01T12:30:00Z",
    "updated_at": "2024-01-01T12:30:00Z",
    "user_id": "123e4567-e89b-12d3-a456-426614174000"
  }
]
```

#### GET /api/chirps/{id}
Retrieve a specific chirp by ID.

**Path Parameters:**
- `id`: Chirp UUID

**Response:**
- **200 OK**: Chirp object
- **400 Bad Request**: Missing or invalid chirp ID
- **404 Not Found**: Chirp not found
- **500 Internal Server Error**: Server error

#### DELETE /api/chirps/{id}
Delete a specific chirp. Requires authentication and ownership.

**Path Parameters:**
- `id`: Chirp UUID

**Headers:**
```
Authorization: Bearer <access-token>
```

**Response:**
- **204 No Content**: Chirp deleted successfully
- **400 Bad Request**: Missing or invalid chirp ID
- **403 Forbidden**: Not authorized (missing token or not chirp owner)
- **404 Not Found**: Chirp not found
- **500 Internal Server Error**: Server error

### Webhooks

#### POST /api/polka/webhooks
Handle webhook events from external services.

**Headers:**
```
Authorization: ApiKey <polka-api-key>
```

**Request Body:**
```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000"
  }
}
```

**Response:**
- **204 No Content**: Event processed successfully
- **401 Unauthorized**: Missing or invalid API key
- **404 Not Found**: User not found
- **500 Internal Server Error**: Server error

**Supported Events:**
- `user.upgraded`: Upgrades a user to Chirpy Red status

### Admin Endpoints

#### GET /admin/metrics
Display admin metrics page showing file server hit count.

**Response:**
- **200 OK**: HTML page with metrics (text/html)

#### POST /admin/reset
Reset the application state (development mode only).

**Response:**
- **200 OK**: Reset successful (development mode)
- **403 Forbidden**: Operation not allowed (production mode)
- **500 Internal Server Error**: Server error

### Static File Serving

#### GET /app/*
Serve static files from the application directory.

**Response:**
- **200 OK**: File content
- **404 Not Found**: File not found

**Headers:**
All static file responses include:
```
Cache-Control: no-cache
```

## Error Response Format

All error responses follow this format:

```json
{
  "error": "Error message description"
}
```

## Common HTTP Status Codes

- **200 OK**: Request successful
- **201 Created**: Resource created successfully
- **204 No Content**: Request successful, no content returned
- **400 Bad Request**: Invalid request data
- **401 Unauthorized**: Authentication required or failed
- **403 Forbidden**: Access denied
- **404 Not Found**: Resource not found
- **500 Internal Server Error**: Server error

## Rate Limiting

Currently, no rate limiting is implemented.

## Data Validation

### Email Validation
- Must be a valid email format according to Go's `net/mail` package

### Password Validation
- Maximum length: 72 characters (bcrypt limitation)
- No minimum length requirement

### Chirp Body Validation
- Maximum length: 140 characters
- Automatic profanity filtering applied

## Environment Variables

The API requires these environment variables:

- `DB_URL`: PostgreSQL connection string
- `PLATFORM`: "dev" or "prod"
- `JWT_TOKEN_SECRET`: Secret for JWT token signing
- `POLKA_KEY`: API key for webhook authentication