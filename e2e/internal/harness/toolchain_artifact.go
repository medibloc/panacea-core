package harness

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	GoToolchainEvidenceSchemaVersion = "1"
	GoToolchainArtifactPath          = "environment/toolchain.json"
)

// GoToolchainEvidence records the concrete host-side Go command and tools used
// to compile and run the E2E harness. Effective Go values are read through the
// selected executable rather than inferred from process environment variables.
type GoToolchainEvidence struct {
	SchemaVersion        string    `json:"schema_version"`
	RecordedAt           time.Time `json:"recorded_at"`
	SelectedGoExecutable string    `json:"selected_go_executable"`
	GoVersionOutput      string    `json:"go_version_output"`
	GOVersion            string    `json:"goversion"`
	GORoot               string    `json:"goroot"`
	GoToolDir            string    `json:"gotooldir"`
	CompilerVersion      string    `json:"compiler_version"`
	GoBin                string    `json:"gobin"`
	GoOS                 string    `json:"goos"`
	GoArch               string    `json:"goarch"`
	HostOS               string    `json:"host_os"`
	HostArch             string    `json:"host_arch"`
}

type goToolchainEnv struct {
	GOVersion string `json:"GOVERSION"`
	GORoot    string `json:"GOROOT"`
	GoToolDir string `json:"GOTOOLDIR"`
	GoBin     string `json:"GOBIN"`
	GoOS      string `json:"GOOS"`
	GoArch    string `json:"GOARCH"`
}

type goToolchainProbe struct {
	lookPath func(string) (string, error)
	run      func(context.Context, string, ...string) ([]byte, error)
	now      func() time.Time
	hostOS   string
	hostArch string
}

// CaptureGoToolchainEvidence probes the Go command selected by PATH and
// GOTOOLCHAIN. It also invokes that command's compiler so a mixed GOROOT can be
// diagnosed from the resulting artifact.
func CaptureGoToolchainEvidence(ctx context.Context) (GoToolchainEvidence, error) {
	return captureGoToolchainEvidence(ctx, goToolchainProbe{
		lookPath: exec.LookPath,
		run:      runGoToolchainCommand,
		now:      time.Now,
		hostOS:   runtime.GOOS,
		hostArch: runtime.GOARCH,
	})
}

func captureGoToolchainEvidence(ctx context.Context, probe goToolchainProbe) (GoToolchainEvidence, error) {
	if ctx == nil {
		return GoToolchainEvidence{}, errors.New("toolchain evidence context is required")
	}
	if probe.lookPath == nil || probe.run == nil || probe.now == nil {
		return GoToolchainEvidence{}, errors.New("toolchain evidence probe is incomplete")
	}

	executable, err := probe.lookPath("go")
	if err != nil {
		return GoToolchainEvidence{}, fmt.Errorf("locate selected go executable: %w", err)
	}
	executable, err = filepath.Abs(executable)
	if err != nil {
		return GoToolchainEvidence{}, fmt.Errorf("resolve selected go executable: %w", err)
	}
	executable = filepath.Clean(executable)

	versionOutput, err := probe.run(ctx, executable, "version")
	if err != nil {
		return GoToolchainEvidence{}, fmt.Errorf("read selected go executable version: %w", err)
	}
	environmentOutput, err := probe.run(
		ctx,
		executable,
		"env", "-json", "GOVERSION", "GOROOT", "GOTOOLDIR", "GOBIN", "GOOS", "GOARCH",
	)
	if err != nil {
		return GoToolchainEvidence{}, fmt.Errorf("read selected go environment: %w", err)
	}
	var environment goToolchainEnv
	if err := json.Unmarshal(environmentOutput, &environment); err != nil {
		return GoToolchainEvidence{}, fmt.Errorf("decode selected go environment: %w", err)
	}
	compilerOutput, err := probe.run(ctx, executable, "tool", "compile", "-V=full")
	if err != nil {
		return GoToolchainEvidence{}, fmt.Errorf("read selected go compiler version: %w", err)
	}

	evidence := GoToolchainEvidence{
		SchemaVersion:        GoToolchainEvidenceSchemaVersion,
		RecordedAt:           probe.now().UTC(),
		SelectedGoExecutable: executable,
		GoVersionOutput:      strings.TrimSpace(string(versionOutput)),
		GOVersion:            strings.TrimSpace(environment.GOVersion),
		GORoot:               strings.TrimSpace(environment.GORoot),
		GoToolDir:            strings.TrimSpace(environment.GoToolDir),
		CompilerVersion:      strings.TrimSpace(string(compilerOutput)),
		GoBin:                strings.TrimSpace(environment.GoBin),
		GoOS:                 strings.TrimSpace(environment.GoOS),
		GoArch:               strings.TrimSpace(environment.GoArch),
		HostOS:               strings.TrimSpace(probe.hostOS),
		HostArch:             strings.TrimSpace(probe.hostArch),
	}
	if err := evidence.Validate(); err != nil {
		return GoToolchainEvidence{}, fmt.Errorf("validate selected go toolchain evidence: %w", err)
	}
	return evidence, nil
}

func runGoToolchainCommand(ctx context.Context, executable string, arguments ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, executable, arguments...)
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w: %s", executable, strings.Join(arguments, " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

// Validate checks that the evidence contains every value needed to diagnose a
// host executable, environment, compiler, or platform mismatch. GOBIN may be
// empty because that is a valid effective Go configuration.
func (e GoToolchainEvidence) Validate() error {
	if e.SchemaVersion != GoToolchainEvidenceSchemaVersion {
		return fmt.Errorf("toolchain evidence schema version %q, want %q", e.SchemaVersion, GoToolchainEvidenceSchemaVersion)
	}
	if e.RecordedAt.IsZero() {
		return errors.New("toolchain evidence recorded_at is required")
	}
	fields := map[string]string{
		"selected_go_executable": e.SelectedGoExecutable,
		"go_version_output":      e.GoVersionOutput,
		"goversion":              e.GOVersion,
		"goroot":                 e.GORoot,
		"gotooldir":              e.GoToolDir,
		"compiler_version":       e.CompilerVersion,
		"goos":                   e.GoOS,
		"goarch":                 e.GoArch,
		"host_os":                e.HostOS,
		"host_arch":              e.HostArch,
	}
	for name, value := range fields {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("toolchain evidence %s is required", name)
		}
	}
	if !filepath.IsAbs(e.SelectedGoExecutable) {
		return fmt.Errorf("selected go executable must be absolute: %q", e.SelectedGoExecutable)
	}
	return nil
}

// RecordGoToolchainEvidence writes already captured evidence to the stable E2E
// artifact path. Separating probing from recording also lets preflight callers
// retain diagnostics when no Docker network exists yet.
func (n *Network) RecordGoToolchainEvidence(evidence GoToolchainEvidence) error {
	if n == nil || n.artifacts == nil {
		return errors.New("toolchain evidence artifact store is unavailable")
	}
	if err := evidence.Validate(); err != nil {
		return fmt.Errorf("validate toolchain evidence: %w", err)
	}
	if err := n.artifacts.writeJSON(GoToolchainArtifactPath, evidence); err != nil {
		return fmt.Errorf("record toolchain evidence: %w", err)
	}
	return nil
}

// CaptureAndRecordGoToolchainEvidence is the one-call runtime helper used by an
// E2E suite after its artifact store has been initialized.
func (n *Network) CaptureAndRecordGoToolchainEvidence(ctx context.Context) (GoToolchainEvidence, error) {
	evidence, err := CaptureGoToolchainEvidence(ctx)
	if err != nil {
		return GoToolchainEvidence{}, err
	}
	if err := n.RecordGoToolchainEvidence(evidence); err != nil {
		return evidence, err
	}
	return evidence, nil
}
