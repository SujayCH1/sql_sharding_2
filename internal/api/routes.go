package api

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *Handler) {

	mux.HandleFunc(
		"/api/query/execute",
		handler.ExecuteQuery,
	)
}
