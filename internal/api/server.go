package api

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
)

//go:embed panel/*
var panelFS embed.FS

// Start starts the API server on the given port
func Start(port int) error {
	mux := http.NewServeMux()
	RegisterRoutes(mux)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("TAVPBox Panel running at http://localhost%s\n", addr)
	fmt.Printf("Dashboard: http://localhost%s\n", addr)
	return http.ListenAndServe(addr, mux)
}

// panelSubFS returns the embedded panel/ subdirectory as an fs.FS
func panelSubFS() fs.FS {
	fsys, err := fs.Sub(panelFS, "panel")
	if err != nil {
		panic(err)
	}
	return fsys
}
