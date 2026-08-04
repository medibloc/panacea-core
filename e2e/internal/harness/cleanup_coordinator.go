package harness

import (
	"errors"
	"fmt"
)

// cleanupSequence guarantees that every cleanup stage gets a chance to run.
// A panic in diagnostics must never prevent the label-scoped Docker removal.
type cleanupSequence struct {
	closeInterchain   func() error
	collectArtifacts  func(failed bool) error
	cleanupDocker     func() error
	finalizeArtifacts func(closeErr, collectErr, dockerErr error) error
}

func (s cleanupSequence) run(testFailed bool) error {
	closeErr := runCleanupStep("interchain close", s.closeInterchain)
	collectErr := runCleanupStep("artifact collection", func() error {
		if s.collectArtifacts == nil {
			return nil
		}
		return s.collectArtifacts(testFailed || closeErr != nil)
	})
	dockerErr := runCleanupStep("Docker cleanup", s.cleanupDocker)
	finalizeErr := runCleanupStep("artifact finalization", func() error {
		if s.finalizeArtifacts == nil {
			return nil
		}
		return s.finalizeArtifacts(closeErr, collectErr, dockerErr)
	})
	return errors.Join(closeErr, collectErr, dockerErr, finalizeErr)
}

func runCleanupStep(name string, step func() error) (err error) {
	if step == nil {
		return nil
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%s panic: %v", name, recovered)
		}
	}()
	return step()
}
