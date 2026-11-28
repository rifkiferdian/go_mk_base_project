package hrdassessment

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"log/slog"
	"mksiaga/internal/auth"
	"mksiaga/internal/navigation"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Handler struct {
	db  *sql.DB
	app string
	nav *navigation.Service
}

func NewHandler(app string, db *sql.DB, nav *navigation.Service) *Handler {
	return &Handler{db, app, nav}
}

type Employee struct {
	ID                     uint64
	Number, Name, Position string
}
type Rule struct {
	ID         uint64
	Condition  string
	Percentage sql.NullFloat64
	Sanction   bool
}
type Category struct {
	ID                uint64
	Name, Description string
	Rules             []Rule
	Selected          string
	Notes             string
	Order             int
}
type Assessment struct {
	ID, EmployeeID                                                                               uint64
	Period, Number, Name, Position, Department, Store, EmploymentStatus, Assessor, Status, Notes string
	Details                                                                                      []Detail
	Violations, Sanctions                                                                        int
}
type Detail struct {
	Category, Description, Condition, Notes string
	Percentage                              sql.NullFloat64
	Sanction                                bool
}

func (h *Handler) categories(c *gin.Context) ([]Category, error) {
	rows, e := h.db.QueryContext(c, "SELECT id,name,description,sort_order FROM hrd_deduction_categories WHERE is_active=1 ORDER BY sort_order,id")
	if e != nil {
		return nil, e
	}
	var out []Category
	for rows.Next() {
		var x Category
		if e = rows.Scan(&x.ID, &x.Name, &x.Description, &x.Order); e != nil {
			rows.Close()
			return nil, e
		}
		out = append(out, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	for i := range out {
		rr, err := h.db.QueryContext(c, "SELECT id,condition_text,deduction_percentage,is_sanction FROM hrd_deduction_rules WHERE category_id=? AND is_active=1 ORDER BY sort_order,id", out[i].ID)
		if err != nil {
			return nil, err
		}
		for rr.Next() {
			var r Rule
			if err = rr.Scan(&r.ID, &r.Condition, &r.Percentage, &r.Sanction); err != nil {
				rr.Close()
				return nil, err
			}
			out[i].Rules = append(out[i].Rules, r)
		}
		err = rr.Err()
		rr.Close()
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
func (h *Handler) Index(c *gin.Context) {
	where := ""
	args := []any{}
	employeeFilter := ""
	if raw := c.Query("employee_id"); raw != "" {
		id, e := strconv.ParseUint(raw, 10, 64)
		if e != nil || id == 0 {
			h.fail(c, sql.ErrNoRows)
			return
		}
		if e = h.db.QueryRowContext(c, "SELECT name FROM employees WHERE id=?", id).Scan(&employeeFilter); e != nil {
			h.fail(c, e)
			return
		}
		where = " WHERE a.employee_id=?"
		args = append(args, id)
	}
	rows, e := h.db.QueryContext(c, `SELECT a.id,a.employee_id,DATE_FORMAT(a.period_month,'%Y-%m'),a.employee_number,a.employee_name,a.position_name,a.department_name,a.store_name,a.employment_status,a.assessor_name,a.status,a.notes,(SELECT COUNT(*) FROM hrd_monthly_assessment_details d WHERE d.assessment_id=a.id AND (d.is_sanction=1 OR d.deduction_percentage>0)),(SELECT COUNT(*) FROM hrd_monthly_assessment_details d WHERE d.assessment_id=a.id AND d.is_sanction=1) FROM hrd_monthly_assessments a`+where+` ORDER BY a.period_month DESC,a.updated_at DESC,a.id DESC`, args...)
	if e != nil {
		h.fail(c, e)
		return
	}
	var out []Assessment
	for rows.Next() {
		var a Assessment
		if e = rows.Scan(&a.ID, &a.EmployeeID, &a.Period, &a.Number, &a.Name, &a.Position, &a.Department, &a.Store, &a.EmploymentStatus, &a.Assessor, &a.Status, &a.Notes, &a.Violations, &a.Sanctions); e != nil {
			rows.Close()
			h.fail(c, e)
			return
		}
		out = append(out, a)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		h.fail(c, e)
		return
	}
	final, draft := 0, 0
	for _, a := range out {
		if a.Status == "final" {
			final++
		} else {
			draft++
		}
	}
	h.render(c, "index", gin.H{"EmployeeFilter": employeeFilter, "Assessments": out, "Count": len(out), "FinalCount": final, "DraftCount": draft})
}
func (h *Handler) load(c *gin.Context, id uint64) (Assessment, error) {
	var a Assessment
	e := h.db.QueryRowContext(c, `SELECT id,employee_id,DATE_FORMAT(period_month,'%Y-%m'),employee_number,employee_name,position_name,department_name,store_name,employment_status,assessor_name,status,notes FROM hrd_monthly_assessments WHERE id=?`, id).Scan(&a.ID, &a.EmployeeID, &a.Period, &a.Number, &a.Name, &a.Position, &a.Department, &a.Store, &a.EmploymentStatus, &a.Assessor, &a.Status, &a.Notes)
	if e != nil {
		return a, e
	}
	rows, e := h.db.QueryContext(c, "SELECT category_name,category_description,condition_text,deduction_percentage,is_sanction,notes FROM hrd_monthly_assessment_details WHERE assessment_id=? ORDER BY sort_order,id", id)
	if e != nil {
		return a, e
	}
	defer rows.Close()
	for rows.Next() {
		var d Detail
		if e = rows.Scan(&d.Category, &d.Description, &d.Condition, &d.Percentage, &d.Sanction, &d.Notes); e != nil {
			return a, e
		}
		if d.Sanction {
			a.Sanctions++
		}
		if d.Sanction || (d.Percentage.Valid && d.Percentage.Float64 > 0) {
			a.Violations++
		}
		a.Details = append(a.Details, d)
	}
	return a, rows.Err()
}
func (h *Handler) Detail(c *gin.Context) {
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	if e != nil {
		h.fail(c, sql.ErrNoRows)
		return
	}
	a, e := h.load(c, id)
	if e != nil {
		h.fail(c, e)
		return
	}
	h.render(c, "detail", gin.H{"Assessment": a})
}
func (h *Handler) Form(c *gin.Context) {
	var a Assessment
	a.Period = time.Now().In(time.FixedZone("WIB", 25200)).Format("2006-01")
	if c.Param("id") != "" {
		id, e := strconv.ParseUint(c.Param("id"), 10, 64)
		if e != nil {
			h.fail(c, sql.ErrNoRows)
			return
		}
		a, e = h.load(c, id)
		if e != nil {
			h.fail(c, e)
			return
		}
		if a.Status == "final" {
			c.Redirect(303, fmt.Sprintf("/hrd-assessments/%d", a.ID))
			return
		}
	}
	cats, e := h.categories(c)
	if e != nil {
		h.fail(c, e)
		return
	}
	if a.ID > 0 {
		rows, err := h.db.QueryContext(c, "SELECT category_id,COALESCE(rule_id,0),notes,deduction_percentage,is_sanction FROM hrd_monthly_assessment_details WHERE assessment_id=?", a.ID)
		if err != nil {
			h.fail(c, err)
			return
		}
		for rows.Next() {
			var cid sql.NullInt64
			var rid uint64
			var note string
			var pct sql.NullFloat64
			var sanction bool
			if err = rows.Scan(&cid, &rid, &note, &pct, &sanction); err != nil {
				rows.Close()
				h.fail(c, err)
				return
			}
			for i := range cats {
				if cid.Valid && cats[i].ID == uint64(cid.Int64) {
					cats[i].Notes = note
					if rid > 0 {
						for _, r := range cats[i].Rules {
							if r.ID == rid {
								cats[i].Selected = strconv.FormatUint(rid, 10)
							}
						}
					} else if pct.Valid && pct.Float64 == 0 && !sanction {
						cats[i].Selected = "0"
					}
				}
			}
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			h.fail(c, err)
			return
		}
	}
	rows, e := h.db.QueryContext(c, "SELECT e.id,e.employee_number,e.name,p.name FROM employees e JOIN employee_assignments a ON a.employee_id=e.id AND a.ended_at IS NULL JOIN positions p ON p.id=a.position_id JOIN job_levels j ON j.id=p.job_level_id WHERE e.status='active' AND j.code='staff' ORDER BY e.name")
	if e != nil {
		h.fail(c, e)
		return
	}
	var employees []Employee
	for rows.Next() {
		var x Employee
		if e = rows.Scan(&x.ID, &x.Number, &x.Name, &x.Position); e != nil {
			rows.Close()
			h.fail(c, e)
			return
		}
		employees = append(employees, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		h.fail(c, e)
		return
	}
	h.render(c, "form", gin.H{"Assessment": a, "Categories": cats, "Employees": employees})
}
func validPeriod(s string) bool {
	d, e := time.Parse("2006-01", s)
	return e == nil && d.Format("2006-01") == s
}
func (h *Handler) Save(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		c.Redirect(303, "/errors/419")
		return
	}
	eid, e := strconv.ParseUint(c.PostForm("employee_id"), 10, 64)
	period, status, notes := c.PostForm("period"), c.PostForm("status"), strings.TrimSpace(c.PostForm("notes"))
	if e != nil || eid == 0 || !validPeriod(period) || (status != "draft" && status != "final") || utf8.RuneCountInString(notes) > 5000 {
		h.message(c, "error", "Karyawan, bulan, status atau catatan tidak valid.")
		return
	}
	id := uint64(0)
	if c.Param("id") != "" {
		id, e = strconv.ParseUint(c.Param("id"), 10, 64)
		if e != nil || id == 0 {
			h.fail(c, sql.ErrNoRows)
			return
		}
	}
	cats, e := h.categories(c)
	if e != nil {
		h.fail(c, e)
		return
	}
	if len(cats) == 0 {
		h.message(c, "error", "Belum ada kategori aktif.")
		return
	}
	type selected struct {
		Category Category
		Rule     *Rule
		Notes    string
	}
	var selectedRows []selected
	for _, cat := range cats {
		raw := c.PostForm(fmt.Sprintf("rule_%d", cat.ID))
		note := strings.TrimSpace(c.PostForm(fmt.Sprintf("note_%d", cat.ID)))
		if utf8.RuneCountInString(note) > 2000 {
			h.message(c, "error", "Catatan kategori maksimal 2000 karakter.")
			return
		}
		if raw == "" {
			if status == "final" {
				h.message(c, "error", "Pilih hasil untuk seluruh kategori sebelum final.")
				return
			}
			continue
		}
		var rule *Rule
		if raw != "0" {
			for _, r := range cat.Rules {
				if strconv.FormatUint(r.ID, 10) == raw {
					copy := r
					rule = &copy
					break
				}
			}
			if rule == nil {
				h.message(c, "error", "Aturan potongan tidak valid atau nonaktif.")
				return
			}
		}
		selectedRows = append(selectedRows, selected{cat, rule, note})
	}
	tx, e := h.db.BeginTx(c, nil)
	if e != nil {
		h.fail(c, e)
		return
	}
	defer tx.Rollback()
	if id > 0 {
		var oldStatus string
		if e = tx.QueryRowContext(c, "SELECT status FROM hrd_monthly_assessments WHERE id=? FOR UPDATE", id).Scan(&oldStatus); e != nil {
			h.fail(c, e)
			return
		}
		if oldStatus == "final" {
			h.message(c, "error", "Penilaian final terkunci.")
			return
		}
	}
	var number, name, position, department, store, employment string
	e = tx.QueryRowContext(c, `SELECT e.employee_number,e.name,COALESCE(p.name,''),COALESCE(d.name,''),COALESCE(s.name,''),COALESCE(ep.employment_status,'') FROM employees e LEFT JOIN employee_profiles ep ON ep.employee_id=e.id LEFT JOIN employee_assignments a ON a.employee_id=e.id AND a.ended_at IS NULL LEFT JOIN positions p ON p.id=a.position_id LEFT JOIN departments d ON d.id=a.department_id LEFT JOIN stores s ON s.id=a.store_id JOIN job_levels j ON j.id=p.job_level_id AND j.code='staff' WHERE e.id=? AND e.status='active'`, eid).Scan(&number, &name, &position, &department, &store, &employment)
	if e != nil {
		h.fail(c, e)
		return
	}
	u, _ := auth.CurrentUser(c)
	if id == 0 {
		result, err := tx.ExecContext(c, `INSERT INTO hrd_monthly_assessments(employee_id,period_month,employee_number,employee_name,position_name,department_name,store_name,employment_status,assessor_user_id,assessor_name,status,notes,finalized_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,IF(?='final',CURRENT_TIMESTAMP(6),NULL))`, eid, period+"-01", number, name, position, department, store, employment, u.ID, u.Name, status, notes, status)
		e = err
		if e == nil {
			v, _ := result.LastInsertId()
			id = uint64(v)
		}
	} else {
		_, e = tx.ExecContext(c, `UPDATE hrd_monthly_assessments SET employee_id=?,period_month=?,employee_number=?,employee_name=?,position_name=?,department_name=?,store_name=?,employment_status=?,assessor_user_id=?,assessor_name=?,status=?,notes=?,finalized_at=IF(?='final',CURRENT_TIMESTAMP(6),NULL) WHERE id=?`, eid, period+"-01", number, name, position, department, store, employment, u.ID, u.Name, status, notes, status, id)
	}
	if e != nil {
		var me *mysql.MySQLError
		if errors.As(e, &me) && me.Number == 1062 {
			h.message(c, "error", "Karyawan sudah memiliki penilaian HRD pada bulan tersebut. Edit draft yang sudah ada.")
			return
		}
		h.fail(c, e)
		return
	}
	if _, e = tx.ExecContext(c, "DELETE FROM hrd_monthly_assessment_details WHERE assessment_id=?", id); e != nil {
		h.fail(c, e)
		return
	}
	for _, x := range selectedRows {
		var rid any
		var pct any = 0
		condition := "Tidak ada pelanggaran"
		sanction := false
		if x.Rule != nil {
			rid = x.Rule.ID
			condition = x.Rule.Condition
			sanction = x.Rule.Sanction
			if x.Rule.Percentage.Valid {
				pct = x.Rule.Percentage.Float64
			} else {
				pct = nil
			}
		}
		_, e = tx.ExecContext(c, "INSERT INTO hrd_monthly_assessment_details(assessment_id,category_id,rule_id,category_name,category_description,condition_text,deduction_percentage,is_sanction,notes,sort_order) VALUES(?,?,?,?,?,?,?,?,?,?)", id, x.Category.ID, rid, x.Category.Name, x.Category.Description, condition, pct, sanction, x.Notes, x.Category.Order)
		if e != nil {
			h.fail(c, e)
			return
		}
	}
	if e = tx.Commit(); e != nil {
		h.fail(c, e)
		return
	}
	c.Redirect(303, fmt.Sprintf("/hrd-assessments/%d?success=Penilaian+berhasil+disimpan.", id))
}
func (h *Handler) message(c *gin.Context, key, msg string) {
	c.Redirect(303, "/hrd-assessments?"+key+"="+url.QueryEscape(msg))
}
func (h *Handler) render(c *gin.Context, page string, data gin.H) {
	u, _ := auth.CurrentUser(c)
	menus, e := h.nav.Menus(c, u, "hrd-assessments")
	if e != nil {
		h.fail(c, e)
		return
	}
	token, e := auth.CSRFToken(c)
	if e != nil {
		h.fail(c, e)
		return
	}
	data["AppName"] = h.app
	data["Title"] = "Penilaian Bulanan HRD"
	data["Section"] = "Penilaian Umum"
	data["Authenticated"] = true
	data["IsDashboard"] = true
	data["UserName"] = u.Name
	data["UserInitials"] = "HR"
	data["RoleName"] = u.RoleName
	data["StoreName"] = u.StoreName
	data["Menus"] = menus
	data["CSRFToken"] = token
	data["SuccessMessage"] = c.Query("success")
	data["ErrorMessage"] = c.Query("error")
	c.HTML(200, "hrdassessment/"+page+".html", data)
}
func (h *Handler) fail(c *gin.Context, e error) {
	if errors.Is(e, sql.ErrNoRows) {
		c.Redirect(303, "/errors/404")
		return
	}
	slog.Error("hrd assessment", "error", e)
	c.Redirect(303, "/errors/500")
}
