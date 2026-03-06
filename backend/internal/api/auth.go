package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/jamespsullivan/pennywise/internal/middleware"
	"github.com/jamespsullivan/pennywise/internal/models"
)

const tokenTTL = 24 * time.Hour

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
}

type AppHandler struct {
	StubHandler
	users    UserRepository
	accounts AccountRepository
	secret   []byte
}

func NewAppHandler(users UserRepository, accounts AccountRepository, secret []byte) *AppHandler {
	return &AppHandler{
		users:    users,
		accounts: accounts,
		secret:   secret,
	}
}

func (h *AppHandler) PostAuthLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		requestID := middleware.GetRequestID(r.Context())
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid request body", requestID)
		return
	}

	requestID := middleware.GetRequestID(r.Context())

	user, err := h.users.GetByEmail(r.Context(), string(req.Email))
	if err != nil {
		WriteError(w, http.StatusUnauthorized, UNAUTHORIZED, "Invalid email or password", requestID)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		WriteError(w, http.StatusUnauthorized, UNAUTHORIZED, "Invalid email or password", requestID)
		return
	}

	claims := middleware.NewClaims(user.ID, user.Email, tokenTTL)
	tokenStr, err := middleware.SignToken(claims, h.secret)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to create token", requestID)
		return
	}

	setTokenCookie(w, tokenStr, int(tokenTTL.Seconds()))

	WriteJSON(w, http.StatusOK, LoginResponse{
		User: UserResponse{
			Id:    ParseID(user.ID),
			Email: user.Email,
			Name:  user.Name,
		},
	})
}

func (h *AppHandler) PostAuthLogout(w http.ResponseWriter, r *http.Request) {
	setTokenCookie(w, "", -1)
	w.WriteHeader(http.StatusNoContent)
}

func setTokenCookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	})
}

func (h *AppHandler) GetAuthMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, UNAUTHORIZED, "User not found", requestID)
		return
	}

	WriteJSON(w, http.StatusOK, UserResponse{
		Id:    ParseID(user.ID),
		Email: user.Email,
		Name:  user.Name,
	})
}
