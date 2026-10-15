package traineeassessment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"mksiaga/internal/assessmentcriterion"
	"mksiaga/internal/assessmentrule"
	"mksiaga/internal/view"
)

func TestCriterionPeriodBoundaries(t *testing.T) {
	periods := []Criterion{{DepartmentID: 2, MonthFrom: 1, MonthTo: 2}, {DepartmentID: 2, MonthFrom: 3, MonthTo: 4}, {DepartmentID: 2, MonthFrom: 5}}
	for _, month := range []uint64{1, 2, 3, 4, 5, 6, 24} {
		count := 0
		for _, c := range periods {
			if c.appliesTo(2, month) {
				count++
			}
			if c.appliesTo(3, month) {
				t.Fatal("Gudang rubric accepted Kasir")
			}
		}
		if count != 1 {
			t.Fatalf("month %d: %d matches", month, count)
		}
	}
}

// Run against a migrated/seeded disposable copy; never defaults to the app DB.
func TestRubricDatabaseIntegration(t *testing.T) {
	name := os.Getenv("TEST_RUBRIC_DB")
	if name == "" {
		t.Skip("set TEST_RUBRIC_DB to a migrated disposable database")
	}
	_ = godotenv.Load("../../.env")
	cfg := mysql.NewConfig()
	cfg.User = os.Getenv("DB_USER")
	cfg.Passwd = os.Getenv("DB_PASSWORD")
	cfg.Net = "tcp"
	cfg.Addr = os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT")
	cfg.DBName = name
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	repo := NewRepository(db)
	service := NewService(repo)
	var gudang, kasir uint64
	if err = db.QueryRow(`SELECT id FROM departments WHERE code='gudang'`).Scan(&gudang); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRow(`SELECT id FROM departments WHERE code='kasir'`).Scan(&kasir); err != nil {
		t.Fatal(err)
	}
	aspects, err := repo.Rubric(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, month := range []uint64{1, 2, 3, 4, 5, 6, 24} {
		x := Input{TraineeEmployeeID: 1, DepartmentID: gudang, TrainingMonth: month, AssessmentDate: "2026-10-15", Status: "final", Notes: "Integration test"}
		for _, a := range aspects {
			for _, c := range a.Criteria {
				if c.appliesTo(gudang, month) {
					x.Details = append(x.Details, DetailInput{CriterionID: c.ID, Score: 5})
				}
			}
		}
		if len(x.Details) != 10 {
			t.Fatalf("month %d: got %d criteria", month, len(x.Details))
		}
		if err = service.validate(ctx, &x); err != nil {
			t.Fatalf("month %d: %v", month, err)
		}
		x.DepartmentID = kasir
		if err = service.validate(ctx, &x); !errors.Is(err, ErrNoRubric) {
			t.Fatalf("Kasir should have no rubric: %v", err)
		}
		x.DepartmentID = gudang
		if month == 1 {
			// The same ten criterion IDs are intentionally valid in every period;
			// only the five score rules selected by training month differ.
			x.TrainingMonth = 3
			if err = service.validate(ctx, &x); err != nil {
				t.Fatalf("shared criteria rejected for another period: %v", err)
			}
		}
	}
	// Every finding count has exactly one score, including all thresholds shown.
	for _, test := range []struct {
		code   string
		month  int
		scores []int
	}{
		{"layanan_1", 1, []int{5, 4, 4, 3, 2, 1, 1}},
		{"kemampuan_1", 1, []int{5, 5, 5, 4, 3, 2, 1}},
		{"kemampuan_1", 3, []int{5, 5, 4, 3, 2, 1, 1}},
		{"etika_1", 3, []int{5, 4, 4, 3, 2, 1, 1}},
		{"kemampuan_1", 5, []int{5, 4, 3, 2, 1, 1, 1}},
		{"kerjasama_1", 5, []int{5, 4, 3, 2, 2, 1, 1}},
		{"keinginan_belajar_2", 5, []int{5, 4, 3, 2, 2, 1, 1}},
	} {
		for findings, want := range test.scores {
			var count, got int
			err = db.QueryRow(`SELECT COUNT(*),COALESCE(MAX(s.score),0) FROM trainee_assessment_score_rules s JOIN trainee_assessment_criteria c ON c.id=s.criterion_id JOIN trainee_assessment_rubrics r ON r.id=s.rubric_id WHERE r.department_id=? AND r.month_from=? AND c.code=? AND s.finding_min<=? AND (s.finding_max IS NULL OR s.finding_max>=?)`, gudang, test.month, test.code, findings, findings).Scan(&count, &got)
			if err != nil || count != 1 || got != want {
				t.Fatalf("%s month=%d findings=%d: count=%d score=%d want=%d err=%v", test.code, test.month, findings, count, got, want, err)
			}
		}
	}
	cr := assessmentcriterion.NewRepository(db)
	criteria, err := cr.Criteria(ctx)
	if err != nil {
		t.Fatal(err)
	}
	rubrics, err := cr.Rubrics(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range criteria {
		if _, err = cr.Criterion(ctx, c.ID); err != nil {
			t.Fatal(err)
		}
	}
	rr := assessmentrule.NewRepository(db)
	rules, err := rr.Rules(ctx)
	if err != nil {
		t.Fatal(err)
	}
	options, err := rr.Criteria(ctx)
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := view.New("../../web/templates")
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("TEST_RUBRIC_WRITES") == "1" {
		trainees, err := repo.Trainees(ctx, 0, true)
		if err != nil {
			t.Fatal(err)
		}
		var trainee Trainee
		for _, candidate := range trainees {
			if candidate.DepartmentID == gudang {
				trainee = candidate
				break
			}
		}
		if trainee.EmployeeID == 0 {
			t.Fatal("test copy needs a Gudang trainee")
		}
		var uid uint64
		if err := db.QueryRow(`SELECT MIN(id) FROM users`).Scan(&uid); err != nil {
			t.Fatal(err)
		}
		x := Input{TraineeEmployeeID: trainee.EmployeeID, DepartmentID: gudang, TrainingMonth: 24, AssessmentDate: "2026-10-15", Status: "final", Notes: "Rubric integration test"}
		for _, a := range aspects {
			for _, c := range a.Criteria {
				if c.appliesTo(gudang, 24) {
					x.Details = append(x.Details, DetailInput{CriterionID: c.ID, Score: 4})
				}
			}
		}
		id, err := service.Create(ctx, uid, "Integration test", "superadmin", x)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Exec(`DELETE FROM trainee_assessments WHERE id=?`, id)
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM trainee_assessment_details WHERE assessment_id=? AND criterion_code IS NOT NULL AND score_rule_description='1x temuan' AND finding_min=1 AND finding_max=1`, id).Scan(&count); err != nil || count != 10 {
			t.Fatalf("snapshots: count=%d err=%v", count, err)
		}
		if _, err = service.Update(ctx, id, uid, "Integration test", "superadmin", x); !errors.Is(err, ErrFinal) {
			t.Fatalf("final assessment editable: %v", err)
		}
		var average float64
		if err = db.QueryRow(`SELECT average_score FROM trainee_assessment_results WHERE assessment_id=?`, id).Scan(&average); err != nil || average != 4 {
			t.Fatalf("result=%v err=%v", average, err)
		}
	}
	for page, data := range map[string]map[string]any{
		"trainee_assessment/form.html":   {"Form": FormData{Aspects: aspects}},
		"assessmentcriterion/index.html": {"Criteria": criteria, "Rubrics": rubrics},
		"assessmentrule/index.html":      {"Rules": rules, "Criteria": options},
	} {
		if err = renderer.Instance(page, data).Render(httptest.NewRecorder()); err != nil {
			t.Fatal(fmt.Errorf("render %s: %w", page, err))
		}
	}
}
