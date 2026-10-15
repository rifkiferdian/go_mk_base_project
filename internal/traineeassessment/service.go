package traineeassessment

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"strings"
	"time"
)

var (
	ErrInvalid   = errors.New("invalid assessment")
	ErrDuplicate = errors.New("duplicate assessment")
	ErrFinal     = errors.New("final assessment locked")
	ErrNoRubric  = errors.New("no unique complete rubric for department and month")
)

type Service struct{ repository *Repository }

func NewService(r *Repository) *Service { return &Service{r} }
func allAccess(roles string) bool {
	for _, r := range strings.Split(roles, ",") {
		if strings.TrimSpace(r) == "superadmin" {
			return true
		}
	}
	return false
}
func (s *Service) List(c context.Context, uid uint64, roles string) ([]Assessment, error) {
	items, err := s.repository.Assessments(c, uid, allAccess(roles))
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index].Details, err = s.repository.AssessmentDetails(c, items[index].ID)
		if err != nil {
			return nil, err
		}
	}
	return items, nil
}
func (s *Service) Form(c context.Context, id, uid uint64, roles string) (FormData, error) {
	trainees, e := s.repository.Trainees(c, uid, allAccess(roles))
	if e != nil {
		return FormData{}, e
	}
	aspects, e := s.repository.Rubric(c)
	if e != nil {
		return FormData{}, e
	}
	f := FormData{Trainees: trainees, Aspects: aspects}
	if id == 0 {
		return f, nil
	}
	a, e := s.repository.Assessment(c, id, uid, allAccess(roles))
	if e != nil {
		return f, e
	}
	if a.Status == "final" {
		return f, ErrFinal
	}
	scores, e := s.repository.DetailScores(c, id)
	if e != nil {
		return f, e
	}
	for ai := range aspects {
		for ci := range aspects[ai].Criteria {
			aspects[ai].Criteria[ci].Score = scores[aspects[ai].Criteria[ci].ID]
		}
	}
	f.Assessment = a
	f.Aspects = aspects
	f.Editing = true
	return f, nil
}
func (s *Service) Create(c context.Context, uid uint64, name, roles string, x Input) (uint64, error) {
	if e := s.prepare(c, uid, roles, &x); e != nil {
		return 0, e
	}
	return s.saveErr(s.repository.Create(c, uid, name, x, allAccess(roles)))
}
func (s *Service) Update(c context.Context, id, uid uint64, name, roles string, x Input) (uint64, error) {
	old, e := s.repository.Assessment(c, id, uid, allAccess(roles))
	if e != nil {
		return 0, e
	}
	if old.Status == "final" {
		return 0, ErrFinal
	}
	if e = s.prepare(c, uid, roles, &x); e != nil {
		return 0, e
	}
	return s.saveErr(s.repository.Update(c, id, uid, name, x, allAccess(roles)))
}
func (s *Service) Delete(c context.Context, id, uid uint64, roles string) error {
	return s.repository.Delete(c, id, uid, allAccess(roles))
}
func (s *Service) saveErr(id uint64, e error) (uint64, error) {
	var me *mysql.MySQLError
	if errors.As(e, &me) {
		if me.Number == 1062 {
			return 0, ErrDuplicate
		}
		if me.Number == 3819 || me.Number == 4025 {
			return 0, ErrInvalid
		}
	}
	return id, e
}
func (s *Service) prepare(ctx context.Context, uid uint64, roles string, x *Input) error {
	trainees, err := s.repository.Trainees(ctx, uid, allAccess(roles))
	if err != nil {
		return err
	}
	for _, t := range trainees {
		if t.EmployeeID == x.TraineeEmployeeID {
			x.DepartmentID = t.DepartmentID
		}
	}
	if x.DepartmentID == 0 {
		return ErrInvalid
	}
	return s.validate(ctx, x)
}

func (s *Service) validate(ctx context.Context, x *Input) error {
	if x.TraineeEmployeeID == 0 || x.TrainingMonth == 0 || x.TrainingMonth > 24 || !validDate(x.AssessmentDate) || !oneOf(x.Status, "draft", "final") {
		return ErrInvalid
	}
	var rubricCount int
	var rubricID uint64
	if err := s.repository.db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(MIN(id),0) FROM trainee_assessment_rubrics WHERE department_id=? AND is_active=1 AND month_from<=? AND (month_to IS NULL OR month_to>=?)`, x.DepartmentID, x.TrainingMonth, x.TrainingMonth).Scan(&rubricCount, &rubricID); err != nil {
		return err
	}
	if rubricCount != 1 {
		return ErrNoRubric
	}
	x.RubricID = rubricID
	aspects, err := s.repository.Rubric(ctx)
	if err != nil {
		return err
	}
	expected := make(map[uint64]bool)
	for _, aspect := range aspects {
		for _, criterion := range aspect.Criteria {
			if !criterion.appliesTo(x.DepartmentID, x.TrainingMonth) {
				continue
			}
			if x.RubricID != 0 && x.RubricID != criterion.RubricID {
				return ErrNoRubric
			}
			x.RubricID = criterion.RubricID
			if len(criterion.Rules) != 5 {
				return ErrNoRubric
			}
			expected[criterion.ID] = true
		}
	}
	if len(expected) == 0 {
		return ErrNoRubric
	}
	if len(x.Details) != len(expected) {
		return ErrInvalid
	}
	filled := 0
	seen := make(map[uint64]bool)
	for _, d := range x.Details {
		if !expected[d.CriterionID] || seen[d.CriterionID] || d.Score < 0 || d.Score > 5 {
			return ErrInvalid
		}
		seen[d.CriterionID] = true
		if d.Score > 0 {
			filled++
		}
	}
	if x.Status == "final" && (filled != len(x.Details) || strings.TrimSpace(x.Notes) == "") {
		return ErrInvalid
	}
	return nil
}

func (c Criterion) appliesTo(departmentID, month uint64) bool {
	return c.DepartmentID == departmentID && month >= c.MonthFrom && (c.MonthTo == 0 || month <= c.MonthTo)
}
func validDate(v string) bool { _, e := time.Parse("2006-01-02", v); return e == nil }
func oneOf(v string, xs ...string) bool {
	for _, x := range xs {
		if v == x {
			return true
		}
	}
	return false
}

var _ = sql.ErrNoRows
