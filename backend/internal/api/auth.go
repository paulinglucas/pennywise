package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/jamespsullivan/pennywise/internal/db/queries"
	"github.com/jamespsullivan/pennywise/internal/middleware"
	"github.com/jamespsullivan/pennywise/internal/models"
)

const tokenTTL = 24 * time.Hour
const maxUsers = 10

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	CountUsers(ctx context.Context) (int, error)
	CreateUser(ctx context.Context, user *models.User) error
}

type AppHandler struct {
	StubHandler
	users             UserRepository
	accounts          AccountRepository
	transactions      TransactionRepository
	transactionGroups TransactionGroupRepository
	assets            AssetRepository
	goals             GoalRepository
	goalContributions GoalContributionRepository
	recurring         RecurringRepository
	alerts            AlertRepository
	dashboard         DashboardRepository
	auditLog          AuditLogWriter
	dlq               FailedRequestWriter
	secret            []byte
	startTime         time.Time
	bcryptCost        int
}

func NewAppHandler(users UserRepository, accounts AccountRepository, transactions TransactionRepository, transactionGroups TransactionGroupRepository, assets AssetRepository, goals GoalRepository, goalContributions GoalContributionRepository, recurring RecurringRepository, alerts AlertRepository, dashboard DashboardRepository, auditLog AuditLogWriter, dlq FailedRequestWriter, secret []byte) *AppHandler {
	return &AppHandler{
		users:             users,
		accounts:          accounts,
		transactions:      transactions,
		transactionGroups: transactionGroups,
		assets:            assets,
		goals:             goals,
		goalContributions: goalContributions,
		recurring:         recurring,
		alerts:            alerts,
		dashboard:         dashboard,
		auditLog:          auditLog,
		dlq:               dlq,
		secret:            secret,
		startTime:         time.Now(),
		bcryptCost:        bcrypt.DefaultCost,
	}
}

func (h *AppHandler) SetBcryptCost(cost int) {
	h.bcryptCost = cost
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

func (h *AppHandler) PostAuthRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	requestID := middleware.GetRequestID(r.Context())

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid request body", requestID)
		return
	}

	if len(req.Password) < 8 {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Password must be at least 8 characters", requestID)
		return
	}

	count, err := h.users.CountUsers(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to check user count", requestID)
		return
	}
	if count >= maxUsers {
		WriteError(w, http.StatusConflict, CONFLICT, "Maximum number of users reached", requestID)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), h.bcryptCost)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to hash password", requestID)
		return
	}

	user := &models.User{
		ID:           uuid.New().String(),
		Email:        string(req.Email),
		Name:         req.Name,
		PasswordHash: string(hash),
	}

	if err := h.users.CreateUser(r.Context(), user); err != nil {
		if errors.Is(err, queries.ErrEmailTaken) {
			WriteError(w, http.StatusConflict, CONFLICT, "Email already taken", requestID)
			return
		}
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to create user", requestID)
		return
	}

	claims := middleware.NewClaims(user.ID, user.Email, tokenTTL)
	tokenStr, err := middleware.SignToken(claims, h.secret)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to create token", requestID)
		return
	}

	setTokenCookie(w, tokenStr, int(tokenTTL.Seconds()))

	WriteJSON(w, http.StatusCreated, LoginResponse{
		User: UserResponse{
			Id:    ParseID(user.ID),
			Email: user.Email,
			Name:  user.Name,
		},
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
