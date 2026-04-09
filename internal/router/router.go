package router

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	"somewebproject/internal/service"
	"somewebproject/internal/transport"
)

func NewRouter(authHandler *transport.Handler, productHandler *transport.ProductHandler, authMiddleware func(http.Handler) http.Handler, requireRoles func(...string) func(http.Handler) http.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /swagger/", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
	mux.HandleFunc("GET /swagger/doc.json", swaggerDocHandler)

	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/refresh", authHandler.Refresh)
	mux.Handle("GET /api/auth/me", authMiddleware(http.HandlerFunc(authHandler.Me)))

	mux.Handle("GET /api/users", chain(http.HandlerFunc(authHandler.ListUsers), authMiddleware, requireRoles(service.RoleAdmin)))
	mux.Handle("GET /api/users/{id}", chain(http.HandlerFunc(authHandler.GetUserByID), authMiddleware, requireRoles(service.RoleAdmin)))
	mux.Handle("PUT /api/users/{id}", chain(http.HandlerFunc(authHandler.UpdateUser), authMiddleware, requireRoles(service.RoleAdmin)))
	mux.Handle("DELETE /api/users/{id}", chain(http.HandlerFunc(authHandler.BlockUser), authMiddleware, requireRoles(service.RoleAdmin)))

	mux.Handle("POST /api/products", chain(http.HandlerFunc(productHandler.CreateProduct), authMiddleware, requireRoles(service.RoleSeller)))
	mux.Handle("GET /api/products", chain(http.HandlerFunc(productHandler.ListProducts), authMiddleware))
	mux.Handle("GET /api/products/{id}", chain(http.HandlerFunc(productHandler.GetProductByID), authMiddleware))
	mux.Handle("PUT /api/products/{id}", chain(http.HandlerFunc(productHandler.UpdateProduct), authMiddleware, requireRoles(service.RoleSeller)))
	mux.Handle("DELETE /api/products/{id}", chain(http.HandlerFunc(productHandler.DeleteProduct), authMiddleware, requireRoles(service.RoleAdmin)))

	return mux
}

func chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}

func swaggerDocHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(swaggerSpec))
}
