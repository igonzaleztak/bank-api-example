package schemas

// HealthResponse is the response for the health check endpoint
type HealthResponse struct {
	Message string `json:"message"`
}

// OkResponse is the response for a successful operation
type OkResponse struct {
	Message string `json:"message"`
}
