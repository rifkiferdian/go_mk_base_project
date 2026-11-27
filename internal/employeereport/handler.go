package employeereport

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
	"mksiaga/internal/auth"
	"mksiaga/internal/navigation"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type Handler struct {
	db  *sql.DB
	nav *navigation.Service
	app string
}

func NewHandler(app string, db *sql.DB, nav *navigation.Service) *Handler {
	return &Handler{db, nav, app}
}

type Employee struct {
	Supervisor                                        string
	ID                                                uint64
	Number, Name, Position, Department, Store, Status string
	Count                                             int
	Latest                                            string
}

func (e Employee) StatusLabel() string {
	switch e.Status {
	case "trainee":
		return "Trainee"
	case "permanent":
		return "Tetap"
	case "post_retirement_daily_worker":
		return "HLPP"
	case "daily_worker":
		return "HL"
	case "probation":
		return "Probation"
	case "contract":
		return "Kontrak"
	}
	return "Belum diisi"
}

type Detail struct {
	CriterionID            uint64
	Aspect, Question, Rule string
	Weight                 float64
	Score                  sql.NullInt64
}
type Assessment struct {
	CreatedAt                                                                    string
	ID                                                                           uint64
	Date, Calendar, Position, Department, Store, Assessor, Notes, Period, Rating string
	Month                                                                        int
	Average                                                                      float64
	Details                                                                      []Detail
}
type PositionReport struct {
	Key, Name           string
	Assessments         []Assessment
	First, Last, Change float64
}

func (a Assessment) TotalScore() float64 {
	total := 0.0
	for _, d := range a.Details {
		if d.Score.Valid {
			total += float64(d.Score.Int64) * d.Weight
		}
	}
	return total
}
func (a Assessment) MaximumScore() float64 {
	total := 0.0
	for _, d := range a.Details {
		total += 5 * d.Weight
	}
	return total
}

func access(c *gin.Context) (uint64, bool) {
	u, _ := auth.CurrentUser(c)
	id, _ := strconv.ParseUint(u.ID, 10, 64)
	for _, r := range strings.Split(u.RoleName, ",") {
		if strings.TrimSpace(r) == "superadmin" {
			return id, true
		}
	}
	return id, false
}
func scope(uid uint64, admin bool) (string, []any) {
	if admin {
		return "", nil
	}
	return ` AND (EXISTS(SELECT 1 FROM employee_assignments ca JOIN users u ON u.id=? JOIN employee_assignments sa ON sa.employee_id=u.employee_id AND sa.ended_at IS NULL JOIN employee_reporting_lines rl ON rl.supervisor_assignment_id=sa.id AND rl.employee_assignment_id=ca.id AND rl.ended_at IS NULL WHERE ca.employee_id=e.id AND ca.ended_at IS NULL) OR EXISTS(SELECT 1 FROM trainee_assessments ah JOIN employee_reporting_lines rh ON rh.employee_assignment_id=ah.trainee_assignment_id JOIN employee_assignments sh ON sh.id=rh.supervisor_assignment_id JOIN users uh ON uh.employee_id=sh.employee_id AND uh.id=? WHERE ah.trainee_employee_id=e.id AND rh.started_at<=ah.assessment_date AND (rh.ended_at IS NULL OR rh.ended_at>=ah.assessment_date)))`, []any{uid, uid}
}
func (h *Handler) employees(ctx context.Context, uid uint64, admin bool) ([]Employee, error) {
	filter, args := scope(uid, admin)
	assessmentScope := ""
	if !admin {
		assessmentScope = " AND EXISTS(SELECT 1 FROM employee_reporting_lines rl JOIN employee_assignments sa ON sa.id=rl.supervisor_assignment_id JOIN users u ON u.employee_id=sa.employee_id AND u.id=" + strconv.FormatUint(uid, 10) + " WHERE rl.employee_assignment_id=a.trainee_assignment_id AND rl.started_at<=a.assessment_date AND (rl.ended_at IS NULL OR rl.ended_at>=a.assessment_date))"
	}
	rows, err := h.db.QueryContext(ctx, `SELECT COALESCE((SELECT GROUP_CONCAT(DISTINCT CONCAT(se.name,' (',se.employee_number,')') ORDER BY se.name SEPARATOR ', ') FROM employee_reporting_lines rl JOIN employee_assignments sa ON sa.id=rl.supervisor_assignment_id AND sa.ended_at IS NULL JOIN employees se ON se.id=sa.employee_id WHERE rl.employee_assignment_id=ca.id AND rl.ended_at IS NULL),'Belum ada atasan'),e.id,e.employee_number,e.name,COALESCE(p.name,'Belum ditempatkan'),COALESCE(d.name,''),COALESCE(s.name,''),COALESCE(ep.employment_status,''),(SELECT COUNT(*) FROM trainee_assessments a WHERE a.trainee_employee_id=e.id AND a.status='final'`+assessmentScope+`),COALESCE((SELECT DATE_FORMAT(MAX(a.assessment_date),'%Y-%m-%d') FROM trainee_assessments a WHERE a.trainee_employee_id=e.id AND a.status='final'`+assessmentScope+`),'') FROM employees e LEFT JOIN employee_profiles ep ON ep.employee_id=e.id LEFT JOIN employee_assignments ca ON ca.employee_id=e.id AND ca.ended_at IS NULL LEFT JOIN positions p ON p.id=ca.position_id LEFT JOIN departments d ON d.id=ca.department_id LEFT JOIN stores s ON s.id=ca.store_id WHERE 1=1`+filter+` ORDER BY (SELECT MAX(a.created_at) FROM trainee_assessments a WHERE a.trainee_employee_id=e.id AND a.status='final'`+assessmentScope+`) DESC,e.name,e.id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Employee
	for rows.Next() {
		var x Employee
		if err = rows.Scan(&x.Supervisor, &x.ID, &x.Number, &x.Name, &x.Position, &x.Department, &x.Store, &x.Status, &x.Count, &x.Latest); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (h *Handler) Index(c *gin.Context) {
	uid, admin := access(c)
	items, err := h.employees(c, uid, admin)
	if err != nil {
		h.fail(c, err)
		return
	}
	assessed, results := 0, 0
	supervisors := map[string]bool{}
	for _, x := range items {
		if x.Count > 0 {
			assessed++
		}
		results += x.Count
		if x.Supervisor != "Belum ada atasan" {
			supervisors[x.Supervisor] = true
		}
	}
	h.render(c, "index", gin.H{"Employees": items, "EmployeeCount": len(items), "AssessedCount": assessed, "UnassessedCount": len(items) - assessed, "ResultCount": results})
}
func (h *Handler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.Redirect(303, "/errors/404")
		return
	}
	uid, admin := access(c)
	employees, err := h.employees(c, uid, admin)
	if err != nil {
		h.fail(c, err)
		return
	}
	var employee Employee
	for _, e := range employees {
		if e.ID == id {
			employee = e
			break
		}
	}
	if employee.ID == 0 {
		c.Redirect(303, "/errors/404")
		return
	}
	items, err := h.assessments(c, id, uid, admin)
	if err != nil {
		h.fail(c, err)
		return
	}
	groups := groupByPosition(items)
	for i := range groups {
		sort.SliceStable(groups[i].Assessments, func(a, b int) bool {
			x, y := groups[i].Assessments[a], groups[i].Assessments[b]
			if x.CreatedAt == y.CreatedAt {
				return x.ID > y.ID
			}
			return x.CreatedAt > y.CreatedAt
		})
	}
	sort.SliceStable(groups, func(a, b int) bool {
		x, y := groups[a].Assessments[0], groups[b].Assessments[0]
		if x.CreatedAt == y.CreatedAt {
			return x.ID > y.ID
		}
		return x.CreatedAt > y.CreatedAt
	})
	h.render(c, "detail", gin.H{"Employee": employee, "Positions": groups, "AssessmentCount": len(items)})
}
func (h *Handler) assessments(ctx context.Context, id, uid uint64, admin bool) ([]Assessment, error) {
	args := []any{id}
	extra := ""
	if !admin {
		extra = ` AND EXISTS(SELECT 1 FROM employee_reporting_lines rl JOIN employee_assignments sa ON sa.id=rl.supervisor_assignment_id JOIN users u ON u.employee_id=sa.employee_id AND u.id=? WHERE rl.employee_assignment_id=a.trainee_assignment_id AND rl.started_at<=a.assessment_date AND (rl.ended_at IS NULL OR rl.ended_at>=a.assessment_date))`
		args = append(args, uid)
	}
	rows, err := h.db.QueryContext(ctx, `SELECT DATE_FORMAT(a.created_at,'%Y-%m-%d %H:%i:%s.%f'),a.id,DATE_FORMAT(a.assessment_date,'%Y-%m-%d'),DATE_FORMAT(a.assessment_date,'%Y-%m'),COALESCE(a.trainee_position_name,'Jabatan tidak tercatat'),a.department_name,a.store_name,a.assessor_name,COALESCE(a.notes,''),COALESCE(r.name,'Periode tidak tercatat'),a.training_month,COALESCE(v.average_score,0),COALESCE(v.rating,'') FROM trainee_assessments a LEFT JOIN trainee_assessment_rubrics r ON r.id=a.rubric_id LEFT JOIN trainee_assessment_results v ON v.assessment_id=a.id WHERE a.trainee_employee_id=? AND a.status='final'`+extra+" ORDER BY a.assessment_date,a.id", args...)
	if err != nil {
		return nil, err
	}
	var out []Assessment
	for rows.Next() {
		var x Assessment
		if err = rows.Scan(&x.CreatedAt, &x.ID, &x.Date, &x.Calendar, &x.Position, &x.Department, &x.Store, &x.Assessor, &x.Notes, &x.Period, &x.Month, &x.Average, &x.Rating); err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, x)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	for i := range out {
		dr, e := h.db.QueryContext(ctx, "SELECT criterion_id,aspect_name,criterion_description,weight,score,COALESCE(score_rule_description,'') FROM trainee_assessment_details WHERE assessment_id=? ORDER BY criterion_id", out[i].ID)
		if e != nil {
			return nil, e
		}
		for dr.Next() {
			var d Detail
			if e = dr.Scan(&d.CriterionID, &d.Aspect, &d.Question, &d.Weight, &d.Score, &d.Rule); e != nil {
				dr.Close()
				return nil, e
			}
			out[i].Details = append(out[i].Details, d)
		}
		e = dr.Err()
		dr.Close()
		if e != nil {
			return nil, e
		}
	}
	return out, nil
}
func groupByPosition(items []Assessment) []PositionReport {
	var out []PositionReport
	index := map[string]int{}
	for _, a := range items {
		key := a.Position + " / " + a.Department
		i, ok := index[key]
		if !ok {
			i = len(out)
			index[key] = i
			out = append(out, PositionReport{Key: strconv.Itoa(i), Name: key, First: a.Average})
		}
		out[i].Assessments = append(out[i].Assessments, a)
		out[i].Last = a.Average
		out[i].Change = out[i].Last - out[i].First
	}
	return out
}
func (h *Handler) render(c *gin.Context, page string, data gin.H) {
	u, _ := auth.CurrentUser(c)
	menus, err := h.nav.Menus(c, u, "employee-reports")
	if err != nil {
		h.fail(c, err)
		return
	}
	_, admin := access(c)
	token, err := auth.CSRFToken(c)
	if err != nil {
		h.fail(c, err)
		return
	}
	data["IsSuperadmin"] = admin
	data["CSRFToken"] = token
	data["SuccessMessage"] = c.Query("success")
	data["AppName"] = h.app
	data["Title"] = "Laporan per Karyawan"
	data["Section"] = "Laporan"
	data["IsDashboard"] = true
	data["Authenticated"] = true
	data["UserName"] = u.Name
	data["UserInitials"] = "US"
	data["RoleName"] = u.RoleName
	data["StoreName"] = u.StoreName
	data["Menus"] = menus
	c.HTML(http.StatusOK, "employeereport/"+page+".html", data)
}
func (h *Handler) fail(c *gin.Context, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		c.Redirect(303, "/errors/404")
		return
	}
	slog.Error("employee report", "error", err)
	c.Redirect(303, "/errors/500")
}
