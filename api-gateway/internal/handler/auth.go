package handler

import (
	"bookshelf/api-gateway/internal/proxy"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
)

// AuthHandler обрабатывает запросы авторизации и пользователей.
type AuthHandler struct {
	proxy *proxy.ServiceProxy
}

// NewAuthHandler создаёт AuthHandler.
func NewAuthHandler(p *proxy.ServiceProxy) *AuthHandler {
	return &AuthHandler{proxy: p}
}

// Register проксирует запрос регистрации.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	h.proxy.ProxyToAuth(w, r, "/api/v1/auth/register")
}

// Login проксирует запрос авторизации.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.proxy.ProxyToAuth(w, r, "/api/v1/auth/login")
}

// Refresh проксирует запрос обновления токена.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	h.proxy.ProxyToAuth(w, r, "/api/v1/auth/refresh")
}

// Logout проксирует запрос выхода.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.proxy.ProxyToAuth(w, r, "/api/v1/auth/logout")
}

// GetCurrentUser проксирует запрос текущего пользователя.
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	h.proxy.ProxyToAuth(w, r, "/api/v1/users/me")
}

// GetUser проксирует запрос пользователя по ID.
func (h *AuthHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	path := "/api/v1/users/" + url.PathEscape(userID)
	h.proxy.ProxyToAuth(w, r, path)
}
