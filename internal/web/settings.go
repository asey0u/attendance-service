package web

import (
	"net/http"

	"github.com/asey0u/attendance-service/internal/auth"
	"github.com/asey0u/attendance-service/internal/httpx"
)

func (h *Handler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())
	threshold := h.settingsSvc.LateThreshold(r.Context())

	h.r.Page(w, http.StatusOK, "manager_settings", map[string]any{
		"Title":     "Настройки",
		"NavKey":    "settings",
		"Claims":    claims,
		"Threshold": threshold,
		"saved":     r.URL.Query().Get("saved"),
	})
}

func (h *Handler) SettingsSave(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		httpx.WriteError(w, httpx.ErrBadRequest)
		return
	}

	value := r.PostFormValue("late_threshold")
	if err := h.settingsSvc.SetLateThreshold(r.Context(), value); err != nil {
		claims := auth.FromContext(r.Context())
		threshold := h.settingsSvc.LateThreshold(r.Context())
		h.r.Page(w, http.StatusBadRequest, "manager_settings", map[string]any{
			"Title":     "Настройки",
			"NavKey":    "settings",
			"Claims":    claims,
			"Threshold": threshold,
			"Error":     err.Error(),
		})
		return
	}

	http.Redirect(w, r, "/manager/settings?saved=1", http.StatusSeeOther)
}
