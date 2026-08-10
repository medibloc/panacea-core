package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/go-bip39"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
)

type upgradeRawSignerTestWallet struct {
	keyName  string
	mnemonic string
	address  []byte
}

func (w upgradeRawSignerTestWallet) KeyName() string          { return w.keyName }
func (w upgradeRawSignerTestWallet) FormattedAddress() string { return "panacea1upgraderawsigner" }
func (w upgradeRawSignerTestWallet) Mnemonic() string         { return w.mnemonic }
func (w upgradeRawSignerTestWallet) Address() []byte          { return w.address }

func TestBuildUpgradeRawTxSignerSuppliesRetainedMnemonicAndMatchingAddress(t *testing.T) {
	t.Parallel()

	const keyName = "upgrade-proposer"
	suppliedMnemonics := make([]string, 0, 2)
	buildWallet := func(_ context.Context, actualKeyName, mnemonic string) (ibc.Wallet, error) {
		require.Equal(t, keyName, actualKeyName)
		require.True(t, bip39.IsMnemonicValid(mnemonic))
		suppliedMnemonics = append(suppliedMnemonics, mnemonic)
		path := hd.CreateHDPath(371, 0, 0).String()
		derived, deriveErr := hd.Secp256k1.Derive()(mnemonic, "", path)
		require.NoError(t, deriveErr)
		privateKey := hd.Secp256k1.Generate()(derived)
		return upgradeRawSignerTestWallet{
			keyName:  actualKeyName,
			mnemonic: mnemonic,
			address:  privateKey.PubKey().Address(),
		}, nil
	}
	wallet, err := buildUpgradeRawTxSigner(context.Background(), keyName, buildWallet)
	require.NoError(t, err)
	secondWallet, err := buildUpgradeRawTxSigner(context.Background(), keyName, buildWallet)
	require.NoError(t, err)
	require.Len(t, suppliedMnemonics, 2)
	require.NotEmpty(t, suppliedMnemonics[0])
	require.Equal(t, suppliedMnemonics[0], suppliedMnemonics[1], "same test key must reproduce the same mnemonic")
	require.Equal(t, suppliedMnemonics[0], wallet.Mnemonic())
	require.Equal(t, wallet.Mnemonic(), secondWallet.Mnemonic())
	require.Equal(t, wallet.Address(), secondWallet.Address())

	path := hd.CreateHDPath(371, 0, 0).String()
	derived, err := hd.Secp256k1.Derive()(wallet.Mnemonic(), "", path)
	require.NoError(t, err)
	privateKey := hd.Secp256k1.Generate()(derived)
	require.True(t, bytes.Equal(privateKey.PubKey().Address(), wallet.Address()))
}

func TestBuildUpgradeRawTxSignerRejectsMnemonicDroppingBuilder(t *testing.T) {
	t.Parallel()

	var suppliedMnemonic string
	_, err := buildUpgradeRawTxSigner(
		context.Background(),
		"upgrade-proposer",
		func(_ context.Context, keyName, mnemonic string) (ibc.Wallet, error) {
			suppliedMnemonic = mnemonic
			return upgradeRawSignerTestWallet{keyName: keyName, address: []byte{1}}, nil
		},
	)
	require.True(t, bip39.IsMnemonicValid(suppliedMnemonic))
	require.ErrorContains(t, err, "did not retain mnemonic")
	require.NotContains(t, err.Error(), suppliedMnemonic, "validation errors must not leak the mnemonic")
}

func TestLegacyPNFTEmptyStoreUsesCompleteBoundedRESTPage(t *testing.T) {
	require.Equal(
		t,
		"/panacea/pnft/v2/denoms?pagination.count_total=true&pagination.limit=100",
		legacyPNFTEmptyRESTPath(),
	)

	count, err := decodeCompleteLegacyPNFTDenomPage(json.RawMessage(
		`{"denoms":[],"pagination":{"next_key":null,"total":"0"}}`,
	))
	require.NoError(t, err)
	require.Zero(t, count)

	for _, test := range []struct {
		name     string
		response string
		contains string
	}{
		{
			name:     "missing pagination",
			response: `{"denoms":[]}`,
			contains: "missing pagination",
		},
		{
			name:     "next page",
			response: `{"denoms":[],"pagination":{"next_key":"bmV4dA==","total":"1"}}`,
			contains: "incomplete",
		},
		{
			name:     "total mismatch",
			response: `{"denoms":[],"pagination":{"next_key":null,"total":"1"}}`,
			contains: "total=1",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := decodeCompleteLegacyPNFTDenomPage(json.RawMessage(test.response))
			require.ErrorContains(t, err, test.contains)
		})
	}
}

func TestCurrentNFTEmptyQueriesUseSupportedCompleteBoundedPages(t *testing.T) {
	t.Parallel()
	require.NoError(t, validateCanonicalNFTClassIDForEmptyQuery("panacea1creator:legacy.isolation"))
	require.ErrorContains(
		t,
		validateCanonicalNFTClassIDForEmptyQuery(legacyPNFTLocalClassID),
		"creator namespace",
	)

	request := currentNFTEmptyPageRequest()
	require.Equal(t, uint64(legacyPNFTEmptyQueryLimit), request.Limit)
	require.False(t, request.CountTotal, "x/nft rejects count_total")
	require.NoError(t, validateCompleteEmptyCurrentNFTPage("classes", 0, &querytypes.PageResponse{}))

	require.ErrorContains(t, validateCompleteEmptyCurrentNFTPage("classes", 1, &querytypes.PageResponse{}), "want zero")
	require.ErrorContains(t, validateCompleteEmptyCurrentNFTPage("classes", 0, nil), "missing pagination")
	require.ErrorContains(
		t,
		validateCompleteEmptyCurrentNFTPage("classes", 0, &querytypes.PageResponse{NextKey: []byte("next")}),
		"next page",
	)

	count, err := decodeCompleteEmptyCurrentPanaceaNFTRecords(json.RawMessage(
		`{"nft_records":[],"pagination":{"next_key":null,"total":"0"}}`,
	))
	require.NoError(t, err)
	require.Zero(t, count)

	_, err = decodeCompleteEmptyCurrentPanaceaNFTRecords(json.RawMessage(`{"nft_records":[]}`))
	require.ErrorContains(t, err, "missing pagination")
	_, err = decodeCompleteEmptyCurrentPanaceaNFTRecords(json.RawMessage(
		`{"nft_records":[],"pagination":{"next_key":"bmV4dA==","total":"0"}}`,
	))
	require.ErrorContains(t, err, "incomplete")
}

func TestBuildNewRawLegacyPNFTTransferEncodesAllSignerFields(t *testing.T) {
	t.Parallel()
	encoded, err := buildNewRawLegacyPNFTTransfer("denom", "asset", "sender", "receiver")
	require.NoError(t, err)
	fields := make(map[protowire.Number]string)
	for len(encoded) > 0 {
		number, wireType, consumed := protowire.ConsumeTag(encoded)
		require.Positive(t, consumed)
		require.Equal(t, protowire.BytesType, wireType)
		encoded = encoded[consumed:]
		value, valueConsumed := protowire.ConsumeBytes(encoded)
		require.Positive(t, valueConsumed)
		fields[number] = string(value)
		encoded = encoded[valueConsumed:]
	}
	require.Equal(t, map[protowire.Number]string{
		1: "denom",
		2: "asset",
		3: "sender",
		4: "receiver",
	}, fields)
}

func TestNewLegacyPNFTFixtureUsesFutureCanonicalClassID(t *testing.T) {
	fixture, err := newLegacyPNFTFixture("panacea1creator")
	require.NoError(t, err)
	require.Equal(t, "legacy.isolation", fixture.LocalClassID)
	require.Equal(t, "panacea1creator:legacy.isolation", fixture.DenomID)
	require.Equal(t, "legacy.asset.1", fixture.PNFTID)
	require.Equal(t, fixture.Creator, fixture.Owner)

	_, err = newLegacyPNFTFixture("  ")
	require.ErrorContains(t, err, "creator")
}

func TestDecodeLegacyPNFTFixtureEvidenceAcceptsExactV221State(t *testing.T) {
	want := legacyPNFTFixture{
		DenomID: "panacea1creator:legacy.isolation",
		PNFTID:  "legacy.asset.1",
		Creator: "panacea1creator",
		Owner:   "panacea1creator",
	}

	evidence, err := decodeLegacyPNFTFixtureEvidence(
		[]byte(`{"denom":{"id":"panacea1creator:legacy.isolation","name":"Legacy Isolation","symbol":"LEGACY","owner":"panacea1creator"}}`),
		[]byte(`{"pnft":{"denom_id":"panacea1creator:legacy.isolation","id":"legacy.asset.1","name":"Legacy Asset","creator":"panacea1creator","owner":"panacea1creator","created_at":"2026-08-04T00:00:00Z"}}`),
		want,
	)
	require.NoError(t, err)
	require.Equal(t, want.DenomID, evidence.Denom.ID)
	require.Equal(t, want.PNFTID, evidence.PNFT.ID)
}

func TestDecodeLegacyPNFTFixtureEvidenceRejectsMismatchedState(t *testing.T) {
	want := legacyPNFTFixture{
		DenomID: "panacea1creator:legacy.isolation",
		PNFTID:  "legacy.asset.1",
		Creator: "panacea1creator",
		Owner:   "panacea1creator",
	}

	tests := []struct {
		name      string
		denomJSON string
		pnftJSON  string
		contains  string
	}{
		{
			name:      "denom id",
			denomJSON: `{"denom":{"id":"other","owner":"panacea1creator"}}`,
			pnftJSON:  `{"pnft":{"denom_id":"panacea1creator:legacy.isolation","id":"legacy.asset.1","creator":"panacea1creator","owner":"panacea1creator"}}`,
			contains:  "legacy denom id",
		},
		{
			name:      "pnft owner",
			denomJSON: `{"denom":{"id":"panacea1creator:legacy.isolation","owner":"panacea1creator"}}`,
			pnftJSON:  `{"pnft":{"denom_id":"panacea1creator:legacy.isolation","id":"legacy.asset.1","creator":"panacea1creator","owner":"panacea1other"}}`,
			contains:  "legacy PNFT owner",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := decodeLegacyPNFTFixtureEvidence([]byte(test.denomJSON), []byte(test.pnftJSON), want)
			require.ErrorContains(t, err, test.contains)
		})
	}
}

func TestDecodeLegacySignedTxEvidenceRequiresOldPNFTMessageAndSignature(t *testing.T) {
	raw := []byte(`{
		"body":{"messages":[{"@type":"/panacea.pnft.v2.MsgTransferPNFTRequest","denom_id":"panacea1signer:legacy.isolation","id":"legacy.asset.1","sender":"panacea1signer","receiver":"panacea1receiver"}]},
		"auth_info":{"signer_infos":[{"public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AQID"},"mode_info":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":"0"}]},
		"signatures":["BAUG"]
	}`)

	evidence, err := decodeLegacySignedTxEvidence(raw, "panacea1signer:legacy.isolation", "panacea1signer")
	require.NoError(t, err)
	require.Equal(t, "/panacea.pnft.v2.MsgTransferPNFTRequest", evidence.TypeURL)
	require.NotEmpty(t, evidence.Signature)

	_, err = decodeLegacySignedTxEvidence(
		[]byte(`{"body":{"messages":[{"@type":"/panacea.pnft.v2.MsgTransferPNFTRequest","denom_id":"panacea1signer:legacy.isolation","id":"legacy.asset.1","sender":"panacea1signer","receiver":"panacea1receiver"}]},"auth_info":{"signer_infos":[{}]},"signatures":[]}`),
		"panacea1signer:legacy.isolation",
		"panacea1signer",
	)
	require.ErrorContains(t, err, "signature")
}
