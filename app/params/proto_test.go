package params

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	gogoproto "github.com/cosmos/gogoproto/proto"
	aoltypes "github.com/medibloc/panacea-core/v2/x/aol/types"
	didtypes "github.com/medibloc/panacea-core/v2/x/did/types"
	pnfttypes "github.com/medibloc/panacea-core/v2/x/pnft/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestCustomMsgsRetainValidateBasic(t *testing.T) {
	testCases := []struct {
		name string
		msg  sdk.Msg
	}{
		{name: "aol/CreateTopic", msg: &aoltypes.MsgCreateTopicRequest{}},
		{name: "aol/AddWriter", msg: &aoltypes.MsgAddWriterRequest{}},
		{name: "aol/DeleteWriter", msg: &aoltypes.MsgDeleteWriterRequest{}},
		{name: "aol/AddRecord", msg: &aoltypes.MsgAddRecordRequest{}},
		{name: "did/CreateDID", msg: &didtypes.MsgCreateDIDRequest{}},
		{name: "did/UpdateDID", msg: &didtypes.MsgUpdateDIDRequest{}},
		{name: "did/DeactivateDID", msg: &didtypes.MsgDeactivateDIDRequest{}},
		{name: "pnft/CreateDenom", msg: &pnfttypes.MsgCreateDenomRequest{}},
		{name: "pnft/UpdateDenom", msg: &pnfttypes.MsgUpdateDenomRequest{}},
		{name: "pnft/DeleteDenom", msg: &pnfttypes.MsgDeleteDenomRequest{}},
		{name: "pnft/TransferDenom", msg: &pnfttypes.MsgTransferDenomRequest{}},
		{name: "pnft/MintPNFT", msg: &pnfttypes.MsgMintPNFTRequest{}},
		{name: "pnft/TransferPNFT", msg: &pnfttypes.MsgTransferPNFTRequest{}},
		{name: "pnft/BurnPNFT", msg: &pnfttypes.MsgBurnPNFTRequest{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Implements(t, (*sdk.HasValidateBasic)(nil), tc.msg)

			validator := tc.msg.(sdk.HasValidateBasic)
			var err error
			require.NotPanics(t, func() {
				err = validator.ValidateBasic()
			})
			require.Error(t, err)
		})
	}
}

func TestCustomMsgLegacyAminoJSONSignBytesEquivalence(t *testing.T) {
	configureTestBech32()

	encodingConfig := MakeEncodingConfig(
		WithCustomGetSigners(aoltypes.CustomGetSigners()...),
		WithAminoJSONEncoderModifiers(didtypes.AminoJSONEncoderModifiers()...),
	)
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	aoltypes.RegisterCodec(encodingConfig.Amino)
	didtypes.RegisterCodec(encodingConfig.Amino)
	pnfttypes.RegisterCodec(encodingConfig.Amino)
	aoltypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	didtypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	pnfttypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	require.NoError(t, encodingConfig.InterfaceRegistry.SigningContext().Validate())

	legacytx.RegressionTestingAminoCodec = encodingConfig.Amino
	t.Cleanup(func() {
		legacytx.RegressionTestingAminoCodec = nil
	})

	addr1 := sdk.AccAddress(repeatedBytes(1)).String()
	addr2 := sdk.AccAddress(repeatedBytes(2)).String()
	did, verificationMethodID, doc := validDIDDocument(t)
	complexDID, complexVerificationMethodID, complexDoc := complexDIDDocument(t)

	testCases := []struct {
		name    string
		msg     sdk.Msg
		signers []signerCase
	}{
		{
			name:    "aol/CreateTopic",
			msg:     &aoltypes.MsgCreateTopicRequest{TopicName: "topic", Description: "desc", OwnerAddress: addr1},
			signers: []signerCase{newSignerCase(addr1, 0)},
		},
		{
			name:    "aol/AddWriter",
			msg:     &aoltypes.MsgAddWriterRequest{TopicName: "topic", Moniker: "writer", Description: "desc", WriterAddress: addr2, OwnerAddress: addr1},
			signers: []signerCase{newSignerCase(addr1, 0)},
		},
		{
			name:    "aol/DeleteWriter",
			msg:     &aoltypes.MsgDeleteWriterRequest{TopicName: "topic", WriterAddress: addr2, OwnerAddress: addr1},
			signers: []signerCase{newSignerCase(addr1, 0)},
		},
		{
			name:    "aol/AddRecord without fee payer",
			msg:     &aoltypes.MsgAddRecordRequest{TopicName: "topic", Key: []byte("key"), Value: []byte("value"), WriterAddress: addr2, OwnerAddress: addr1},
			signers: []signerCase{newSignerCase(addr2, 0)},
		},
		{
			name: "aol/AddRecord with fee payer",
			msg:  &aoltypes.MsgAddRecordRequest{TopicName: "topic", Key: []byte("key"), Value: []byte("value"), WriterAddress: addr2, OwnerAddress: addr1, FeePayerAddress: addr1},
			signers: []signerCase{
				newSignerCase(addr1, 0),
				newSignerCase(addr2, 1),
			},
		},
		{
			name:    "did/CreateDID",
			msg:     &didtypes.MsgCreateDIDRequest{Did: did, Document: doc, VerificationMethodId: verificationMethodID, Signature: []byte("signature"), FromAddress: addr1},
			signers: []signerCase{newSignerCase(addr1, 0)},
		},
		{
			name:    "did/UpdateDID",
			msg:     &didtypes.MsgUpdateDIDRequest{Did: did, Document: doc, VerificationMethodId: verificationMethodID, Signature: []byte("signature"), FromAddress: addr1},
			signers: []signerCase{newSignerCase(addr1, 0)},
		},
		{
			name:    "did/CreateDID with JSON-LD document edge cases",
			msg:     &didtypes.MsgCreateDIDRequest{Did: complexDID, Document: complexDoc, VerificationMethodId: complexVerificationMethodID, Signature: []byte("signature"), FromAddress: addr1},
			signers: []signerCase{newSignerCase(addr1, 0)},
		},
		{
			name:    "did/UpdateDID with JSON-LD document edge cases",
			msg:     &didtypes.MsgUpdateDIDRequest{Did: complexDID, Document: complexDoc, VerificationMethodId: complexVerificationMethodID, Signature: []byte("signature"), FromAddress: addr1},
			signers: []signerCase{newSignerCase(addr1, 0)},
		},
		{
			name:    "did/DeactivateDID",
			msg:     &didtypes.MsgDeactivateDIDRequest{Did: did, VerificationMethodId: verificationMethodID, Signature: []byte("signature"), FromAddress: addr1},
			signers: []signerCase{newSignerCase(addr1, 0)},
		},
		{
			name:    "pnft/CreateDenom",
			msg:     &pnfttypes.MsgCreateDenomRequest{Id: "denom", Name: "name", Symbol: "symbol", Description: "desc", Uri: "uri", UriHash: "hash", Data: "data", Creator: addr1},
			signers: []signerCase{newSignerCase(addr1, 0)},
		},
		{
			name:    "pnft/UpdateDenom",
			msg:     &pnfttypes.MsgUpdateDenomRequest{Id: "denom", Name: "name", Symbol: "symbol", Description: "desc", Uri: "uri", UriHash: "hash", Data: "data", Updater: addr1},
			signers: []signerCase{newSignerCase(addr1, 0)},
		},
		{
			name:    "pnft/DeleteDenom",
			msg:     &pnfttypes.MsgDeleteDenomRequest{Id: "denom", Remover: addr1},
			signers: []signerCase{newSignerCase(addr1, 0)},
		},
		{
			name:    "pnft/TransferDenom",
			msg:     &pnfttypes.MsgTransferDenomRequest{Id: "denom", Sender: addr1, Receiver: addr2},
			signers: []signerCase{newSignerCase(addr1, 0)},
		},
		{
			name:    "pnft/MintPNFT",
			msg:     &pnfttypes.MsgMintPNFTRequest{DenomId: "denom", Id: "pnft", Name: "name", Description: "desc", Uri: "uri", UriHash: "hash", Data: "data", Creator: addr1},
			signers: []signerCase{newSignerCase(addr1, 0)},
		},
		{
			name:    "pnft/TransferPNFT",
			msg:     &pnfttypes.MsgTransferPNFTRequest{DenomId: "denom", Id: "pnft", Sender: addr1, Receiver: addr2},
			signers: []signerCase{newSignerCase(addr1, 0)},
		},
		{
			name:    "pnft/BurnPNFT",
			msg:     &pnfttypes.MsgBurnPNFTRequest{DenomId: "denom", Id: "pnft", Burner: addr1},
			signers: []signerCase{newSignerCase(addr1, 0)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, signer := range tc.signers {
				legacySignBytes, newSignBytes := legacyAndAminoJSONSignBytes(t, encodingConfig, tc.msg, signer)
				require.Equal(t, string(legacySignBytes), string(newSignBytes))
			}
		})
	}
}

func TestDIDMsgAnnotatedSigners(t *testing.T) {
	configureTestBech32()

	encodingConfig := MakeEncodingConfig(
		WithCustomGetSigners(aoltypes.CustomGetSigners()...),
		WithAminoJSONEncoderModifiers(didtypes.AminoJSONEncoderModifiers()...),
	)
	didtypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	require.NoError(t, encodingConfig.InterfaceRegistry.SigningContext().Validate())

	fromAddress := sdk.AccAddress(repeatedBytes(1)).String()
	expectedSigner := sdk.MustAccAddressFromBech32(fromAddress)

	testCases := []protoreflect.FullName{
		"panacea.did.v2.MsgCreateDIDRequest",
		"panacea.did.v2.MsgUpdateDIDRequest",
		"panacea.did.v2.MsgDeactivateDIDRequest",
	}

	for _, typeName := range testCases {
		t.Run(string(typeName), func(t *testing.T) {
			signers := annotatedSigners(t, encodingConfig, typeName, "from_address", fromAddress)
			require.Equal(t, [][]byte{expectedSigner}, signers)
		})
	}
}

type signerCase struct {
	address       string
	accountNumber uint64
	sequence      uint64
}

func newSignerCase(address string, offset uint64) signerCase {
	return signerCase{
		address:       address,
		accountNumber: 7 + offset,
		sequence:      11 + offset,
	}
}

func legacyAndAminoJSONSignBytes(t *testing.T, encodingConfig EncodingConfig, msg sdk.Msg, signer signerCase) ([]byte, []byte) {
	t.Helper()

	const (
		chainID       = "test-chain"
		gas           = uint64(200000)
		timeoutHeight = uint64(123)
		memo          = "memo"
	)

	fee := sdk.NewCoins(sdk.NewInt64Coin("umed", 10))
	stdFee := legacytx.StdFee{Amount: fee, Gas: gas}

	expected := legacytx.StdSignBytes(chainID, signer.accountNumber, signer.sequence, timeoutHeight, stdFee, []sdk.Msg{msg}, memo)

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	txBuilder.SetFeeAmount(fee)
	txBuilder.SetGasLimit(gas)
	txBuilder.SetMemo(memo)
	txBuilder.SetTimeoutHeight(timeoutHeight)
	require.NoError(t, txBuilder.SetMsgs(msg))

	signerData := authsigning.SignerData{
		Address:       signer.address,
		ChainID:       chainID,
		AccountNumber: signer.accountNumber,
		Sequence:      signer.sequence,
	}
	actual, err := authsigning.GetSignBytesAdapter(
		context.Background(),
		encodingConfig.TxConfig.SignModeHandler(),
		signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		signerData,
		txBuilder.GetTx(),
	)
	require.NoError(t, err)

	return expected, actual
}

func annotatedSigners(t *testing.T, encodingConfig EncodingConfig, typeName protoreflect.FullName, fieldName protoreflect.Name, address string) [][]byte {
	t.Helper()

	desc, err := gogoproto.HybridResolver.FindDescriptorByName(typeName)
	require.NoError(t, err)
	msgDesc, ok := desc.(protoreflect.MessageDescriptor)
	require.True(t, ok)

	msg := dynamicpb.NewMessage(msgDesc)
	field := msgDesc.Fields().ByName(fieldName)
	require.NotNil(t, field)
	msg.Set(field, protoreflect.ValueOfString(address))

	signers, err := encodingConfig.InterfaceRegistry.SigningContext().GetSigners(msg)
	require.NoError(t, err)
	return signers
}

func validDIDDocument(t *testing.T) (string, string, *didtypes.DIDDocument) {
	t.Helper()

	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	verificationMethodID := didtypes.NewVerificationMethodID(did, "key1")
	verificationMethod := didtypes.NewVerificationMethod(verificationMethodID, didtypes.ES256K_2019, did, []byte{1, 2, 3, 4, 5})
	service := didtypes.NewService("service1", "LinkedDomains", "https://service.org")
	doc := didtypes.NewDIDDocument(
		did,
		didtypes.WithController(did),
		didtypes.WithVerificationMethods([]*didtypes.VerificationMethod{&verificationMethod}),
		didtypes.WithAuthentications([]didtypes.VerificationRelationship{didtypes.NewVerificationRelationship(verificationMethodID)}),
		didtypes.WithAssertionMethods([]didtypes.VerificationRelationship{didtypes.NewVerificationRelationshipDedicated(verificationMethod)}),
		didtypes.WithServices([]*didtypes.Service{&service}),
	)
	require.True(t, doc.Valid())

	return did, verificationMethodID, &doc
}

func complexDIDDocument(t *testing.T) (string, string, *didtypes.DIDDocument) {
	t.Helper()

	did := "did:panacea:8Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	controllerDID := "did:panacea:9Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	verificationMethodID := didtypes.NewVerificationMethodID(did, "key1")
	verificationMethodID2 := didtypes.NewVerificationMethodID(did, "key2")
	verificationMethod := didtypes.NewVerificationMethod(verificationMethodID, didtypes.ES256K_2019, did, []byte{1, 2, 3, 4, 5})
	verificationMethod2 := didtypes.NewVerificationMethod(verificationMethodID2, didtypes.ES256K_2018, did, []byte{6, 7, 8, 9, 10})
	service1 := didtypes.NewService("service1", "LinkedDomains", "https://service1.org")
	service2 := didtypes.NewService("service2", "CredentialRepository", "https://service2.org")

	doc := didtypes.NewDIDDocument(
		did,
		didtypes.WithVerificationMethods([]*didtypes.VerificationMethod{&verificationMethod, &verificationMethod2}),
		didtypes.WithAuthentications([]didtypes.VerificationRelationship{
			didtypes.NewVerificationRelationship(verificationMethodID),
			didtypes.NewVerificationRelationshipDedicated(verificationMethod2),
		}),
		didtypes.WithAssertionMethods([]didtypes.VerificationRelationship{
			didtypes.NewVerificationRelationshipDedicated(verificationMethod),
		}),
		didtypes.WithKeyAgreements([]didtypes.VerificationRelationship{
			didtypes.NewVerificationRelationship(verificationMethodID2),
		}),
		didtypes.WithCapabilityInvocations([]didtypes.VerificationRelationship{
			didtypes.NewVerificationRelationship(verificationMethodID),
		}),
		didtypes.WithCapabilityDelegations([]didtypes.VerificationRelationship{
			didtypes.NewVerificationRelationshipDedicated(verificationMethod2),
		}),
		didtypes.WithServices([]*didtypes.Service{&service1, &service2}),
	)
	doc.Contexts = &didtypes.JSONStringOrStrings{didtypes.ContextDIDV1, "https://example.com/did/context/v1"}
	doc.Controller = &didtypes.JSONStringOrStrings{did, controllerDID}
	require.True(t, doc.Valid())

	return did, verificationMethodID, &doc
}

func configureTestBech32() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("panacea", "panaceapub")
	config.SetBech32PrefixForValidator("panaceavaloper", "panaceavaloperpub")
}

func repeatedBytes(v byte) []byte {
	bz := make([]byte, 20)
	for i := range bz {
		bz[i] = v
	}
	return bz
}
