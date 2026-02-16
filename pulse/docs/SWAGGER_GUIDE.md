# Swagger/OpenAPI Documentation Guide

## Overview

The Node Pulse API now includes interactive API documentation powered by Swagger/OpenAPI. The documentation is automatically generated from code annotations and provides a user-friendly interface for exploring and testing the API.

## Accessing Swagger UI

Once the server is running, access the Swagger UI at:

```
http://localhost:6532/swagger/index.html
```

## Adding Swagger Annotations to New Endpoints

To add documentation to your API endpoints, use the following Swagger annotation format above your handler functions:

### Basic Annotation Structure

```go
// @Summary		Short description
// @Description	Longer description with details
// @Tags			category
// @Accept			json
// @Produce		json
// @Param			name		type		required	description
// @Success		200			{object}	TypeName	"Success message"
// @Failure		400			{object}	models.ErrorResponse	"Error description"
// @Security		BearerAuth
// @Router			/path [method]
func (h *Handler) MethodName(c *gin.Context) {
    // Handler implementation
}
```

### Common Annotations

#### Summary
A one-line description of what the endpoint does.

```go
// @Summary		Create a new node
```

#### Description
Detailed explanation. Can use multiple lines for formatted text.

```go
// @Description	Creates a new monitoring node with the provided configuration.
// @Description
// @Description	**Validation rules:**
// @Description	- Name must be unique per region
// @Description	- IP must be a valid IPv4 or IPv6 address
```

#### Tags
Used to group endpoints in the Swagger UI.

```go
// @Tags		nodes
// @Tags		auth
// @Tags		health
```

#### Accept/Produce
Content types accepted and produced.

```go
// @Accept		json
// @Produce		json
```

#### Parameters

**Query parameter:**
```go
// @Param		region		query		string		false	"Filter by region"	example(us-west)
```

**Path parameter:**
```go
// @Param		id			path		string		true	"Node UUID"
```

**Body parameter:**
```go
// @Param		request		body		models.CreateNodeRequest	true	"Node creation request"
```

#### Responses

**Success response:**
```go
// @Success		200			{object}	models.Node	"Node details"
// @Success		201			{object}	models.Node	"Node created"
```

**Error responses:**
```go
// @Failure		400			{object}	models.ErrorResponse	"Invalid request"
// @Failure		401			{object}	models.ErrorResponse	"Unauthorized"
// @Failure		404			{object}	models.ErrorResponse	"Not found"
// @Failure		500			{object}	models.ErrorResponse	"Internal server error"
```

#### Security
Indicates if the endpoint requires authentication.

```go
// @Security		BearerAuth
```

For public endpoints (no auth required), omit the `@Security` annotation.

#### Router
Defines the endpoint path and HTTP method.

```go
// @Router			/nodes [post]
// @Router			/nodes/{id} [get]
// @Router			/auth/login [post]
```

### Complete Examples

#### POST Endpoint (With Auth)
```go
// CreateNodeHandler handles POST /api/v1/nodes
// @Summary		Create a new node
// @Description	Creates a new monitoring node. If a node with the same name and IP exists, it will be updated instead.
// @Tags			nodes
// @Accept			json
// @Produce		json
// @Param			request	body		models.CreateNodeRequest	true	"Node creation request"
// @Success		201		{object}	models.Node	"Node created successfully"
// @Success		200		{object}	models.Node	"Node updated successfully (duplicate found)"
// @Failure		400		{object}	models.ErrorResponse	"Invalid request parameters"
// @Failure		401		{object}	models.ErrorResponse	"Unauthorized"
// @Failure		403		{object}	models.ErrorResponse	"Forbidden (requires admin or operator role)"
// @Failure		500		{object}	models.ErrorResponse	"Internal server error"
// @Security		BearerAuth
// @Router			/nodes [post]
func (h *NodeHandler) CreateNodeHandler(c *gin.Context) {
    // Implementation
}
```

#### GET Endpoint (Public, No Auth)
```go
// Handler returns a Gin handler for health check
// @Summary		Health check
// @Description	Returns the health status of the API service.
// @Tags			health
// @Accept			json
// @Produce		json
// @Success		200	{object}	HealthResponse	"Health status"
// @Router			/health [get]
func (h *HealthChecker) Handler(c *gin.Context) {
    // Implementation
}
```

#### GET Endpoint with Query Parameters
```go
// GetNodesHandler handles GET /api/v1/nodes
// @Summary		List all nodes
// @Description	Retrieves a list of all nodes. Supports filtering by region.
// @Tags			nodes
// @Accept			json
// @Produce		json
// @Param			region		query		string	false	"Filter by region"	example(us-west)
// @Success		200		{array}		models.Node	"List of nodes"
// @Failure		401		{object}	models.ErrorResponse	"Unauthorized"
// @Failure		500		{object}	models.ErrorResponse	"Internal server error"
// @Security		BearerAuth
// @Router			/nodes [get]
func (h *NodeHandler) GetNodesHandler(c *gin.Context) {
    // Implementation
}
```

## Regenerating Documentation

After adding or modifying Swagger annotations, regenerate the documentation:

```bash
# Using Makefile (recommended)
make swag

# Or directly using swag command
swag init -g cmd/server/main.go -o api/docs --parseDependency --parseInternal
```

This will update:
- `api/docs/docs.go` - Go code for serving Swagger UI
- `api/docs/swagger.json` - OpenAPI specification (JSON)
- `api/docs/swagger.yaml` - OpenAPI specification (YAML)

## Project-Specific Notes

### Main API Configuration

The main API information (title, version, description, host) is defined in `cmd/server/main.go`:

```go
// @title			Node Pulse API
// @version		1.0
// @description	Node Pulse is a distributed monitoring system...
// @host			localhost:6532
// @BasePath		/api/v1
```

### Authentication

The project uses Bearer token authentication. The security definition is configured in `main.go`:

```go
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name							Authorization
// @description					Enter the token with the `Bearer ` prefix
```

### Route Configuration

Swagger UI is served at `/swagger/*any` in `internal/api/routes.go`:

```go
router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
```

## Best Practices

1. **Always add annotations for new endpoints** - Keep documentation in sync with code
2. **Use descriptive summaries** - Make the API easy to understand
3. **Include all error cases** - Document 400, 401, 403, 404, 500 responses
4. **Group related endpoints with tags** - Use consistent tag names
5. **Provide examples in descriptions** - Show expected input/output formats
6. **Regenerate docs after changes** - Run `make swag` before committing

## Resources

- [Swagger Annotation Guide](https://swaggo.github.io/swag/go/decorator/)
- [OpenAPI Specification](https://swagger.io/specification/)
- [gin-swagger GitHub](https://github.com/swaggo/gin-swagger)

## Troubleshooting

### Import errors
If you see errors about missing imports, ensure you have the dependencies:
```bash
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files
go install github.com/swaggo/swag/cmd/swag@latest
```

### Docs not updating
- Make sure to run `make swag` after adding annotations
- Check that the import path in `main.go` is correct: `_ "github.com/whg517/node-pulse/pulse/api/docs"`
- Restart the server after regenerating docs

### Swagger UI not accessible
- Verify the route is registered in `routes.go`
- Check the server is running on the correct port (default 6532)
- Ensure no middleware is blocking the `/swagger/*` path
