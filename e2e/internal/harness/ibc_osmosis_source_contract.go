package harness

import (
	"errors"
	"fmt"
)

const (
	OsmosisPinnedSourceContractArtifactPath = "ibc/osmosis-source-contract.json"
	osmosisPinnedSourceContractSchema       = "panacea.osmosis-pinned-source-contract/v1"
	osmosisSourceRepository                 = "https://github.com/osmosis-labs/osmosis"
	osmosisGoModSHA256                      = "26d7ba123c5fa97d3725d3dd22e39954fa8770fd3497599232e0d9b97727ff19"
	osmosisTransferWiringSHA256             = "c53472e962f3b4ac1547a8940f9b524cb2c5719ff9fb367d6d747197fbd45282"
)

// IBCPinnedSourceFileContract identifies source bytes at an immutable commit.
// The hash is recorded as provenance; the E2E lane does not fetch GitHub or
// treat a network response as a runtime assertion.
type IBCPinnedSourceFileContract struct {
	Path      string `json:"path"`
	Reference string `json:"reference"`
	SHA256    string `json:"sha256"`
}

// OsmosisPinnedSourceContract records facts that are only available from the
// v31.0.2 source tree. In particular, live node-info does not expose Go module
// replacement metadata or the transfer middleware wiring/activation state.
type OsmosisPinnedSourceContract struct {
	SchemaVersion  string                      `json:"schema_version"`
	Repository     string                      `json:"repository"`
	ReleaseTag     string                      `json:"release_tag"`
	Commit         string                      `json:"commit"`
	GoMod          IBCPinnedSourceFileContract `json:"go_mod"`
	TransferWiring IBCPinnedSourceFileContract `json:"transfer_wiring"`
	SDKDependency  IBCDependencyContract       `json:"sdk_dependency"`
	RecvStack      []string                    `json:"recv_stack"`
	SendStack      []string                    `json:"send_stack"`
}

func pinnedOsmosisSourceContract() OsmosisPinnedSourceContract {
	return OsmosisPinnedSourceContract{
		SchemaVersion: osmosisPinnedSourceContractSchema,
		Repository:    osmosisSourceRepository,
		ReleaseTag:    PinnedIBCProvenance().Osmosis.Tag,
		Commit:        osmosisSourceCommit,
		GoMod: IBCPinnedSourceFileContract{
			Path:      "go.mod",
			Reference: osmosisRawSourceReference("go.mod"),
			SHA256:    osmosisGoModSHA256,
		},
		TransferWiring: IBCPinnedSourceFileContract{
			Path:      "app/keepers/keepers.go",
			Reference: osmosisRawSourceReference("app/keepers/keepers.go"),
			SHA256:    osmosisTransferWiringSHA256,
		},
		SDKDependency: IBCDependencyContract{
			Path:    cosmosSDKModulePath,
			Version: "v0.50.14",
			Replacement: &IBCDependencyReplacement{
				Path:    osmosisSDKModulePath,
				Version: "v0.50.14-v30-osmo",
			},
		},
		RecvStack: []string{"ibc-hooks", "ibc-rate-limit", "packet-forward-middleware", "ics20-transfer"},
		SendStack: []string{"ics20-transfer", "ibc-rate-limit", "ibc-hooks", "ibc-channel"},
	}
}

func osmosisRawSourceReference(path string) string {
	return "https://raw.githubusercontent.com/osmosis-labs/osmosis/" + osmosisSourceCommit + "/" + path
}

func (c OsmosisPinnedSourceContract) Validate() error {
	want := pinnedOsmosisSourceContract()
	var validationErrors []error
	if c.SchemaVersion != want.SchemaVersion || c.Repository != want.Repository ||
		c.ReleaseTag != want.ReleaseTag || c.Commit != want.Commit {
		validationErrors = append(validationErrors, errors.New("Osmosis source contract is not pinned to the v31.0.2 release commit"))
	}
	if err := validatePinnedSourceFile(c.GoMod, want.GoMod); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("Osmosis go.mod: %w", err))
	}
	if err := validatePinnedSourceFile(c.TransferWiring, want.TransferWiring); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("Osmosis transfer wiring: %w", err))
	}
	if !dependencyContractEqual(c.SDKDependency, want.SDKDependency) {
		validationErrors = append(validationErrors, fmt.Errorf(
			"Osmosis SDK dependency = %s, want %s",
			formatDependencyContract(c.SDKDependency, true),
			formatDependencyContract(want.SDKDependency, true),
		))
	}
	if !stringSlicesEqual(c.RecvStack, want.RecvStack) || !stringSlicesEqual(c.SendStack, want.SendStack) {
		validationErrors = append(validationErrors, errors.New("Osmosis transfer middleware order does not match the pinned source contract"))
	}
	return errors.Join(validationErrors...)
}

func validatePinnedSourceFile(got, want IBCPinnedSourceFileContract) error {
	if got.Path != want.Path || got.Reference != want.Reference || got.SHA256 != want.SHA256 {
		return errors.New("path, immutable reference, or SHA-256 does not match")
	}
	if err := validateSHA256(got.SHA256); err != nil {
		return err
	}
	return nil
}
