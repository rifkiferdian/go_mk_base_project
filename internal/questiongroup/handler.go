package questiongroup

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"log/slog"
	"mksiaga/internal/auth"
	"mksiaga/internal/navigation"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Handler struct {
	db  *sql.DB
	nav *navigation.Service
	app string
}

func NewHandler(app string, db *sql.DB, nav *navigation.Service) *Handler {
	return &Handler{db, nav, app}
}

type Group struct {
	PositionItems                                   []Position
	ID                                              uint64
	Name, Description, Statuses, Positions, Periods string
	Criteria                                        []Criterion
	PositionCount                                   int
}
type Criterion struct {
	ID                                uint64
	Code, Description, Aspect, Status string
}

func statusLabel(s string) string {
	switch s {
	case "trainee":
		return "Trainee"
	case "probation":
		return "Probation"
	case "contract":
		return "Kontrak"
	case "permanent":
		return "Tetap"
	case "daily_worker":
		return "HL"
	case "post_retirement_daily_worker":
		return "HLPP"
	}
	return s
}
func (h *Handler) Index(c *gin.Context) {
	rows, err := h.db.QueryContext(c, "SELECT id,name,description FROM assessment_question_groups ORDER BY name")
	if err != nil {
		h.fail(c, err)
		return
	}
	var groups []Group
	for rows.Next() {
		var g Group
		if err = rows.Scan(&g.ID, &g.Name, &g.Description); err != nil {
			rows.Close()
			h.fail(c, err)
			return
		}
		groups = append(groups, g)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		h.fail(c, err)
		return
	}
	total, trainee, hlpp := 0, 0, 0
	for i := range groups {
		g := &groups[i]
		cr, e := h.db.QueryContext(c, `SELECT c.id,c.code,c.description,a.name,c.employment_status FROM trainee_assessment_criteria c JOIN trainee_assessment_aspects a ON a.id=c.aspect_id WHERE c.question_group_id=? ORDER BY a.sort_order,c.sort_order,c.id`, g.ID)
		if e != nil {
			h.fail(c, e)
			return
		}
		statuses := map[string]bool{}
		for cr.Next() {
			var x Criterion
			if e = cr.Scan(&x.ID, &x.Code, &x.Description, &x.Aspect, &x.Status); e != nil {
				cr.Close()
				h.fail(c, e)
				return
			}
			statuses[x.Status] = true
			x.Status = statusLabel(x.Status)
			g.Criteria = append(g.Criteria, x)
		}
		e = cr.Err()
		cr.Close()
		if e != nil {
			h.fail(c, e)
			return
		}
		if statuses["trainee"] {
			trainee++
		}
		if statuses["post_retirement_daily_worker"] {
			hlpp++
		}
		var labels []string
		for _, s := range []string{"trainee", "probation", "contract", "permanent", "daily_worker", "post_retirement_daily_worker"} {
			if statuses[s] {
				labels = append(labels, statusLabel(s))
			}
		}
		g.Statuses = strings.Join(labels, ", ")
		positions, pe := h.db.QueryContext(c, "SELECT DISTINCT p.id,p.name FROM positions p JOIN trainee_assessment_criterion_positions cp ON cp.position_id=p.id JOIN trainee_assessment_criteria c ON c.id=cp.criterion_id WHERE c.question_group_id=? ORDER BY p.name", g.ID)
		if pe != nil {
			h.fail(c, pe)
			return
		}
		for positions.Next() {
			var p Position
			if pe = positions.Scan(&p.ID, &p.Name); pe != nil {
				positions.Close()
				h.fail(c, pe)
				return
			}
			g.PositionItems = append(g.PositionItems, p)
		}
		pe = positions.Err()
		positions.Close()
		if pe != nil {
			h.fail(c, pe)
			return
		}
		total += len(g.Criteria)
		e = h.db.QueryRowContext(c, `SELECT COALESCE(GROUP_CONCAT(DISTINCT p.name ORDER BY p.name SEPARATOR ', '),''),COUNT(DISTINCT p.id) FROM trainee_assessment_criteria c JOIN trainee_assessment_criterion_positions cp ON cp.criterion_id=c.id JOIN positions p ON p.id=cp.position_id WHERE c.question_group_id=?`, g.ID).Scan(&g.Positions, &g.PositionCount)
		if e != nil {
			h.fail(c, e)
			return
		}
		e = h.db.QueryRowContext(c, `SELECT COALESCE(GROUP_CONCAT(DISTINCT CONCAT('Bulan ',r.month_from,'-',COALESCE(r.month_to,'+')) ORDER BY r.month_from SEPARATOR ', '),'') FROM trainee_assessment_criteria c JOIN trainee_assessment_score_rules s ON s.criterion_id=c.id JOIN trainee_assessment_rubrics r ON r.id=s.rubric_id AND r.is_active=1 WHERE c.question_group_id=?`, g.ID).Scan(&g.Periods)
		if e != nil {
			h.fail(c, e)
			return
		}
	}
	var ungrouped int
	if err = h.db.QueryRowContext(c, "SELECT COUNT(*) FROM trainee_assessment_criteria WHERE question_group_id IS NULL").Scan(&ungrouped); err != nil {
		h.fail(c, err)
		return
	}
	var positions []Position
	pr, err := h.db.QueryContext(c, "SELECT id,name FROM positions ORDER BY name")
	if err != nil {
		h.fail(c, err)
		return
	}
	for pr.Next() {
		var p Position
		if err = pr.Scan(&p.ID, &p.Name); err != nil {
			pr.Close()
			h.fail(c, err)
			return
		}
		positions = append(positions, p)
	}
	err = pr.Err()
	pr.Close()
	if err != nil {
		h.fail(c, err)
		return
	}
	u, _ := auth.CurrentUser(c)
	menus, err := h.nav.Menus(c, u, "question-groups")
	if err != nil {
		h.fail(c, err)
		return
	}
	token, err := auth.CSRFToken(c)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.HTML(http.StatusOK, "questiongroup/index.html", gin.H{"AppName": h.app, "IsDashboard": true, "Authenticated": true, "Title": "Kelompok Soal", "ActivePage": "question-groups", "Section": "Pengaturan Penilaian", "UserName": u.Name, "UserInitials": "US", "StoreName": u.StoreName, "RoleName": u.RoleName, "Menus": menus, "CSRFToken": token, "Positions": positions, "Groups": groups, "GroupCount": len(groups), "CriterionCount": total, "TraineeCount": trainee, "HLPPCount": hlpp, "UngroupedCount": ungrouped, "ErrorMessage": c.Query("error"), "SuccessMessage": c.Query("success")})
}
func (h *Handler) Save(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "Sesi form berakhir.")
		return
	}
	name, description := strings.TrimSpace(c.PostForm("name")), strings.TrimSpace(c.PostForm("description"))
	if name == "" || utf8.RuneCountInString(name) > 150 || utf8.RuneCountInString(description) > 1000 {
		h.redirect(c, "error", "Nama atau keterangan tidak valid.")
		return
	}
	var err error
	if c.Param("id") == "" {
		_, err = h.db.ExecContext(c, "INSERT INTO assessment_question_groups(code,name,description) VALUES(UUID(),?,?)", name, description)
	} else {
		id, e := strconv.ParseUint(c.Param("id"), 10, 64)
		if e != nil || id == 0 {
			h.redirect(c, "error", "Kelompok tidak valid.")
			return
		}
		_, err = h.db.ExecContext(c, "UPDATE assessment_question_groups SET name=?,description=? WHERE id=?", name, description, id)
	}
	if err != nil {
		h.fail(c, err)
		return
	}
	h.redirect(c, "success", "Kelompok soal berhasil disimpan.")
}
func (h *Handler) Delete(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "Sesi form berakhir.")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		h.redirect(c, "error", "Kelompok tidak valid.")
		return
	}
	result, err := h.db.ExecContext(c, "DELETE FROM assessment_question_groups WHERE id=? AND NOT EXISTS(SELECT 1 FROM trainee_assessment_criteria WHERE question_group_id=?)", id, id)
	if err != nil {
		h.fail(c, err)
		return
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		h.redirect(c, "error", "Kelompok masih berisi soal. Pindahkan subaspeknya terlebih dahulu.")
		return
	}
	h.redirect(c, "success", "Kelompok kosong berhasil dihapus.")
}
func (h *Handler) redirect(c *gin.Context, key, value string) {
	c.Redirect(http.StatusSeeOther, "/master/question-groups?"+key+"="+url.QueryEscape(value))
}
func (h *Handler) fail(c *gin.Context, err error) {
	slog.Error("question group", "error", err)
	c.Redirect(http.StatusSeeOther, "/errors/500")
}
