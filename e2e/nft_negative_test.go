package e2e_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/go-bip39"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	negativePrimaryNFTID = "negative.1"
	negativeLockedNFTID  = "locked.1"
	negativeNoRevokeID   = "no-revoke.1"
	negativeCapacityID   = "capacity.1"
)

var negativeURIHash = "sha256:" + strings.Repeat("c", 64)

type negativeNFTSnapshot struct {
	AllClasses       any
	ClassRecord      any
	CustomLiveNFTs   any
	StandardLiveNFTs any
	Records          map[string]any
	Supply           uint64
	Balances         map[string]uint64
	Owners           map[string]string
}

func decodeModuleAccountAddress(raw []byte) (string, error) {
	var response struct {
		Account struct {
			BaseAccount struct {
				Address string `json:"address"`
			} `json:"base_account"`
			Value struct {
				Address string `json:"address"`
			} `json:"value"`
		} `json:"account"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return "", err
	}
	address := response.Account.BaseAccount.Address
	if address == "" {
		address = response.Account.Value.Address
	}
	if address == "" {
		return "", errors.New("module account query returned no base account address")
	}
	return address, nil
}

func TestDecodeModuleAccountAddress(t *testing.T) {
	t.Parallel()

	t.Run("module account", func(t *testing.T) {
		raw := []byte(`{
			"account": {
				"@type": "/cosmos.auth.v1beta1.ModuleAccount",
				"base_account": {
					"address": "panacea1module"
				}
			}
		}`)

		address, err := decodeModuleAccountAddress(raw)
		require.NoError(t, err)
		require.Equal(t, "panacea1module", address)
	})

	t.Run("legacy module account CLI envelope", func(t *testing.T) {
		raw := []byte(`{
			"account": {
				"type": "/cosmos.auth.v1beta1.ModuleAccount",
				"value": {
					"address": "panacea1legacy"
				}
			}
		}`)

		address, err := decodeModuleAccountAddress(raw)
		require.NoError(t, err)
		require.Equal(t, "panacea1legacy", address)
	})

	t.Run("missing base account", func(t *testing.T) {
		_, err := decodeModuleAccountAddress([]byte(`{"account":{}}`))
		require.ErrorContains(t, err, "no base account address")
	})

	t.Run("malformed response", func(t *testing.T) {
		_, err := decodeModuleAccountAddress([]byte(`{"account":`))
		require.Error(t, err)
	})
}

func TestAppendRawProtoBytesPreservesVerbatimWireValue(t *testing.T) {
	t.Parallel()

	encoded := appendRawProtoBytes(nil, 1, []byte("x"))
	encoded = appendRawProtoBytes(encoded, 7, []byte{0xff, 0x00})
	require.Equal(t, "0a01783a02ff00", hex.EncodeToString(encoded))
}

// TestNFTNegativeStateIntegrity drives every message through a real signed
// transaction, CheckTx, consensus, FinalizeBlock, commit, and full-node query.
// Client-side Any parsing failures are asserted separately because no
// transaction exists for them by design.
func TestNFTNegativeStateIntegrity(t *testing.T) {
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

	creator := buildAndFundNegativeIntegrityWallet(t, ctx, network, "negative-integrity-creator")
	controller := buildAndFundNegativeIntegrityWallet(t, ctx, network, "negative-integrity-controller")
	owner := buildAndFundNegativeIntegrityWallet(t, ctx, network, "negative-integrity-owner")
	receiver := buildAndFundNegativeIntegrityWallet(t, ctx, network, "negative-integrity-receiver")
	attacker := buildAndFundNegativeIntegrityWallet(t, ctx, network, "negative-integrity-attacker")
	trackedAddresses := []string{
		creator.FormattedAddress(),
		controller.FormattedAddress(),
		owner.FormattedAddress(),
		receiver.FormattedAddress(),
		attacker.FormattedAddress(),
	}

	moduleJSON, err := network.FullNodeCLIQuery(ctx, "negative-fee-collector", "auth", "module-account", "fee_collector")
	require.NoError(t, err)
	moduleAddress, err := decodeModuleAccountAddress(moduleJSON)
	require.NoError(t, err)
	require.NotEmpty(t, moduleAddress)

	classID := createNegativeClass(
		t, ctx, network, creator,
		"negative.owner", "owner-transferable", true, 4,
	)
	expectNegativeSuccess(t, ctx, network, "negative-update-controller", creator.KeyName(),
		"nft", "update-controller", classID, controller.FormattedAddress(),
	)
	t.Run("no-op controller update is committed and atomic", func(t *testing.T) {
		result := assertDeliverFailureLeavesNFTState(
			t,
			ctx,
			network,
			"negative-no-op-controller-update",
			creator.KeyName(),
			"sdk",
			18,
			classID,
			nil,
			trackedAddresses,
			"nft", "update-controller", classID, controller.FormattedAddress(),
		)
		require.Contains(t, result.RawLog, "new controller must differ from current controller")
	})
	mintNegativeNFT(t, ctx, network, "negative-mint-primary", controller, classID, negativePrimaryNFTID, owner.FormattedAddress())

	lockedClassID := createNegativeClass(
		t, ctx, network, creator,
		"negative.locked", "locked", true, 2,
	)
	mintNegativeNFT(t, ctx, network, "negative-mint-locked", creator, lockedClassID, negativeLockedNFTID, owner.FormattedAddress())

	noRevokeClassID := createNegativeClass(
		t, ctx, network, creator,
		"negative.no-revoke", "owner-transferable", false, 2,
	)
	mintNegativeNFT(t, ctx, network, "negative-mint-no-revoke", creator, noRevokeClassID, negativeNoRevokeID, owner.FormattedAddress())

	capacityClassID := createNegativeClass(
		t, ctx, network, creator,
		"negative.capacity", "owner-transferable", true, 1,
	)
	mintNegativeNFT(t, ctx, network, "negative-mint-capacity", creator, capacityClassID, negativeCapacityID, owner.FormattedAddress())

	t.Run("module account cannot authenticate as class creator", func(t *testing.T) {
		assertModuleCreatorCheckTxFailure(
			t, ctx, network, attacker, moduleAddress,
			classID, []string{negativePrimaryNFTID}, trackedAddresses,
		)
	})

	t.Run("raw Any wire failures occur at the correct boundary and are atomic", func(t *testing.T) {
		buildMint := func(nftID, dataTypeURL string, dataValue []byte) harness.RawProtoMessage {
			var packedData []byte
			packedData = appendRawProtoString(packedData, 1, dataTypeURL)
			packedData = appendRawProtoBytes(packedData, 2, dataValue)

			var message []byte
			message = appendRawProtoString(message, 1, classID)
			message = appendRawProtoString(message, 2, nftID)
			message = appendRawProtoString(message, 3, controller.FormattedAddress())
			message = appendRawProtoString(message, 4, owner.FormattedAddress())
			message = appendRawProtoBytes(message, 7, packedData)
			return harness.RawProtoMessage{
				TypeURL: "/panacea.nft.v1.MsgMintRequest",
				Value:   message,
			}
		}

		var nonCanonicalBasicData []byte
		nonCanonicalBasicData = appendRawProtoString(nonCanonicalBasicData, 1, "first")
		nonCanonicalBasicData = appendRawProtoString(nonCanonicalBasicData, 1, "canonical-last-value")
		for _, tc := range []struct {
			name          string
			nftID         string
			dataType      string
			dataValue     []byte
			wantCommitted bool
			wantCode      uint32
			logContains   string
		}{
			{
				name: "non-canonical BasicNFTData wire bytes", nftID: "raw-noncanonical.1",
				dataType: "/panacea.nft.v1.BasicNFTData", dataValue: nonCanonicalBasicData,
				wantCommitted: true, wantCode: 18,
				logContains: "canonical deterministic protobuf encoding",
			},
			{
				name: "unknown NFTData type URL", nftID: "raw-unknown.1",
				dataType: "/panacea.nft.v1.UnknownNFTData", dataValue: []byte{0x0a, 0x01, 'x'},
				wantCommitted: false, wantCode: 2,
				logContains: "unable to resolve type URL /panacea.nft.v1.UnknownNFTData",
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				assertNFTRecordAbsent(t, ctx, network, "raw-any-absent-before-"+tc.nftID, classID, tc.nftID)
				before := captureNegativeNFTSnapshot(t, ctx, network, "raw-any-before-"+tc.nftID, classID, []string{negativePrimaryNFTID}, trackedAddresses)
				result, rawErr := network.BroadcastRawTx(ctx, "raw-any-"+tc.nftID, harness.RawTxRequest{
					Signer:    controller,
					Message:   buildMint(tc.nftID, tc.dataType, tc.dataValue),
					GasLimit:  500_000,
					FeeAmount: sdkmath.NewInt(2_500_000),
				})
				require.NoError(t, rawErr)
				require.NotNil(t, result)
				if tc.wantCommitted {
					require.Positive(t, result.HeightInt64())
				} else {
					require.Zero(t, result.HeightInt64())
				}
				require.Equal(t, "sdk", result.Codespace)
				require.Equal(t, tc.wantCode, result.Code)
				require.Contains(t, result.RawLog, tc.logContains)
				assertFailureHasNoNFTEvents(t, result)
				after := captureNegativeNFTSnapshot(t, ctx, network, "raw-any-after-"+tc.nftID, classID, []string{negativePrimaryNFTID}, trackedAddresses)
				require.Equal(t, before, after)
				assertNFTRecordAbsent(t, ctx, network, "raw-any-absent-"+tc.nftID, classID, tc.nftID)
			})
		}
	})

	t.Run("signer-bound malformed and module accounts fail in CheckTx", func(t *testing.T) {
		buildMint := func(controllerAddress, nftID string) harness.RawProtoMessage {
			var message []byte
			message = appendRawProtoString(message, 1, classID)
			message = appendRawProtoString(message, 2, nftID)
			message = appendRawProtoString(message, 3, controllerAddress)
			message = appendRawProtoString(message, 4, owner.FormattedAddress())
			return harness.RawProtoMessage{TypeURL: "/panacea.nft.v1.MsgMintRequest", Value: message}
		}
		buildBurn := func(ownerAddress string) harness.RawProtoMessage {
			var message []byte
			message = appendRawProtoString(message, 1, classID)
			message = appendRawProtoString(message, 2, negativePrimaryNFTID)
			message = appendRawProtoString(message, 3, ownerAddress)
			return harness.RawProtoMessage{TypeURL: "/panacea.nft.v1.MsgBurnRequest", Value: message}
		}

		// In SDK v0.50, malformed signer fields survive TxDecoder and fail when
		// the auth Tx's ValidateBasic extracts annotated signers. The bare
		// address-codec error is therefore classified as undefined/1, not the
		// sdk/2 classification used for a transaction decoding failure.
		for _, tc := range []struct {
			name          string
			signer        ibc.Wallet
			signerAccount string
			feePayer      ibc.Wallet
			message       harness.RawProtoMessage
			wantCodespace string
			wantCode      uint32
			logContains   string
		}{
			{
				name: "malformed controller", signer: controller,
				message:       buildMint("not-an-address", "raw-invalid-controller.1"),
				wantCodespace: "undefined", wantCode: 1,
				logContains: "decoding bech32 failed",
			},
			{
				name: "module controller", signer: controller,
				signerAccount: moduleAddress, feePayer: controller,
				message:       buildMint(moduleAddress, "raw-module-controller.1"),
				wantCodespace: "sdk", wantCode: 8,
				logContains: "pubKey does not match signer address",
			},
			{
				name: "malformed owner", signer: owner,
				message:       buildBurn("not-an-address"),
				wantCodespace: "undefined", wantCode: 1,
				logContains: "decoding bech32 failed",
			},
			{
				name: "module owner", signer: owner,
				signerAccount: moduleAddress, feePayer: owner,
				message:       buildBurn(moduleAddress),
				wantCodespace: "sdk", wantCode: 8,
				logContains: "pubKey does not match signer address",
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				before := captureNegativeNFTSnapshot(t, ctx, network, "raw-signer-before-"+slugNegativeStep(tc.name), classID, []string{negativePrimaryNFTID}, trackedAddresses)
				result, rawErr := network.BroadcastRawTx(ctx, "raw-signer-"+slugNegativeStep(tc.name), harness.RawTxRequest{
					Signer:               tc.signer,
					SignerAccountAddress: tc.signerAccount,
					FeePayer:             tc.feePayer,
					Message:              tc.message,
					GasLimit:             500_000,
					FeeAmount:            sdkmath.NewInt(2_500_000),
				})
				require.NoError(t, rawErr)
				require.NotNil(t, result)
				require.Zero(t, result.HeightInt64())
				require.NotZero(t, result.Code)
				if tc.wantCodespace != "" {
					require.Equal(t, tc.wantCodespace, result.Codespace)
					require.Equal(t, tc.wantCode, result.Code)
				}
				require.Contains(t, result.RawLog, tc.logContains)
				after := captureNegativeNFTSnapshot(t, ctx, network, "raw-signer-after-"+slugNegativeStep(tc.name), classID, []string{negativePrimaryNFTID}, trackedAddresses)
				require.Equal(t, before, after)
			})
		}
	})

	t.Run("legacy PNFT direct and authz-nested execution are disabled", func(t *testing.T) {
		buildLegacyMint := func(legacyCreator string) harness.RawProtoMessage {
			var message []byte
			message = appendRawProtoString(message, 1, "legacy-denom")
			message = appendRawProtoString(message, 2, "legacy.1")
			message = appendRawProtoString(message, 8, legacyCreator)
			return harness.RawProtoMessage{TypeURL: "/panacea.pnft.v2.MsgMintPNFTRequest", Value: message}
		}
		assertLegacyDisabled := func(step string, signer ibc.Wallet, message harness.RawProtoMessage) {
			t.Helper()
			before := captureNegativeNFTSnapshot(t, ctx, network, step+"-before", classID, []string{negativePrimaryNFTID}, trackedAddresses)
			result, rawErr := network.BroadcastRawTx(ctx, step, harness.RawTxRequest{
				Signer:    signer,
				Message:   message,
				GasLimit:  500_000,
				FeeAmount: sdkmath.NewInt(2_500_000),
			})
			require.NoError(t, rawErr)
			require.NotNil(t, result)
			require.Positive(t, result.HeightInt64())
			require.Equal(t, "sdk", result.Codespace)
			require.EqualValues(t, 18, result.Code)
			require.Contains(t, result.RawLog, "legacy PNFT messages are disabled")
			assertFailureHasNoNFTEvents(t, result)
			after := captureNegativeNFTSnapshot(t, ctx, network, step+"-after", classID, []string{negativePrimaryNFTID}, trackedAddresses)
			require.Equal(t, before, after)
		}

		legacyMint := buildLegacyMint(creator.FormattedAddress())
		assertLegacyDisabled("legacy-pnft-direct", creator, legacyMint)

		expectNegativeSuccess(
			t, ctx, network, "legacy-pnft-authz-grant", creator.KeyName(),
			"authz", "grant", attacker.FormattedAddress(), "generic",
			"--msg-type", legacyMint.TypeURL,
		)
		var nestedLegacyAny []byte
		nestedLegacyAny = appendRawProtoString(nestedLegacyAny, 1, legacyMint.TypeURL)
		nestedLegacyAny = appendRawProtoBytes(nestedLegacyAny, 2, legacyMint.Value)
		var execMessage []byte
		execMessage = appendRawProtoString(execMessage, 1, attacker.FormattedAddress())
		execMessage = appendRawProtoBytes(execMessage, 2, nestedLegacyAny)
		assertLegacyDisabled("legacy-pnft-authz-nested", attacker, harness.RawProtoMessage{
			TypeURL: "/cosmos.authz.v1beta1.MsgExec",
			Value:   execMessage,
		})
	})

	t.Run("authorization failures are committed and atomic", func(t *testing.T) {
		for _, tc := range []struct {
			name    string
			keyName string
			command []string
		}{
			{
				name: "previous controller cannot update controller", keyName: creator.KeyName(),
				command: []string{"nft", "update-controller", classID, receiver.FormattedAddress()},
			},
			{
				name: "arbitrary account cannot update controller", keyName: attacker.KeyName(),
				command: []string{"nft", "update-controller", classID, receiver.FormattedAddress()},
			},
			{
				name: "previous controller cannot mint", keyName: creator.KeyName(),
				command: negativeMintCommand(classID, "old-controller.1", owner.FormattedAddress()),
			},
			{
				name: "arbitrary account cannot mint", keyName: attacker.KeyName(),
				command: negativeMintCommand(classID, "attacker.1", owner.FormattedAddress()),
			},
			{
				name: "previous controller cannot revoke", keyName: creator.KeyName(),
				command: []string{"nft", "revoke", classID, negativePrimaryNFTID},
			},
			{
				name: "arbitrary account cannot revoke", keyName: attacker.KeyName(),
				command: []string{"nft", "revoke", classID, negativePrimaryNFTID},
			},
			{
				name: "non-owner cannot send", keyName: attacker.KeyName(),
				command: []string{"nft", "send", classID, negativePrimaryNFTID, receiver.FormattedAddress()},
			},
			{
				name: "non-owner cannot burn", keyName: attacker.KeyName(),
				command: []string{"nft", "burn", classID, negativePrimaryNFTID},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				assertDeliverFailureLeavesNFTState(
					t, ctx, network, "auth-"+slugNegativeStep(tc.name), tc.keyName,
					"sdk", 4, classID, []string{negativePrimaryNFTID}, trackedAddresses,
					tc.command...,
				)
			})
		}
		for step, nftID := range map[string]string{
			"old-controller-mint-absent": "old-controller.1",
			"attacker-mint-absent":       "attacker.1",
		} {
			assertNFTRecordAbsent(t, ctx, network, step, classID, nftID)
		}
	})

	t.Run("validation and module-account failures are rejected at their boundary and atomic", func(t *testing.T) {
		for _, tc := range []struct {
			name       string
			controller string
		}{
			{name: "invalid new controller", controller: "not-an-address"},
			{name: "mixed-case new controller", controller: strings.ToUpper(receiver.FormattedAddress()[:1]) + receiver.FormattedAddress()[1:]},
		} {
			t.Run(tc.name, func(t *testing.T) {
				step := "validation-" + slugNegativeStep(tc.name)
				before := captureNegativeNFTSnapshot(t, ctx, network, step+"-before", classID, []string{negativePrimaryNFTID}, trackedAddresses)
				err := network.BroadcastTxExpectCLIRejection(
					ctx,
					step,
					network.Chain.Validators[0],
					controller.KeyName(),
					"decoding bech32 failed",
					"nft", "update-controller", classID, tc.controller,
				)
				require.NoError(t, err)
				after := captureNegativeNFTSnapshot(t, ctx, network, step+"-after", classID, []string{negativePrimaryNFTID}, trackedAddresses)
				require.Equal(t, before, after)
			})
		}

		for _, tc := range []struct {
			name      string
			keyName   string
			codespace string
			code      uint32
			command   []string
		}{
			{
				name: "invalid local class id", keyName: creator.KeyName(), codespace: "sdk", code: 18,
				command: []string{"nft", "create-class", "UPPERCASE", "Invalid", "BAD", "owner-transferable", "true", "1"},
			},
			{
				name: "overlong local class id", keyName: creator.KeyName(), codespace: "sdk", code: 18,
				command: []string{"nft", "create-class", strings.Repeat("a", 65), "Invalid", "BAD", "owner-transferable", "true", "1"},
			},
			{
				name: "control character in class metadata", keyName: creator.KeyName(), codespace: "sdk", code: 18,
				command: []string{"nft", "create-class", "bad-control", "bad\nname", "BAD", "owner-transferable", "true", "1"},
			},
			{
				name: "invalid class uri hash", keyName: creator.KeyName(), codespace: "sdk", code: 18,
				command: []string{"nft", "create-class", "bad-hash", "Bad Hash", "BAD", "owner-transferable", "true", "1", "--uri", "https://example.test/bad", "--uri-hash", "sha256:bad"},
			},
			{
				name: "invalid full class id", keyName: controller.KeyName(), codespace: "sdk", code: 18,
				command: negativeMintCommand("invalid", "invalid-class.1", owner.FormattedAddress()),
			},
			{
				name: "non-canonical class creator address", keyName: controller.KeyName(), codespace: "sdk", code: 18,
				command: negativeMintCommand(strings.ToUpper(creator.FormattedAddress()[:1])+creator.FormattedAddress()[1:]+":negative.owner", "mixed-class.1", owner.FormattedAddress()),
			},
			{
				name: "invalid nft id", keyName: controller.KeyName(), codespace: "sdk", code: 18,
				command: negativeMintCommand(classID, ".", owner.FormattedAddress()),
			},
			{
				name: "overlong nft id", keyName: controller.KeyName(), codespace: "sdk", code: 18,
				command: negativeMintCommand(classID, strings.Repeat("a", 65), owner.FormattedAddress()),
			},
			{
				name: "uri hash without uri", keyName: controller.KeyName(), codespace: "sdk", code: 18,
				command: []string{"nft", "mint", classID, "hash-without-uri.1", owner.FormattedAddress(), "--uri-hash", negativeURIHash},
			},
			{
				name: "empty basic nft data", keyName: controller.KeyName(), codespace: "sdk", code: 18,
				command: []string{"nft", "mint", classID, "empty-data.1", owner.FormattedAddress(), "--data", `{"@type":"/panacea.nft.v1.BasicNFTData"}`},
			},
			{
				name: "invalid mint recipient", keyName: controller.KeyName(), codespace: "sdk", code: 7,
				command: negativeMintCommand(classID, "invalid-recipient.1", "not-an-address"),
			},
			{
				name: "mixed-case mint recipient", keyName: controller.KeyName(), codespace: "sdk", code: 7,
				command: negativeMintCommand(classID, "mixed-recipient.1", strings.ToUpper(receiver.FormattedAddress()[:1])+receiver.FormattedAddress()[1:]),
			},
			{
				name: "invalid send receiver", keyName: owner.KeyName(), codespace: "sdk", code: 7,
				command: []string{"nft", "send", classID, negativePrimaryNFTID, "not-an-address"},
			},
			{
				name: "mixed-case send receiver", keyName: owner.KeyName(), codespace: "sdk", code: 7,
				command: []string{"nft", "send", classID, negativePrimaryNFTID, strings.ToUpper(receiver.FormattedAddress()[:1]) + receiver.FormattedAddress()[1:]},
			},
			{
				name: "module account controller", keyName: controller.KeyName(), codespace: "sdk", code: 18,
				command: []string{"nft", "update-controller", classID, moduleAddress},
			},
			{
				name: "module account recipient", keyName: controller.KeyName(), codespace: "sdk", code: 18,
				command: negativeMintCommand(classID, "module-recipient.1", moduleAddress),
			},
			{
				name: "module account receiver", keyName: owner.KeyName(), codespace: "sdk", code: 18,
				command: []string{"nft", "send", classID, negativePrimaryNFTID, moduleAddress},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				assertDeliverFailureLeavesNFTState(
					t, ctx, network, "validation-"+slugNegativeStep(tc.name), tc.keyName,
					tc.codespace, tc.code, classID, []string{negativePrimaryNFTID}, trackedAddresses,
					tc.command...,
				)
			})
		}
		for step, nftID := range map[string]string{
			"hash-without-uri-absent":  "hash-without-uri.1",
			"empty-data-absent":        "empty-data.1",
			"invalid-recipient-absent": "invalid-recipient.1",
			"mixed-recipient-absent":   "mixed-recipient.1",
			"module-recipient-absent":  "module-recipient.1",
		} {
			assertNFTRecordAbsent(t, ctx, network, step, classID, nftID)
		}
	})

	t.Run("duplicate live ids and classes are rejected without partial state", func(t *testing.T) {
		assertDeliverFailureLeavesNFTState(
			t, ctx, network, "duplicate-class", creator.KeyName(),
			"nft", 3, classID, []string{negativePrimaryNFTID}, trackedAddresses,
			"nft", "create-class", "negative.owner", "Duplicate", "DUP", "owner-transferable", "true", "4",
		)
		assertDeliverFailureLeavesNFTState(
			t, ctx, network, "duplicate-live-nft", controller.KeyName(),
			"nft", 5, classID, []string{negativePrimaryNFTID}, trackedAddresses,
			negativeMintCommand(classID, negativePrimaryNFTID, receiver.FormattedAddress())...,
		)
	})

	t.Run("class policies and lifecycle failures are permanent and atomic", func(t *testing.T) {
		assertDeliverFailureLeavesNFTState(
			t, ctx, network, "locked-send", owner.KeyName(),
			"panacea_nft", 2, lockedClassID, []string{negativeLockedNFTID}, trackedAddresses,
			"nft", "send", lockedClassID, negativeLockedNFTID, receiver.FormattedAddress(),
		)
		assertDeliverFailureLeavesNFTState(
			t, ctx, network, "non-revocable-revoke", creator.KeyName(),
			"panacea_nft", 3, noRevokeClassID, []string{negativeNoRevokeID}, trackedAddresses,
			"nft", "revoke", noRevokeClassID, negativeNoRevokeID,
		)

		expectNegativeSuccess(t, ctx, network, "burn-capacity", owner.KeyName(),
			"nft", "burn", capacityClassID, negativeCapacityID,
		)
		assertDeliverFailureLeavesNFTState(
			t, ctx, network, "capacity-not-restored-after-burn", creator.KeyName(),
			"panacea_nft", 5, capacityClassID, []string{negativeCapacityID}, trackedAddresses,
			negativeMintCommand(capacityClassID, "capacity.2", owner.FormattedAddress())...,
		)
		assertNFTRecordAbsent(t, ctx, network, "capacity-second-id-absent", capacityClassID, "capacity.2")
		assertDeliverFailureLeavesNFTState(
			t, ctx, network, "burned-id-permanent", creator.KeyName(),
			"panacea_nft", 6, capacityClassID, []string{negativeCapacityID}, trackedAddresses,
			negativeMintCommand(capacityClassID, negativeCapacityID, owner.FormattedAddress())...,
		)

		expectNegativeSuccess(t, ctx, network, "revoke-primary", controller.KeyName(),
			"nft", "revoke", classID, negativePrimaryNFTID,
		)
		assertDeliverFailureLeavesNFTState(
			t, ctx, network, "revoked-send", owner.KeyName(),
			"panacea_nft", 4, classID, []string{negativePrimaryNFTID}, trackedAddresses,
			"nft", "send", classID, negativePrimaryNFTID, receiver.FormattedAddress(),
		)
		assertDeliverFailureLeavesNFTState(
			t, ctx, network, "duplicate-revoke", controller.KeyName(),
			"panacea_nft", 4, classID, []string{negativePrimaryNFTID}, trackedAddresses,
			"nft", "revoke", classID, negativePrimaryNFTID,
		)
	})

	t.Run("competing transactions cannot spend one account sequence twice", func(t *testing.T) {
		before := captureNegativeNFTSnapshot(t, ctx, network, "competing-sequence-before", classID, []string{negativePrimaryNFTID}, trackedAddresses)
		accountJSON, err := network.FullNodeCLIQuery(
			ctx, "competing-sequence-account", "auth", "account-info", attacker.FormattedAddress(),
		)
		require.NoError(t, err)
		var accountResponse struct {
			Info struct {
				AccountNumber string `json:"account_number"`
				Sequence      string `json:"sequence"`
			} `json:"info"`
		}
		require.NoError(t, json.Unmarshal(accountJSON, &accountResponse))
		require.NotEmpty(t, accountResponse.Info.AccountNumber)
		sequence, err := strconv.ParseUint(accountResponse.Info.Sequence, 10, 64)
		require.NoError(t, err)

		type broadcastOutcome struct {
			stdout []byte
			stderr []byte
			err    error
		}
		outcomes := make(chan broadcastOutcome, 2)
		for _, newController := range []string{owner.FormattedAddress(), receiver.FormattedAddress()} {
			command := network.Chain.Validators[0].TxCommand(
				attacker.KeyName(),
				append(
					negativeTxCommand("nft", "update-controller", classID, newController),
					"--account-number", accountResponse.Info.AccountNumber,
					"--sequence", strconv.FormatUint(sequence, 10),
				)...,
			)
			go func(command []string) {
				stdout, stderr, execErr := network.Chain.Validators[0].Exec(ctx, command, network.Chain.Config().Env)
				outcomes <- broadcastOutcome{stdout: stdout, stderr: stderr, err: execErr}
			}(command)
		}

		type txResponse struct {
			Height    string            `json:"height"`
			TxHash    string            `json:"txhash"`
			Codespace string            `json:"codespace"`
			Code      uint32            `json:"code"`
			RawLog    string            `json:"raw_log"`
			Events    []harness.TxEvent `json:"events"`
		}
		responses := make([]txResponse, 0, 2)
		for range 2 {
			outcome := <-outcomes
			require.NoError(t, outcome.err, string(outcome.stderr))
			var response txResponse
			require.NoError(t, json.Unmarshal(outcome.stdout, &response), string(outcome.stdout))
			responses = append(responses, response)
		}
		var accepted *txResponse
		var rejected *txResponse
		for i := range responses {
			switch responses[i].Code {
			case 0:
				accepted = &responses[i]
			case 32:
				rejected = &responses[i]
			}
		}
		require.NotNil(t, accepted, "one same-sequence transaction must pass CheckTx")
		require.NotNil(t, rejected, "one same-sequence transaction must fail CheckTx")
		require.Equal(t, "sdk", rejected.Codespace)
		require.Equal(t, "0", rejected.Height)
		require.Contains(t, rejected.RawLog, "account sequence mismatch")

		startHeight, err := network.Chain.Height(ctx)
		require.NoError(t, err)
		require.NoError(t, network.WaitForFullNode(ctx, startHeight+2))
		committedJSON, err := network.FullNodeCLIQuery(ctx, "competing-sequence-committed", "tx", accepted.TxHash)
		require.NoError(t, err)
		var committed txResponse
		require.NoError(t, json.Unmarshal(committedJSON, &committed))
		require.Equal(t, accepted.TxHash, committed.TxHash)
		require.NotEqual(t, "0", committed.Height)
		require.Equal(t, "sdk", committed.Codespace)
		require.EqualValues(t, 4, committed.Code)
		require.Contains(t, committed.RawLog, "does not control class")
		assertFailureHasNoNFTEvents(t, &harness.TxResult{Events: committed.Events})

		after := captureNegativeNFTSnapshot(t, ctx, network, "competing-sequence-after", classID, []string{negativePrimaryNFTID}, trackedAddresses)
		require.Equal(t, before, after)
	})

	t.Run("ante failures stop at CheckTx and preserve application state", func(t *testing.T) {
		currentHeight, err := network.Chain.Height(ctx)
		require.NoError(t, err)
		accountJSON, err := network.FullNodeCLIQuery(
			ctx, "checktx-stale-sequence-account", "auth", "account-info", creator.FormattedAddress(),
		)
		require.NoError(t, err)
		var accountResponse struct {
			Info struct {
				Sequence string `json:"sequence"`
			} `json:"info"`
		}
		require.NoError(t, json.Unmarshal(accountJSON, &accountResponse))
		currentSequence, err := strconv.ParseUint(accountResponse.Info.Sequence, 10, 64)
		require.NoError(t, err)
		require.Positive(t, currentSequence, "stale-sequence probe requires a nonzero committed sequence")
		staleSequence := currentSequence - 1
		for _, tc := range []struct {
			name      string
			codespace string
			code      uint32
			command   []string
		}{
			{
				name: "out of gas", codespace: "sdk", code: 11,
				command: append([]string{"nft", "update-controller", classID, receiver.FormattedAddress()}, "--gas", "1", "--broadcast-mode", "sync"),
			},
			{
				name: "timeout height", codespace: "sdk", code: 30,
				command: append(negativeTxCommand("nft", "update-controller", classID, receiver.FormattedAddress()), "--timeout-height", fmt.Sprint(currentHeight)),
			},
			{
				name: "insufficient fee", codespace: "sdk", code: 13,
				command: []string{"nft", "update-controller", classID, receiver.FormattedAddress(), "--gas", "500000", "--fees", "0umed", "--broadcast-mode", "sync"},
			},
			{
				name: "fee exceeds balance", codespace: "sdk", code: 5,
				command: []string{"nft", "update-controller", classID, receiver.FormattedAddress(), "--gas", "500000", "--fees", "999999999999umed", "--broadcast-mode", "sync"},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				before := captureNegativeNFTSnapshot(t, ctx, network, "checktx-before-"+slugNegativeStep(tc.name), classID, []string{negativePrimaryNFTID}, trackedAddresses)
				result, txErr := network.BroadcastTxExpectCheckTxFailure(
					ctx,
					"checktx-"+slugNegativeStep(tc.name),
					network.Chain.Validators[0],
					creator.KeyName(),
					tc.codespace,
					tc.code,
					tc.command...,
				)
				require.NoError(t, txErr)
				require.NotNil(t, result, "CheckTx rejection must return its ABCI response")
				require.Equal(t, tc.codespace, result.Codespace)
				require.Equal(t, tc.code, result.Code)
				require.Zero(t, result.HeightInt64(), "CheckTx-rejected transaction must not commit")
				after := captureNegativeNFTSnapshot(t, ctx, network, "checktx-after-"+slugNegativeStep(tc.name), classID, []string{negativePrimaryNFTID}, trackedAddresses)
				require.Equal(t, before, after)
			})
		}

		t.Run("stale account sequence", func(t *testing.T) {
			before := captureNegativeNFTSnapshot(t, ctx, network, "checktx-before-stale-account-sequence", classID, []string{negativePrimaryNFTID}, trackedAddresses)
			var message []byte
			message = appendRawProtoString(message, 1, classID)
			message = appendRawProtoString(message, 2, creator.FormattedAddress())
			message = appendRawProtoString(message, 3, receiver.FormattedAddress())
			result, rawErr := network.BroadcastRawTx(ctx, "checktx-stale-account-sequence", harness.RawTxRequest{
				Signer: creator,
				Message: harness.RawProtoMessage{
					TypeURL: "/panacea.nft.v1.MsgUpdateControllerRequest",
					Value:   message,
				},
				GasLimit:  500_000,
				FeeAmount: sdkmath.NewInt(2_500_000),
				Sequence:  &staleSequence,
			})
			require.NoError(t, rawErr)
			require.NotNil(t, result)
			require.Equal(t, "sdk", result.Codespace)
			require.EqualValues(t, 32, result.Code)
			require.Zero(t, result.HeightInt64())
			require.Contains(t, result.RawLog, "account sequence mismatch")
			after := captureNegativeNFTSnapshot(t, ctx, network, "checktx-after-stale-account-sequence", classID, []string{negativePrimaryNFTID}, trackedAddresses)
			require.Equal(t, before, after)
		})
	})

	t.Run("malformed and unknown Any JSON never reaches CheckTx", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			data string
		}{
			{name: "malformed json", data: `{"@type":`},
			{name: "unknown type url", data: `{"@type":"/panacea.nft.v1.UnknownNFTData","name":"x"}`},
		} {
			t.Run(tc.name, func(t *testing.T) {
				before := captureNegativeNFTSnapshot(t, ctx, network, "client-before-"+slugNegativeStep(tc.name), classID, []string{negativePrimaryNFTID}, trackedAddresses)
				err := network.BroadcastTxExpectCLIRejection(
					ctx,
					"client-"+slugNegativeStep(tc.name),
					network.Chain.Validators[0],
					controller.KeyName(),
					"NFT data",
					negativeTxCommand("nft", "mint", classID, "bad-any.1", owner.FormattedAddress(), "--data", tc.data)...,
				)
				require.NoError(t, err)
				after := captureNegativeNFTSnapshot(t, ctx, network, "client-after-"+slugNegativeStep(tc.name), classID, []string{negativePrimaryNFTID}, trackedAddresses)
				require.Equal(t, before, after)
			})
		}
	})
}

func buildAndFundNegativeIntegrityWallet(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	keyName string,
) ibc.Wallet {
	t.Helper()
	entropy := sha256.Sum256([]byte("panacea negative e2e wallet: " + keyName))
	mnemonic, err := bip39.NewMnemonic(entropy[:])
	require.NoError(t, err)
	wallet, err := network.BuildWallet(ctx, keyName, mnemonic)
	require.NoError(t, err)
	_, err = network.BroadcastAndWaitTx(
		ctx,
		"fund-"+keyName,
		network.Chain.Validators[0],
		interchaintest.FaucetAccountKeyName,
		"bank", "send",
		interchaintest.FaucetAccountKeyName,
		wallet.FormattedAddress(),
		sdkmath.NewInt(200_000_000).String()+"umed",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	return wallet
}

func createNegativeClass(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	creator ibc.Wallet,
	localID string,
	transferPolicy string,
	revocable bool,
	maxSupply uint64,
) string {
	t.Helper()
	expectNegativeSuccess(
		t, ctx, network, "create-"+localID, creator.KeyName(),
		"nft", "create-class", localID, "Negative "+localID, "NEG", transferPolicy,
		fmt.Sprint(revocable), fmt.Sprint(maxSupply),
		"--uri", "https://example.test/classes/"+localID,
		"--uri-hash", negativeURIHash,
	)
	return creator.FormattedAddress() + ":" + localID
}

func mintNegativeNFT(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	controller ibc.Wallet,
	classID string,
	nftID string,
	recipient string,
) {
	t.Helper()
	expectNegativeSuccess(t, ctx, network, step, controller.KeyName(), negativeMintCommand(classID, nftID, recipient)...)
}

func negativeMintCommand(classID, nftID, recipient string) []string {
	return []string{
		"nft", "mint", classID, nftID, recipient,
		"--uri", "https://example.test/nfts/" + nftID,
		"--uri-hash", negativeURIHash,
		"--data", `{"@type":"/panacea.nft.v1.BasicNFTData","name":"Negative fixture"}`,
	}
}

func negativeTxCommand(command ...string) []string {
	return append(append([]string(nil), command...), "--gas", "500000", "--broadcast-mode", "sync")
}

func expectNegativeSuccess(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	keyName string,
	command ...string,
) *harness.TxResult {
	t.Helper()
	result, err := network.BroadcastAndWaitTx(
		ctx, step, network.Chain.Validators[0], keyName, negativeTxCommand(command...)...,
	)
	require.NoError(t, err)
	return result
}

func assertDeliverFailureLeavesNFTState(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	keyName string,
	wantCodespace string,
	wantCode uint32,
	classID string,
	recordIDs []string,
	trackedAddresses []string,
	command ...string,
) *harness.TxResult {
	t.Helper()
	before := captureNegativeNFTSnapshot(t, ctx, network, step+"-before", classID, recordIDs, trackedAddresses)
	result, err := network.BroadcastAndWaitTxExpectDeliverFailure(
		ctx,
		step,
		network.Chain.Validators[0],
		keyName,
		wantCodespace,
		wantCode,
		negativeTxCommand(command...)...,
	)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Positive(t, result.HeightInt64(), "DeliverTx failure must be committed")
	assertFailureHasNoNFTEvents(t, result)
	after := captureNegativeNFTSnapshot(t, ctx, network, step+"-after", classID, recordIDs, trackedAddresses)
	require.Equal(t, before, after, "failed transaction changed NFT state or derived indexes")
	return result
}

func captureNegativeNFTSnapshot(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	classID string,
	recordIDs []string,
	trackedAddresses []string,
) negativeNFTSnapshot {
	t.Helper()
	query := func(suffix string, command ...string) any {
		raw, err := network.FullNodeCLIQuery(ctx, step+"-"+suffix, command...)
		require.NoError(t, err)
		var value any
		require.NoError(t, json.Unmarshal(raw, &value))
		return value
	}

	snapshot := negativeNFTSnapshot{
		AllClasses:       query("all-classes", "nft", "classes"),
		ClassRecord:      query("class-record", "nft", "class-record", classID),
		CustomLiveNFTs:   query("custom-live", "nft", "nft-records", "--class-id", classID),
		StandardLiveNFTs: query("standard-live", "nft", "nfts", classID),
		Records:          make(map[string]any, len(recordIDs)),
		Balances:         make(map[string]uint64, len(trackedAddresses)),
		Owners:           make(map[string]string, len(recordIDs)),
	}
	for _, nftID := range recordIDs {
		snapshot.Records[nftID] = query("record-"+nftID, "nft", "nft-record", classID, nftID)
		ownerResponse, err := network.QueryNFTOwnerGRPC(ctx, step+"-owner-"+nftID, classID, nftID)
		require.NoError(t, err)
		snapshot.Owners[nftID] = ownerResponse.Owner
	}
	supply, err := network.QueryNFTSupplyGRPC(ctx, step+"-supply", classID)
	require.NoError(t, err)
	snapshot.Supply = supply.Amount
	for _, address := range trackedAddresses {
		balance, err := network.QueryNFTBalanceGRPC(ctx, step+"-balance-"+address, classID, address)
		require.NoError(t, err)
		snapshot.Balances[address] = balance.Amount
	}
	return snapshot
}

func assertNFTRecordAbsent(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	classID string,
	nftID string,
) {
	t.Helper()
	_, err := network.FullNodeRESTGet(
		ctx,
		&http.Client{Timeout: 5 * time.Second},
		step,
		"/panacea/nft/v1/nfts/"+strings.Replace(classID, ":", "%3A", 1)+"/"+nftID,
	)
	require.ErrorContains(t, err, "HTTP 404")
}

func assertFailureHasNoNFTEvents(t *testing.T, result *harness.TxResult) {
	t.Helper()
	for _, eventType := range []string{
		"panacea.nft.v1.EventClassCreated",
		"panacea.nft.v1.EventControllerUpdated",
		"cosmos.nft.v1beta1.EventMint",
		"cosmos.nft.v1beta1.EventSend",
		"panacea.nft.v1.EventNFTRevoked",
		"cosmos.nft.v1beta1.EventBurn",
	} {
		_, found := result.FindEvent(eventType)
		require.False(t, found, "failed transaction emitted partial %s", eventType)
	}
}

// assertModuleCreatorCheckTxFailure uses a funded explicit fee payer so fee
// deduction cannot mask the signer-bound module-account failure. The message's
// creator is the module account while an arbitrary user key occupies its
// signer slot, so SetPubKey rejects the mismatch before application execution.
func assertModuleCreatorCheckTxFailure(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	signer ibc.Wallet,
	moduleAddress string,
	snapshotClassID string,
	recordIDs []string,
	trackedAddresses []string,
) {
	t.Helper()
	const step = "module-account-class-creator"
	before := captureNegativeNFTSnapshot(t, ctx, network, step+"-before", snapshotClassID, recordIDs, trackedAddresses)
	var message []byte
	message = appendRawProtoString(message, 1, moduleAddress)
	message = appendRawProtoString(message, 2, "module-creator")
	message = appendRawProtoString(message, 3, "Module Creator")
	message = appendRawProtoString(message, 4, "MOD")
	message = appendRawProtoUvarint(message, uint64(8<<3))
	message = appendRawProtoUvarint(message, 2)
	message = appendRawProtoUvarint(message, uint64(9<<3))
	message = appendRawProtoUvarint(message, 1)
	message = appendRawProtoUvarint(message, uint64(10<<3))
	message = appendRawProtoUvarint(message, 1)
	result, err := network.BroadcastRawTx(ctx, step, harness.RawTxRequest{
		Signer:               signer,
		SignerAccountAddress: moduleAddress,
		FeePayer:             signer,
		Message: harness.RawProtoMessage{
			TypeURL: "/panacea.nft.v1.MsgCreateClassRequest",
			Value:   message,
		},
		GasLimit:  500_000,
		FeeAmount: sdkmath.NewInt(2_500_000),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Zero(t, result.HeightInt64())
	require.Equal(t, "sdk", result.Codespace)
	require.EqualValues(t, 8, result.Code)
	require.Contains(t, result.RawLog, "pubKey does not match signer address")
	assertFailureHasNoNFTEvents(t, result)

	after := captureNegativeNFTSnapshot(t, ctx, network, step+"-after", snapshotClassID, recordIDs, trackedAddresses)
	require.Equal(t, before, after)
}

func slugNegativeStep(value string) string {
	value = strings.ToLower(value)
	value = strings.NewReplacer(" ", "-", "/", "-", "_", "-").Replace(value)
	return value
}

func appendRawProtoBytes(destination []byte, fieldNumber int, value []byte) []byte {
	destination = appendRawProtoUvarint(destination, uint64(fieldNumber<<3|2))
	destination = appendRawProtoUvarint(destination, uint64(len(value)))
	return append(destination, value...)
}

func appendRawProtoString(destination []byte, fieldNumber int, value string) []byte {
	return appendRawProtoBytes(destination, fieldNumber, []byte(value))
}

func appendRawProtoUvarint(destination []byte, value uint64) []byte {
	for value >= 0x80 {
		destination = append(destination, byte(value)|0x80)
		value >>= 7
	}
	return append(destination, byte(value))
}
