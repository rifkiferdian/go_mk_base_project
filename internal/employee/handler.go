package employee

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

func NewHandler(a string, s *Service, n *navigation.Service) *Handler { return &Handler{a, s, n} }
func (h *Handler) Employees(c *gin.Context) {
	items, e := h.service.Employees(c)
	if e != nil {
		h.fail(c, e)
		return
	}
	active, inactive, assigned := 0, 0, 0
	for _, x := range items {
		if x.Status == "active" {
			active++
		} else {
			inactive++
		}
		if x.HasCurrentAssignment {
			assigned++
		}
	}
	h.render(c, "employee/employees.html", "employees", gin.H{"Title": "Karyawan", "Employees": items, "EmployeeCount": len(items), "ActiveCount": active, "InactiveCount": inactive, "AssignedCount": assigned})
}
func (h *Handler) CreateEmployee(c *gin.Context) {
	if !h.csrf(c, "employees") {
		return
	}
	h.result(c, "employees", h.service.CreateEmployee(c, employeeInput(c)), "employee-created")
}
func (h *Handler) UpdateEmployee(c *gin.Context) {
	if !h.csrf(c, "employees") {
		return
	}
	id, ok := id(c)
	if !ok {
		h.redirect(c, "employees", "error", "invalid")
		return
	}
	h.result(c, "employees", h.service.UpdateEmployee(c, id, employeeInput(c)), "employee-updated")
}
func (h *Handler) DeleteEmployee(c *gin.Context) {
	if !h.csrf(c, "employees") {
		return
	}
	id, ok := id(c)
	if !ok {
		h.redirect(c, "employees", "error", "invalid")
		return
	}
	h.result(c, "employees", h.service.DeleteEmployee(c, id), "employee-deleted")
}
func employeeInput(c *gin.Context) EmployeeInput {
	return EmployeeInput{c.PostForm("employee_number"), c.PostForm("name"), c.PostForm("email"), c.PostForm("phone"), c.PostForm("status")}
}

func (h *Handler) Profiles(c *gin.Context) {
	items, e := h.service.Profiles(c)
	if e != nil {
		h.fail(c, e)
		return
	}
	available, e := h.service.EmployeesWithoutProfile(c)
	if e != nil {
		h.fail(c, e)
		return
	}
	h.render(c, "employee/profiles.html", "employee-profiles", gin.H{"Title": "Biodata Karyawan", "Profiles": items, "AvailableEmployees": available, "ProfileCount": len(items), "IncompleteCount": len(available)})
}
func (h *Handler) CreateProfile(c *gin.Context) {
	if !h.csrf(c, "employee-profiles") {
		return
	}
	x, ok := profileInput(c)
	if !ok {
		h.redirect(c, "employee-profiles", "error", "invalid")
		return
	}
	h.result(c, "employee-profiles", h.service.CreateProfile(c, x), "profile-created")
}
func (h *Handler) UpdateProfile(c *gin.Context) {
	if !h.csrf(c, "employee-profiles") {
		return
	}
	id, ok := id(c)
	x, xok := profileInput(c)
	if !ok || !xok {
		h.redirect(c, "employee-profiles", "error", "invalid")
		return
	}
	h.result(c, "employee-profiles", h.service.UpdateProfile(c, id, x), "profile-updated")
}
func (h *Handler) DeleteProfile(c *gin.Context) {
	if !h.csrf(c, "employee-profiles") {
		return
	}
	id, ok := id(c)
	if !ok {
		h.redirect(c, "employee-profiles", "error", "invalid")
		return
	}
	h.result(c, "employee-profiles", h.service.DeleteProfile(c, id), "profile-deleted")
}
func profileInput(c *gin.Context) (ProfileInput, bool) {
	v, e := strconv.ParseUint(c.PostForm("employee_id"), 10, 64)
	if e != nil && c.PostForm("employee_id") != "" {
		return ProfileInput{}, false
	}
	return ProfileInput{EmployeeID: v, BirthPlace: c.PostForm("birth_place"), BirthDate: c.PostForm("birth_date"), Gender: c.PostForm("gender"), Address: c.PostForm("address"), JoinedDate: c.PostForm("joined_date"), EmploymentStatus: c.PostForm("employment_status"), PhotoPath: c.PostForm("photo_path"), EmergencyContactName: c.PostForm("emergency_contact_name"), EmergencyContactPhone: c.PostForm("emergency_contact_phone"), EmergencyContactRelationship: c.PostForm("emergency_contact_relationship")}, true
}

func (h *Handler) Assignments(c *gin.Context) {
	items, e := h.service.Assignments(c)
	if e != nil {
		h.fail(c, e)
		return
	}
	employees, _ := h.service.EmployeeOptions(c)
	positions, _ := h.service.PositionOptions(c)
	departments, _ := h.service.DepartmentOptions(c)
	stores, _ := h.service.StoreOptions(c)
	current := 0
	for _, x := range items {
		if x.Current {
			current++
		}
	}
	h.render(c, "employee/assignments.html", "employee-assignments", gin.H{"Title": "Penempatan Karyawan", "Assignments": items, "Employees": employees, "Positions": positions, "Departments": departments, "Stores": stores, "AssignmentCount": len(items), "CurrentCount": current})
}
func (h *Handler) CreateAssignment(c *gin.Context) {
	if !h.csrf(c, "employee-assignments") {
		return
	}
	x, ok := assignmentInput(c)
	if !ok {
		h.redirect(c, "employee-assignments", "error", "invalid")
		return
	}
	h.result(c, "employee-assignments", h.service.CreateAssignment(c, x), "assignment-created")
}
func (h *Handler) UpdateAssignment(c *gin.Context) {
	if !h.csrf(c, "employee-assignments") {
		return
	}
	id, ok := id(c)
	x, xok := assignmentInput(c)
	if !ok || !xok {
		h.redirect(c, "employee-assignments", "error", "invalid")
		return
	}
	h.result(c, "employee-assignments", h.service.UpdateAssignment(c, id, x), "assignment-updated")
}
func (h *Handler) DeleteAssignment(c *gin.Context) {
	if !h.csrf(c, "employee-assignments") {
		return
	}
	id, ok := id(c)
	if !ok {
		h.redirect(c, "employee-assignments", "error", "invalid")
		return
	}
	h.result(c, "employee-assignments", h.service.DeleteAssignment(c, id), "assignment-deleted")
}
func assignmentInput(c *gin.Context) (AssignmentInput, bool) {
	vals := make([]uint64, 4)
	for i, k := range []string{"employee_id", "position_id", "department_id", "store_id"} {
		v, e := strconv.ParseUint(c.PostForm(k), 10, 64)
		if e != nil {
			return AssignmentInput{}, false
		}
		vals[i] = v
	}
	return AssignmentInput{vals[0], vals[1], vals[2], vals[3], c.PostForm("started_at"), c.PostForm("ended_at")}, true
}

func (h *Handler) ReportingLines(c *gin.Context) {
	items, e := h.service.ReportingLines(c)
	if e != nil {
		h.fail(c, e)
		return
	}
	options, e := h.service.AssignmentOptions(c)
	if e != nil {
		h.fail(c, e)
		return
	}
	current := 0
	for _, x := range items {
		if x.Current {
			current++
		}
	}
	h.render(c, "employee/reporting_lines.html", "employee-reporting-lines", gin.H{"Title": "Atasan dan Bawahan", "ReportingLines": items, "Assignments": options, "ReportingCount": len(items), "CurrentCount": current})
}
func (h *Handler) CreateReportingLine(c *gin.Context) {
	if !h.csrf(c, "employee-reporting-lines") {
		return
	}
	x, ok := reportingInput(c)
	if !ok {
		h.redirect(c, "employee-reporting-lines", "error", "invalid")
		return
	}
	h.result(c, "employee-reporting-lines", h.service.CreateReportingLine(c, x), "reporting-created")
}
func (h *Handler) UpdateReportingLine(c *gin.Context) {
	if !h.csrf(c, "employee-reporting-lines") {
		return
	}
	id, ok := id(c)
	x, xok := reportingInput(c)
	if !ok || !xok {
		h.redirect(c, "employee-reporting-lines", "error", "invalid")
		return
	}
	h.result(c, "employee-reporting-lines", h.service.UpdateReportingLine(c, id, x), "reporting-updated")
}
func (h *Handler) DeleteReportingLine(c *gin.Context) {
	if !h.csrf(c, "employee-reporting-lines") {
		return
	}
	id, ok := id(c)
	if !ok {
		h.redirect(c, "employee-reporting-lines", "error", "invalid")
		return
	}
	h.result(c, "employee-reporting-lines", h.service.DeleteReportingLine(c, id), "reporting-deleted")
}
func reportingInput(c *gin.Context) (ReportingLineInput, bool) {
	a, e := strconv.ParseUint(c.PostForm("employee_assignment_id"), 10, 64)
	b, e2 := strconv.ParseUint(c.PostForm("supervisor_assignment_id"), 10, 64)
	return ReportingLineInput{a, b, c.PostForm("started_at"), c.PostForm("ended_at")}, e == nil && e2 == nil
}

func id(c *gin.Context) (uint64, bool) {
	v, e := strconv.ParseUint(c.Param("id"), 10, 64)
	return v, e == nil && v > 0
}
func (h *Handler) csrf(c *gin.Context, page string) bool {
	if auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		return true
	}
	h.redirect(c, page, "error", "csrf")
	return false
}
func (h *Handler) render(c *gin.Context, t, page string, d gin.H) {
	u, _ := auth.CurrentUser(c)
	menus, e := h.navigation.Menus(c, u, page)
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
	d["ActivePage"], d["Section"] = page, "Data Karyawan"
	d["UserName"], d["UserInitials"], d["StoreName"], d["RoleName"] = u.Name, initials(u.Name), u.StoreName, u.RoleName
	d["CSRFToken"], d["Menus"] = token, menus
	d["SuccessMessage"], d["ErrorMessage"] = message(c.Query("success")), message(c.Query("error"))
	c.HTML(http.StatusOK, t, d)
}
func (h *Handler) result(c *gin.Context, page string, e error, success string) {
	if e == nil {
		h.redirect(c, page, "success", success)
		return
	}
	code := "failed"
	switch {
	case errors.Is(e, ErrInvalid):
		code = "invalid"
	case errors.Is(e, ErrDuplicate):
		code = "duplicate"
	case errors.Is(e, ErrInUse):
		code = "in-use"
	case errors.Is(e, sql.ErrNoRows):
		code = "not-found"
	default:
		slog.Error("employee operation failed", "page", page, "error", e)
	}
	h.redirect(c, page, "error", code)
}
func (h *Handler) redirect(c *gin.Context, page, key, value string) {
	c.Redirect(http.StatusSeeOther, "/master/"+page+"?"+key+"="+url.QueryEscape(value))
}
func (h *Handler) fail(c *gin.Context, e error) {
	slog.Error("employee module", "error", e)
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
func message(c string) string {
	return map[string]string{"employee-created": "Karyawan berhasil ditambahkan.", "employee-updated": "Data karyawan berhasil diperbarui.", "employee-deleted": "Karyawan berhasil dihapus.", "profile-created": "Biodata berhasil ditambahkan.", "profile-updated": "Biodata berhasil diperbarui.", "profile-deleted": "Biodata berhasil dihapus.", "assignment-created": "Penempatan berhasil ditambahkan.", "assignment-updated": "Penempatan berhasil diperbarui.", "assignment-deleted": "Penempatan berhasil dihapus.", "reporting-created": "Hubungan atasan berhasil ditambahkan.", "reporting-updated": "Hubungan atasan berhasil diperbarui.", "reporting-deleted": "Hubungan atasan berhasil dihapus.", "invalid": "Data belum lengkap atau periode tidak valid.", "duplicate": "Data aktif atau nomor tersebut sudah digunakan.", "in-use": "Data masih digunakan oleh data lain.", "not-found": "Data tidak ditemukan.", "csrf": "Sesi form berakhir. Muat ulang halaman.", "failed": "Operasi gagal diproses."}[c]
}
