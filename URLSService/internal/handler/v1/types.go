package v1

// CreateURLRequest represents the request to create a short URL
type CreateURLRequest struct {
	URL string `json:"url" example:"https://example.com"`
	TTL int    `json:"ttl" example:"3600"`
}

// URLResponse represents the response after creating a short URL
type URLResponse struct {
	ShortURL  string `json:"short_url" example:"abc123"`
	ExpiresAt string `json:"expires_at" example:"2025-11-10T12:00:00Z"`
}

// URLDataResponse represents the response when retrieving URL data
type URLDataResponse struct {
	OriginalURL string `json:"original_url" example:"https://example.com"`
	ExpiresAt   string `json:"expires_at" example:"2025-11-10T12:00:00Z"`
}

// URLListItem represents a single URL in the list response
type URLListItem struct {
	ShortCode   string `json:"short_code" example:"abc123"`
	OriginalURL string `json:"original_url" example:"https://example.com"`
	ExpiresAt   string `json:"expires_at" example:"2025-11-10T12:00:00Z"`
	CreatedAt   string `json:"created_at" example:"2025-11-10T10:00:00Z"`
}

// URLListResponse represents the response when listing user URLs
type URLListResponse struct {
	URLs []URLListItem `json:"urls"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid request"`
	Message string `json:"message,omitempty" example:"URL is required"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// ReadinessResponse represents readiness check response
type ReadinessResponse struct {
	Status   string `json:"status" example:"ok"`
	Postgres string `json:"postgres,omitempty" example:"up"`
	Redis    string `json:"redis,omitempty" example:"up"`
	Reason   string `json:"reason,omitempty" example:"shutting down"`
}
