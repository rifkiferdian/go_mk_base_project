package assessmentcriterion

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Criteria(ctx context.Context) ([]Criterion, error) {
	const query = `SELECT c.id,c.aspect_id,a.code,a.name,c.code,c.description,c.weight,c.sort_order,
		COUNT(DISTINCT sr.score),COUNT(DISTINCT d.assessment_id)
		FROM trainee_assessment_criteria c
		JOIN trainee_assessment_aspects a ON a.id=c.aspect_id
		LEFT JOIN trainee_assessment_score_rules sr ON sr.criterion_id=c.id
		LEFT JOIN trainee_assessment_details d ON d.criterion_id=c.id
		GROUP BY c.id,c.aspect_id,a.code,a.name,c.code,c.description,c.weight,c.sort_order
		ORDER BY a.sort_order,c.sort_order,c.id`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list assessment criteria: %w", err)
	}
	defer rows.Close()
	var result []Criterion
	for rows.Next() {
		var item Criterion
		if err := rows.Scan(&item.ID, &item.AspectID, &item.AspectCode, &item.AspectName, &item.Code, &item.Description, &item.Weight, &item.SortOrder, &item.RuleCount, &item.AssessmentCount); err != nil {
			return nil, fmt.Errorf("scan assessment criterion: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Aspects(ctx context.Context) ([]AspectOption, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,name FROM trainee_assessment_aspects ORDER BY sort_order,name`)
	if err != nil {
		return nil, fmt.Errorf("list aspect options: %w", err)
	}
	defer rows.Close()
	var result []AspectOption
	for rows.Next() {
		var item AspectOption
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Criterion(ctx context.Context, id uint64) (Criterion, error) {
	var item Criterion
	err := r.db.QueryRowContext(ctx, `SELECT c.id,c.aspect_id,a.code,a.name,c.code,c.description,c.weight,c.sort_order,
		(SELECT COUNT(*) FROM trainee_assessment_score_rules sr WHERE sr.criterion_id=c.id),
		(SELECT COUNT(*) FROM trainee_assessment_details d WHERE d.criterion_id=c.id)
		FROM trainee_assessment_criteria c JOIN trainee_assessment_aspects a ON a.id=c.aspect_id WHERE c.id=?`, id).
		Scan(&item.ID, &item.AspectID, &item.AspectCode, &item.AspectName, &item.Code, &item.Description, &item.Weight, &item.SortOrder, &item.RuleCount, &item.AssessmentCount)
	return item, err
}

func (r *Repository) AspectExists(ctx context.Context, id uint64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM trainee_assessment_aspects WHERE id=?)`, id).Scan(&exists)
	return exists, err
}

func (r *Repository) Create(ctx context.Context, input Input) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO trainee_assessment_criteria(aspect_id,code,description,weight,sort_order) VALUES(?,?,?,?,?)`, input.AspectID, input.Code, input.Description, input.Weight, input.SortOrder)
	return err
}

func (r *Repository) Update(ctx context.Context, id uint64, input Input) error {
	_, err := r.db.ExecContext(ctx, `UPDATE trainee_assessment_criteria SET aspect_id=?,code=?,description=?,weight=?,sort_order=? WHERE id=?`, input.AspectID, input.Code, input.Description, input.Weight, input.SortOrder, id)
	return err
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM trainee_assessment_score_rules WHERE criterion_id=?`, id); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM trainee_assessment_criteria WHERE id=?`, id)
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
	return tx.Commit()
}
