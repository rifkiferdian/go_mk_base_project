package assessmentrule

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
	rules, err := h.service.Rules(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	criteria, err := h.service.Criteria(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	rubrics, err := h.service.Rubrics(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	groups := make([]RuleGroup, 0)
	groupIndex := make(map[string]int)
	for _, rule := range rules {
		key := strconv.FormatUint(rule.RubricID, 10) + ":" + strconv.FormatUint(rule.CriterionID, 10)
		index, found := groupIndex[key]
		if !found {
			var departmentID uint64
			for _, rubric := range rubrics {
				if rubric.ID == rule.RubricID {
					departmentID = rubric.DepartmentID
					break
				}
			}
			groups = append(groups, RuleGroup{RubricID: rule.RubricID, CriterionID: rule.CriterionID, AspectID: rule.AspectID, DepartmentID: departmentID, DepartmentName: rule.DepartmentName, MonthLabel: rule.MonthLabel, CriterionCode: rule.CriterionCode, CriterionDescription: rule.CriterionDescription, AspectName: rule.AspectName})
			index = len(groups) - 1
			groupIndex[key] = index
		}
		groups[index].Rules = append(groups[index].Rules, rule)
	}
	configured := 0
	departmentSeen := map[uint64]bool{}
	departments := make([]AspectOption, 0)
	for _, rubric := range rubrics {
		if !departmentSeen[rubric.DepartmentID] {
			departmentSeen[rubric.DepartmentID] = true
			departments = append(departments, AspectOption{ID: rubric.DepartmentID, Name: rubric.DepartmentName})
		}
	}
	aspectSeen := map[uint64]bool{}
	aspects := make([]AspectOption, 0)
	for _, item := range criteria {
		if item.ExpectedRuleCount > 0 && item.RuleCount == item.ExpectedRuleCount {
			configured++
		}
		if !aspectSeen[item.AspectID] {
			aspectSeen[item.AspectID] = true
			aspects = append(aspects, AspectOption{ID: item.AspectID, Name: item.AspectName})
		}
	}
	h.render(c, http.StatusOK, gin.H{"Title": "Aturan Skor", "ActivePage": "assessment-score-rules", "Rules": rules, "RuleGroups": groups, "Criteria": criteria, "Rubrics": rubrics, "RuleDepartments": departments, "RuleAspects": aspects, "RuleCount": len(rules), "CriterionCount": len(criteria), "ConfiguredCount": configured, "IncompleteCount": len(criteria) - configured, "AspectCount": len(aspects)})
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
	h.redirectResult(c, h.service.Create(c.Request.Context(), input), "rule-created")
}
func (h *Handler) Update(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "csrf")
		return
	}
	key, ok := routeKey(c)
	if !ok {
		h.redirect(c, "error", "invalid")
		return
	}
	input, ok := formInput(c)
	if !ok {
		h.redirect(c, "error", "invalid")
		return
	}
	h.redirectResult(c, h.service.Update(c.Request.Context(), key, input), "rule-updated")
}
func (h *Handler) Delete(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "csrf")
		return
	}
	key, ok := routeKey(c)
	if !ok {
		h.redirect(c, "error", "invalid")
		return
	}
	h.redirectResult(c, h.service.Delete(c.Request.Context(), key), "rule-deleted")
}
func formInput(c *gin.Context) (Input, bool) {
	rubricID, e0 := strconv.ParseUint(c.PostForm("rubric_id"), 10, 64)
	criterionID, e1 := strconv.ParseUint(c.PostForm("criterion_id"), 10, 64)
	score, e2 := strconv.ParseUint(c.PostForm("score"), 10, 8)
	minimum, e3 := strconv.ParseUint(c.PostForm("finding_min"), 10, 16)
	if e0 != nil || e1 != nil || e2 != nil || e3 != nil {
		return Input{}, false
	}
	var maximum *uint16
	if raw := strings.TrimSpace(c.PostForm("finding_max")); raw != "" {
		value, err := strconv.ParseUint(raw, 10, 16)
		if err != nil {
			return Input{}, false
		}
		parsed := uint16(value)
		maximum = &parsed
	}
	return Input{RubricID: rubricID, CriterionID: criterionID, Score: uint8(score), FindingMin: uint16(minimum), FindingMax: maximum, Description: c.PostForm("description")}, true
}
func routeKey(c *gin.Context) (Key, bool) {
	rubricID, e0 := strconv.ParseUint(c.Param("rubricID"), 10, 64)
	criterionID, e1 := strconv.ParseUint(c.Param("criterionID"), 10, 64)
	score, e2 := strconv.ParseUint(c.Param("score"), 10, 8)
	return Key{RubricID: rubricID, CriterionID: criterionID, Score: uint8(score)}, e0 == nil && e1 == nil && e2 == nil && rubricID > 0 && criterionID > 0 && score >= 1 && score <= 5
}
func (h *Handler) render(c *gin.Context, status int, data gin.H) {
	current, _ := auth.CurrentUser(c)
	menus, err := h.navigation.Menus(c.Request.Context(), current, "assessment-score-rules")
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
	c.HTML(status, "assessmentrule/index.html", data)
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
	case errors.Is(err, ErrOverlap):
		code = "overlap"
	case errors.Is(err, sql.ErrNoRows):
		code = "not-found"
	default:
		slog.Error("assessment score rule operation failed", "error", err)
	}
	h.redirect(c, "error", code)
}
func (h *Handler) redirect(c *gin.Context, key, value string) {
	c.Redirect(http.StatusSeeOther, "/master/assessment-score-rules?"+key+"="+url.QueryEscape(value))
}
func (h *Handler) internalError(c *gin.Context, err error) {
	slog.Error("render assessment score rules", "error", err)
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
	return map[string]string{"rule-created": "Aturan skor berhasil dibuat.", "rule-updated": "Aturan skor berhasil diperbarui.", "rule-deleted": "Aturan skor berhasil dihapus.", "csrf": "Sesi form berakhir. Silakan muat ulang halaman.", "invalid": "Data aturan skor belum lengkap atau tidak valid.", "duplicate": "Skor tersebut sudah memiliki aturan pada subaspek yang dipilih.", "overlap": "Rentang jumlah temuan bertumpang tindih dengan aturan lain.", "not-found": "Aturan skor tidak ditemukan.", "failed": "Operasi aturan skor gagal diproses."}[code]
}
