package traineeassessment

import (
	"context"
	"errors"
	"testing"
)

func TestDeleteRejectsNonSuperadminBeforeDatabase(t *testing.T) {
	service := NewService(nil)
	for _, role := range []string{"", "supervisor", "admin", "superadmin_copy", "supervisor,admin"} {
		if err := service.Delete(context.Background(), 1, 1, role); !errors.Is(err, ErrInvalid) {
			t.Fatalf("role %q should be denied: %v", role, err)
		}
	}
}
