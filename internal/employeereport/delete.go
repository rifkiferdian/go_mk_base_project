package employeereport

import (
	"github.com/gin-gonic/gin"
	"mksiaga/internal/auth"
	"mksiaga/internal/traineeassessment"
	"strconv"
)

func (h *Handler) DeleteAssessment(c *gin.Context) {
	uid, admin := access(c)
	if !admin {
		c.Redirect(303, "/errors/403")
		return
	}
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		c.Redirect(303, "/errors/419")
		return
	}
	employeeID, e := strconv.ParseUint(c.Param("id"), 10, 64)
	assessmentID, e2 := strconv.ParseUint(c.Param("assessmentID"), 10, 64)
	if e != nil || e2 != nil || employeeID == 0 || assessmentID == 0 {
		c.Redirect(303, "/errors/404")
		return
	}
	var owner uint64
	if e = h.db.QueryRowContext(c, "SELECT trainee_employee_id FROM trainee_assessments WHERE id=?", assessmentID).Scan(&owner); e != nil {
		h.fail(c, e)
		return
	}
	if owner != employeeID {
		c.Redirect(303, "/errors/404")
		return
	}
	u, _ := auth.CurrentUser(c)
	service := traineeassessment.NewService(traineeassessment.NewRepository(h.db))
	if e = service.Delete(c, assessmentID, uid, u.RoleName); e != nil {
		h.fail(c, e)
		return
	}
	c.Redirect(303, "/reports/employees/"+strconv.FormatUint(employeeID, 10)+"?success=Penilaian+berhasil+dihapus.")
}
