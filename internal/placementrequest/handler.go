package placementrequest

import (
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
	u, _ := auth.CurrentUser(c)
	uid, _ := strconv.ParseUint(u.ID, 10, 64)
	items, employees, positions, stores, e := h.s.Page(c, uid, u.RoleName)
	if e != nil {
		h.fail(c, e)
		return
	}
	pending := 0
	for _, x := range items {
		if x.Status == "pending" {
			pending++
		}
	}
	h.render(c, gin.H{"Title": "Pengajuan Penempatan", "ActivePage": "placement-change-requests", "Requests": items, "Employees": employees, "Positions": positions, "Stores": stores, "PendingCount": pending, "RequestCount": len(items), "CanCreate": strings.Contains(u.RoleName, "superadmin") || strings.Contains(u.RoleName, "supervisor"), "CanApprove": strings.Contains(u.RoleName, "superadmin") || strings.Contains(u.RoleName, "manager")})
}
func (h *Handler) Create(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "csrf")
		return
	}
	u, _ := auth.CurrentUser(c)
	uid, _ := strconv.ParseUint(u.ID, 10, 64)
	eid, e1 := strconv.ParseUint(c.PostForm("employee_id"), 10, 64)
	pid, e2 := strconv.ParseUint(c.PostForm("position_id"), 10, 64)
	sid, e3 := strconv.ParseUint(c.PostForm("store_id"), 10, 64)
	if e1 != nil || e2 != nil || e3 != nil {
		h.redirect(c, "error", "invalid")
		return
	}
	e := h.s.Create(c, uid, u.RoleName, Input{eid, pid, sid, c.PostForm("effective_date"), c.PostForm("reason")})
	h.result(c, e, "created")
}
func (h *Handler) Approve(c *gin.Context) { h.decide(c, true) }
func (h *Handler) Reject(c *gin.Context)  { h.decide(c, false) }
func (h *Handler) decide(c *gin.Context, approve bool) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "csrf")
		return
	}
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	if e != nil {
		h.redirect(c, "error", "invalid")
		return
	}
	u, _ := auth.CurrentUser(c)
	uid, _ := strconv.ParseUint(u.ID, 10, 64)
	h.result(c, h.s.Decide(c, id, uid, approve, c.PostForm("decision_notes")), map[bool]string{true: "approved", false: "rejected"}[approve])
}
func (h *Handler) render(c *gin.Context, d gin.H) {
	u, _ := auth.CurrentUser(c)
	menus, e := h.n.Menus(c, u, "placement-change-requests")
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
	d["Section"] = "Kepegawaian"
	d["UserName"], d["UserInitials"], d["StoreName"], d["RoleName"] = u.Name, initials(u.Name), u.StoreName, u.RoleName
	d["CSRFToken"], d["Menus"] = token, menus
	d["SuccessMessage"], d["ErrorMessage"] = msg(c.Query("success")), msg(c.Query("error"))
	c.HTML(http.StatusOK, "placementrequest/index.html", d)
}
func (h *Handler) result(c *gin.Context, e error, success string) {
	if e == nil {
		h.redirect(c, "success", success)
		return
	}
	code := "failed"
	if errors.Is(e, ErrInvalid) {
		code = "invalid"
	} else if errors.Is(e, ErrFuture) {
		code = "future"
	}
	h.redirect(c, "error", code)
}
func (h *Handler) redirect(c *gin.Context, k, v string) {
	c.Redirect(http.StatusSeeOther, "/placement-change-requests?"+k+"="+url.QueryEscape(v))
}
func (h *Handler) fail(c *gin.Context, e error) {
	slog.Error("placement request", "error", e)
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
	return map[string]string{"created": "Pengajuan perubahan penempatan berhasil dikirim.", "approved": "Pengajuan disetujui dan penempatan berhasil diperbarui.", "rejected": "Pengajuan berhasil ditolak.", "future": "Pengajuan belum dapat disetujui sebelum tanggal efektif.", "invalid": "Data pengajuan tidak valid.", "csrf": "Sesi form berakhir. Muat ulang halaman.", "failed": "Pengajuan gagal diproses."}[c]
}
