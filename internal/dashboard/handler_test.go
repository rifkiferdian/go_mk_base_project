package dashboard

import (
	"context"
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"mksiaga/internal/assessmentreport"
	"mksiaga/internal/view"
	"net/http/httptest"
	"os"
	"testing"
)

func TestDashboardRenders(t *testing.T) {
	renderer, e := view.New("../../web/templates")
	if e != nil {
		t.Fatal(e)
	}
	for _, admin := range []bool{true, false} {
		data := gin.H{"Admin": admin, "Overview": Overview{Employees: 10, Covered: 8, Missing: 2, Coverage: 80}, "Report": assessmentreport.Report{Recent: []assessmentreport.RecentResult{{EmployeeName: "Contoh"}}}}
		w := httptest.NewRecorder()
		if e = renderer.Instance("dashboard/index.html", data).Render(w); e != nil {
			t.Fatal(e)
		}
	}
}
func TestOverviewReadOnly(t *testing.T) {
	dsn := os.Getenv("TEST_DASHBOARD_DSN")
	if dsn == "" {
		t.Skip("requires read-only dashboard database check")
	}
	db, e := sql.Open("mysql", dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	for _, admin := range []bool{true, false} {
		overview, e := loadOverview(context.Background(), db, 0, admin)
		if e != nil {
			t.Fatal(e)
		}
		if overview.Covered+overview.Missing != overview.Employees {
			t.Fatal("invalid coverage totals")
		}
		if overview.Coverage < 0 || overview.Coverage > 100 {
			t.Fatal("invalid percentage")
		}
	}
}
