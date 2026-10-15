package dashboard

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"mksiaga/internal/assessmentreport"
	"mksiaga/internal/auth"
	"mksiaga/internal/navigation"
)

type Handler struct {
	appName    string
	reports    *assessmentreport.Service
	navigation *navigation.Service
}

func NewHandler(appName string, reportService *assessmentreport.Service, navigationService *navigation.Service) *Handler {
	return &Handler{appName: appName, reports: reportService, navigation: navigationService}
}

func (h *Handler) Index(c *gin.Context) {
	departmentID, _ := strconv.ParseUint(c.Query("department_id"), 10, 64)
	rubricID, _ := strconv.ParseUint(c.Query("rubric_id"), 10, 64)
	filter := assessmentreport.Filter{DepartmentID: departmentID, RubricID: rubricID}
	report, err := h.reports.ReportWithFilter(c.Request.Context(), filter)
	if err != nil {
		slog.Error("prepare dashboard report", "error", err)
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	departments, rubrics, err := h.reports.FilterOptions(c.Request.Context())
	if err != nil {
		slog.Error("prepare dashboard filters", "error", err)
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	user, _ := auth.CurrentUser(c)
	menus, err := h.navigation.Menus(c.Request.Context(), user, "dashboard")
	if err != nil {
		slog.Error("prepare dashboard navigation", "error", err)
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	csrfToken, err := auth.CSRFToken(c)
	if err != nil {
		slog.Error("prepare dashboard csrf", "error", err)
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	c.HTML(http.StatusOK, "dashboard/index.html", gin.H{
		"Title":         "Dashboard",
		"Section":       "Ruang kerja",
		"AppName":       h.appName,
		"Authenticated": true,
		"IsDashboard":   true,
		"UserName":      user.Name,
		"UserInitials":  initials(user.Name),
		"StoreName":     user.StoreName,
		"RoleName":      user.RoleName,
		"DateLabel":     indonesianDate(time.Now()),
		"Report":        report,
		"Filter":        filter,
		"Departments":   departments,
		"Rubrics":       rubrics,
		"Menus":         menus,
		"CSRFToken":     csrfToken,
	})
}

func initials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "US"
	}
	result := []rune(strings.ToUpper(parts[0]))[:1]
	if len(parts) > 1 {
		result = append(result, []rune(strings.ToUpper(parts[len(parts)-1]))[0])
	}
	return string(result)
}

func indonesianDate(value time.Time) string {
	days := [...]string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	months := [...]string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	return strings.ToUpper(days[value.Weekday()] + ", " + value.Format("02") + " " + months[value.Month()-1] + " " + value.Format("2006"))
}
