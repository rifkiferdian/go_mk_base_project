package assessmentreport

import (
	"log/slog"
	"net/http"
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

func NewHandler(appName string, service *Service, navigation *navigation.Service) *Handler {
	return &Handler{appName: appName, service: service, navigation: navigation}
}

func (h *Handler) Index(c *gin.Context) {
	report, err := h.service.Report(c)
	if err != nil {
		slog.Error("assessment report", "error", err)
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	user, _ := auth.CurrentUser(c)
	menus, err := h.navigation.Menus(c, user, "trainee-assessment-reports")
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	token, err := auth.CSRFToken(c)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	c.HTML(http.StatusOK, "assessment_report/index.html", gin.H{
		"AppName": h.appName, "Authenticated": true, "IsDashboard": true,
		"Title": "Laporan Penilaian Trainee", "Section": "Laporan", "ActivePage": "trainee-assessment-reports",
		"UserName": user.Name, "UserInitials": reportInitials(user.Name), "StoreName": user.StoreName, "RoleName": user.RoleName,
		"CSRFToken": token, "Menus": menus, "Report": report,
	})
}

func reportInitials(name string) string {
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
