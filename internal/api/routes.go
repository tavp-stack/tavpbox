package api

import "net/http"

// RegisterRoutes registers all API routes
func RegisterRoutes(mux *http.ServeMux) {
	// API endpoints
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("GET /api/projects", handleListProjects)
	mux.HandleFunc("GET /api/projects/{name}", handleGetProject)
	mux.HandleFunc("POST /api/projects", handleCreateProject)
	mux.HandleFunc("POST /api/projects/{name}/start", handleStartProject)
	mux.HandleFunc("POST /api/projects/{name}/stop", handleStopProject)
	mux.HandleFunc("POST /api/projects/{name}/restart", handleRestartProject)
	mux.HandleFunc("DELETE /api/projects/{name}", handleDestroyProject)
	mux.HandleFunc("GET /api/projects/{name}/logs", handleProjectLogs)
	mux.HandleFunc("GET /api/recipes", handleListRecipes)
	mux.HandleFunc("GET /api/services", handleListServices)

	// Serve embedded panel frontend
	panelFS := panelSubFS()
	fileServer := http.FileServerFS(panelFS)
	mux.Handle("/", fileServer)
}
