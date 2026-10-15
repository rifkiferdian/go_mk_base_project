package traineeassessment

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
	appName    string
	service    *Service
	navigation *navigation.Service
}

func NewHandler(a string, s *Service, n *navigation.Service) *Handler { return &Handler{a, s, n} }
func (h *Handler) Index(c *gin.Context) {
	u, _ := auth.CurrentUser(c)
	uid, _ := strconv.ParseUint(u.ID, 10, 64)
	items, e := h.service.List(c, uid, u.RoleName)
	if e != nil {
		h.fail(c, e)
		return
	}
	draft, final := 0, 0
	for _, x := range items {
		if x.Status == "final" {
			final++
		} else {
			draft++
		}
	}
	h.render(c, "trainee_assessment/index.html", gin.H{"Title": "Penilaian Trainee", "Assessments": items, "AssessmentCount": len(items), "DraftCount": draft, "FinalCount": final})
}
func (h *Handler) New(c *gin.Context) { h.form(c, 0) }
func (h *Handler) Edit(c *gin.Context) {
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "error", "invalid")
		return
	}
	h.form(c, id)
}
func (h *Handler) form(c *gin.Context, id uint64) {
	u, _ := auth.CurrentUser(c)
	uid, _ := strconv.ParseUint(u.ID, 10, 64)
	f, e := h.service.Form(c, id, uid, u.RoleName)
	if e != nil {
		if errors.Is(e, ErrFinal) {
			h.redirect(c, "error", "final")
			return
		}
		h.fail(c, e)
		return
	}
	h.render(c, "trainee_assessment/form.html", gin.H{"Title": map[bool]string{true: "Edit Penilaian", false: "Input Penilaian"}[f.Editing], "Form": f})
}
func (h *Handler) Create(c *gin.Context) {
	u, _ := auth.CurrentUser(c)
	uid, _ := strconv.ParseUint(u.ID, 10, 64)
	x, ok := input(c)
	if !ok {
		h.redirect(c, "error", "invalid")
		return
	}
	_, e := h.service.Create(c, uid, u.Name, u.RoleName, x)
	h.result(c, e, "created")
}
func (h *Handler) Update(c *gin.Context) {
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "error", "invalid")
		return
	}
	u, _ := auth.CurrentUser(c)
	uid, _ := strconv.ParseUint(u.ID, 10, 64)
	x, ok := input(c)
	if !ok {
		h.redirect(c, "error", "invalid")
		return
	}
	_, e := h.service.Update(c, id, uid, u.Name, u.RoleName, x)
	h.result(c, e, "updated")
}
func (h *Handler) Delete(c *gin.Context) {
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "error", "invalid")
		return
	}
	u, _ := auth.CurrentUser(c)
	uid, _ := strconv.ParseUint(u.ID, 10, 64)
	h.result(c, h.service.Delete(c, id, uid, u.RoleName), "deleted")
}
func input(c *gin.Context) (Input, bool) {
	eid, e := strconv.ParseUint(c.PostForm("trainee_employee_id"), 10, 64)
	month, e2 := strconv.ParseUint(c.PostForm("training_month"), 10, 64)
	if e != nil || e2 != nil {
		return Input{}, false
	}
	x := Input{TraineeEmployeeID: eid, TrainingMonth: month, AssessmentDate: c.PostForm("assessment_date"), Notes: c.PostForm("notes"), Status: "final"}
	for k, values := range c.Request.PostForm {
		if !strings.HasPrefix(k, "score_") {
			continue
		}
		keyParts := strings.Split(strings.TrimPrefix(k, "score_"), "_")
		cid, e := strconv.ParseUint(keyParts[0], 10, 64)
		if e != nil {
			continue
		}
		score := 0
		if len(values) > 0 && values[len(values)-1] != "" {
			score, _ = strconv.Atoi(values[len(values)-1])
		}
		x.Details = append(x.Details, DetailInput{CriterionID: cid, Score: score})
	}
	return x, true
}
func routeID(c *gin.Context) (uint64, bool) {
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	return id, e == nil && id > 0
}
func (h *Handler) render(c *gin.Context, t string, d gin.H) {
	u, _ := auth.CurrentUser(c)
	menus, e := h.navigation.Menus(c, u, "trainee-assessments")
	if e != nil {
		h.fail(c, e)
		return
	}
	token, e := auth.CSRFToken(c)
	if e != nil {
		h.fail(c, e)
		return
	}
	d["AppName"], d["IsDashboard"], d["Authenticated"] = h.appName, true, true
	d["ActivePage"], d["Section"] = "trainee-assessments", "Penilaian"
	d["UserName"], d["UserInitials"], d["StoreName"], d["RoleName"] = u.Name, initials(u.Name), u.StoreName, u.RoleName
	d["CSRFToken"], d["Menus"] = token, menus
	d["SuccessMessage"], d["ErrorMessage"] = msg(c.Query("success")), msg(c.Query("error"))
	c.HTML(http.StatusOK, t, d)
}
func (h *Handler) result(c *gin.Context, e error, success string) {
	if e == nil {
		h.redirect(c, "success", success)
		return
	}
	code := "failed"
	switch {
	case errors.Is(e, ErrInvalid):
		code = "invalid"
	case errors.Is(e, ErrDuplicate):
		code = "duplicate"
	case errors.Is(e, ErrFinal):
		code = "final"
	case errors.Is(e, ErrNoRubric):
		code = "no-rubric"
	case errors.Is(e, sql.ErrNoRows):
		code = "not-found"
	default:
		slog.Error("assessment operation", "error", e)
	}
	h.redirect(c, "error", code)
}
func (h *Handler) redirect(c *gin.Context, k, v string) {
	c.Redirect(http.StatusSeeOther, "/assessments?"+k+"="+url.QueryEscape(v))
}
func (h *Handler) fail(c *gin.Context, e error) {
	slog.Error("assessment module", "error", e)
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
	if c == "no-rubric" {
		return "Rubrik bagian dan bulan ini belum tersedia, belum lengkap, atau periodenya bertumpang tindih. Hubungi pengelola master penilaian."
	}
	return map[string]string{"created": "Penilaian berhasil disimpan.", "updated": "Penilaian berhasil diperbarui.", "deleted": "Draft penilaian berhasil dihapus.", "invalid": "Lengkapi data, semua nilai, dan catatan sebelum diajukan.", "duplicate": "Trainee tersebut sudah dinilai pada bulan pelatihan yang sama.", "final": "Penilaian yang sudah diajukan tidak dapat diubah.", "not-found": "Penilaian tidak ditemukan atau tidak dapat diakses.", "failed": "Operasi penilaian gagal diproses."}[c]
}
