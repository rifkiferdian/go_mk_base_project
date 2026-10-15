package assessmentcriterion

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
	criteria, err := h.service.Criteria(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	aspects, err := h.service.Aspects(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	configured, used := 0, 0
	rubrics, err := h.service.Rubrics(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	departments, err := h.service.Departments(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	totalWeight := 0.0
	for _, item := range criteria {
		totalWeight += item.Weight
		if item.ExpectedRuleCount > 0 && item.RuleCount == item.ExpectedRuleCount {
			configured++
		}
		if item.AssessmentCount > 0 {
			used++
		}
	}
	h.render(c, http.StatusOK, gin.H{"Title": "Subaspek Penilaian", "ActivePage": "assessment-criteria", "Criteria": criteria, "Aspects": aspects, "Departments": departments, "Rubrics": rubrics, "CriterionCount": len(criteria), "ConfiguredCount": configured, "UsedCount": used, "RubricCount": len(rubrics), "TotalWeight": totalWeight})
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
	h.redirectResult(c, h.service.Create(c.Request.Context(), input), "criterion-created")
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
	h.redirectResult(c, h.service.Update(c.Request.Context(), id, input), "criterion-updated")
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
	h.redirectResult(c, h.service.Delete(c.Request.Context(), id), "criterion-deleted")
}
func formInput(c *gin.Context) (Input, bool) {
	departmentID, departmentErr := strconv.ParseUint(c.PostForm("department_id"), 10, 64)
	if departmentErr != nil || departmentID == 0 {
		return Input{}, false
	}
	aspectID, e1 := strconv.ParseUint(c.PostForm("aspect_id"), 10, 64)
	weight, e2 := strconv.ParseFloat(c.PostForm("weight"), 64)
	order, e3 := strconv.ParseUint(c.PostForm("sort_order"), 10, 16)
	if e1 != nil || e2 != nil || e3 != nil || aspectID == 0 || order == 0 {
		return Input{}, false
	}
	return Input{DepartmentID: departmentID, AspectID: aspectID, Code: c.PostForm("code"), Description: c.PostForm("description"), Weight: weight, SortOrder: uint16(order)}, true
}
func routeID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return id, err == nil && id > 0
}
func (h *Handler) render(c *gin.Context, status int, data gin.H) {
	current, _ := auth.CurrentUser(c)
	menus, err := h.navigation.Menus(c.Request.Context(), current, "assessment-criteria")
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
	data["Section"] = "Master Data"
	data["UserName"], data["UserInitials"] = current.Name, initials(current.Name)
	data["StoreName"], data["RoleName"], data["CSRFToken"] = current.StoreName, current.RoleName, token
	data["Menus"] = menus
	data["SuccessMessage"], data["ErrorMessage"] = message(c.Query("success")), message(c.Query("error"))
	c.HTML(status, "assessmentcriterion/index.html", data)
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
	case errors.Is(err, ErrInUse):
		code = "in-use"
	case errors.Is(err, sql.ErrNoRows):
		code = "not-found"
	default:
		slog.Error("assessment criterion operation failed", "error", err)
	}
	h.redirect(c, "error", code)
}
func (h *Handler) redirect(c *gin.Context, key, value string) {
	c.Redirect(http.StatusSeeOther, "/master/assessment-criteria?"+key+"="+url.QueryEscape(value))
}
func (h *Handler) internalError(c *gin.Context, err error) {
	slog.Error("render assessment criteria", "error", err)
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
	return map[string]string{"criterion-created": "Subaspek berhasil dibuat.", "criterion-updated": "Subaspek berhasil diperbarui.", "criterion-deleted": "Subaspek dan aturan skornya berhasil dihapus.", "csrf": "Sesi form berakhir. Silakan muat ulang halaman.", "invalid": "Data subaspek belum lengkap atau tidak valid.", "duplicate": "Kode subaspek sudah digunakan.", "in-use": "Subaspek sudah digunakan dalam penilaian dan tidak dapat dihapus.", "not-found": "Subaspek tidak ditemukan.", "failed": "Operasi subaspek gagal diproses."}[code]
}
