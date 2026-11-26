package questiongroup

import (
	"github.com/gin-gonic/gin"
	"mksiaga/internal/view"
	"net/http/httptest"
	"testing"
)

func TestGroupPageRenders(t *testing.T) {
	renderer, err := view.New("../../web/templates")
	if err != nil {
		t.Fatal(err)
	}
	data := gin.H{"Groups": []Group{{ID: 1, Name: "Kafe", Statuses: "Trainee", PositionItems: []Position{{ID: 2, Name: "COOKING"}}, Criteria: []Criterion{{Description: "Contoh", Status: "Trainee"}}}}, "Positions": []Position{{ID: 1, Name: "COOKING"}}, "CSRFToken": "test", "GroupCount": 1, "CriterionCount": 1}
	w := httptest.NewRecorder()
	if err = renderer.Instance("questiongroup/index.html", data).Render(w); err != nil {
		t.Fatal(err)
	}
	if w.Body.Len() == 0 {
		t.Fatal("empty page")
	}
}
