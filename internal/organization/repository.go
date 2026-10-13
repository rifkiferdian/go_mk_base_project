package organization

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) JobLevels(ctx context.Context) ([]JobLevel, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT l.id,l.code,l.name,l.rank_order,COUNT(p.id) FROM job_levels l LEFT JOIN positions p ON p.job_level_id=l.id GROUP BY l.id,l.code,l.name,l.rank_order ORDER BY l.rank_order,l.name`)
	if err != nil {
		return nil, fmt.Errorf("list job levels: %w", err)
	}
	defer rows.Close()
	var result []JobLevel
	for rows.Next() {
		var item JobLevel
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.RankOrder, &item.PositionCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (r *Repository) JobLevel(ctx context.Context, id uint64) (JobLevel, error) {
	var item JobLevel
	err := r.db.QueryRowContext(ctx, `SELECT l.id,l.code,l.name,l.rank_order,COUNT(p.id) FROM job_levels l LEFT JOIN positions p ON p.job_level_id=l.id WHERE l.id=? GROUP BY l.id,l.code,l.name,l.rank_order`, id).Scan(&item.ID, &item.Code, &item.Name, &item.RankOrder, &item.PositionCount)
	return item, err
}
func (r *Repository) CreateJobLevel(ctx context.Context, input JobLevelInput) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO job_levels(code,name,rank_order) VALUES(?,?,?)`, input.Code, input.Name, input.RankOrder)
	return err
}
func (r *Repository) UpdateJobLevel(ctx context.Context, id uint64, input JobLevelInput) error {
	_, err := r.db.ExecContext(ctx, `UPDATE job_levels SET code=?,name=?,rank_order=? WHERE id=?`, input.Code, input.Name, input.RankOrder, id)
	return err
}
func (r *Repository) DeleteJobLevel(ctx context.Context, id uint64) error {
	return deleteByID(ctx, r.db, `DELETE FROM job_levels WHERE id=?`, id)
}

func (r *Repository) Positions(ctx context.Context) ([]Position, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT p.id,p.job_level_id,l.name,l.rank_order,p.code,p.name,COUNT(a.id) FROM positions p JOIN job_levels l ON l.id=p.job_level_id LEFT JOIN employee_assignments a ON a.position_id=p.id GROUP BY p.id,p.job_level_id,l.name,l.rank_order,p.code,p.name ORDER BY l.rank_order,p.name`)
	if err != nil {
		return nil, fmt.Errorf("list positions: %w", err)
	}
	defer rows.Close()
	var result []Position
	for rows.Next() {
		var item Position
		if err := rows.Scan(&item.ID, &item.JobLevelID, &item.JobLevelName, &item.JobLevelRank, &item.Code, &item.Name, &item.AssignmentCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (r *Repository) Position(ctx context.Context, id uint64) (Position, error) {
	var item Position
	err := r.db.QueryRowContext(ctx, `SELECT p.id,p.job_level_id,l.name,l.rank_order,p.code,p.name,COUNT(a.id) FROM positions p JOIN job_levels l ON l.id=p.job_level_id LEFT JOIN employee_assignments a ON a.position_id=p.id WHERE p.id=? GROUP BY p.id,p.job_level_id,l.name,l.rank_order,p.code,p.name`, id).Scan(&item.ID, &item.JobLevelID, &item.JobLevelName, &item.JobLevelRank, &item.Code, &item.Name, &item.AssignmentCount)
	return item, err
}
func (r *Repository) CreatePosition(ctx context.Context, input PositionInput) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO positions(job_level_id,code,name) VALUES(?,?,?)`, input.JobLevelID, input.Code, input.Name)
	return err
}
func (r *Repository) UpdatePosition(ctx context.Context, id uint64, input PositionInput) error {
	_, err := r.db.ExecContext(ctx, `UPDATE positions SET job_level_id=?,code=?,name=? WHERE id=?`, input.JobLevelID, input.Code, input.Name, id)
	return err
}
func (r *Repository) DeletePosition(ctx context.Context, id uint64) error {
	return deleteByID(ctx, r.db, `DELETE FROM positions WHERE id=?`, id)
}

func (r *Repository) Departments(ctx context.Context) ([]Department, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT d.id,d.code,d.name,COUNT(a.id) FROM departments d LEFT JOIN employee_assignments a ON a.department_id=d.id GROUP BY d.id,d.code,d.name ORDER BY d.name`)
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}
	defer rows.Close()
	var result []Department
	for rows.Next() {
		var item Department
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.AssignmentCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (r *Repository) Department(ctx context.Context, id uint64) (Department, error) {
	var item Department
	err := r.db.QueryRowContext(ctx, `SELECT d.id,d.code,d.name,COUNT(a.id) FROM departments d LEFT JOIN employee_assignments a ON a.department_id=d.id WHERE d.id=? GROUP BY d.id,d.code,d.name`, id).Scan(&item.ID, &item.Code, &item.Name, &item.AssignmentCount)
	return item, err
}
func (r *Repository) CreateDepartment(ctx context.Context, input DepartmentInput) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO departments(code,name) VALUES(?,?)`, input.Code, input.Name)
	return err
}
func (r *Repository) UpdateDepartment(ctx context.Context, id uint64, input DepartmentInput) error {
	_, err := r.db.ExecContext(ctx, `UPDATE departments SET code=?,name=? WHERE id=?`, input.Code, input.Name, id)
	return err
}
func (r *Repository) DeleteDepartment(ctx context.Context, id uint64) error {
	return deleteByID(ctx, r.db, `DELETE FROM departments WHERE id=?`, id)
}

func deleteByID(ctx context.Context, db *sql.DB, query string, id uint64) error {
	result, err := db.ExecContext(ctx, query, id)
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
