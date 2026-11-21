package team

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"mksiaga/internal/auth"
	"mksiaga/internal/navigation"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	app string
	r   *Repository
	n   *navigation.Service
}

func NewHandler(a string, r *Repository, n *navigation.Service) *Handler { return &Handler{a, r, n} }
func (h *Handler) Index(c *gin.Context) {
	u, _ := auth.CurrentUser(c)
	uid, _ := strconv.ParseUint(u.ID, 10, 64)
	all := strings.Contains(u.RoleName, "superadmin")
	items, e := h.r.Members(c, uid, all)
	if e != nil {
		slog.Error("team list", "error", e)
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	menus, e := h.n.Menus(c, u, "my-team")
	if e != nil {
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	token, _ := auth.CSRFToken(c)
	departments := make(map[string]struct{})
	stores := make(map[string]struct{})
	traineeCount := 0
	for _, item := range items {
		departments[item.Department] = struct{}{}
		stores[item.Store] = struct{}{}
		if item.Employment == "trainee" {
			traineeCount++
		}
	}
	c.HTML(http.StatusOK, "team/index.html", gin.H{"Title": "Tim Saya", "ActivePage": "my-team", "AppName": h.app, "IsDashboard": true, "Authenticated": true, "Section": "Kepegawaian", "UserName": u.Name, "UserInitials": initials(u.Name), "StoreName": u.StoreName, "RoleName": u.RoleName, "CSRFToken": token, "Menus": menus, "Members": items, "MemberCount": len(items), "TraineeCount": traineeCount, "DepartmentCount": len(departments), "StoreCount": len(stores)})
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
