package harness

import (
	"fmt"

	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
)

const (
	osmosisImageRepository = "docker.io/osmolabs/osmosis"
	osmosisImageTag        = "31.0.2"
	osmosisImageDigest     = "sha256:8de930072fef03ea034b5a38f3cf93e5f47b6ccb8b1776a34e402aa47c819e0e"
	osmosisSourceCommit    = "a56c05b0e83341b9a3c0e6e3508520f15e9f2e49"

	hermesImageRepository = "ghcr.io/informalsystems/hermes"
	hermesImageTag        = "1.8.2"
	hermesImageDigest     = "sha256:5422e8a26bf42db4a6223e999823df9269428bc936c9bc5826221632304d28b1"
	hermesSourceCommit    = "06dfbafb4893255a79043ec4032034a83ebd53df"
)

// PinnedOsmosisImage returns the exact multi-architecture Osmosis release used
// as the local IBC counterparty. The tag keeps artifacts readable while the
// digest prevents the registry tag from changing the executed bytes.
func PinnedOsmosisImage() ibc.DockerImage {
	return ibc.DockerImage{
		Repository: osmosisImageRepository,
		Version:    osmosisImageTag + "@" + osmosisImageDigest,
		UIDGID:     "0:0",
	}
}

// PinnedHermesImage returns the exact Hermes release deployed by the topology.
func PinnedHermesImage() ibc.DockerImage {
	return ibc.DockerImage{
		Repository: hermesImageRepository,
		Version:    hermesImageTag + "@" + hermesImageDigest,
		UIDGID:     "1000:1000",
	}
}

// IBCReleaseProvenance is the immutable compatibility input recorded beside
// every topology run. Runtime inspection artifacts are kept separately so a
// requested image reference is never mistaken for Docker's resolved image ID.
type IBCReleaseProvenance struct {
	Osmosis OsmosisReleaseProvenance `json:"osmosis"`
	Hermes  HermesReleaseProvenance  `json:"hermes"`
}

type OsmosisReleaseProvenance struct {
	Repository       string `json:"repository"`
	Tag              string `json:"tag"`
	Digest           string `json:"digest"`
	Reference        string `json:"reference"`
	UIDGID           string `json:"uid_gid"`
	SourceCommit     string `json:"source_commit"`
	CosmosSDKVersion string `json:"cosmos_sdk_version"`
	CometBFTVersion  string `json:"cometbft_version"`
	IBCGoVersion     string `json:"ibc_go_version"`
}

type HermesReleaseProvenance struct {
	Repository        string `json:"repository"`
	Tag               string `json:"tag"`
	Digest            string `json:"digest"`
	Reference         string `json:"reference"`
	UIDGID            string `json:"uid_gid"`
	SourceCommit      string `json:"source_commit"`
	ReleaseIdentifier string `json:"release_identifier"`
}

// PinnedIBCProvenance returns a value copy suitable for artifact serialization.
func PinnedIBCProvenance() IBCReleaseProvenance {
	osmosisImage := PinnedOsmosisImage()
	hermesImage := PinnedHermesImage()
	return IBCReleaseProvenance{
		Osmosis: OsmosisReleaseProvenance{
			Repository:       osmosisImage.Repository,
			Tag:              osmosisImageTag,
			Digest:           osmosisImageDigest,
			Reference:        osmosisImage.Ref(),
			UIDGID:           osmosisImage.UIDGID,
			SourceCommit:     osmosisSourceCommit,
			CosmosSDKVersion: "v0.50.14-v30-osmo",
			CometBFTVersion:  "v0.38.22",
			IBCGoVersion:     "v8.7.0",
		},
		Hermes: HermesReleaseProvenance{
			Repository:        hermesImage.Repository,
			Tag:               hermesImageTag,
			Digest:            hermesImageDigest,
			Reference:         hermesImage.Ref(),
			UIDGID:            hermesImage.UIDGID,
			SourceCommit:      hermesSourceCommit,
			ReleaseIdentifier: hermesImageTag + "+" + hermesSourceCommit[:7],
		},
	}
}

// NewOsmosisChainSpec builds a small local chain from the exact Osmosis
// v31.0.2 production release. It deliberately uses a run-owned chain ID rather
// than the public osmosis-1 identity.
func NewOsmosisChainSpec(runID string) (*interchaintest.ChainSpec, error) {
	if !runIDPattern.MatchString(runID) {
		return nil, fmt.Errorf("run ID %q must match %s", runID, runIDPattern)
	}

	chainName := "osmosis-" + runID
	coinDecimals := int64(6)
	validators := 1
	fullNodes := 1
	image := PinnedOsmosisImage()

	return &interchaintest.ChainSpec{
		Name:          chainName,
		ChainName:     chainName,
		Version:       image.Version,
		NumValidators: &validators,
		NumFullNodes:  &fullNodes,
		ChainConfig: ibc.ChainConfig{
			Type:         ibc.Cosmos,
			Name:         chainName,
			ChainID:      chainName,
			Bin:          "osmosisd",
			Bech32Prefix: "osmo",
			Denom:        "uosmo",
			CoinType:     "118",
			CoinDecimals: &coinDecimals,
			// Osmosis v31 enables the feemarket module. Its initial local-chain
			// base-gas-price is approximately 0.03uosmo, so using the historical
			// 0.0025 value makes Hermes emit a JSON ChainError even though the
			// command exits zero and leaves the IBC path only half configured.
			GasPrices:      "0.03uosmo",
			GasAdjustment:  2.0,
			Gas:            "auto",
			TrustingPeriod: "336h",
			Images:         []ibc.DockerImage{image},
			ConfigFileOverrides: map[string]any{
				"config/config.toml": testutil.Toml{
					"consensus": testutil.Toml{
						"timeout_commit": "1s",
					},
				},
			},
		},
	}, nil
}
