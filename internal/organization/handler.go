package organization

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

func (h *Handler) JobLevels(c *gin.Context) {
	items, err := h.service.JobLevels(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	used, positions := 0, 0
	for _, item := range items {
		positions += item.PositionCount
		if item.PositionCount > 0 {
			used++
		}
	}
	h.render(c, http.StatusOK, "organization/job_levels.html", "job-levels", gin.H{"Title": "Tingkatan Jabatan", "JobLevels": items, "JobLevelCount": len(items), "UsedCount": used, "PositionCount": positions})
}
func (h *Handler) CreateJobLevel(c *gin.Context) {
	if !h.validCSRF(c) {
		return
	}
	rank, err := strconv.ParseUint(c.PostForm("rank_order"), 10, 16)
	if err != nil {
		h.redirect(c, "job-levels", "error", "invalid")
		return
	}
	h.redirectResult(c, "job-levels", h.service.CreateJobLevel(c.Request.Context(), JobLevelInput{Code: c.PostForm("code"), Name: c.PostForm("name"), RankOrder: uint16(rank)}), "job-level-created")
}
func (h *Handler) UpdateJobLevel(c *gin.Context) {
	if !h.validCSRF(c) {
		return
	}
	id, ok := routeID(c)
	rank, err := strconv.ParseUint(c.PostForm("rank_order"), 10, 16)
	if !ok || err != nil {
		h.redirect(c, "job-levels", "error", "invalid")
		return
	}
	h.redirectResult(c, "job-levels", h.service.UpdateJobLevel(c.Request.Context(), id, JobLevelInput{Code: c.PostForm("code"), Name: c.PostForm("name"), RankOrder: uint16(rank)}), "job-level-updated")
}
func (h *Handler) DeleteJobLevel(c *gin.Context) {
	if !h.validCSRF(c) {
		return
	}
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "job-levels", "error", "invalid")
		return
	}
	h.redirectResult(c, "job-levels", h.service.DeleteJobLevel(c.Request.Context(), id), "job-level-deleted")
}

func (h *Handler) Positions(c *gin.Context) {
	items, err := h.service.Positions(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	levels, err := h.service.JobLevels(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	used, assignments := 0, 0
	for _, item := range items {
		assignments += item.AssignmentCount
		if item.AssignmentCount > 0 {
			used++
		}
	}
	h.render(c, http.StatusOK, "organization/positions.html", "positions", gin.H{"Title": "Jabatan", "Positions": items, "JobLevels": levels, "PositionCount": len(items), "UsedCount": used, "AssignmentCount": assignments})
}
func (h *Handler) CreatePosition(c *gin.Context) {
	if !h.validCSRF(c) {
		return
	}
	levelID, err := strconv.ParseUint(c.PostForm("job_level_id"), 10, 64)
	if err != nil {
		h.redirect(c, "positions", "error", "invalid")
		return
	}
	h.redirectResult(c, "positions", h.service.CreatePosition(c.Request.Context(), PositionInput{JobLevelID: levelID, Code: c.PostForm("code"), Name: c.PostForm("name")}), "position-created")
}
func (h *Handler) UpdatePosition(c *gin.Context) {
	if !h.validCSRF(c) {
		return
	}
	id, ok := routeID(c)
	levelID, err := strconv.ParseUint(c.PostForm("job_level_id"), 10, 64)
	if !ok || err != nil {
		h.redirect(c, "positions", "error", "invalid")
		return
	}
	h.redirectResult(c, "positions", h.service.UpdatePosition(c.Request.Context(), id, PositionInput{JobLevelID: levelID, Code: c.PostForm("code"), Name: c.PostForm("name")}), "position-updated")
}
func (h *Handler) DeletePosition(c *gin.Context) {
	if !h.validCSRF(c) {
		return
	}
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "positions", "error", "invalid")
		return
	}
	h.redirectResult(c, "positions", h.service.DeletePosition(c.Request.Context(), id), "position-deleted")
}

func (h *Handler) Departments(c *gin.Context) {
	items, err := h.service.Departments(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	used, assignments := 0, 0
	for _, item := range items {
		assignments += item.AssignmentCount
		if item.AssignmentCount > 0 {
			used++
		}
	}
	h.render(c, http.StatusOK, "organization/departments.html", "departments", gin.H{"Title": "Bagian", "Departments": items, "DepartmentCount": len(items), "UsedCount": used, "AssignmentCount": assignments})
}
func (h *Handler) CreateDepartment(c *gin.Context) {
	if !h.validCSRF(c) {
		return
	}
	h.redirectResult(c, "departments", h.service.CreateDepartment(c.Request.Context(), DepartmentInput{Code: c.PostForm("code"), Name: c.PostForm("name")}), "department-created")
}
func (h *Handler) UpdateDepartment(c *gin.Context) {
	if !h.validCSRF(c) {
		return
	}
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "departments", "error", "invalid")
		return
	}
	h.redirectResult(c, "departments", h.service.UpdateDepartment(c.Request.Context(), id, DepartmentInput{Code: c.PostForm("code"), Name: c.PostForm("name")}), "department-updated")
}
func (h *Handler) DeleteDepartment(c *gin.Context) {
	if !h.validCSRF(c) {
		return
	}
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "departments", "error", "invalid")
		return
	}
	h.redirectResult(c, "departments", h.service.DeleteDepartment(c.Request.Context(), id), "department-deleted")
}

func (h *Handler) validCSRF(c *gin.Context) bool {
	if auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		return true
	}
	page := strings.TrimPrefix(c.FullPath(), "/master/")
	page = strings.Split(page, "/")[0]
	h.redirect(c, page, "error", "csrf")
	return false
}
func routeID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return id, err == nil && id > 0
}
func (h *Handler) render(c *gin.Context, status int, templateName, activePage string, data gin.H) {
	current, _ := auth.CurrentUser(c)
	menus, err := h.navigation.Menus(c.Request.Context(), current, activePage)
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
	data["ActivePage"], data["Section"] = activePage, "Master Data"
	data["UserName"], data["UserInitials"] = current.Name, initials(current.Name)
	data["StoreName"], data["RoleName"], data["CSRFToken"] = current.StoreName, current.RoleName, token
	data["Menus"] = menus
	data["SuccessMessage"], data["ErrorMessage"] = message(c.Query("success")), message(c.Query("error"))
	c.HTML(status, templateName, data)
}
func (h *Handler) redirectResult(c *gin.Context, page string, err error, success string) {
	if err == nil {
		h.redirect(c, page, "success", success)
		return
	}
	code := "failed"
	switch {
	case errors.Is(err, ErrInvalid):
		code = "invalid"
	case errors.Is(err, ErrDuplicate):
		code = "duplicate"
	case errors.Is(err, ErrInUse):
		code = "in-use"
	case errors.Is(err, sql.ErrNoRows):
		code = "not-found"
	default:
		slog.Error("organization master operation failed", "page", page, "error", err)
	}
	h.redirect(c, page, "error", code)
}
func (h *Handler) redirect(c *gin.Context, page, key, value string) {
	c.Redirect(http.StatusSeeOther, "/master/"+page+"?"+key+"="+url.QueryEscape(value))
}
func (h *Handler) internalError(c *gin.Context, err error) {
	slog.Error("render organization master", "error", err)
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
	return map[string]string{"job-level-created": "Tingkatan jabatan berhasil dibuat.", "job-level-updated": "Tingkatan jabatan berhasil diperbarui.", "job-level-deleted": "Tingkatan jabatan berhasil dihapus.", "position-created": "Jabatan berhasil dibuat.", "position-updated": "Jabatan berhasil diperbarui.", "position-deleted": "Jabatan berhasil dihapus.", "department-created": "Bagian berhasil dibuat.", "department-updated": "Bagian berhasil diperbarui.", "department-deleted": "Bagian berhasil dihapus.", "csrf": "Sesi form berakhir. Silakan muat ulang halaman.", "invalid": "Data belum lengkap atau tidak valid.", "duplicate": "Kode tersebut sudah digunakan.", "in-use": "Data masih digunakan dan tidak dapat dihapus.", "not-found": "Data tidak ditemukan.", "failed": "Operasi gagal diproses."}[code]
}
