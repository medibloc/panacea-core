package e2e_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestNFTNegativeProtocolBoundaries(t *testing.T) {
	if os.Getenv("PANACEA_E2E") != "1" {
		t.Skip("set PANACEA_E2E=1 or use ./scripts/e2e/run.sh negative")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	network, err := harness.Start(ctx, t, harness.Config{
		Image:         harness.CurrentImage(),
		NumValidators: 1,
		NumFullNodes:  1,
	})
	require.NoError(t, err)
	defer network.RecordTestPanic()

	creator := buildAndFundNegativeIntegrityWallet(t, ctx, network, "negative-boundary-creator")
	owner := buildAndFundNegativeIntegrityWallet(t, ctx, network, "negative-boundary-owner")
	receiver := buildAndFundNegativeIntegrityWallet(t, ctx, network, "negative-boundary-receiver")
	trackedAddresses := []string{
		creator.FormattedAddress(),
		owner.FormattedAddress(),
		receiver.FormattedAddress(),
	}

	maximumLocalClassID := strings.Repeat("c", 64)
	maximumClassURI := fixedLengthNegativeURI(256)
	expectNegativeSuccess(
		t,
		ctx,
		network,
		"negative-boundary-create-maximum",
		creator.KeyName(),
		"nft", "create-class",
		maximumLocalClassID,
		strings.Repeat("n", 128),
		strings.Repeat("s", 32),
		"owner-transferable", "true", "10",
		"--description", strings.Repeat("d", 1024),
		"--uri", maximumClassURI,
		"--uri-hash", negativeURIHash,
	)
	classID := creator.FormattedAddress() + ":" + maximumLocalClassID

	maximumNFTID := strings.Repeat("n", 64)
	maximumDataName := strings.Repeat("m", 1021) // tag + two-byte length + value = 1024 bytes
	expectNegativeSuccess(
		t,
		ctx,
		network,
		"negative-boundary-mint-maximum",
		creator.KeyName(),
		"nft", "mint", classID, maximumNFTID, owner.FormattedAddress(),
		"--uri", fixedLengthNegativeURI(256),
		"--uri-hash", negativeURIHash,
		"--data", fmt.Sprintf(`{"@type":"/panacea.nft.v1.BasicNFTData","name":%q}`, maximumDataName),
	)
	recordIDs := []string{maximumNFTID}

	t.Run("class metadata byte and control boundaries", func(t *testing.T) {
		for _, testCase := range []struct {
			name    string
			localID string
			command []string
		}{
			{
				name: "name exceeds 128 bytes", localID: "name-too-long",
				command: []string{"nft", "create-class", "name-too-long", strings.Repeat("n", 129), "SYM", "owner-transferable", "true", "1"},
			},
			{
				name: "symbol exceeds 32 bytes", localID: "symbol-too-long",
				command: []string{"nft", "create-class", "symbol-too-long", "Name", strings.Repeat("s", 33), "owner-transferable", "true", "1"},
			},
			{
				name: "description exceeds 1024 bytes", localID: "description-too-long",
				command: []string{"nft", "create-class", "description-too-long", "Name", "SYM", "owner-transferable", "true", "1", "--description", strings.Repeat("d", 1025)},
			},
			{
				name: "URI exceeds 256 bytes", localID: "uri-too-long",
				command: []string{"nft", "create-class", "uri-too-long", "Name", "SYM", "owner-transferable", "true", "1", "--uri", fixedLengthNegativeURI(257)},
			},
			{
				name: "control character in local class ID", localID: "",
				command: []string{"nft", "create-class", "bad\nid", "Name", "SYM", "owner-transferable", "true", "1"},
			},
			{
				name: "control character in URI", localID: "uri-control",
				command: []string{"nft", "create-class", "uri-control", "Name", "SYM", "owner-transferable", "true", "1", "--uri", "https://example.test/bad\nuri"},
			},
			{
				name: "uppercase URI hash", localID: "hash-uppercase",
				command: []string{"nft", "create-class", "hash-uppercase", "Name", "SYM", "owner-transferable", "true", "1", "--uri", "https://example.test/hash", "--uri-hash", "sha256:" + strings.Repeat("A", 64)},
			},
			{
				name: "short URI hash", localID: "hash-short",
				command: []string{"nft", "create-class", "hash-short", "Name", "SYM", "owner-transferable", "true", "1", "--uri", "https://example.test/hash", "--uri-hash", "sha256:" + strings.Repeat("a", 63)},
			},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				assertDeliverFailureLeavesNFTState(
					t,
					ctx,
					network,
					"boundary-class-"+slugNegativeStep(testCase.name),
					creator.KeyName(),
					"sdk",
					18,
					classID,
					recordIDs,
					trackedAddresses,
					testCase.command...,
				)
				if testCase.localID != "" {
					assertClassRecordAbsent(
						t, ctx, network, "boundary-class-absent-"+slugNegativeStep(testCase.name),
						creator.FormattedAddress()+":"+testCase.localID,
					)
				}
			})
		}
	})

	t.Run("NFT metadata byte hash and control boundaries", func(t *testing.T) {
		for _, testCase := range []struct {
			name    string
			nftID   string
			command []string
		}{
			{
				name: "URI exceeds 256 bytes", nftID: "uri-too-long.1",
				command: []string{"nft", "mint", classID, "uri-too-long.1", owner.FormattedAddress(), "--uri", fixedLengthNegativeURI(257)},
			},
			{
				name: "control character in NFT ID", nftID: "",
				command: []string{"nft", "mint", classID, "bad\nid", owner.FormattedAddress()},
			},
			{
				name: "control character in URI", nftID: "uri-control.1",
				command: []string{"nft", "mint", classID, "uri-control.1", owner.FormattedAddress(), "--uri", "https://example.test/bad\nuri"},
			},
			{
				name: "uppercase URI hash", nftID: "hash-uppercase.1",
				command: []string{"nft", "mint", classID, "hash-uppercase.1", owner.FormattedAddress(), "--uri", "https://example.test/hash", "--uri-hash", "sha256:" + strings.Repeat("A", 64)},
			},
			{
				name: "short URI hash", nftID: "hash-short.1",
				command: []string{"nft", "mint", classID, "hash-short.1", owner.FormattedAddress(), "--uri", "https://example.test/hash", "--uri-hash", "sha256:" + strings.Repeat("a", 63)},
			},
			{
				name: "encoded data exceeds 1024 bytes", nftID: "data-too-long.1",
				command: []string{"nft", "mint", classID, "data-too-long.1", owner.FormattedAddress(), "--data", fmt.Sprintf(`{"@type":"/panacea.nft.v1.BasicNFTData","name":%q}`, strings.Repeat("m", 1022))},
			},
			{
				name: "control character in data", nftID: "data-control.1",
				command: []string{"nft", "mint", classID, "data-control.1", owner.FormattedAddress(), "--data", `{"@type":"/panacea.nft.v1.BasicNFTData","name":"bad\u0000name"}`},
			},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				if testCase.nftID != "" {
					assertNFTRecordAbsent(t, ctx, network, "boundary-nft-before-"+slugNegativeStep(testCase.name), classID, testCase.nftID)
				}
				assertDeliverFailureLeavesNFTState(
					t,
					ctx,
					network,
					"boundary-nft-"+slugNegativeStep(testCase.name),
					creator.KeyName(),
					"sdk",
					18,
					classID,
					recordIDs,
					trackedAddresses,
					testCase.command...,
				)
				if testCase.nftID != "" {
					assertNFTRecordAbsent(t, ctx, network, "boundary-nft-after-"+slugNegativeStep(testCase.name), classID, testCase.nftID)
				}
			})
		}
	})

	t.Run("wrong network address prefix is rejected for every destination role", func(t *testing.T) {
		wrongPrefix, err := bech32.ConvertAndEncode("cosmos", receiver.Address())
		require.NoError(t, err)
		for _, testCase := range []struct {
			name         string
			nftID        string
			keyName      string
			cliPreflight bool
			command      []string
		}{
			{
				name: "controller", keyName: creator.KeyName(), cliPreflight: true,
				command: []string{"nft", "update-controller", classID, wrongPrefix},
			},
			{
				name: "recipient", nftID: "wrong-prefix-recipient.1", keyName: creator.KeyName(),
				command: negativeMintCommand(classID, "wrong-prefix-recipient.1", wrongPrefix),
			},
			{
				name: "receiver", keyName: owner.KeyName(),
				command: []string{"nft", "send", classID, maximumNFTID, wrongPrefix},
			},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				if testCase.nftID != "" {
					assertNFTRecordAbsent(t, ctx, network, "boundary-address-before-"+testCase.name, classID, testCase.nftID)
				}
				if testCase.cliPreflight {
					assertCLIRejectionLeavesNFTState(
						t,
						ctx,
						network,
						"boundary-address-"+testCase.name,
						testCase.keyName,
						"hrp does not match bech32 prefix",
						classID,
						recordIDs,
						trackedAddresses,
						testCase.command...,
					)
				} else {
					assertDeliverFailureLeavesNFTState(
						t,
						ctx,
						network,
						"boundary-address-"+testCase.name,
						testCase.keyName,
						"sdk",
						7,
						classID,
						recordIDs,
						trackedAddresses,
						testCase.command...,
					)
				}
				if testCase.nftID != "" {
					assertNFTRecordAbsent(t, ctx, network, "boundary-address-after-"+testCase.name, classID, testCase.nftID)
				}
			})
		}
	})
}

func assertCLIRejectionLeavesNFTState(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	keyName string,
	expectedDiagnostic string,
	classID string,
	recordIDs []string,
	trackedAddresses []string,
	command ...string,
) {
	t.Helper()
	before := captureNegativeNFTSnapshot(t, ctx, network, step+"-before", classID, recordIDs, trackedAddresses)
	require.NoError(t, network.BroadcastTxExpectCLIRejection(
		ctx,
		step,
		network.Chain.Validators[0],
		keyName,
		expectedDiagnostic,
		negativeTxCommand(command...)...,
	))
	after := captureNegativeNFTSnapshot(t, ctx, network, step+"-after", classID, recordIDs, trackedAddresses)
	require.Equal(t, before, after, "CLI-rejected transaction changed NFT state or derived indexes")
}

func fixedLengthNegativeURI(length int) string {
	const prefix = "https://example.test/"
	if length <= len(prefix) {
		return prefix[:length]
	}
	return prefix + strings.Repeat("u", length-len(prefix))
}

func assertClassRecordAbsent(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	classID string,
) {
	t.Helper()
	_, err := network.FullNodeRESTGet(
		ctx,
		&http.Client{Timeout: 5 * time.Second},
		step,
		"/panacea/nft/v1/classes/"+strings.Replace(classID, ":", "%3A", 1),
	)
	require.ErrorContains(t, err, "HTTP 404")
}

func TestFixedLengthNegativeURIUsesByteLength(t *testing.T) {
	t.Parallel()
	for _, length := range []int{0, 1, 21, 22, 256, 257} {
		require.Len(t, []byte(fixedLengthNegativeURI(length)), length)
	}
}
