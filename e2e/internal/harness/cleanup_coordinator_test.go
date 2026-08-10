package harness

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestCleanupSequenceRunsEveryStageAfterErrorsAndPanics(t *testing.T) {
	var events []string
	collectErr := errors.New("artifact sink unavailable")
	finalizeErr := errors.New("manifest write failed")

	sequence := cleanupSequence{
		closeInterchain: func() error {
			events = append(events, "close")
			panic("close panic")
		},
		collectArtifacts: func(failed bool) error {
			events = append(events, "collect")
			if !failed {
				t.Fatal("close panic was not propagated to artifact failure state")
			}
			return collectErr
		},
		cleanupDocker: func() error {
			events = append(events, "docker")
			panic("docker panic")
		},
		finalizeArtifacts: func(closeErr, gotCollectErr, dockerErr error) error {
			events = append(events, "finalize")
			if closeErr == nil || !strings.Contains(closeErr.Error(), "close panic") {
				t.Fatalf("close error = %v", closeErr)
			}
			if !errors.Is(gotCollectErr, collectErr) {
				t.Fatalf("collect error = %v", gotCollectErr)
			}
			if dockerErr == nil || !strings.Contains(dockerErr.Error(), "docker panic") {
				t.Fatalf("docker error = %v", dockerErr)
			}
			return finalizeErr
		},
	}

	err := sequence.run(false)
	if !reflect.DeepEqual(events, []string{"close", "collect", "docker", "finalize"}) {
		t.Fatalf("cleanup order = %#v", events)
	}
	for _, fragment := range []string{"close panic", collectErr.Error(), "docker panic", finalizeErr.Error()} {
		if err == nil || !strings.Contains(err.Error(), fragment) {
			t.Fatalf("cleanup error %q does not contain %q", err, fragment)
		}
	}
}

func TestCleanupSequenceKeepsSuccessfulCollectionNonFailure(t *testing.T) {
	collectFailed := true
	sequence := cleanupSequence{
		closeInterchain: func() error { return nil },
		collectArtifacts: func(failed bool) error {
			collectFailed = failed
			return nil
		},
		cleanupDocker:     func() error { return nil },
		finalizeArtifacts: func(error, error, error) error { return nil },
	}
	if err := sequence.run(false); err != nil {
		t.Fatalf("cleanup sequence: %v", err)
	}
	if collectFailed {
		t.Fatal("successful test was reported as failed to artifact collection")
	}
}
