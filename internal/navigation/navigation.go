package navigation

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"mksiaga/internal/auth"
)

type Item struct {
	Key      string
	Label    string
	URL      string
	Active   bool
	Open     bool
	Children []Item
}

type Group struct {
	Label string
	Items []Item
}

type definition struct {
	group      string
	key        string
	label      string
	url        string
	permission string
	parent     string
}

var definitions = []definition{
	{group: "Menu utama", key: "dashboard", label: "Dashboard", url: "/"},
	{group: "Master Data", parent: "master", key: "stores", label: "Store", url: "/master/stores", permission: "stores.view"},
	{group: "Master Data", parent: "master", key: "job-levels", label: "Tingkatan Jabatan", url: "/master/job-levels", permission: "job_levels.view"},
	{group: "Master Data", parent: "master", key: "positions", label: "Jabatan", url: "/master/positions", permission: "positions.view"},
	{group: "Master Data", parent: "master", key: "departments", label: "Bagian", url: "/master/departments", permission: "departments.view"},
	{group: "Master Data", parent: "master", key: "employees", label: "Karyawan", url: "/master/employees", permission: "employees.view"},
	{group: "Master Data", parent: "master", key: "employee-profiles", label: "Biodata Karyawan", url: "/master/employee-profiles", permission: "employee_profiles.view"},
	{group: "Master Data", parent: "master", key: "employee-assignments", label: "Penempatan Karyawan", url: "/master/employee-assignments", permission: "employee_assignments.view"},
	{group: "Master Data", parent: "master", key: "employee-reporting-lines", label: "Atasan & Bawahan", url: "/master/employee-reporting-lines", permission: "employee_reporting_lines.view"},
	{group: "Pengaturan", parent: "assessment-settings", key: "assessment-aspects", label: "Aspek Penilaian", url: "/master/assessment-aspects", permission: "assessment_aspects.view"},
	{group: "Pengaturan", parent: "assessment-settings", key: "assessment-criteria", label: "Subaspek Penilaian", url: "/master/assessment-criteria", permission: "assessment_criteria.view"},
	{group: "Pengaturan", parent: "assessment-settings", key: "assessment-score-rules", label: "Aturan Skor", url: "/master/assessment-score-rules", permission: "assessment_score_rules.view"},
	{group: "Pengelolaan", parent: "access", key: "users", label: "User", url: "/settings/users", permission: "users.view"},
	{group: "Pengelolaan", parent: "access", key: "roles", label: "Role", url: "/settings/roles", permission: "roles.view"},
	{group: "Pengelolaan", parent: "access", key: "permissions", label: "Permission", url: "/settings/permissions", permission: "permissions.view"},
}

type Service struct{ db *sql.DB }

func NewService(db *sql.DB) *Service { return &Service{db: db} }

func (s *Service) Menus(ctx context.Context, user auth.SessionUser, activePage string) ([]Group, error) {
	allowed := make(map[string]bool)
	if hasRole(user.RoleName, "superadmin") {
		for _, item := range definitions {
			allowed[item.permission] = true
		}
	} else {
		userStoreID, err := strconv.ParseUint(user.UserStoreID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid user store session: %w", err)
		}
		rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT p.name
			FROM permissions p
			WHERE EXISTS(SELECT 1 FROM user_store_permissions usp WHERE usp.user_store_id=? AND usp.permission_id=p.id)
			   OR EXISTS(SELECT 1 FROM user_store_roles usr JOIN role_permissions rp ON rp.role_id=usr.role_id WHERE usr.user_store_id=? AND rp.permission_id=p.id)`, userStoreID, userStoreID)
		if err != nil {
			return nil, fmt.Errorf("load navigation permissions: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return nil, err
			}
			allowed[name] = true
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	groups := make([]Group, 0, 2)
	indexes := make(map[string]int)
	children := make(map[string][]Item)
	for _, definition := range definitions {
		if definition.permission != "" && !allowed[definition.permission] {
			continue
		}
		item := Item{Key: definition.key, Label: definition.label, URL: definition.url, Active: definition.key == activePage}
		if definition.parent != "" {
			children[definition.parent] = append(children[definition.parent], item)
			continue
		}
		index, exists := indexes[definition.group]
		if !exists {
			index = len(groups)
			indexes[definition.group] = index
			groups = append(groups, Group{Label: definition.group})
		}
		groups[index].Items = append(groups[index].Items, item)
	}
	if accessItems := children["access"]; len(accessItems) > 0 {
		index, exists := indexes["Pengelolaan"]
		if !exists {
			index = len(groups)
			indexes["Pengelolaan"] = index
			groups = append(groups, Group{Label: "Pengelolaan"})
		}
		open := false
		for _, item := range accessItems {
			open = open || item.Active
		}
		groups[index].Items = append(groups[index].Items, Item{Key: "access", Label: "Manajemen Akses", Open: open, Active: open, Children: accessItems})
	}
	if masterItems := children["master"]; len(masterItems) > 0 {
		index, exists := indexes["Master Data"]
		if !exists {
			index = len(groups)
			indexes["Master Data"] = index
			groups = append(groups, Group{Label: "Master Data"})
		}
		open := false
		for _, item := range masterItems {
			open = open || item.Active
		}
		groups[index].Items = append(groups[index].Items, Item{Key: "master", Label: "Master Data", Open: open, Active: open, Children: masterItems})
	}
	if assessmentItems := children["assessment-settings"]; len(assessmentItems) > 0 {
		index, exists := indexes["Pengaturan"]
		if !exists {
			index = len(groups)
			indexes["Pengaturan"] = index
			groups = append(groups, Group{Label: "Pengaturan"})
		}
		open := false
		for _, item := range assessmentItems {
			open = open || item.Active
		}
		groups[index].Items = append(groups[index].Items, Item{Key: "assessment-settings", Label: "Pengaturan Penilaian", Open: open, Active: open, Children: assessmentItems})
	}
	return groups, nil
}

func hasRole(roleNames, required string) bool {
	for _, role := range strings.Split(roleNames, ",") {
		if strings.TrimSpace(role) == required {
			return true
		}
	}
	return false
}
