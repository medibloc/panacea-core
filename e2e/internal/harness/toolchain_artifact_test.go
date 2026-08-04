package harness

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCaptureGoToolchainEvidence(t *testing.T) {
	recordedAt := time.Date(2026, time.August, 4, 1, 2, 3, 0, time.FixedZone("fixture", 9*60*60))
	executable := filepath.Join(string(filepath.Separator), "opt", "go1.23.12", "bin", "go")
	responses := map[string][]byte{
		"version": []byte("go version go1.23.12 linux/amd64\n"),
		"env": []byte(`{
  "GOVERSION": "go1.23.12",
  "GOROOT": "/opt/go1.23.12",
  "GOTOOLDIR": "/opt/go1.23.12/pkg/tool/linux_amd64",
  "GOBIN": "/work/bin",
  "GOOS": "linux",
  "GOARCH": "amd64"
}`),
		"compiler": []byte("compile version go1.23.12\n"),
	}
	var invocations [][]string
	probe := goToolchainProbe{
		lookPath: func(name string) (string, error) {
			require.Equal(t, "go", name)
			return executable, nil
		},
		run: func(_ context.Context, gotExecutable string, arguments ...string) ([]byte, error) {
			require.Equal(t, executable, gotExecutable)
			invocations = append(invocations, append([]string(nil), arguments...))
			switch {
			case len(arguments) == 1 && arguments[0] == "version":
				return responses["version"], nil
			case len(arguments) > 1 && arguments[0] == "env":
				return responses["env"], nil
			case len(arguments) == 3 && arguments[0] == "tool" && arguments[1] == "compile":
				return responses["compiler"], nil
			default:
				return nil, errors.New("unexpected go invocation")
			}
		},
		now:      func() time.Time { return recordedAt },
		hostOS:   "darwin",
		hostArch: "arm64",
	}

	evidence, err := captureGoToolchainEvidence(context.Background(), probe)
	require.NoError(t, err)
	require.Equal(t, GoToolchainEvidenceSchemaVersion, evidence.SchemaVersion)
	require.Equal(t, recordedAt.UTC(), evidence.RecordedAt)
	require.Equal(t, executable, evidence.SelectedGoExecutable)
	require.Equal(t, "go version go1.23.12 linux/amd64", evidence.GoVersionOutput)
	require.Equal(t, "go1.23.12", evidence.GOVersion)
	require.Equal(t, "/opt/go1.23.12", evidence.GORoot)
	require.Equal(t, "/opt/go1.23.12/pkg/tool/linux_amd64", evidence.GoToolDir)
	require.Equal(t, "compile version go1.23.12", evidence.CompilerVersion)
	require.Equal(t, "/work/bin", evidence.GoBin)
	require.Equal(t, "linux", evidence.GoOS)
	require.Equal(t, "amd64", evidence.GoArch)
	require.Equal(t, "darwin", evidence.HostOS)
	require.Equal(t, "arm64", evidence.HostArch)
	require.Len(t, invocations, 3)
}

func TestCaptureGoToolchainEvidenceReportsProbeFailures(t *testing.T) {
	executable := filepath.Join(string(filepath.Separator), "opt", "go", "bin", "go")
	base := goToolchainProbe{
		lookPath: func(string) (string, error) { return executable, nil },
		run: func(_ context.Context, _ string, arguments ...string) ([]byte, error) {
			switch arguments[0] {
			case "version":
				return []byte("go version go1.23.12 linux/amd64"), nil
			case "env":
				return []byte(`{"GOVERSION":"go1.23.12","GOROOT":"/opt/go","GOTOOLDIR":"/opt/go/pkg/tool/linux_amd64","GOOS":"linux","GOARCH":"amd64"}`), nil
			case "tool":
				return []byte("compile version go1.23.12"), nil
			default:
				return nil, errors.New("unexpected invocation")
			}
		},
		now:      time.Now,
		hostOS:   "linux",
		hostArch: "amd64",
	}

	tests := []struct {
		name      string
		mutate    func(*goToolchainProbe)
		wantError string
	}{
		{
			name: "go executable is absent",
			mutate: func(probe *goToolchainProbe) {
				probe.lookPath = func(string) (string, error) { return "", errors.New("not found") }
			},
			wantError: "locate selected go executable",
		},
		{
			name: "go version command fails",
			mutate: func(probe *goToolchainProbe) {
				probe.run = func(context.Context, string, ...string) ([]byte, error) { return nil, errors.New("version failed") }
			},
			wantError: "read selected go executable version",
		},
		{
			name: "go environment is invalid JSON",
			mutate: func(probe *goToolchainProbe) {
				original := probe.run
				probe.run = func(ctx context.Context, executable string, arguments ...string) ([]byte, error) {
					if arguments[0] == "env" {
						return []byte("{"), nil
					}
					return original(ctx, executable, arguments...)
				}
			},
			wantError: "decode selected go environment",
		},
		{
			name: "compiler version is absent",
			mutate: func(probe *goToolchainProbe) {
				original := probe.run
				probe.run = func(ctx context.Context, executable string, arguments ...string) ([]byte, error) {
					if arguments[0] == "tool" {
						return []byte("  "), nil
					}
					return original(ctx, executable, arguments...)
				}
			},
			wantError: "compiler_version is required",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			probe := base
			test.mutate(&probe)
			_, err := captureGoToolchainEvidence(context.Background(), probe)
			require.ErrorContains(t, err, test.wantError)
		})
	}
}

func TestRecordGoToolchainEvidence(t *testing.T) {
	store, err := newArtifactStore(
		"toolchain-artifact",
		"run-toolchain",
		artifactTestConfig(filepath.Join(trustedArtifactTempDir(t), "artifacts")),
	)
	require.NoError(t, err)
	network := &Network{artifacts: store}
	evidence := validGoToolchainEvidence()

	require.NoError(t, network.RecordGoToolchainEvidence(evidence))
	contents, err := os.ReadFile(filepath.Join(store.dir, filepath.FromSlash(GoToolchainArtifactPath)))
	require.NoError(t, err)
	var recorded GoToolchainEvidence
	require.NoError(t, json.Unmarshal(contents, &recorded))
	require.Equal(t, evidence, recorded)
}

func TestRecordGoToolchainEvidenceRejectsInvalidEvidence(t *testing.T) {
	store, err := newArtifactStore(
		"toolchain-artifact-invalid",
		"run-toolchain-invalid",
		artifactTestConfig(filepath.Join(trustedArtifactTempDir(t), "artifacts")),
	)
	require.NoError(t, err)
	network := &Network{artifacts: store}
	evidence := validGoToolchainEvidence()
	evidence.CompilerVersion = ""

	err = network.RecordGoToolchainEvidence(evidence)
	require.ErrorContains(t, err, "compiler_version is required")
	_, statErr := os.Stat(filepath.Join(store.dir, filepath.FromSlash(GoToolchainArtifactPath)))
	require.ErrorIs(t, statErr, os.ErrNotExist)
}

func validGoToolchainEvidence() GoToolchainEvidence {
	return GoToolchainEvidence{
		SchemaVersion:        GoToolchainEvidenceSchemaVersion,
		RecordedAt:           time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC),
		SelectedGoExecutable: filepath.Join(string(filepath.Separator), "opt", "go", "bin", "go"),
		GoVersionOutput:      "go version go1.23.12 linux/amd64",
		GOVersion:            "go1.23.12",
		GORoot:               "/opt/go",
		GoToolDir:            "/opt/go/pkg/tool/linux_amd64",
		CompilerVersion:      "compile version go1.23.12",
		GoOS:                 "linux",
		GoArch:               "amd64",
		HostOS:               "linux",
		HostArch:             "amd64",
	}
}
