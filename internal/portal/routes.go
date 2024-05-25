// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package portal

import (
	"net/http"
)

// register provided routes to http.ServerMux
func registerRoutes(
	mux *http.ServeMux,
	routes map[string]http.Handler,
) {
	for route, handler := range routes {
		mux.Handle(route, handler)
	}
}

func (p *Portal) addRoutes() map[string]http.Handler {
	routes := make(map[string]http.Handler)

	//NOTE: if middleware is needed, it can be added right here

	routes["GET /"] = http.HandlerFunc(p.home)
	routes["GET /login"] = http.HandlerFunc(p.login)
	routes["GET /register"] = http.HandlerFunc(p.register)

	return routes
}
