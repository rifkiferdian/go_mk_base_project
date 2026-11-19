UPDATE users u
JOIN employees e ON e.employee_number=u.employee_number
SET u.employee_id=e.id
WHERE u.employee_id IS NULL AND u.deleted_at IS NULL;
