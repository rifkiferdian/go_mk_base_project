package hrdassessment

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"mksiaga/internal/view"
	"net/http/httptest"
	"os"
	"testing"
)

func TestPeriodValidation(t *testing.T) {
	for _, s := range []string{"2026-11", "2027-01"} {
		if !validPeriod(s) {
			t.Fatal(s)
		}
	}
	for _, s := range []string{"", "2026-13", "2026-1", "2026-11-01"} {
		if validPeriod(s) {
			t.Fatal(s)
		}
	}
}
func TestHRDTemplates(t *testing.T) {
	renderer, e := view.New("../../web/templates")
	if e != nil {
		t.Fatal(e)
	}
	a := Assessment{ID: 1, EmployeeID: 1, Name: "Contoh", Period: "2026-11", Status: "draft", Details: []Detail{{Category: "Terlambat", Condition: "1x", Percentage: sql.NullFloat64{Float64: 25, Valid: true}}, {Category: "HP", Sanction: true}}}
	data := gin.H{"Assessment": a, "Assessments": []Assessment{a}, "Categories": []Category{{ID: 1, Name: "Contoh", Selected: "1", Rules: []Rule{{ID: 1, Condition: "1x", Percentage: sql.NullFloat64{Float64: 25, Valid: true}}, {ID: 2, Condition: "SANKSI", Sanction: true}}}}, "Employees": []Employee{{ID: 1, Name: "Contoh"}}, "CSRFToken": "test"}
	for _, page := range []string{"index", "form", "detail"} {
		w := httptest.NewRecorder()
		if e = renderer.Instance("hrdassessment/"+page+".html", data).Render(w); e != nil {
			t.Fatal(e)
		}
	}
}
func TestMasterDatabaseReadOnly(t *testing.T) {
	dsn := os.Getenv("TEST_HRD_DSN")
	if dsn == "" {
		t.Skip("requires database for read-only verification")
	}
	db, e := sql.Open("mysql", dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	h := NewHandler("", db, nil)
	cats, e := h.categories(c)
	if e != nil {
		t.Fatal(e)
	}
	if len(cats) < 21 {
		t.Fatalf("expected imported categories: %d", len(cats))
	}
	sanctions := 0
	for _, cat := range cats {
		for _, r := range cat.Rules {
			if r.Sanction {
				sanctions++
				if r.Percentage.Valid {
					t.Fatal("sanction converted into percentage")
				}
			}
		}
	}
	if sanctions == 0 {
		t.Fatal("missing sanction")
	}
}
