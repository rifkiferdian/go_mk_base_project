package assessmentcategory

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
	items, err := h.service.Categories(c.Request.Context())
	if err != nil {
		h.fail(c, err)
		return
	}
	active := 0
	for _, x := range items {
		if x.IsActive {
			active++
		}
	}
	h.render(c, gin.H{"Title": "Kategori Hasil", "ActivePage": "assessment-result-categories", "Categories": items, "CategoryCount": len(items), "ActiveCount": active})
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
	h.result(c, h.service.Create(c.Request.Context(), input), "created")
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
	h.result(c, h.service.Update(c.Request.Context(), id, input), "updated")
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
	h.result(c, h.service.Delete(c.Request.Context(), id), "deleted")
}
func formInput(c *gin.Context) (Input, bool) {
	min, e1 := strconv.ParseFloat(c.PostForm("minimum_score"), 64)
	max, e2 := strconv.ParseFloat(c.PostForm("maximum_score"), 64)
	cut, e3 := strconv.ParseFloat(c.PostForm("deduction_percentage"), 64)
	order, e4 := strconv.ParseUint(c.PostForm("sort_order"), 10, 16)
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
		return Input{}, false
	}
	return Input{Name: c.PostForm("name"), Minimum: min, Maximum: max, Deduction: cut, SortOrder: uint16(order), IsActive: c.PostForm("is_active") == "1"}, true
}
func routeID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return id, err == nil && id > 0
}
func (h *Handler) render(c *gin.Context, data gin.H) {
	current, _ := auth.CurrentUser(c)
	menus, err := h.navigation.Menus(c.Request.Context(), current, "assessment-result-categories")
	if err != nil {
		h.fail(c, err)
		return
	}
	token, err := auth.CSRFToken(c)
	if err != nil {
		h.fail(c, err)
		return
	}
	data["AppName"], data["IsDashboard"], data["Authenticated"] = h.appName, true, true
	data["Section"] = "Pengaturan Penilaian"
	data["UserName"], data["UserInitials"] = current.Name, initials(current.Name)
	data["StoreName"], data["RoleName"], data["CSRFToken"] = current.StoreName, current.RoleName, token
	data["Menus"] = menus
	data["SuccessMessage"], data["ErrorMessage"] = message(c.Query("success")), message(c.Query("error"))
	c.HTML(http.StatusOK, "assessmentcategory/index.html", data)
}
func (h *Handler) result(c *gin.Context, err error, success string) {
	if err == nil {
		h.redirect(c, "success", success)
		return
	}
	code := "failed"
	if errors.Is(err, ErrInvalid) {
		code = "invalid"
	} else if errors.Is(err, ErrOverlap) {
		code = "overlap"
	} else if errors.Is(err, sql.ErrNoRows) {
		code = "not-found"
	} else {
		slog.Error("assessment category operation failed", "error", err)
	}
	h.redirect(c, "error", code)
}
func (h *Handler) redirect(c *gin.Context, key, value string) {
	c.Redirect(http.StatusSeeOther, "/master/assessment-result-categories?"+key+"="+url.QueryEscape(value))
}
func (h *Handler) fail(c *gin.Context, err error) {
	slog.Error("assessment category failed", "error", err)
	c.Redirect(http.StatusSeeOther, "/errors/500")
}
func initials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "US"
	}
	v := string([]rune(strings.ToUpper(parts[0]))[0])
	if len(parts) > 1 {
		v += string([]rune(strings.ToUpper(parts[len(parts)-1]))[0])
	}
	return v
}
func message(code string) string {
	return map[string]string{"created": "Kategori hasil berhasil dibuat.", "updated": "Kategori hasil berhasil diperbarui.", "deleted": "Kategori hasil berhasil dihapus.", "csrf": "Sesi form berakhir. Silakan muat ulang halaman.", "invalid": "Data kategori tidak valid. Nilai harus berada pada skala 1–5 dan potongan 0–100%.", "overlap": "Rentang nilai bertumpang tindih dengan kategori lain.", "not-found": "Kategori hasil tidak ditemukan.", "failed": "Operasi kategori hasil gagal diproses."}[code]
}
