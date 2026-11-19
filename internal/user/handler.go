package user

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"mksiaga/internal/auth"
	"mksiaga/internal/navigation"
)

type Handler struct {
	appName    string
	service    *Service
	navigation *navigation.Service
}

func NewHandler(appName string, service *Service, navigationService *navigation.Service) *Handler {
	return &Handler{appName: appName, service: service, navigation: navigationService}
}

func (h *Handler) Index(c *gin.Context) {
	users, err := h.service.Users(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	stores, roles, err := h.service.Options(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	employees, err := h.service.Employees(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	active, inactive, assignments := 0, 0, 0
	for _, item := range users {
		if item.Status == "active" {
			active++
		} else {
			inactive++
		}
		assignments += item.StoreCount
	}
	h.render(c, http.StatusOK, "user/index.html", gin.H{"Title": "User", "ActivePage": "users", "Users": users, "Stores": stores, "Roles": roles, "Employees": employees, "UserCount": len(users), "ActiveUserCount": active, "InactiveUserCount": inactive, "StoreAssignmentCount": assignments})
}
func (h *Handler) Create(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "csrf")
		return
	}
	input, ok := formInput(c)
	if !ok {
		h.redirect(c, "error", "invalid")
		return
	}
	h.redirectResult(c, h.service.Create(c.Request.Context(), input), "user-created")
}
func (h *Handler) Update(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "csrf")
		return
	}
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "error", "invalid")
		return
	}
	input, ok := formInput(c)
	if !ok {
		h.redirect(c, "error", "invalid")
		return
	}
	h.redirectResult(c, h.service.Update(c.Request.Context(), id, input), "user-updated")
}
func (h *Handler) Delete(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "csrf")
		return
	}
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "error", "invalid")
		return
	}
	current, _ := auth.CurrentUser(c)
	currentID, _ := strconv.ParseUint(current.ID, 10, 64)
	h.redirectResult(c, h.service.Delete(c.Request.Context(), id, currentID), "user-deleted")
}

func (h *Handler) StoreAssignments(c *gin.Context) {
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "error", "invalid")
		return
	}
	item, assignments, roles, err := h.service.StoreAssignments(c.Request.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		h.redirect(c, "error", "not-found")
		return
	}
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.render(c, http.StatusOK, "user/stores.html", gin.H{"Title": "Store user", "ActivePage": "users", "ManagedUser": item, "Assignments": assignments, "Roles": roles})
}

func (h *Handler) UpdateStoreAssignments(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "csrf")
		return
	}
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "error", "invalid")
		return
	}
	defaultStoreID, err := strconv.ParseUint(c.PostForm("default_store_id"), 10, 64)
	if err != nil {
		h.redirect(c, "error", "invalid")
		return
	}
	inputs := make([]AssignmentInput, 0)
	for _, rawStoreID := range c.PostFormArray("store_ids") {
		storeID, storeErr := strconv.ParseUint(rawStoreID, 10, 64)
		roleID, roleErr := strconv.ParseUint(c.PostForm("role_"+rawStoreID), 10, 64)
		if storeErr != nil || roleErr != nil {
			h.redirect(c, "error", "invalid")
			return
		}
		inputs = append(inputs, AssignmentInput{StoreID: storeID, RoleID: roleID, IsDefault: storeID == defaultStoreID})
	}
	if err := h.service.SyncStoreAssignments(c.Request.Context(), id, inputs); err != nil {
		h.redirectResult(c, err, "user-stores-updated")
		return
	}
	c.Redirect(http.StatusSeeOther, "/settings/users/"+strconv.FormatUint(id, 10)+"/stores?success=user-stores-updated")
}

func formInput(c *gin.Context) (Input, bool) {
	storeID, e1 := strconv.ParseUint(c.PostForm("store_id"), 10, 64)
	roleID, e2 := strconv.ParseUint(c.PostForm("role_id"), 10, 64)
	if e1 != nil || e2 != nil {
		return Input{}, false
	}
	return Input{EmployeeNumber: c.PostForm("employee_number"), Name: c.PostForm("name"), Username: c.PostForm("username"), Email: c.PostForm("email"), Phone: c.PostForm("phone"), Password: c.PostForm("password"), Status: c.PostForm("status"), StoreID: storeID, RoleID: roleID}, true
}
func routeID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return id, err == nil && id > 0
}

func (h *Handler) render(c *gin.Context, status int, templateName string, data gin.H) {
	current, _ := auth.CurrentUser(c)
	menus, err := h.navigation.Menus(c.Request.Context(), current, "users")
	if err != nil {
		h.internalError(c, err)
		return
	}
	token, err := auth.CSRFToken(c)
	if err != nil {
		h.internalError(c, err)
		return
	}
	data["AppName"], data["IsDashboard"], data["Authenticated"] = h.appName, true, true
	data["Section"] = "Pengelolaan"
	data["UserName"], data["UserInitials"] = current.Name, initials(current.Name)
	data["StoreName"], data["RoleName"], data["CSRFToken"] = current.StoreName, current.RoleName, token
	data["Menus"] = menus
	data["SuccessMessage"] = message(c.Query("success"))
	data["ErrorMessage"] = message(c.Query("error"))
	c.HTML(status, templateName, data)
}
func (h *Handler) redirectResult(c *gin.Context, err error, success string) {
	if err == nil {
		h.redirect(c, "success", success)
		return
	}
	code := "failed"
	switch {
	case errors.Is(err, ErrInvalid):
		code = "invalid"
	case errors.Is(err, ErrDuplicate):
		code = "duplicate"
	case errors.Is(err, ErrProtected):
		code = "protected"
	case errors.Is(err, sql.ErrNoRows):
		code = "not-found"
	default:
		slog.Error("user operation failed", "error", err)
	}
	h.redirect(c, "error", code)
}
func (h *Handler) redirect(c *gin.Context, key, value string) {
	c.Redirect(http.StatusSeeOther, "/settings/users?"+key+"="+url.QueryEscape(value))
}
func (h *Handler) internalError(c *gin.Context, err error) {
	slog.Error("render users", "error", err)
	c.Redirect(http.StatusSeeOther, "/errors/500")
}
func initials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "US"
	}
	value := string([]rune(strings.ToUpper(parts[0]))[0])
	if len(parts) > 1 {
		value += string([]rune(strings.ToUpper(parts[len(parts)-1]))[0])
	}
	return value
}
func message(code string) string {
	return map[string]string{"user-created": "User berhasil dibuat.", "user-updated": "User berhasil diperbarui.", "user-deleted": "User berhasil dinonaktifkan dan dihapus dari daftar.", "user-stores-updated": "Assignment store dan role user berhasil disimpan.", "csrf": "Sesi form berakhir. Silakan muat ulang halaman.", "invalid": "Data user belum lengkap atau tidak valid.", "duplicate": "Username, email, atau nomor pegawai sudah digunakan.", "protected": "Akun yang sedang digunakan tidak dapat dihapus.", "not-found": "User tidak ditemukan.", "failed": "Operasi user gagal diproses."}[code]
}
