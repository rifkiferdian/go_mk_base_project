package hrdmaster

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
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

type Category struct {
	ID                uint64
	Name, Description string
	Order             int
	Active            bool
	Rules             int
}
type Rule struct {
	ID, CategoryID      uint64
	Category, Condition string
	Percentage          sql.NullFloat64
	Sanction, Active    bool
	Order               int
}

func (h *Handler) categories(c *gin.Context) ([]Category, error) {
	rows, e := h.db.QueryContext(c, "SELECT c.id,c.name,c.description,c.sort_order,c.is_active,(SELECT COUNT(*) FROM hrd_deduction_rules r WHERE r.category_id=c.id) FROM hrd_deduction_categories c ORDER BY c.sort_order,c.id")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []Category
	for rows.Next() {
		var x Category
		if e = rows.Scan(&x.ID, &x.Name, &x.Description, &x.Order, &x.Active, &x.Rules); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (h *Handler) Categories(c *gin.Context) { h.page(c, false) }
func (h *Handler) Rules(c *gin.Context)      { h.page(c, true) }
func (h *Handler) page(c *gin.Context, rules bool) {
	cats, e := h.categories(c)
	if e != nil {
		h.fail(c, e)
		return
	}
	var items []Rule
	if rules {
		rows, err := h.db.QueryContext(c, "SELECT r.id,r.category_id,c.name,r.condition_text,r.deduction_percentage,r.is_sanction,r.sort_order,r.is_active FROM hrd_deduction_rules r JOIN hrd_deduction_categories c ON c.id=r.category_id ORDER BY c.sort_order,c.id,r.sort_order,r.id")
		if err != nil {
			h.fail(c, err)
			return
		}
		for rows.Next() {
			var x Rule
			if err = rows.Scan(&x.ID, &x.CategoryID, &x.Category, &x.Condition, &x.Percentage, &x.Sanction, &x.Order, &x.Active); err != nil {
				rows.Close()
				h.fail(c, err)
				return
			}
			items = append(items, x)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			h.fail(c, err)
			return
		}
	}
	key, title := "hrd-categories", "Kategori Potongan HRD"
	if rules {
		key, title = "hrd-rules", "Aturan Potongan HRD"
	}
	u, _ := auth.CurrentUser(c)
	menus, e := h.nav.Menus(c, u, key)
	if e != nil {
		h.fail(c, e)
		return
	}
	token, e := auth.CSRFToken(c)
	if e != nil {
		h.fail(c, e)
		return
	}
	active := 0
	if rules {
		for _, x := range items {
			if x.Active {
				active++
			}
		}
	} else {
		for _, x := range cats {
			if x.Active {
				active++
			}
		}
	}
	count := len(cats)
	if rules {
		count = len(items)
	}
	c.HTML(http.StatusOK, "hrdmaster/index.html", gin.H{"Title": title, "Section": "Master Umum HRD", "AppName": h.app, "IsDashboard": true, "Authenticated": true, "UserName": u.Name, "UserInitials": "HR", "StoreName": u.StoreName, "RoleName": u.RoleName, "Menus": menus, "CSRFToken": token, "Categories": cats, "Rules": items, "IsRules": rules, "PageKey": key, "Count": count, "ActiveCount": active, "SuccessMessage": c.Query("success"), "ErrorMessage": c.Query("error")})
}
func (h *Handler) SaveCategory(c *gin.Context) { h.save(c, false) }
func (h *Handler) SaveRule(c *gin.Context)     { h.save(c, true) }
func (h *Handler) save(c *gin.Context, rules bool) {
	key := "hrd-categories"
	if rules {
		key = "hrd-rules"
	}
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, key, "error", "Sesi form berakhir.")
		return
	}
	id := uint64(0)
	var e error
	if c.Param("id") != "" {
		id, e = strconv.ParseUint(c.Param("id"), 10, 64)
		if e != nil || id == 0 {
			h.redirect(c, key, "error", "Data tidak valid.")
			return
		}
	}
	order, e := strconv.Atoi(c.PostForm("sort_order"))
	active := c.PostForm("is_active") == "1"
	if e != nil || order < 1 || order > 65535 {
		h.redirect(c, key, "error", "Urutan harus 1 sampai 65535.")
		return
	}
	if !rules {
		name, description := strings.TrimSpace(c.PostForm("name")), strings.TrimSpace(c.PostForm("description"))
		if name == "" || utf8.RuneCountInString(name) > 255 || utf8.RuneCountInString(description) > 3000 {
			h.redirect(c, key, "error", "Nama atau keterangan kategori tidak valid.")
			return
		}
		if id == 0 {
			_, e = h.db.ExecContext(c, "INSERT INTO hrd_deduction_categories(name,description,sort_order,is_active) VALUES(?,?,?,?)", name, description, order, active)
		} else {
			_, e = h.db.ExecContext(c, "UPDATE hrd_deduction_categories SET name=?,description=?,sort_order=?,is_active=? WHERE id=?", name, description, order, active, id)
		}
	} else {
		cid, err := strconv.ParseUint(c.PostForm("category_id"), 10, 64)
		condition := strings.TrimSpace(c.PostForm("condition_text"))
		sanction := c.PostForm("is_sanction") == "1"
		percentage, err2 := strconv.ParseFloat(c.PostForm("percentage"), 64)
		if err != nil || cid == 0 || condition == "" || utf8.RuneCountInString(condition) > 1000 || (!sanction && (err2 != nil || !(percentage == 25 || percentage == 50 || percentage == 75 || percentage == 100))) {
			h.redirect(c, key, "error", "Kategori, kondisi atau persentase tidak valid.")
			return
		}
		var value any = percentage
		if sanction {
			value = nil
		}
		if id == 0 {
			_, e = h.db.ExecContext(c, "INSERT INTO hrd_deduction_rules(category_id,condition_text,deduction_percentage,is_sanction,sort_order,is_active) VALUES(?,?,?,?,?,?)", cid, condition, value, sanction, order, active)
		} else {
			_, e = h.db.ExecContext(c, "UPDATE hrd_deduction_rules SET category_id=?,condition_text=?,deduction_percentage=?,is_sanction=?,sort_order=?,is_active=? WHERE id=?", cid, condition, value, sanction, order, active, id)
		}
	}
	h.result(c, key, e, "Data master berhasil disimpan.")
}
func (h *Handler) DeleteCategory(c *gin.Context) { h.remove(c, false) }
func (h *Handler) DeleteRule(c *gin.Context)     { h.remove(c, true) }
func (h *Handler) remove(c *gin.Context, rules bool) {
	key, table := "hrd-categories", "hrd_deduction_categories"
	if rules {
		key, table = "hrd-rules", "hrd_deduction_rules"
	}
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, key, "error", "Sesi form berakhir.")
		return
	}
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	if e != nil || id == 0 {
		h.redirect(c, key, "error", "Data tidak valid.")
		return
	}
	_, e = h.db.ExecContext(c, "DELETE FROM "+table+" WHERE id=?", id)
	h.result(c, key, e, "Data master berhasil dihapus.")
}
func (h *Handler) result(c *gin.Context, key string, e error, msg string) {
	if e == nil {
		h.redirect(c, key, "success", msg)
		return
	}
	var me *mysql.MySQLError
	if errors.As(e, &me) {
		if me.Number == 1062 {
			h.redirect(c, key, "error", "Nama kategori sudah digunakan.")
			return
		}
		if me.Number == 1451 {
			h.redirect(c, key, "error", "Kategori masih memiliki aturan potongan. Nonaktifkan atau hapus aturannya terlebih dahulu.")
			return
		}
		if me.Number == 1452 {
			h.redirect(c, key, "error", "Kategori tidak ditemukan.")
			return
		}
	}
	h.fail(c, e)
}
func (h *Handler) redirect(c *gin.Context, key, param, msg string) {
	c.Redirect(303, "/master/"+key+"?"+param+"="+url.QueryEscape(msg))
}
func (h *Handler) fail(c *gin.Context, e error) {
	slog.Error("hrd master", "error", e)
	c.Redirect(303, "/errors/500")
}
