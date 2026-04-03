# API Documentation for Franchise Management Platform

## Base URL
`http://localhost:8080/api/v1` (development)
`https://api.yourdomain.com/api/v1` (production)

## Authentication
All protected endpoints require a Bearer token in the Authorization header:
```
Authorization: Bearer <jwt_token>
```

## Endpoints

### Authentication

#### POST /auth/register
Register a new user
```json
{
  "email": "user@example.com",
  "password": "securepassword",
  "role": "dealer",
  "tenant_id": "tenant-123",
  "first_name": "John",
  "last_name": "Doe"
}
```

#### POST /auth/login
Authenticate user and get tokens
```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

Response:
```json
{
  "user": { ... },
  "token": "jwt_token",
  "refresh_token": "refresh_token"
}
```

#### POST /auth/logout
Invalidate user session (requires authentication)

#### POST /auth/refresh
Refresh authentication token
```json
{
  "refresh_token": "refresh_token"
}
```

### Users

#### GET /users/profile
Get authenticated user's profile (requires authentication)

#### PUT /users/profile
Update authenticated user's profile
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+7 (999) 123-45-67"
}
```

### Checklists

#### GET /checklists
Get all checklists for authenticated user (requires authentication)
Query parameters:
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 10, max: 100)

#### GET /checklists/:id
Get a specific checklist by ID (requires authentication)

#### POST /checklists
Create a new checklist (requires authentication)
```json
{
  "title": "Daily Tasks",
  "description": "Tasks to complete today",
  "tasks": [
    {
      "title": "Call client",
      "description": "Call important client",
      "order": 1
    }
  ]
}
```

#### PUT /checklists/:id
Update an existing checklist (requires authentication)

#### DELETE /checklists/:id
Delete a checklist (requires authentication)

#### POST /checklists/:id/complete
Mark a checklist as completed (requires authentication)

### Dealers (Franchiser only)

#### GET /dealers
Get all dealers in the franchise network (requires franchiser role)
Query parameters:
- `type`: Filter by type (default: "dealer")

#### GET /dealers/:id
Get a specific dealer by ID (requires franchiser role)

## Error Responses

All error responses follow this format:
```json
{
  "error": "Error type",
  "message": "Human-readable error message"
}
```

## Success Responses

Most successful responses follow this format:
```json
{
  "message": "Success message"
}
```

Or return the created/updated resource object.