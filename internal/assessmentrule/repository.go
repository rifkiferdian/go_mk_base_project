package assessmentrule

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Rules(ctx context.Context) ([]Rule, error) {
	const query = `SELECT r.criterion_id,c.code,c.description,a.id,a.name,r.score,r.finding_min,r.finding_max,r.description
		FROM trainee_assessment_score_rules r
		JOIN trainee_assessment_criteria c ON c.id=r.criterion_id
		JOIN trainee_assessment_aspects a ON a.id=c.aspect_id
		ORDER BY a.sort_order,c.sort_order,r.score DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list assessment score rules: %w", err)
	}
	defer rows.Close()
	var result []Rule
	for rows.Next() {
		var item Rule
		var maximum sql.NullInt64
		if err := rows.Scan(&item.CriterionID, &item.CriterionCode, &item.CriterionDescription, &item.AspectID, &item.AspectName, &item.Score, &item.FindingMin, &maximum, &item.Description); err != nil {
			return nil, fmt.Errorf("scan assessment score rule: %w", err)
		}
		item.HasFindingMax = maximum.Valid
		if maximum.Valid {
			item.FindingMax = uint16(maximum.Int64)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Criteria(ctx context.Context) ([]CriterionOption, error) {
	const query = `SELECT c.id,a.id,a.name,c.code,c.description,COUNT(r.score)
		FROM trainee_assessment_criteria c JOIN trainee_assessment_aspects a ON a.id=c.aspect_id
		LEFT JOIN trainee_assessment_score_rules r ON r.criterion_id=c.id
		GROUP BY c.id,a.id,a.name,c.code,c.description,a.sort_order,c.sort_order
		ORDER BY a.sort_order,c.sort_order,c.id`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list criterion options: %w", err)
	}
	defer rows.Close()
	var result []CriterionOption
	for rows.Next() {
		var item CriterionOption
		if err := rows.Scan(&item.ID, &item.AspectID, &item.AspectName, &item.Code, &item.Description, &item.RuleCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Rule(ctx context.Context, key Key) (Rule, error) {
	var item Rule
	var maximum sql.NullInt64
	err := r.db.QueryRowContext(ctx, `SELECT r.criterion_id,c.code,c.description,a.id,a.name,r.score,r.finding_min,r.finding_max,r.description FROM trainee_assessment_score_rules r JOIN trainee_assessment_criteria c ON c.id=r.criterion_id JOIN trainee_assessment_aspects a ON a.id=c.aspect_id WHERE r.criterion_id=? AND r.score=?`, key.CriterionID, key.Score).Scan(&item.CriterionID, &item.CriterionCode, &item.CriterionDescription, &item.AspectID, &item.AspectName, &item.Score, &item.FindingMin, &maximum, &item.Description)
	item.HasFindingMax = maximum.Valid
	if maximum.Valid {
		item.FindingMax = uint16(maximum.Int64)
	}
	return item, err
}

func (r *Repository) CriterionExists(ctx context.Context, id uint64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM trainee_assessment_criteria WHERE id=?)`, id).Scan(&exists)
	return exists, err
}
func (r *Repository) RulesForCriterion(ctx context.Context, id uint64) ([]Rule, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT criterion_id,score,finding_min,finding_max,description FROM trainee_assessment_score_rules WHERE criterion_id=?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Rule
	for rows.Next() {
		var item Rule
		var maximum sql.NullInt64
		if err := rows.Scan(&item.CriterionID, &item.Score, &item.FindingMin, &maximum, &item.Description); err != nil {
			return nil, err
		}
		item.HasFindingMax = maximum.Valid
		if maximum.Valid {
			item.FindingMax = uint16(maximum.Int64)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (r *Repository) Create(ctx context.Context, input Input) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO trainee_assessment_score_rules(criterion_id,score,finding_min,finding_max,description) VALUES(?,?,?,?,?)`, input.CriterionID, input.Score, input.FindingMin, input.FindingMax, input.Description)
	return err
}
func (r *Repository) Update(ctx context.Context, old Key, input Input) error {
	result, err := r.db.ExecContext(ctx, `UPDATE trainee_assessment_score_rules SET criterion_id=?,score=?,finding_min=?,finding_max=?,description=? WHERE criterion_id=? AND score=?`, input.CriterionID, input.Score, input.FindingMin, input.FindingMax, input.Description, old.CriterionID, old.Score)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		if _, findErr := r.Rule(ctx, Key{CriterionID: input.CriterionID, Score: input.Score}); findErr == nil {
			return nil
		}
		return sql.ErrNoRows
	}
	return nil
}
func (r *Repository) Delete(ctx context.Context, key Key) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM trainee_assessment_score_rules WHERE criterion_id=? AND score=?`, key.CriterionID, key.Score)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
