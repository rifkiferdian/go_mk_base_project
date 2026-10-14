package router

import (
	"context"
	"database/sql"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"mksiaga/internal/access"
	"mksiaga/internal/assessmentaspect"
	"mksiaga/internal/assessmentcriterion"
	"mksiaga/internal/assessmentreport"
	"mksiaga/internal/assessmentrule"
	"mksiaga/internal/auth"
	"mksiaga/internal/config"
	"mksiaga/internal/dashboard"
	"mksiaga/internal/employee"
	"mksiaga/internal/httperror"
	"mksiaga/internal/navigation"
	"mksiaga/internal/organization"
	"mksiaga/internal/profile"
	"mksiaga/internal/sessionmanagement"
	storeManagement "mksiaga/internal/store"
	"mksiaga/internal/traineeassessment"
	userManagement "mksiaga/internal/user"
	"mksiaga/internal/view"
)

func New(cfg config.Config, db *sql.DB, webRoot string) (*gin.Engine, error) {
	renderer, err := view.New(filepath.Join(webRoot, "templates"))
	if err != nil {
		return nil, err
	}
	gin.SetMode(cfg.GinMode)
	r := gin.New()
	r.Use(gin.Logger(), httperror.Recovery(cfg.AppName))
	store := cookie.NewStore([]byte(cfg.Session.Secret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   12 * 60 * 60,
		HttpOnly: true,
		Secure:   cfg.Session.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions(auth.SessionName, store))
	if err := r.SetTrustedProxies(nil); err != nil {
		return nil, err
	}
	r.HTMLRender = renderer
	r.Static("/static", filepath.Join(webRoot, "static"))
	fontAwesomeRoot := filepath.Join(filepath.Dir(webRoot), "node_modules", "@fortawesome", "fontawesome-free")
	r.StaticFile("/vendor/fontawesome/css/all.min.css", filepath.Join(fontAwesomeRoot, "css", "all.min.css"))
	r.Static("/vendor/fontawesome/webfonts", filepath.Join(fontAwesomeRoot, "webfonts"))
	sweetAlertRoot := filepath.Join(filepath.Dir(webRoot), "node_modules", "sweetalert2", "dist")
	r.StaticFile("/vendor/sweetalert2/sweetalert2.all.min.js", filepath.Join(sweetAlertRoot, "sweetalert2.all.min.js"))
	r.GET("/errors/403", httperror.Handler(cfg.AppName, "403"))
	r.GET("/errors/404", httperror.Handler(cfg.AppName, "404"))
	r.GET("/errors/419", httperror.Handler(cfg.AppName, "419"))
	r.GET("/errors/500", httperror.Handler(cfg.AppName, "500"))
	r.GET("/errors/maintenance", httperror.Handler(cfg.AppName, "maintenance"))
	if cfg.Maintenance.Enabled {
		r.Use(httperror.Maintenance(cfg.AppName, cfg.Maintenance.Message))
	}
	r.Use(auth.RequireCSRF())
	var authService *auth.Service
	if db != nil {
		authService = auth.NewService(auth.NewRepository(db))
	}
	authHandler := auth.NewHandler(cfg.AppName, authService, cfg.Session.CookieSecure)
	r.GET("/login", authHandler.ShowLogin)
	r.POST("/login", authHandler.Login)
	authorized := r.Group("/")
	authorized.Use(auth.RequireLogin(db))
	var navigationService *navigation.Service
	if db != nil {
		navigationService = navigation.NewService(db)
	}
	authorized.GET("/", dashboard.NewHandler(cfg.AppName, navigationService).Index)
	authorized.POST("/logout", authHandler.Logout)
	if db != nil {
		accessHandler := access.NewHandler(cfg.AppName, access.NewService(access.NewRepository(db)), navigationService)
		assessmentAspectHandler := assessmentaspect.NewHandler(cfg.AppName, assessmentaspect.NewService(assessmentaspect.NewRepository(db)), navigationService)
		assessmentCriterionHandler := assessmentcriterion.NewHandler(cfg.AppName, assessmentcriterion.NewService(assessmentcriterion.NewRepository(db)), navigationService)
		assessmentRuleHandler := assessmentrule.NewHandler(cfg.AppName, assessmentrule.NewService(assessmentrule.NewRepository(db)), navigationService)
		organizationHandler := organization.NewHandler(cfg.AppName, organization.NewService(organization.NewRepository(db)), navigationService)
		employeeHandler := employee.NewHandler(cfg.AppName, employee.NewService(employee.NewRepository(db)), navigationService)
		assessmentHandler := traineeassessment.NewHandler(cfg.AppName, traineeassessment.NewService(traineeassessment.NewRepository(db)), navigationService)
		assessmentReportHandler := assessmentreport.NewHandler(cfg.AppName, assessmentreport.NewService(assessmentreport.NewRepository(db)), navigationService)
		profileHandler := profile.NewHandler(cfg.AppName, profile.NewService(profile.NewRepository(db)), navigationService)
		storeHandler := storeManagement.NewHandler(cfg.AppName, storeManagement.NewService(storeManagement.NewRepository(db)), navigationService)
		sessionHandler := sessionmanagement.NewHandler(cfg.AppName, sessionmanagement.NewService(sessionmanagement.NewRepository(db)), navigationService)
		userHandler := userManagement.NewHandler(cfg.AppName, userManagement.NewService(userManagement.NewRepository(db)), navigationService)
		authorized.GET("/profile", profileHandler.Index)
		authorized.POST("/profile/name", profileHandler.UpdateName)
		authorized.POST("/profile/password", profileHandler.UpdatePassword)
		authorized.GET("/assessments", auth.RequirePermission(db, "trainee_assessments.view"), assessmentHandler.Index)
		authorized.GET("/assessments/new", auth.RequirePermission(db, "trainee_assessments.create"), assessmentHandler.New)
		authorized.POST("/assessments", auth.RequirePermission(db, "trainee_assessments.create"), assessmentHandler.Create)
		authorized.GET("/assessments/:id/edit", auth.RequirePermission(db, "trainee_assessments.update"), assessmentHandler.Edit)
		authorized.POST("/assessments/:id/update", auth.RequirePermission(db, "trainee_assessments.update"), assessmentHandler.Update)
		authorized.POST("/assessments/:id/delete", auth.RequirePermission(db, "trainee_assessments.delete"), assessmentHandler.Delete)
		authorized.GET("/reports/trainee-assessments", auth.RequirePermission(db, "trainee_assessment_reports.view"), assessmentReportHandler.Index)
		authorized.GET("/profile/sessions", sessionHandler.Index)
		authorized.POST("/profile/sessions/revoke-others", sessionHandler.RevokeOthers)
		authorized.POST("/profile/sessions/:id/revoke", sessionHandler.Revoke)
		settings := authorized.Group("/settings")
		settings.GET("/users", auth.RequirePermission(db, "users.view"), userHandler.Index)
		settings.POST("/users", auth.RequirePermission(db, "users.create"), userHandler.Create)
		settings.POST("/users/:id/update", auth.RequirePermission(db, "users.update"), userHandler.Update)
		settings.POST("/users/:id/delete", auth.RequirePermission(db, "users.delete"), userHandler.Delete)
		settings.GET("/users/:id/stores", auth.RequirePermission(db, "users.update"), userHandler.StoreAssignments)
		settings.POST("/users/:id/stores", auth.RequirePermission(db, "users.update"), userHandler.UpdateStoreAssignments)
		settings.GET("/roles", auth.RequirePermission(db, "roles.view"), accessHandler.Roles)
		settings.POST("/roles", auth.RequirePermission(db, "roles.create"), accessHandler.CreateRole)
		settings.POST("/roles/:id/update", auth.RequirePermission(db, "roles.update"), accessHandler.UpdateRole)
		settings.POST("/roles/:id/delete", auth.RequirePermission(db, "roles.delete"), accessHandler.DeleteRole)
		settings.GET("/roles/:id/permissions", auth.RequirePermission(db, "roles.update"), accessHandler.RolePermissions)
		settings.POST("/roles/:id/permissions", auth.RequirePermission(db, "roles.update"), accessHandler.UpdateRolePermissions)
		settings.GET("/permissions", auth.RequirePermission(db, "permissions.view"), accessHandler.Permissions)
		settings.POST("/permissions", auth.RequirePermission(db, "permissions.create"), accessHandler.CreatePermission)
		settings.POST("/permissions/:id/update", auth.RequirePermission(db, "permissions.update"), accessHandler.UpdatePermission)
		settings.POST("/permissions/:id/delete", auth.RequirePermission(db, "permissions.delete"), accessHandler.DeletePermission)
		master := authorized.Group("/master")
		master.GET("/stores", auth.RequirePermission(db, "stores.view"), storeHandler.Index)
		master.POST("/stores", auth.RequirePermission(db, "stores.create"), storeHandler.Create)
		master.POST("/stores/:id/update", auth.RequirePermission(db, "stores.update"), storeHandler.Update)
		master.POST("/stores/:id/delete", auth.RequirePermission(db, "stores.delete"), storeHandler.Delete)
		master.GET("/assessment-aspects", auth.RequirePermission(db, "assessment_aspects.view"), assessmentAspectHandler.Index)
		master.POST("/assessment-aspects", auth.RequirePermission(db, "assessment_aspects.create"), assessmentAspectHandler.Create)
		master.POST("/assessment-aspects/:id/update", auth.RequirePermission(db, "assessment_aspects.update"), assessmentAspectHandler.Update)
		master.POST("/assessment-aspects/:id/delete", auth.RequirePermission(db, "assessment_aspects.delete"), assessmentAspectHandler.Delete)
		master.GET("/assessment-criteria", auth.RequirePermission(db, "assessment_criteria.view"), assessmentCriterionHandler.Index)
		master.POST("/assessment-criteria", auth.RequirePermission(db, "assessment_criteria.create"), assessmentCriterionHandler.Create)
		master.POST("/assessment-criteria/:id/update", auth.RequirePermission(db, "assessment_criteria.update"), assessmentCriterionHandler.Update)
		master.POST("/assessment-criteria/:id/delete", auth.RequirePermission(db, "assessment_criteria.delete"), assessmentCriterionHandler.Delete)
		master.GET("/assessment-score-rules", auth.RequirePermission(db, "assessment_score_rules.view"), assessmentRuleHandler.Index)
		master.POST("/assessment-score-rules", auth.RequirePermission(db, "assessment_score_rules.create"), assessmentRuleHandler.Create)
		master.POST("/assessment-score-rules/:criterionID/:score/update", auth.RequirePermission(db, "assessment_score_rules.update"), assessmentRuleHandler.Update)
		master.POST("/assessment-score-rules/:criterionID/:score/delete", auth.RequirePermission(db, "assessment_score_rules.delete"), assessmentRuleHandler.Delete)
		master.GET("/job-levels", auth.RequirePermission(db, "job_levels.view"), organizationHandler.JobLevels)
		master.POST("/job-levels", auth.RequirePermission(db, "job_levels.create"), organizationHandler.CreateJobLevel)
		master.POST("/job-levels/:id/update", auth.RequirePermission(db, "job_levels.update"), organizationHandler.UpdateJobLevel)
		master.POST("/job-levels/:id/delete", auth.RequirePermission(db, "job_levels.delete"), organizationHandler.DeleteJobLevel)
		master.GET("/positions", auth.RequirePermission(db, "positions.view"), organizationHandler.Positions)
		master.POST("/positions", auth.RequirePermission(db, "positions.create"), organizationHandler.CreatePosition)
		master.POST("/positions/:id/update", auth.RequirePermission(db, "positions.update"), organizationHandler.UpdatePosition)
		master.POST("/positions/:id/delete", auth.RequirePermission(db, "positions.delete"), organizationHandler.DeletePosition)
		master.GET("/departments", auth.RequirePermission(db, "departments.view"), organizationHandler.Departments)
		master.POST("/departments", auth.RequirePermission(db, "departments.create"), organizationHandler.CreateDepartment)
		master.POST("/departments/:id/update", auth.RequirePermission(db, "departments.update"), organizationHandler.UpdateDepartment)
		master.POST("/departments/:id/delete", auth.RequirePermission(db, "departments.delete"), organizationHandler.DeleteDepartment)
		master.GET("/employees", auth.RequirePermission(db, "employees.view"), employeeHandler.Employees)
		master.POST("/employees", auth.RequirePermission(db, "employees.create"), employeeHandler.CreateEmployee)
		master.POST("/employees/:id/update", auth.RequirePermission(db, "employees.update"), employeeHandler.UpdateEmployee)
		master.POST("/employees/:id/delete", auth.RequirePermission(db, "employees.delete"), employeeHandler.DeleteEmployee)
		master.GET("/employee-profiles", auth.RequirePermission(db, "employee_profiles.view"), employeeHandler.Profiles)
		master.POST("/employee-profiles", auth.RequirePermission(db, "employee_profiles.create"), employeeHandler.CreateProfile)
		master.POST("/employee-profiles/:id/update", auth.RequirePermission(db, "employee_profiles.update"), employeeHandler.UpdateProfile)
		master.POST("/employee-profiles/:id/delete", auth.RequirePermission(db, "employee_profiles.delete"), employeeHandler.DeleteProfile)
		master.GET("/employee-assignments", auth.RequirePermission(db, "employee_assignments.view"), employeeHandler.Assignments)
		master.POST("/employee-assignments", auth.RequirePermission(db, "employee_assignments.create"), employeeHandler.CreateAssignment)
		master.POST("/employee-assignments/:id/update", auth.RequirePermission(db, "employee_assignments.update"), employeeHandler.UpdateAssignment)
		master.POST("/employee-assignments/:id/delete", auth.RequirePermission(db, "employee_assignments.delete"), employeeHandler.DeleteAssignment)
		master.GET("/employee-reporting-lines", auth.RequirePermission(db, "employee_reporting_lines.view"), employeeHandler.ReportingLines)
		master.POST("/employee-reporting-lines", auth.RequirePermission(db, "employee_reporting_lines.create"), employeeHandler.CreateReportingLine)
		master.POST("/employee-reporting-lines/:id/update", auth.RequirePermission(db, "employee_reporting_lines.update"), employeeHandler.UpdateReportingLine)
		master.POST("/employee-reporting-lines/:id/delete", auth.RequirePermission(db, "employee_reporting_lines.delete"), employeeHandler.DeleteReportingLine)
	}
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "database": "disabled"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	r.NoRoute(httperror.Handler(cfg.AppName, "404"))
	return r, nil
}
