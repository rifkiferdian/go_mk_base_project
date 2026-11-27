package employeereport

import (
	"context"
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"mksiaga/internal/view"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGroupByPositionKeepsTimeline(t *testing.T) {
	items := []Assessment{{ID: 1, Position: "Kasir", Department: "Operational", Month: 2, Average: 4}, {ID: 2, Position: "Cooking", Department: "Operational", Month: 1, Average: 3}, {ID: 3, Position: "Kasir", Department: "Operational", Month: 8, Average: 4.5}}
	groups := groupByPosition(items)
	if len(groups) != 2 || len(groups[0].Assessments) != 2 || groups[0].Change != 0.5 || groups[0].Assessments[1].Month != 8 {
		t.Fatalf("incorrect position timelines: %+v", groups)
	}
}
func TestTemplates(t *testing.T) {
	renderer, e := view.New("../../web/templates")
	if e != nil {
		t.Fatal(e)
	}
	a := Assessment{ID: 1, Date: "2026-11-27", Calendar: "2026-11", Position: "COOKING", Period: "Bulan 5-6", Month: 8, Average: 4, Details: []Detail{{Question: "Contoh", Score: sql.NullInt64{Int64: 4, Valid: true}}}}
	data := gin.H{"Employee": Employee{ID: 1, Name: "Contoh", Status: "trainee"}, "Employees": []Employee{{ID: 1, Name: "Contoh", Status: "trainee"}}, "Positions": groupByPosition([]Assessment{a}), "AssessmentCount": 1}
	for _, page := range []string{"index", "detail"} {
		w := httptest.NewRecorder()
		if e = renderer.Instance("employeereport/"+page+".html", data).Render(w); e != nil {
			t.Fatal(e)
		}
	}
}
func TestDatabaseReadOnly(t *testing.T) {
	dsn := os.Getenv("TEST_EMPLOYEE_REPORT_DSN")
	if dsn == "" {
		t.Skip("requires read-only database check")
	}
	db, e := sql.Open("mysql", dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	h := NewHandler("", db, nil)
	employees, e := h.employees(context.Background(), 0, true)
	if e != nil {
		t.Fatal(e)
	}
	if len(employees) > 0 {
		if _, e = h.assessments(context.Background(), employees[0].ID, 0, true); e != nil {
			t.Fatal(e)
		}
	}
	denied, e := h.employees(context.Background(), 0, false)
	if e != nil {
		t.Fatal(e)
	}
	if len(denied) != 0 {
		t.Fatal("unknown user can see employees")
	}
	rows, e := db.Query("SELECT DISTINCT trainee_employee_id FROM trainee_assessments WHERE status='final' LIMIT 5")
	if e != nil {
		t.Fatal(e)
	}
	var ids []uint64
	for rows.Next() {
		var id uint64
		if e = rows.Scan(&id); e != nil {
			t.Fatal(e)
		}
		ids = append(ids, id)
	}
	rows.Close()
	for _, id := range ids {
		if _, e = h.assessments(context.Background(), id, 0, true); e != nil {
			t.Fatal(e)
		}
		denied, e := h.assessments(context.Background(), id, 0, false)
		if e != nil {
			t.Fatal(e)
		}
		if len(denied) != 0 {
			t.Fatal("unknown user can see results")
		}
	}
}
