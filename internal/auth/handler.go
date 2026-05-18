package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/asey0u/attendance-service/internal/httpx"
)

const CookieName = "auth_token"

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	token, claims, err := h.svc.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	setAuthCookie(w, token)
	httpx.JSON(w, http.StatusOK, loginResponse{Token: token, Role: claims.Role})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	clearAuthCookie(w)
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type setupAdminRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *Handler) SetupAdmin(w http.ResponseWriter, r *http.Request) {
	var req setupAdminRequest
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	if err := h.svc.BootstrapAdmin(r.Context(), req.Login, req.Password); err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims := FromContext(r.Context())
	if claims == nil {
		httpx.WriteError(w, httpx.ErrUnauthorized)
		return
	}

	user, err := h.svc.GetUser(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, user)
}

type createUserRequest struct {
	Login      string `json:"login"`
	Password   string `json:"password"`
	Role       string `json:"role"`
	EmployeeID *int   `json:"employee_id,omitempty"`
}

func (h *Handler) AdminCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	id, err := h.svc.CreateUser(r.Context(), req.Login, req.Password, req.Role, req.EmployeeID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusCreated, map[string]int{"id": id})
}

func (h *Handler) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.ListUsers(r.Context(), 0, 0)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, users)
}

func (h *Handler) AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseIntParam(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	if err = h.svc.DeleteUser(r.Context(), id); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			httpx.WriteError(w, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

type updateRoleRequest struct {
	Role string `json:"role"`
}

func (h *Handler) AdminUpdateRole(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseIntParam(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	var req updateRoleRequest
	if err = httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	if err = h.svc.UpdateRole(r.Context(), id, req.Role); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			httpx.WriteError(w, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(TokenTTL),
	})
}

func clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
