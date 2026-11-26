package questiongroup

import (
	"github.com/gin-gonic/gin"
	"mksiaga/internal/auth"
	"strconv"
)

// Unused rules are removed; rubrics referenced by assessments remain intact.
func (h *Handler) RemovePosition(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "Sesi form berakhir.")
		return
	}
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	pid, e2 := strconv.ParseUint(c.Param("positionID"), 10, 64)
	if e != nil || e2 != nil || id == 0 || pid == 0 {
		h.redirect(c, "error", "Kelompok atau jabatan tidak valid.")
		return
	}
	tx, e := h.db.BeginTx(c, nil)
	if e != nil {
		h.fail(c, e)
		return
	}
	defer tx.Rollback()
	var name string
	if e = tx.QueryRowContext(c, "SELECT name FROM assessment_question_groups WHERE id=? FOR UPDATE", id).Scan(&name); e != nil {
		h.fail(c, e)
		return
	}
	result, e := tx.ExecContext(c, "DELETE cp FROM trainee_assessment_criterion_positions cp JOIN trainee_assessment_criteria c ON c.id=cp.criterion_id WHERE cp.position_id=? AND c.question_group_id=?", pid, id)
	if e != nil {
		h.fail(c, e)
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		h.redirect(c, "error", "Jabatan sudah tidak terhubung ke kelompok ini.")
		return
	}
	// Lock headers before checking usage so assessment inserts cannot race cleanup.
	rows, e := tx.QueryContext(c, "SELECT id FROM trainee_assessment_rubrics WHERE position_id=? FOR UPDATE", pid)
	if e != nil {
		h.fail(c, e)
		return
	}
	var rubricIDs []uint64
	for rows.Next() {
		var rubricID uint64
		if e = rows.Scan(&rubricID); e != nil {
			rows.Close()
			h.fail(c, e)
			return
		}
		rubricIDs = append(rubricIDs, rubricID)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		h.fail(c, e)
		return
	}
	for _, rubricID := range rubricIDs {
		_, e = tx.ExecContext(c, `DELETE s FROM trainee_assessment_score_rules s JOIN trainee_assessment_criteria c ON c.id=s.criterion_id WHERE s.rubric_id=? AND c.question_group_id=? AND NOT EXISTS(SELECT 1 FROM trainee_assessments a WHERE a.rubric_id=s.rubric_id)`, rubricID, id)
		if e != nil {
			h.fail(c, e)
			return
		}
		_, e = tx.ExecContext(c, `DELETE FROM trainee_assessment_rubrics WHERE id=? AND NOT EXISTS(SELECT 1 FROM trainee_assessment_score_rules WHERE rubric_id=?) AND NOT EXISTS(SELECT 1 FROM trainee_assessments WHERE rubric_id=?)`, rubricID, rubricID, rubricID)
		if e != nil {
			h.fail(c, e)
			return
		}
	}
	if e = tx.Commit(); e != nil {
		h.fail(c, e)
		return
	}
	h.redirect(c, "success", "Jabatan berhasil dilepas. Aturan yang belum dipakai dan periode kosong dihapus; data karyawan, soal bersama, serta riwayat penilaian tetap tersimpan.")
}
