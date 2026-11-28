package hrdmaster

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"mksiaga/internal/view"
	"net/http/httptest"
	"testing"
)

func TestMasterPagesRender(t *testing.T) {
	renderer, e := view.New("../../web/templates")
	if e != nil {
		t.Fatal(e)
	}
	for _, rules := range []bool{true, false} {
		data := gin.H{"IsRules": rules, "Categories": []Category{{ID: 1, Name: "Keterlambatan", Order: 1, Active: true}}, "Rules": []Rule{{ID: 1, CategoryID: 1, Category: "Keterlambatan", Condition: "Contoh", Percentage: sql.NullFloat64{Float64: 25, Valid: true}, Order: 1, Active: true}, {ID: 2, CategoryID: 1, Sanction: true, Order: 2}}, "PageKey": "hrd-rules", "CSRFToken": "test"}
		w := httptest.NewRecorder()
		if e = renderer.Instance("hrdmaster/index.html", data).Render(w); e != nil {
			t.Fatal(e)
		}
	}
}
