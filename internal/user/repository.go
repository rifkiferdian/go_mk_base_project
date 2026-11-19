package user

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Users(ctx context.Context) ([]User, error) {
	const query = `SELECT u.id,COALESCE(u.employee_number,''),u.name,u.username,COALESCE(u.email,''),COALESCE(u.phone,''),u.status,
		COALESCE(MAX(CASE WHEN us.is_default=1 THEN s.id END),MAX(s.id),0),
		COALESCE(MAX(CASE WHEN us.is_default=1 THEN s.name END),MAX(s.name),''),
		COALESCE(MAX(CASE WHEN us.is_default=1 THEN ro.id END),MAX(ro.id),0),
		COALESCE(GROUP_CONCAT(DISTINCT ro.name ORDER BY ro.name SEPARATOR ', '),''),
		COUNT(DISTINCT us.id),COALESCE(DATE_FORMAT(u.last_login_at,'%d %b %Y %H:%i'),'Belum pernah')
		FROM users u
		LEFT JOIN user_stores us ON us.user_id=u.id
		LEFT JOIN stores s ON s.id=us.store_id
		LEFT JOIN user_store_roles usr ON usr.user_store_id=us.id
		LEFT JOIN roles ro ON ro.id=usr.role_id
		WHERE u.deleted_at IS NULL
		GROUP BY u.id,u.employee_number,u.name,u.username,u.email,u.phone,u.status,u.last_login_at
		ORDER BY u.name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	var result []User
	for rows.Next() {
		var item User
		if err := rows.Scan(&item.ID, &item.EmployeeNumber, &item.Name, &item.Username, &item.Email, &item.Phone, &item.Status, &item.StoreID, &item.StoreName, &item.RoleID, &item.RoleNames, &item.StoreCount, &item.LastLogin); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) User(ctx context.Context, id uint64) (User, error) {
	const query = `SELECT u.id,COALESCE(u.employee_number,''),u.name,u.username,COALESCE(u.email,''),COALESCE(u.phone,''),u.status,
		COALESCE(us.store_id,0),COALESCE(s.name,''),COALESCE((SELECT role_id FROM user_store_roles WHERE user_store_id=us.id ORDER BY role_id LIMIT 1),0)
		FROM users u
		LEFT JOIN user_stores us ON us.user_id=u.id AND us.id=(SELECT us2.id FROM user_stores us2 WHERE us2.user_id=u.id ORDER BY us2.is_default DESC,us2.id LIMIT 1)
		LEFT JOIN stores s ON s.id=us.store_id
		WHERE u.id=? AND u.deleted_at IS NULL`
	var item User
	err := r.db.QueryRowContext(ctx, query, id).Scan(&item.ID, &item.EmployeeNumber, &item.Name, &item.Username, &item.Email, &item.Phone, &item.Status, &item.StoreID, &item.StoreName, &item.RoleID)
	return item, err
}

func (r *Repository) Options(ctx context.Context) ([]StoreOption, []RoleOption, error) {
	storeRows, err := r.db.QueryContext(ctx, `SELECT id,name FROM stores WHERE status='active' AND deleted_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, nil, fmt.Errorf("list stores: %w", err)
	}
	defer storeRows.Close()
	var stores []StoreOption
	for storeRows.Next() {
		var item StoreOption
		if err := storeRows.Scan(&item.ID, &item.Name); err != nil {
			return nil, nil, err
		}
		stores = append(stores, item)
	}
	if err := storeRows.Err(); err != nil {
		return nil, nil, err
	}
	roleRows, err := r.db.QueryContext(ctx, `SELECT id,name FROM roles WHERE guard_name='web' ORDER BY name`)
	if err != nil {
		return nil, nil, fmt.Errorf("list roles: %w", err)
	}
	defer roleRows.Close()
	var roles []RoleOption
	for roleRows.Next() {
		var item RoleOption
		if err := roleRows.Scan(&item.ID, &item.Name); err != nil {
			return nil, nil, err
		}
		roles = append(roles, item)
	}
	return stores, roles, roleRows.Err()
}

func (r *Repository) Employees(ctx context.Context) ([]EmployeeOption, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT employee_number,name FROM employees WHERE status='active' ORDER BY employee_number`)
	if err != nil {
		return nil, fmt.Errorf("list employees for user lookup: %w", err)
	}
	defer rows.Close()
	var result []EmployeeOption
	for rows.Next() {
		var item EmployeeOption
		if err := rows.Scan(&item.Number, &item.Name); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) StoreAssignments(ctx context.Context, userID uint64) ([]StoreAssignment, error) {
	const query = `SELECT s.id,s.name,us.id IS NOT NULL,COALESCE(us.is_default,0),
		COALESCE((SELECT role_id FROM user_store_roles WHERE user_store_id=us.id ORDER BY role_id LIMIT 1),0)
		FROM stores s
		LEFT JOIN user_stores us ON us.store_id=s.id AND us.user_id=? AND us.status='active'
		WHERE s.status='active' AND s.deleted_at IS NULL ORDER BY s.name`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list user store assignments: %w", err)
	}
	defer rows.Close()
	var result []StoreAssignment
	for rows.Next() {
		var item StoreAssignment
		if err := rows.Scan(&item.StoreID, &item.StoreName, &item.Assigned, &item.IsDefault, &item.RoleID); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) SyncStoreAssignments(ctx context.Context, userID uint64, assignments []AssignmentInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `UPDATE user_stores SET status='inactive',is_default=0 WHERE user_id=?`, userID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE usr FROM user_store_roles usr JOIN user_stores us ON us.id=usr.user_store_id WHERE us.user_id=?`, userID); err != nil {
		return err
	}
	for _, assignment := range assignments {
		result, execErr := tx.ExecContext(ctx, `INSERT INTO user_stores(user_id,store_id,is_default,status,joined_at)
			SELECT ?,s.id,?,'active',CURRENT_TIMESTAMP(6) FROM stores s WHERE s.id=? AND s.status='active' AND s.deleted_at IS NULL
			ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(user_stores.id),is_default=VALUES(is_default),status='active'`, userID, assignment.IsDefault, assignment.StoreID)
		if execErr != nil {
			return execErr
		}
		userStoreID, idErr := result.LastInsertId()
		if idErr != nil {
			return idErr
		}
		if userStoreID == 0 {
			return sql.ErrNoRows
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM user_store_roles WHERE user_store_id=?`, userStoreID); err != nil {
			return err
		}
		result, err = tx.ExecContext(ctx, `INSERT INTO user_store_roles(user_store_id,role_id) SELECT ?,id FROM roles WHERE id=? AND guard_name='web'`, userStoreID, assignment.RoleID)
		if err != nil {
			return err
		}
		affected, affectedErr := result.RowsAffected()
		if affectedErr != nil {
			return affectedErr
		}
		if affected == 0 {
			return sql.ErrNoRows
		}
	}
	return tx.Commit()
}

func (r *Repository) Create(ctx context.Context, input Input, passwordHash string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `INSERT INTO users(employee_id,employee_number,name,username,email,password_hash,phone,status) VALUES((SELECT e.id FROM employees e WHERE e.employee_number=NULLIF(?,'')),NULLIF(?,''),?,?,NULLIF(?,''),?,NULLIF(?,''),?)`, input.EmployeeNumber, input.EmployeeNumber, input.Name, input.Username, input.Email, passwordHash, input.Phone, input.Status)
	if err != nil {
		return err
	}
	userID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	result, err = tx.ExecContext(ctx, `INSERT INTO user_stores(user_id,store_id,is_default,status,joined_at) VALUES(?,?,1,'active',CURRENT_TIMESTAMP(6))`, userID, input.StoreID)
	if err != nil {
		return err
	}
	userStoreID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO user_store_roles(user_store_id,role_id) VALUES(?,?)`, userStoreID, input.RoleID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) Update(ctx context.Context, id uint64, input Input, passwordHash string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if passwordHash == "" {
		_, err = tx.ExecContext(ctx, `UPDATE users SET employee_id=(SELECT e.id FROM employees e WHERE e.employee_number=NULLIF(?,'')),employee_number=NULLIF(?,''),name=?,username=?,email=NULLIF(?,''),phone=NULLIF(?,''),status=? WHERE id=? AND deleted_at IS NULL`, input.EmployeeNumber, input.EmployeeNumber, input.Name, input.Username, input.Email, input.Phone, input.Status, id)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE users SET employee_id=(SELECT e.id FROM employees e WHERE e.employee_number=NULLIF(?,'')),employee_number=NULLIF(?,''),name=?,username=?,email=NULLIF(?,''),phone=NULLIF(?,''),status=?,password_hash=? WHERE id=? AND deleted_at IS NULL`, input.EmployeeNumber, input.EmployeeNumber, input.Name, input.Username, input.Email, input.Phone, input.Status, passwordHash, id)
	}
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE user_stores SET is_default=0 WHERE user_id=?`, id); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO user_stores(user_id,store_id,is_default,status,joined_at) VALUES(?,?,1,'active',CURRENT_TIMESTAMP(6)) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),is_default=1,status='active'`, id, input.StoreID)
	if err != nil {
		return err
	}
	userStoreID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM user_store_roles WHERE user_store_id=?`, userStoreID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO user_store_roles(user_store_id,role_id) VALUES(?,?)`, userStoreID, input.RoleID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users u LEFT JOIN user_stores us ON us.user_id=u.id SET u.deleted_at=CURRENT_TIMESTAMP(6),u.status='inactive',us.status='inactive' WHERE u.id=? AND u.deleted_at IS NULL`, id)
	return err
}
