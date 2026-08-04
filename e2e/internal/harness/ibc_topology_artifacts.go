package harness

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	dockerclient "github.com/docker/docker/client"
	"github.com/pelletier/go-toml"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/relayer/hermes"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
)

// IBCTopologyDescriptor is the non-secret topology identity persisted before
// Docker build begins. It makes setup failures attributable even when no node
// reached a queryable height.
type IBCTopologyDescriptor struct {
	Path              string `json:"path"`
	PanaceaChainID    string `json:"panacea_chain_id"`
	OsmosisChainID    string `json:"osmosis_chain_id"`
	PanaceaValidators int    `json:"panacea_validators"`
	PanaceaFullNodes  int    `json:"panacea_full_nodes"`
	OsmosisValidators int    `json:"osmosis_validators"`
	OsmosisFullNodes  int    `json:"osmosis_full_nodes"`
	SkipPathCreation  bool   `json:"skip_path_creation"`
}

type ibcTopologyArtifacts struct {
	base       *artifactStore
	descriptor IBCTopologyDescriptor
	reporter   *testreporter.Reporter

	closeOnce sync.Once
	closeErr  error
}

func newIBCTopologyArtifacts(base *artifactStore, descriptor IBCTopologyDescriptor) (*ibcTopologyArtifacts, error) {
	if base == nil {
		return nil, errors.New("IBC artifact base store is required")
	}
	if strings.TrimSpace(descriptor.Path) == "" {
		return nil, errors.New("IBC path name is required")
	}
	if strings.TrimSpace(descriptor.PanaceaChainID) == "" || strings.TrimSpace(descriptor.OsmosisChainID) == "" {
		return nil, errors.New("both IBC chain IDs are required")
	}
	if descriptor.PanaceaValidators < 1 || descriptor.OsmosisValidators < 1 {
		return nil, errors.New("both IBC chains require at least one validator")
	}
	if descriptor.PanaceaFullNodes < 0 || descriptor.OsmosisFullNodes < 0 {
		return nil, errors.New("IBC full-node counts cannot be negative")
	}

	if err := base.writeJSON("ibc/provenance.json", PinnedIBCProvenance()); err != nil {
		return nil, fmt.Errorf("write IBC provenance: %w", err)
	}
	sourceContract := pinnedOsmosisSourceContract()
	if err := sourceContract.Validate(); err != nil {
		return nil, fmt.Errorf("validate Osmosis pinned source contract: %w", err)
	}
	if err := base.writeJSON(OsmosisPinnedSourceContractArtifactPath, sourceContract); err != nil {
		return nil, fmt.Errorf("write Osmosis pinned source contract: %w", err)
	}
	if err := base.writeJSON("ibc/topology.json", descriptor); err != nil {
		return nil, fmt.Errorf("write IBC topology: %w", err)
	}

	reporterPath, err := base.safePath(filepath.Join("ibc", "hermes", "exec-report.jsonl"))
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(reporterPath), 0o700); err != nil {
		return nil, fmt.Errorf("create Hermes artifact directory: %w", err)
	}
	reporterFile, err := os.OpenFile(reporterPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("create Hermes execution reporter: %w", err)
	}

	return &ibcTopologyArtifacts{
		base:       base,
		descriptor: descriptor,
		reporter:   testreporter.NewReporter(reporterFile),
	}, nil
}

func (a *ibcTopologyArtifacts) closeReporter() error {
	if a == nil {
		return nil
	}
	a.closeOnce.Do(func() {
		if a.reporter != nil {
			a.closeErr = a.reporter.Close()
		}
	})
	return a.closeErr
}

func pinHermesCompatMode(config []byte, chainIDs []string) ([]byte, error) {
	tree, err := toml.LoadBytes(config)
	if err != nil {
		return nil, fmt.Errorf("parse generated Hermes config: %w", err)
	}
	chains, ok := tree.Get("chains").([]*toml.Tree)
	if !ok || len(chains) == 0 {
		return nil, errors.New("generated Hermes config has no chain tables")
	}

	expected := make(map[string]struct{}, len(chainIDs))
	for _, chainID := range chainIDs {
		chainID = strings.TrimSpace(chainID)
		if chainID == "" {
			return nil, errors.New("Hermes compatibility chain ID is required")
		}
		expected[chainID] = struct{}{}
	}
	for _, chain := range chains {
		chainID, ok := chain.Get("id").(string)
		if !ok || strings.TrimSpace(chainID) == "" {
			return nil, errors.New("generated Hermes chain table has no ID")
		}
		if _, ok := expected[chainID]; !ok {
			continue
		}
		chain.Set("compat_mode", "0.37")
		delete(expected, chainID)
	}
	if len(expected) != 0 {
		missing := make([]string, 0, len(expected))
		for chainID := range expected {
			missing = append(missing, chainID)
		}
		sort.Strings(missing)
		return nil, fmt.Errorf("generated Hermes config is missing chains: %s", strings.Join(missing, ", "))
	}

	output, err := tree.ToTomlString()
	if err != nil {
		return nil, fmt.Errorf("serialize generated Hermes config: %w", err)
	}
	return []byte(output), nil
}

func pinHermesPacketFilters(config []byte, channelByChain map[string]string) ([]byte, error) {
	tree, err := toml.LoadBytes(config)
	if err != nil {
		return nil, fmt.Errorf("parse generated Hermes config for packet filters: %w", err)
	}
	chains, ok := tree.Get("chains").([]*toml.Tree)
	if !ok || len(chains) == 0 {
		return nil, errors.New("generated Hermes config has no chain tables")
	}
	expected := make(map[string]string, len(channelByChain))
	for chainID, channelID := range channelByChain {
		chainID = strings.TrimSpace(chainID)
		channelID = strings.TrimSpace(channelID)
		if chainID == "" || channelID == "" {
			return nil, errors.New("Hermes packet filter requires chain and channel IDs")
		}
		expected[chainID] = channelID
	}
	for _, chain := range chains {
		chainID, ok := chain.Get("id").(string)
		if !ok || strings.TrimSpace(chainID) == "" {
			return nil, errors.New("generated Hermes chain table has no ID")
		}
		channelID, ok := expected[chainID]
		if !ok {
			continue
		}
		filter, err := toml.Load(fmt.Sprintf(
			"policy = \"allow\"\nlist = [[\"transfer\", %q]]\n",
			channelID,
		))
		if err != nil {
			return nil, fmt.Errorf("construct Hermes packet filter for %s: %w", chainID, err)
		}
		chain.Set("packet_filter", filter)
		delete(expected, chainID)
	}
	if len(expected) != 0 {
		missing := make([]string, 0, len(expected))
		for chainID := range expected {
			missing = append(missing, chainID)
		}
		sort.Strings(missing)
		return nil, fmt.Errorf("generated Hermes config is missing packet-filter chains: %s", strings.Join(missing, ", "))
	}
	output, err := tree.ToTomlString()
	if err != nil {
		return nil, fmt.Errorf("serialize packet-filtered Hermes config: %w", err)
	}
	return []byte(output), nil
}

type hermesCommandEvidence struct {
	Command  []string `json:"command"`
	ExitCode int      `json:"exit_code"`
	Stdout   string   `json:"stdout"`
	Stderr   string   `json:"stderr"`
	Error    string   `json:"error,omitempty"`
}

// HermesRuntimeIdentityEvidence joins the immutable release provenance with
// observations made inside the exact relayer image. The version output alone
// does not expose the source commit, while the multi-architecture image digest
// alone does not make the executed binary checksum visible.
type HermesRuntimeIdentityEvidence struct {
	ReleaseIdentifier string `json:"release_identifier"`
	VersionOutput     string `json:"version_output"`
	SourceCommit      string `json:"source_commit"`
	ImageReference    string `json:"image_reference"`
	ImageDigest       string `json:"image_digest"`
	BinaryPath        string `json:"binary_path"`
	BinarySHA256      string `json:"binary_sha256"`
}

func validateHermesRuntimeIdentity(
	versionOutput string,
	checksumOutput string,
	provenance IBCReleaseProvenance,
) (HermesRuntimeIdentityEvidence, error) {
	var zero HermesRuntimeIdentityEvidence
	expectedProvenance := PinnedIBCProvenance().Hermes
	if provenance.Hermes != expectedProvenance {
		return zero, fmt.Errorf("Hermes provenance is not the pinned %s release", expectedProvenance.ReleaseIdentifier)
	}

	versionOutput = strings.TrimSpace(versionOutput)
	expectedVersion := "hermes " + hermesImageTag
	if versionOutput != expectedVersion {
		return zero, fmt.Errorf("Hermes runtime version = %q, want exact %q", versionOutput, expectedVersion)
	}

	checksumFields := strings.Fields(checksumOutput)
	if len(checksumFields) != 2 {
		return zero, errors.New("Hermes binary checksum output must contain one SHA-256 and one path")
	}
	checksum := checksumFields[0]
	decodedChecksum, err := hex.DecodeString(checksum)
	if err != nil || len(decodedChecksum) != 32 || checksum != strings.ToLower(checksum) {
		return zero, errors.New("Hermes binary checksum is not a lowercase SHA-256")
	}
	binaryPath := checksumFields[1]
	if !filepath.IsAbs(binaryPath) || filepath.Base(binaryPath) != "hermes" {
		return zero, fmt.Errorf("Hermes binary checksum path %q is not an absolute Hermes binary path", binaryPath)
	}

	return HermesRuntimeIdentityEvidence{
		ReleaseIdentifier: expectedProvenance.ReleaseIdentifier,
		VersionOutput:     versionOutput,
		SourceCommit:      expectedProvenance.SourceCommit,
		ImageReference:    expectedProvenance.Reference,
		ImageDigest:       expectedProvenance.Digest,
		BinaryPath:        binaryPath,
		BinarySHA256:      checksum,
	}, nil
}

func (a *ibcTopologyArtifacts) recordHermesRuntimeEvidence(
	ctx context.Context,
	relayer *hermes.Relayer,
	reporter ibc.RelayerExecReporter,
	client *dockerclient.Client,
) (HermesRuntimeIdentityEvidence, error) {
	var zero HermesRuntimeIdentityEvidence
	if relayer == nil {
		return zero, errors.New("Hermes relayer is required")
	}
	if runtimeImage := relayer.ContainerImage(); runtimeImage != PinnedHermesImage() {
		return zero, fmt.Errorf("Hermes runtime image = %s, want exact %s", runtimeImage.Ref(), PinnedHermesImage().Ref())
	}

	config, err := relayer.ReadFileFromHomeDir(ctx, ".hermes/config.toml")
	if err != nil {
		return zero, fmt.Errorf("read generated Hermes config: %w", err)
	}
	config, err = pinHermesCompatMode(config, []string{
		a.descriptor.PanaceaChainID,
		a.descriptor.OsmosisChainID,
	})
	if err != nil {
		return zero, err
	}
	if err := relayer.WriteFileToHomeDir(ctx, ".hermes/config.toml", config); err != nil {
		return zero, fmt.Errorf("write pinned Hermes config: %w", err)
	}
	if err := a.base.write("ibc/hermes/config.toml", config); err != nil {
		return zero, err
	}

	version, err := a.execHermesEvidence(ctx, relayer, reporter, "version", []string{"hermes", "version"})
	if err != nil {
		return zero, err
	}
	if err := a.base.write("ibc/hermes/version.txt", []byte(strings.TrimSpace(version)+"\n")); err != nil {
		return zero, err
	}
	checksum, err := a.execHermesEvidence(
		ctx,
		relayer,
		reporter,
		"binary-sha256",
		[]string{"sh", "-c", `set -eu; hermes_path="$(command -v hermes)"; sha256sum "$hermes_path"`},
	)
	if err != nil {
		return zero, err
	}
	if err := a.base.write("ibc/hermes/binary-sha256.txt", []byte(strings.TrimSpace(checksum)+"\n")); err != nil {
		return zero, err
	}
	identity, err := validateHermesRuntimeIdentity(version, checksum, PinnedIBCProvenance())
	if err != nil {
		return zero, err
	}
	if err := a.base.writeJSON("ibc/hermes/runtime-identity.json", identity); err != nil {
		return zero, err
	}
	if _, err := a.execHermesEvidence(ctx, relayer, reporter, "config-validate", []string{"hermes", "config", "validate"}); err != nil {
		return zero, err
	}
	if _, err := a.execHermesEvidence(ctx, relayer, reporter, "health-check", []string{"hermes", "health-check"}); err != nil {
		return zero, err
	}
	if err := a.recordResolvedIBCImages(ctx, client); err != nil {
		return zero, err
	}
	return identity, nil
}

func (a *ibcTopologyArtifacts) pinHermesTransferChannels(
	ctx context.Context,
	relayer *hermes.Relayer,
	reporter ibc.RelayerExecReporter,
	handshake IBCChannelHandshake,
) error {
	if a == nil || a.base == nil {
		return errors.New("IBC artifact store is required")
	}
	if relayer == nil {
		return errors.New("Hermes relayer is required")
	}
	if err := handshake.Validate(); err != nil {
		return fmt.Errorf("validate channel before pinning Hermes packet filters: %w", err)
	}

	config, err := relayer.ReadFileFromHomeDir(ctx, ".hermes/config.toml")
	if err != nil {
		return fmt.Errorf("read Hermes config before packet filtering: %w", err)
	}
	config, err = pinHermesPacketFilters(config, map[string]string{
		handshake.Panacea.ChainID: handshake.Panacea.ChannelID,
		handshake.Osmosis.ChainID: handshake.Osmosis.ChannelID,
	})
	if err != nil {
		return err
	}
	if err := relayer.WriteFileToHomeDir(ctx, ".hermes/config.toml", config); err != nil {
		return fmt.Errorf("write channel-filtered Hermes config: %w", err)
	}
	if err := a.base.write("ibc/hermes/config.toml", config); err != nil {
		return err
	}
	if err := a.base.write("ibc/hermes/config-post-channel.toml", config); err != nil {
		return err
	}
	if _, err := a.execHermesEvidence(ctx, relayer, reporter, "config-validate-post-channel", []string{"hermes", "config", "validate"}); err != nil {
		return err
	}
	if _, err := a.execHermesEvidence(ctx, relayer, reporter, "health-check-post-channel", []string{"hermes", "health-check"}); err != nil {
		return err
	}
	return nil
}

func (a *ibcTopologyArtifacts) execHermesEvidence(
	ctx context.Context,
	relayer *hermes.Relayer,
	reporter ibc.RelayerExecReporter,
	name string,
	command []string,
) (string, error) {
	result := relayer.Exec(ctx, reporter, command, nil)
	evidence := hermesCommandEvidence{
		Command:  append([]string(nil), command...),
		ExitCode: result.ExitCode,
		Stdout:   string(result.Stdout),
		Stderr:   string(result.Stderr),
		Error:    errorString(result.Err),
	}
	if err := a.base.writeJSON(filepath.Join("ibc", "hermes", name+".json"), evidence); err != nil {
		return "", err
	}
	if result.Err != nil {
		return "", fmt.Errorf("Hermes %s execution: %w", name, result.Err)
	}
	if result.ExitCode != 0 {
		return "", fmt.Errorf("Hermes %s exited with code %d: %s", name, result.ExitCode, strings.TrimSpace(string(result.Stderr)))
	}
	return string(result.Stdout), nil
}

type resolvedImageEvidence struct {
	RequestedReference string   `json:"requested_reference"`
	ImageID            string   `json:"image_id"`
	RepoDigests        []string `json:"repo_digests,omitempty"`
	Architecture       string   `json:"architecture,omitempty"`
	OS                 string   `json:"os,omitempty"`
}

func (a *ibcTopologyArtifacts) recordResolvedIBCImages(ctx context.Context, client *dockerclient.Client) error {
	if client == nil {
		return errors.New("Docker client is required for IBC image provenance")
	}
	images := map[string]ibc.DockerImage{
		"osmosis": PinnedOsmosisImage(),
		"hermes":  PinnedHermesImage(),
	}
	resolved := make(map[string]resolvedImageEvidence, len(images))
	for name, image := range images {
		inspect, _, err := client.ImageInspectWithRaw(ctx, image.Ref())
		if err != nil {
			return fmt.Errorf("inspect pinned %s image %s: %w", name, image.Ref(), err)
		}
		resolved[name] = resolvedImageEvidence{
			RequestedReference: image.Ref(),
			ImageID:            inspect.ID,
			RepoDigests:        append([]string(nil), inspect.RepoDigests...),
			Architecture:       inspect.Architecture,
			OS:                 inspect.Os,
		}
	}
	return a.base.writeJSON("ibc/resolved-images.json", resolved)
}

func (a *ibcTopologyArtifacts) collect(
	failed bool,
	osmosis *cosmos.CosmosChain,
	client *dockerclient.Client,
) error {
	var osmosisErr error
	if osmosis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), artifactCollectionDeadline(len(osmosis.Nodes())))
		osmosisErr = a.collectOsmosis(ctx, osmosis, client)
		cancel()
		if osmosisErr != nil {
			a.base.recordFailure("ibc-osmosis-artifact-collection", osmosisErr)
		}
	}
	baseErr := a.base.collect(failed || osmosisErr != nil)
	return errors.Join(osmosisErr, baseErr)
}

func (a *ibcTopologyArtifacts) collectOsmosis(
	ctx context.Context,
	chain *cosmos.CosmosChain,
	client *dockerclient.Client,
) error {
	var collectionErrors []error
	const chainRoot = "chains/osmosis"
	if len(chain.Validators) > 0 {
		genesis, err := chain.Validators[0].ReadFile(ctx, "config/genesis.json")
		if err != nil {
			collectionErrors = append(collectionErrors, fmt.Errorf("read Osmosis genesis: %w", err))
		} else if err := a.base.write(filepath.Join(chainRoot, "genesis.json"), genesis); err != nil {
			collectionErrors = append(collectionErrors, err)
		}
		stdout, stderr, err := chain.Validators[0].ExecBin(ctx, "version", "--long")
		if err != nil {
			collectionErrors = append(collectionErrors, fmt.Errorf("read Osmosis version: %w: %s", err, strings.TrimSpace(string(stderr))))
		} else if err := a.base.write(filepath.Join(chainRoot, "versions.txt"), stdout); err != nil {
			collectionErrors = append(collectionErrors, err)
		}
	}
	for _, node := range chain.Nodes() {
		base := filepath.Join(chainRoot, "nodes", node.Name())
		for _, name := range []string{"app.toml", "config.toml", "client.toml"} {
			contents, err := node.ReadFile(ctx, filepath.Join("config", name))
			if err != nil {
				collectionErrors = append(collectionErrors, fmt.Errorf("Osmosis node %s read %s: %w", node.Name(), name, err))
				continue
			}
			if err := a.base.write(filepath.Join(base, "config", name), contents); err != nil {
				collectionErrors = append(collectionErrors, err)
			}
		}
		statusArtifact := artifactStatus{RecordedAt: time.Now().UTC()}
		if node.Client == nil {
			statusArtifact.Error = "CometBFT RPC client is not initialized"
			collectionErrors = append(collectionErrors, fmt.Errorf("Osmosis node %s status: %s", node.Name(), statusArtifact.Error))
		} else {
			status, err := node.Client.Status(ctx)
			if err != nil {
				statusArtifact.Error = err.Error()
				collectionErrors = append(collectionErrors, fmt.Errorf("Osmosis node %s status: %w", node.Name(), err))
			} else if status == nil {
				statusArtifact.Error = "CometBFT RPC returned an empty status"
				collectionErrors = append(collectionErrors, fmt.Errorf("Osmosis node %s status: %s", node.Name(), statusArtifact.Error))
			} else {
				statusArtifact.OK = true
				statusArtifact.LastHeight = status.SyncInfo.LatestBlockHeight
				statusArtifact.CometBFTStatus = status
				a.base.recordNodeHeight(node.Name(), statusArtifact.LastHeight)
			}
		}
		if err := a.base.writeJSON(filepath.Join(base, "status.json"), statusArtifact); err != nil {
			collectionErrors = append(collectionErrors, err)
		}
		if client != nil && node.ContainerID() != "" {
			if err := a.base.collectLogs(ctx, client, node.ContainerID(), filepath.Join(base, "logs", "container.log")); err != nil {
				collectionErrors = append(collectionErrors, fmt.Errorf("Osmosis node %s logs: %w", node.Name(), err))
			}
		}
	}
	return errors.Join(collectionErrors...)
}
