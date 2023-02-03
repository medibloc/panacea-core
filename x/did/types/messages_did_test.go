package types_test

import (
	"testing"

	ariesdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/did/internal/secp256k1util"
	"github.com/medibloc/panacea-core/v2/x/did/types"
)

func TestMsgCreateDID(t *testing.T) {
	doc := newDIDDocument()
	sig := []byte("my-sig")
	fromAddr := getFromAddress(t)
	document, err := ariesdid.ParseDocument(doc.Document)
	require.NoError(t, err)

	msg := types.NewMsgCreateDID(document.ID, doc, document.VerificationMethod[0].ID, sig, fromAddr.String())
	require.Equal(t, document.ID, msg.Did)
	require.Equal(t, doc, *msg.Document)
	require.Equal(t, document.VerificationMethod[0].ID, msg.VerificationMethodId)
	require.Equal(t, sig, msg.Signature)
	require.Equal(t, fromAddr.String(), msg.FromAddress)

	require.Equal(t, types.RouterKey, msg.Route())
	require.Equal(t, "create_did", msg.Type())
	require.Nil(t, msg.ValidateBasic())
	require.Equal(t, 1, len(msg.GetSigners()))
	require.Equal(t, fromAddr, msg.GetSigners()[0])
}

func getFromAddress(t *testing.T) sdk.AccAddress {
	fromAddr, err := sdk.AccAddressFromBech32("panacea154p6kyu9kqgvcmq63w3vpn893ssy6anpu8ykfq")
	require.NoError(t, err)
	return fromAddr
}

func newDIDDocument() types.DIDDocument {
	did, _ := types.ValidateDID("did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm")
	pubKey, _ := secp256k1util.PubKeyFromBase58("qoRmLNBEXoaKDE8dKffMq2DBNxacTEfvbKRuFrccYW1b")

	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, secp256k1util.PubKeyBytes(pubKey))
	authentication := types.NewVerification(verificationMethod, ariesdid.Authentication)

	service := types.NewService("service1", "LinkedDomains", "https://example.org")

	document := types.NewDocument(did,
		ariesdid.WithVerificationMethod([]ariesdid.VerificationMethod{verificationMethod}),
		ariesdid.WithAuthentication([]ariesdid.Verification{authentication}),
		ariesdid.WithService([]ariesdid.Service{service}))

	didDocument, _ := types.NewDIDDocument(document, "aries-framework-go@v0.1.8")

	return didDocument
}
