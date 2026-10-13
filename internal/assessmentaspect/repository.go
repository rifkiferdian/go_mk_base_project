package assessmentaspect

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Aspects(ctx context.Context) ([]Aspect, error) {
	const query = `SELECT a.id,a.code,a.name,a.sort_order,COUNT(c.id)
		FROM trainee_assessment_aspects a
		LEFT JOIN trainee_assessment_criteria c ON c.aspect_id=a.id
		GROUP BY a.id,a.code,a.name,a.sort_order
		ORDER BY a.sort_order,a.name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list assessment aspects: %w", err)
	}
	defer rows.Close()
	var result []Aspect
	for rows.Next() {
		var item Aspect
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.SortOrder, &item.CriterionCount); err != nil {
			return nil, fmt.Errorf("scan assessment aspect: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Aspect(ctx context.Context, id uint64) (Aspect, error) {
	var item Aspect
	err := r.db.QueryRowContext(ctx, `SELECT a.id,a.code,a.name,a.sort_order,COUNT(c.id)
		FROM trainee_assessment_aspects a LEFT JOIN trainee_assessment_criteria c ON c.aspect_id=a.id
		WHERE a.id=? GROUP BY a.id,a.code,a.name,a.sort_order`, id).
		Scan(&item.ID, &item.Code, &item.Name, &item.SortOrder, &item.CriterionCount)
	return item, err
}

func (r *Repository) Create(ctx context.Context, input Input) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO trainee_assessment_aspects(code,name,sort_order) VALUES(?,?,?)`, input.Code, input.Name, input.SortOrder)
	return err
}

func (r *Repository) Update(ctx context.Context, id uint64, input Input) error {
	_, err := r.db.ExecContext(ctx, `UPDATE trainee_assessment_aspects SET code=?,name=?,sort_order=? WHERE id=?`, input.Code, input.Name, input.SortOrder, id)
	return err
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM trainee_assessment_aspects WHERE id=?`, id)
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
