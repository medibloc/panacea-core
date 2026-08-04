package harness

import (
	"testing"
	"time"
)

func TestDockerCleanupDeadline(t *testing.T) {
	t.Parallel()

	if dockerCleanupTimeout != 45*time.Second {
		t.Fatalf("dockerCleanupTimeout = %s, want 45s", dockerCleanupTimeout)
	}
	if dockerOperationTimeout >= dockerCleanupTimeout {
		t.Fatalf(
			"dockerOperationTimeout = %s must be less than cleanup deadline %s",
			dockerOperationTimeout,
			dockerCleanupTimeout,
		)
	}
}
