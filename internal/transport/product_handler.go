package transport

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"somewebproject/internal/auth"
	"somewebproject/internal/models"
	"somewebproject/internal/service"
)

type ProductHandler struct {
	ProductService service.ProductService
}

func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{ProductService: productService}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req ProductCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.ProductService.Create(r.Context(), principal.ID, req.Name, req.Category, req.Description, req.Price, req.Stock)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toProductResponse(product))
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	filter := service.ProductListFilter{
		Query:       strings.TrimSpace(r.URL.Query().Get("search")),
		Category:    strings.TrimSpace(r.URL.Query().Get("category")),
		OnlyInStock: strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("in_stock")), "true"),
	}

	if minPriceText := strings.TrimSpace(r.URL.Query().Get("min_price")); minPriceText != "" {
		minPrice, err := strconv.ParseFloat(minPriceText, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid min_price")
			return
		}
		filter.MinPrice = &minPrice
	}

	if maxPriceText := strings.TrimSpace(r.URL.Query().Get("max_price")); maxPriceText != "" {
		maxPrice, err := strconv.ParseFloat(maxPriceText, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid max_price")
			return
		}
		filter.MaxPrice = &maxPrice
	}

	products, err := h.ProductService.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := make([]ProductResponse, 0, len(products))
	for i := range products {
		response = append(response, toProductResponse(&products[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseProductID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	product, err := h.ProductService.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toProductResponse(product))
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseProductID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req ProductUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updates := make(map[string]any)
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.Stock != nil {
		updates["stock"] = *req.Stock
	}

	updated, err := h.ProductService.Update(r.Context(), id, updates)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toProductResponse(updated))
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseProductID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.ProductService.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toProductResponse(product *models.Product) ProductResponse {
	return ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Category:    product.Category,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		OwnerID:     product.OwnerID,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}

func parseProductID(r *http.Request) (uint, error) {
	value := r.PathValue("id")
	if value == "" {
		return 0, errors.New("missing id")
	}

	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, errors.New("invalid id")
	}

	return uint(id), nil
}
