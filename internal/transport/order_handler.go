package transport

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"somewebproject/internal/auth"
	"somewebproject/internal/models"
	"somewebproject/internal/service"

	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/webhook"
)

type OrderHandler struct {
	CartService  service.CartService
	OrderService service.OrderService
	WebhookKey   string
}

func NewOrderHandler(cartService service.CartService, orderService service.OrderService, webhookKey string) *OrderHandler {
	return &OrderHandler{CartService: cartService, OrderService: orderService, WebhookKey: webhookKey}
}

func (h *OrderHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.CartService.GetByUserID(r.Context(), principal.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toCartResponse(items))
}

func (h *OrderHandler) SyncCart(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req SyncCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	lines := make([]service.CartLineInput, 0, len(req.Items))
	for i := range req.Items {
		lines = append(lines, service.CartLineInput{
			ProductID: req.Items[i].ProductID,
			Quantity:  req.Items[i].Quantity,
		})
	}

	items, err := h.CartService.Sync(r.Context(), principal.ID, lines)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toCartResponse(items))
}

func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := h.OrderService.Checkout(r.Context(), principal.ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, CheckoutResponse{
		OrderID:            result.Order.ID,
		Status:             result.Order.Status,
		TotalAmount:        result.Order.TotalAmount,
		PaymentReference:   result.Order.PaymentRef,
		StripeClientSecret: result.ClientSecret,
	})
}

func (h *OrderHandler) ListMyOrders(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	orders, err := h.OrderService.ListMyOrders(r.Context(), principal.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toOrderListResponse(orders))
}

func (h *OrderHandler) ListAllOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.OrderService.ListAllOrders(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toOrderListResponse(orders))
}

func (h *OrderHandler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(h.WebhookKey) == "" {
		log.Println("[webhook] ERROR: webhook key not configured")
		writeError(w, http.StatusServiceUnavailable, "stripe webhook is not configured")
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[webhook] ERROR: failed to read payload: %v", err)
		writeError(w, http.StatusBadRequest, "invalid webhook payload")
		return
	}

	signature := r.Header.Get("Stripe-Signature")
	log.Printf("[webhook] Processing - payload: %d bytes, signature present: %v, key length: %d",
		len(payload), signature != "", len(h.WebhookKey))

	if signature != "" {
		log.Printf("[webhook] Signature header: %s...", signature[:min(50, len(signature))])
	}

	// Use ConstructEventWithOptions to ignore API version mismatch
	// stripe-go version might lag behind actual Stripe API versions
	event, err := webhook.ConstructEventWithOptions(payload, signature, h.WebhookKey, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if event.Type != "payment_intent.succeeded" {
		log.Printf("[webhook] ✓ Event type %s ignored (only payment_intent.succeeded is processed)", event.Type)
		w.WriteHeader(http.StatusOK)
		return
	}

	var intent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &intent); err != nil {
		log.Printf("[webhook] ERROR: failed to parse payment intent: %v", err)
		writeError(w, http.StatusBadRequest, "invalid payment intent payload")
		return
	}

	if strings.TrimSpace(intent.ID) == "" {
		log.Println("[webhook] ERROR: payment intent has no ID")
		writeError(w, http.StatusBadRequest, "missing payment reference")
		return
	}

	log.Printf("[webhook] Processing payment_intent.succeeded for payment_ref=%s, amount=%d", intent.ID, intent.Amount)

	if err := h.OrderService.MarkPaidByPaymentRef(r.Context(), intent.ID); err != nil {
		log.Printf("[webhook] ERROR: failed to mark payment: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to update order status")
		return
	}

	log.Printf("[webhook] ✓ Successfully processed payment_ref=%s", intent.ID)
	w.WriteHeader(http.StatusOK)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func toCartResponse(items []models.CartItem) CartResponse {
	response := CartResponse{Items: make([]CartItemResponse, 0, len(items))}

	for i := range items {
		lineTotal := items[i].Price * float64(items[i].Quantity)
		response.Total += lineTotal
		response.Items = append(response.Items, CartItemResponse{
			ID:       items[i].ID,
			Product:  toProductResponse(&items[i].Product),
			Price:    items[i].Price,
			Quantity: items[i].Quantity,
		})
	}

	return response
}

func toOrderListResponse(orders []models.Order) []OrderResponse {
	response := make([]OrderResponse, 0, len(orders))

	for i := range orders {
		items := make([]OrderItemResponse, 0, len(orders[i].Items))
		for j := range orders[i].Items {
			items = append(items, OrderItemResponse{
				ProductID:   orders[i].Items[j].ProductID,
				ProductName: orders[i].Items[j].ProductName,
				UnitPrice:   orders[i].Items[j].UnitPrice,
				Quantity:    orders[i].Items[j].Quantity,
				LineTotal:   orders[i].Items[j].LineTotal,
			})
		}

		response = append(response, OrderResponse{
			ID:              orders[i].ID,
			Status:          orders[i].Status,
			TotalAmount:     orders[i].TotalAmount,
			PaymentProvider: orders[i].PaymentProvider,
			PaymentRef:      orders[i].PaymentRef,
			Items:           items,
			CreatedAt:       orders[i].CreatedAt,
		})
	}

	return response
}
