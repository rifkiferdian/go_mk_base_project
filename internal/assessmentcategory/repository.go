package assessmentcategory

import (
	"context"
	"database/sql"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Categories(ctx context.Context) ([]Category, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,name,minimum_score,maximum_score,deduction_percentage,sort_order,is_active FROM trainee_assessment_result_categories ORDER BY sort_order,minimum_score`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Category
	for rows.Next() {
		var item Category
		if err = rows.Scan(&item.ID, &item.Name, &item.Minimum, &item.Maximum, &item.Deduction, &item.SortOrder, &item.IsActive); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Category(ctx context.Context, id uint64) (Category, error) {
	var item Category
	err := r.db.QueryRowContext(ctx, `SELECT id,name,minimum_score,maximum_score,deduction_percentage,sort_order,is_active FROM trainee_assessment_result_categories WHERE id=?`, id).Scan(&item.ID, &item.Name, &item.Minimum, &item.Maximum, &item.Deduction, &item.SortOrder, &item.IsActive)
	return item, err
}

func (r *Repository) Overlaps(ctx context.Context, id uint64, minimum, maximum float64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM trainee_assessment_result_categories WHERE id<>? AND minimum_score<=? AND maximum_score>=?`, id, maximum, minimum).Scan(&count)
	return count > 0, err
}

func (r *Repository) Create(ctx context.Context, input Input) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO trainee_assessment_result_categories(name,minimum_score,maximum_score,deduction_percentage,sort_order,is_active) VALUES(?,?,?,?,?,?)`, input.Name, input.Minimum, input.Maximum, input.Deduction, input.SortOrder, input.IsActive)
	return err
}

func (r *Repository) Update(ctx context.Context, id uint64, input Input) error {
	result, err := r.db.ExecContext(ctx, `UPDATE trainee_assessment_result_categories SET name=?,minimum_score=?,maximum_score=?,deduction_percentage=?,sort_order=?,is_active=? WHERE id=?`, input.Name, input.Minimum, input.Maximum, input.Deduction, input.SortOrder, input.IsActive, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err == nil && n == 0 {
		_, err = r.Category(ctx, id)
	}
	return err
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM trainee_assessment_result_categories WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
