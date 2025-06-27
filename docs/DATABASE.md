# Database Documentation

ChirpStack uses PostgreSQL as its database with SQLC for type-safe query generation.

## Database Schema

### Tables

#### users
Stores user account information.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    email VARCHAR(255) UNIQUE NOT NULL,
    hashed_password VARCHAR(255) NOT NULL,
    is_chirpy_red BOOLEAN NOT NULL DEFAULT FALSE
);
```

**Columns:**
- `id`: Unique identifier (UUID)
- `created_at`: Timestamp when user was created
- `updated_at`: Timestamp when user was last updated
- `email`: User's email address (unique)
- `hashed_password`: Bcrypt-hashed password
- `is_chirpy_red`: Premium status flag

#### chirps
Stores user-generated content (chirps/tweets).

```sql
CREATE TABLE chirps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    body VARCHAR(140) NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE
);
```

**Columns:**
- `id`: Unique identifier (UUID)
- `created_at`: Timestamp when chirp was created
- `updated_at`: Timestamp when chirp was last updated
- `body`: Chirp content (max 140 characters)
- `user_id`: Foreign key to users table

#### refresh_tokens
Stores refresh tokens for authentication.

```sql
CREATE TABLE refresh_tokens (
    token VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL DEFAULT (NOW() + INTERVAL '60 days'),
    revoked_at TIMESTAMP
);
```

**Columns:**
- `token`: Primary key, the refresh token string
- `created_at`: Timestamp when token was created
- `updated_at`: Timestamp when token was last updated
- `user_id`: Foreign key to users table
- `expires_at`: Token expiration timestamp (60 days from creation)
- `revoked_at`: Timestamp when token was revoked (NULL if active)

## Migration Files

The database schema is managed through numbered migration files in `sql/schema/`:

1. `001_users.sql` - Creates users table
2. `002_chirps.sql` - Creates chirps table
3. `003_add_hashed_password_to_users.sql` - Adds password hashing
4. `004_refresh_tokens.sql` - Creates refresh tokens table
5. `005_add_is_chirpy_red_to_users.sql` - Adds premium status

## SQLC Configuration

SQLC generates type-safe Go code from SQL queries. Configuration is in `sqlc.yaml`:

```yaml
version: "2"
sql:
  - schema: "sql/schema"
    queries: "sql/queries"
    engine: "postgresql"
    gen:
      go:
        out: "internal/database"
        emit_json_tags: true
        overrides:
          - column: users.hashed_password
            go_struct_tag: json:"-"
```

### Generated Code

SQLC generates:
- `internal/database/db.go` - Database connection interface
- `internal/database/models.go` - Go structs for database tables
- `internal/database/*.sql.go` - Type-safe query functions

## Queries

SQL queries are defined in `sql/queries/` and categorized by table:

### User Queries (`users.sql`)

```sql
-- name: CreateUser :one
INSERT INTO users (email, hashed_password)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET email = $1, hashed_password = $2, updated_at = NOW()
WHERE id = $3
RETURNING *;

-- name: UpgradeUser :one
UPDATE users
SET is_chirpy_red = TRUE, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: TruncateUsers :exec
DELETE FROM users;
```

### Chirp Queries (`chirps.sql`)

```sql
-- name: CreateChirp :one
INSERT INTO chirps (body, user_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirps
ORDER BY created_at ASC;

-- name: GetChirpByID :one
SELECT * FROM chirps WHERE id = $1;

-- name: GetChirpsByAuthorID :many
SELECT * FROM chirps WHERE user_id = $1
ORDER BY created_at ASC;

-- name: DeleteChirpByID :exec
DELETE FROM chirps WHERE id = $1;
```

### Refresh Token Queries (`refresh_tokens.sql`)

```sql
-- name: StoreRefreshToken :one
INSERT INTO refresh_tokens (token, user_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT token, created_at, updated_at, user_id, expires_at, revoked_at
FROM refresh_tokens
WHERE 1=1
AND token = $1
AND revoked_at IS NULL;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE token = $1
RETURNING *;
```

## Database Connection

The application connects to PostgreSQL using the `DB_URL` environment variable:

```go
db, err := sql.Open("postgres", os.Getenv("DB_URL"))
```

Example connection string:
```
postgres://username:password@localhost/dbname?sslmode=disable
```

## Development Setup

1. Create a PostgreSQL database
2. Set the `DB_URL` environment variable
3. Run migrations in order from `sql/schema/`
4. Generate SQLC code: `sqlc generate`

## Data Types

### UUID Generation
All ID fields use PostgreSQL's `gen_random_uuid()` function for UUID generation.

### Timestamps
All timestamp fields use PostgreSQL's `NOW()` function for default values.

### JSON Serialization
The `hashed_password` field is excluded from JSON serialization for security.

## Indexes and Performance

Currently, the schema relies on:
- Primary key indexes on `id` columns
- Unique index on `users.email`
- Foreign key indexes on reference columns

For production, consider adding indexes on:
- `chirps.created_at` for chronological ordering
- `chirps.user_id` for author-based queries
- `refresh_tokens.expires_at` for token cleanup

## Security Considerations

1. **Password Storage**: Passwords are hashed using bcrypt before storage
2. **Token Security**: Refresh tokens have expiration and revocation capabilities
3. **Foreign Keys**: Cascade deletes ensure data consistency
4. **JSON Exclusion**: Sensitive fields are excluded from JSON serialization
