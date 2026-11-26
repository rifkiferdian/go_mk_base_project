package dashboard

import (
	"database/sql"
	"log/slog"
	"mksiaga/internal/traineeassessment"
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
	db         *sql.DB
	appName    string
	reports    *assessmentreport.Service
	navigation *navigation.Service
}

func NewHandler(appName string, reportService *assessmentreport.Service, navigationService *navigation.Service, db *sql.DB) *Handler {
	return &Handler{db: db, appName: appName, reports: reportService, navigation: navigationService}
}

func (h *Handler) Index(c *gin.Context) {
	user, _ := auth.CurrentUser(c)
	admin := false
	for _, role := range strings.Split(user.RoleName, ",") {
		if strings.TrimSpace(role) == "superadmin" {
			admin = true
		}
	}
	userID, err := strconv.ParseUint(user.ID, 10, 64)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	overview, err := loadOverview(c, h.db, userID, admin)
	if err != nil {
		slog.Error("dashboard overview", "error", err)
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	items, err := traineeassessment.NewService(traineeassessment.NewRepository(h.db)).List(c, userID, user.RoleName, traineeassessment.ListFilter{})
	if err != nil {
		slog.Error("dashboard assessments", "error", err)
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	var report assessmentreport.Report
	if admin {
		report, err = h.reports.Report(c.Request.Context())
	} else {
		seen := map[uint64]bool{}
		for _, x := range items {
			if x.Status != "final" {
				continue
			}
			report.Summary.AssessmentCount++
			seen[x.TraineeEmployeeID] = true
			report.Summary.AverageScore += x.AverageScore
			if x.AverageScore > report.Summary.HighestScore {
				report.Summary.HighestScore = x.AverageScore
			}
			if len(report.Recent) < 5 {
				report.Recent = append(report.Recent, assessmentreport.RecentResult{EmployeeName: x.EmployeeName, EmployeeNumber: x.EmployeeNumber, PositionName: x.PositionName, DepartmentName: x.DepartmentName, AssessmentDate: x.AssessmentDate, TrainingMonth: int(x.TrainingMonth), AverageScore: x.AverageScore, Rating: x.Rating})
			}
		}
		report.Summary.TraineeCount = len(seen)
		if report.Summary.AssessmentCount > 0 {
			report.Summary.AverageScore /= float64(report.Summary.AssessmentCount)
		}
	}
	for _, x := range items {
		if x.Status == "draft" {
			overview.Draft++
		} else if x.Status == "final" {
			overview.Final++
		}
	}

	if err != nil {
		slog.Error("prepare dashboard report", "error", err)
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
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
		"Overview": overview, "Admin": admin,
		"Title":         "Dashboard",
		"Section":       "Ruang kerja",
		"AppName":       h.appName,
		"Authenticated": true,
		"IsDashboard":   true,
		"UserName":      user.Name,
		"UserInitials":  initials(user.Name),
		"StoreName":     user.StoreName,
		"RoleName":      user.RoleName,
		"DateLabel":     indonesianDate(time.Now().In(time.FixedZone("WIB", 7*60*60))),
		"Report":        report,
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
