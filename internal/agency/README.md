# Agency Management Package

This package implements the agency management system for CodeValdCortex. Agencies represent use cases (e.g., water distribution, logistics, etc.) as first-class entities with their own configuration, agents, and context.

## Architecture

The package follows a clean architecture pattern with the following components:

### Core Components

- **types.go**: Data models and types for agencies
- **service.go**: Business logic and service interface
- **repository.go**: Data persistence interface
- **repository_arango.go**: ArangoDB implementation
- **validator.go**: Validation logic for agency data
- **context.go**: Context management for request scoping

### Data Model

The `Agency` type represents a use case with the following key fields:

```go
type Agency struct {
    ID          string            // Unique identifier (e.g., UC-INFRA-001)
    Name        string            // Human-readable name
    DisplayName string            // UI display name
    Description string            // Description
    Category    string            // Category (infrastructure, agriculture, etc.)
    Icon        string            // Emoji icon for UI
    Status      AgencyStatus      // active, inactive, paused, archived
    Metadata    AgencyMetadata    // Additional metadata
    Settings    AgencySettings    // Configuration settings
    CreatedAt   time.Time         // Creation timestamp
    UpdatedAt   time.Time         // Last update timestamp
}
```

## Usage

### Service Initialization

```go
// Create repository
repo, err := agency.NewArangoRepository(db)
if err != nil {
    log.Fatal(err)
}

// Create validator
validator := agency.NewValidator()

// Create service
service := agency.NewService(repo, validator)
```

### Creating an Agency

```go
newAgency := &models.Agency{
    ID:          "UC-INFRA-001",
    Name:        "Water Distribution Network",
    DisplayName: "💧 Water Distribution",
    Category:    "infrastructure",
    Status:      models.AgencyStatusActive,
}

err := service.CreateAgency(ctx, newAgency)
```

### Retrieving Agencies

```go
// Get single agency
agency, err := service.GetAgency(ctx, "UC-INFRA-001")

// List with filters
filters := models.AgencyFilters{
    Category: "infrastructure",
    Status:   models.AgencyStatusActive,
    Search:   "water",
    Limit:    10,
}
agencies, err := service.ListAgencies(ctx, filters)
```

### Managing Active Agency

```go
// Set active agency
err := service.SetActiveAgency(ctx, "UC-INFRA-001")

// Get active agency
active, err := service.GetActiveAgency(ctx)
```

### Context Management

```go
// Create context manager
cm := agency.NewContextManager(service)

// Inject agency into context
ctx, err := cm.WithAgency(ctx, "UC-INFRA-001")

// Retrieve from context
agency, err := agency.GetAgencyFromContext(ctx)
agencyID, err := agency.GetAgencyIDFromContext(ctx)
```

## API Handlers

The package provides HTTP handlers using Gin framework:

```go
// Create handler
handler := handlers.NewAgencyHandler(service)

// Register routes
api := router.Group("/api/v1")
handler.RegisterRoutes(api)
```

### Available Endpoints

- `GET /api/v1/agencies` - List all agencies
- `POST /api/v1/agencies` - Create new agency
- `GET /api/v1/agencies/:id` - Get agency details
- `PUT /api/v1/agencies/:id` - Update agency
- `DELETE /api/v1/agencies/:id` - Delete agency
- `POST /api/v1/agencies/:id/activate` - Set as active agency
- `GET /api/v1/agencies/active` - Get current active agency
- `GET /api/v1/agencies/:id/statistics` - Get agency statistics

## Database Schema

The agencies are stored in ArangoDB with the following indexes:

- Unique index on `id` field
- Index on `category` field
- Index on `status` field
- Compound index on `category` and `status`

## Migration

Use the migration script to import existing use cases:

```bash
go run scripts/migrate-agencies.go
```

The script will:
1. Scan the `/usecases` directory
2. Parse use case folder names
3. Create agency records in the database
4. Skip existing agencies

## Validation Rules

- Agency ID must start with "UC-"
- Required fields: ID, Name, DisplayName, Category
- Status must be one of: active, inactive, paused, archived
- Config path must be absolute or start with "/usecases/"

## Testing

Run tests with:

```bash
go test ./internal/agency/...
```

## Future Enhancements

- Agent count tracking
- Real-time statistics
- Agency templates
- Bulk operations
- Agency cloning
- Import/export functionality
