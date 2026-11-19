package assessmentrubric

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
	"mksiaga/internal/auth"
	"mksiaga/internal/navigation"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Handler struct {
	app string
	s   *Service
	n   *navigation.Service
}

func NewHandler(a string, s *Service, n *navigation.Service) *Handler { return &Handler{a, s, n} }
func (h *Handler) Index(c *gin.Context) {
	items, positions, e := h.s.Page(c)
	if e != nil {
		h.fail(c, e)
		return
	}
	active := 0
	for _, x := range items {
		if x.IsActive {
			active++
		}
	}
	h.render(c, gin.H{"Title": "Periode Rubrik", "ActivePage": "assessment-rubrics", "Rubrics": items, "Positions": positions, "RubricCount": len(items), "ActiveCount": active})
}
func (h *Handler) Create(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "csrf")
		return
	}
	x, ok := input(c)
	if !ok {
		h.redirect(c, "error", "invalid")
		return
	}
	h.result(c, h.s.Create(c, x), "created")
}
func (h *Handler) Update(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "csrf")
		return
	}
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	x, ok := input(c)
	if e != nil || !ok {
		h.redirect(c, "error", "invalid")
		return
	}
	h.result(c, h.s.Update(c, id, x), "updated")
}
func (h *Handler) Delete(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "csrf")
		return
	}
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	if e != nil {
		h.redirect(c, "error", "invalid")
		return
	}
	h.result(c, h.s.Delete(c, id), "deleted")
}
func input(c *gin.Context) (Input, bool) {
	p, e1 := strconv.ParseUint(c.PostForm("position_id"), 10, 64)
	from, e2 := strconv.ParseUint(c.PostForm("month_from"), 10, 16)
	to, e3 := strconv.ParseUint(c.PostForm("month_to"), 10, 16)
	if e1 != nil || e2 != nil || e3 != nil {
		return Input{}, false
	}
	return Input{p, uint16(from), uint16(to), c.PostForm("is_active") == "1"}, true
}
func (h *Handler) result(c *gin.Context, e error, success string) {
	if e == nil {
		h.redirect(c, "success", success)
		return
	}
	code := "failed"
	if errors.Is(e, ErrInvalid) {
		code = "invalid"
	} else if errors.Is(e, ErrOverlap) {
		code = "overlap"
	} else if errors.Is(e, ErrInUse) {
		code = "in-use"
	} else if errors.Is(e, sql.ErrNoRows) {
		code = "not-found"
	}
	h.redirect(c, "error", code)
}
func (h *Handler) render(c *gin.Context, d gin.H) {
	u, _ := auth.CurrentUser(c)
	menus, e := h.n.Menus(c, u, "assessment-rubrics")
	if e != nil {
		h.fail(c, e)
		return
	}
	token, e := auth.CSRFToken(c)
	if e != nil {
		h.fail(c, e)
		return
	}
	d["AppName"], d["IsDashboard"], d["Authenticated"] = h.app, true, true
	d["Section"] = "Pengaturan Penilaian"
	d["UserName"], d["UserInitials"], d["StoreName"], d["RoleName"] = u.Name, initials(u.Name), u.StoreName, u.RoleName
	d["CSRFToken"], d["Menus"] = token, menus
	d["SuccessMessage"], d["ErrorMessage"] = msg(c.Query("success")), msg(c.Query("error"))
	c.HTML(http.StatusOK, "assessmentrubric/index.html", d)
}
func (h *Handler) redirect(c *gin.Context, k, v string) {
	c.Redirect(http.StatusSeeOther, "/master/assessment-rubrics?"+k+"="+url.QueryEscape(v))
}
func (h *Handler) fail(c *gin.Context, e error) {
	slog.Error("assessment rubric", "error", e)
	c.Redirect(http.StatusSeeOther, "/errors/500")
}
func initials(n string) string {
	p := strings.Fields(n)
	if len(p) == 0 {
		return "US"
	}
	v := string([]rune(strings.ToUpper(p[0]))[0])
	if len(p) > 1 {
		v += string([]rune(strings.ToUpper(p[len(p)-1]))[0])
	}
	return v
}
func msg(c string) string {
	return map[string]string{"created": "Periode rubrik berhasil dibuat.", "updated": "Periode rubrik berhasil diperbarui.", "deleted": "Periode rubrik berhasil dihapus.", "overlap": "Rentang bulan bertumpang tindih dengan periode lain pada jabatan yang sama.", "in-use": "Periode sudah memiliki aturan skor atau dipakai penilaian.", "invalid": "Jabatan atau rentang bulan tidak valid.", "not-found": "Periode rubrik tidak ditemukan.", "csrf": "Sesi form berakhir. Muat ulang halaman.", "failed": "Operasi periode rubrik gagal."}[c]
}
