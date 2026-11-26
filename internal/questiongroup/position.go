package questiongroup

import (
	"github.com/gin-gonic/gin"
	"mksiaga/internal/auth"
	"strconv"
)

type Position struct {
	ID   uint64
	Name string
}

// AddPosition copies the reference rubric and links the entire set atomically.
func (h *Handler) AddPosition(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "Sesi form berakhir.")
		return
	}
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	pid, e2 := strconv.ParseUint(c.PostForm("position_id"), 10, 64)
	if e != nil || e2 != nil || id == 0 || pid == 0 {
		h.redirect(c, "error", "Pilih jabatan yang valid.")
		return
	}
	tx, e := h.db.BeginTx(c, nil)
	if e != nil {
		h.fail(c, e)
		return
	}
	defer tx.Rollback()
	var groupName string
	if e = tx.QueryRowContext(c, "SELECT name FROM assessment_question_groups WHERE id=? FOR UPDATE", id).Scan(&groupName); e != nil {
		h.fail(c, e)
		return
	}
	var source uint64
	e = tx.QueryRowContext(c, `SELECT r.position_id FROM trainee_assessment_criteria c JOIN trainee_assessment_score_rules s ON s.criterion_id=c.id JOIN trainee_assessment_rubrics r ON r.id=s.rubric_id AND r.is_active=1 WHERE c.question_group_id=? GROUP BY r.position_id HAVING COUNT(DISTINCT c.id)=(SELECT COUNT(*) FROM trainee_assessment_criteria WHERE question_group_id=?) ORDER BY r.position_id LIMIT 1`, id, id).Scan(&source)
	if e != nil {
		h.redirect(c, "error", "Kelompok belum memiliki satu jabatan referensi dengan soal lengkap.")
		return
	}
	var conflicts int
	e = tx.QueryRowContext(c, `SELECT COUNT(*) FROM trainee_assessment_criterion_positions cp JOIN trainee_assessment_criteria c ON c.id=cp.criterion_id WHERE cp.position_id=? AND c.employment_status IN (SELECT employment_status FROM trainee_assessment_criteria WHERE question_group_id=?)`, pid, id).Scan(&conflicts)
	if e != nil {
		h.fail(c, e)
		return
	}
	if conflicts > 0 {
		h.redirect(c, "error", "Jabatan sudah memiliki soal untuk status kerja ini. Periksa subaspek sebelum mengganti kelompok.")
		return
	}
	var dept uint64
	if e = tx.QueryRowContext(c, "SELECT department_id FROM positions WHERE id=? FOR UPDATE", pid).Scan(&dept); e != nil {
		h.fail(c, e)
		return
	}
	_, e = tx.ExecContext(c, `INSERT INTO trainee_assessment_rubrics(department_id,position_id,code,name,month_from,month_to,is_active)
 SELECT ?,?,CONCAT('group_',?,'_',r.month_from,'_',COALESCE(r.month_to,0)),CONCAT(LEFT(?,100),' - Bulan ',r.month_from,'-',COALESCE(r.month_to,'+')),r.month_from,r.month_to,1
 FROM trainee_assessment_rubrics r WHERE r.position_id=? AND r.is_active=1 AND EXISTS(SELECT 1 FROM trainee_assessment_score_rules s JOIN trainee_assessment_criteria c ON c.id=s.criterion_id WHERE s.rubric_id=r.id AND c.question_group_id=?) ON DUPLICATE KEY UPDATE is_active=1`, dept, pid, id, groupName, source, id)
	if e != nil {
		h.fail(c, e)
		return
	}
	_, e = tx.ExecContext(c, "INSERT INTO trainee_assessment_criterion_positions(criterion_id,position_id) SELECT id,? FROM trainee_assessment_criteria WHERE question_group_id=?", pid, id)
	if e != nil {
		h.fail(c, e)
		return
	}
	result, e := tx.ExecContext(c, `INSERT INTO trainee_assessment_score_rules(rubric_id,criterion_id,score,finding_min,finding_max,description)
 SELECT target.id,s.criterion_id,s.score,s.finding_min,s.finding_max,s.description
 FROM trainee_assessment_rubrics target JOIN trainee_assessment_rubrics source ON source.position_id=? AND source.is_active=1 AND source.month_from=target.month_from AND COALESCE(source.month_to,0)=COALESCE(target.month_to,0)
 JOIN trainee_assessment_score_rules s ON s.rubric_id=source.id JOIN trainee_assessment_criteria c ON c.id=s.criterion_id
 WHERE target.position_id=? AND target.code=CONCAT('group_',?,'_',target.month_from,'_',COALESCE(target.month_to,0)) AND c.question_group_id=? ON DUPLICATE KEY UPDATE finding_min=VALUES(finding_min),finding_max=VALUES(finding_max),description=VALUES(description)`, source, pid, id, id)
	if e != nil {
		h.fail(c, e)
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		h.redirect(c, "error", "Kelompok belum memiliki aturan skor yang dapat disalin.")
		return
	}
	if e = tx.Commit(); e != nil {
		h.fail(c, e)
		return
	}
	h.redirect(c, "success", "Seluruh soal dan aturan periode sudah terhubung ke jabatan baru.")
}
