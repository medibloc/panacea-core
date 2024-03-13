package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/did/internal/secp256k1util"
	"github.com/medibloc/panacea-core/v2/x/did/types"
)

func TestMsgCreateDID(t *testing.T) {
	doc := newDIDDocument()
	sig := []byte("my-sig")
	fromAddr := getFromAddress(t)

	msg := types.NewMsgCreateDIDResponse(doc.Id, doc, doc.VerificationMethods[0].Id, sig, fromAddr.String())
	require.Equal(t, doc.Id, msg.Did)
	require.Equal(t, doc, *msg.Document)
	require.Equal(t, doc.VerificationMethods[0].Id, msg.VerificationMethodId)
	require.Equal(t, sig, msg.Signature)
	require.Equal(t, fromAddr.String(), msg.FromAddress)

	require.Equal(t, types.RouterKey, msg.Route())
	require.Equal(t, "create_did", msg.Type())
	require.Nil(t, msg.ValidateBasic())
	require.Equal(t, 1, len(msg.GetSigners()))
	require.Equal(t, fromAddr, msg.GetSigners()[0])

	// The legacy GetSignBytes() would be deprecated by cosmos-sdk soon.
	require.Equal(t, `{"did":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm","document":{"assertion_methods":[{"controller":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm","id":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key2","public_key_base58":"qoRmLNBEXoaKDE8dKffMq2DBNxacTEfvbKRuFrccYW1b","type":"EcdsaSecp256k1VerificationKey2019"}],"authentications":["did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key1"],"contexts":"https://www.w3.org/ns/did/v1","id":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm","services":[{"id":"service1","service_endpoint":"https://example.org","type":"LinkedDomains"}],"verification_methods":[{"controller":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm","id":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key1","public_key_base58":"qoRmLNBEXoaKDE8dKffMq2DBNxacTEfvbKRuFrccYW1b","type":"EcdsaSecp256k1VerificationKey2019"}]},"from_address":"panacea154p6kyu9kqgvcmq63w3vpn893ssy6anpu8ykfq","signature":"bXktc2ln","verification_method_id":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key1"}`,
		string(msg.GetSignBytes()),
	)
}

func getFromAddress(t *testing.T) sdk.AccAddress {
	fromAddr, err := sdk.AccAddressFromBech32("panacea154p6kyu9kqgvcmq63w3vpn893ssy6anpu8ykfq")
	require.NoError(t, err)
	return fromAddr
}

func newDIDDocument() types.DIDDocument {
	did, _ := types.ParseDID("did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm")
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	pubKey, _ := secp256k1util.PubKeyFromBase58("qoRmLNBEXoaKDE8dKffMq2DBNxacTEfvbKRuFrccYW1b")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, secp256k1util.PubKeyBytes(pubKey))
	verificationMethods := []*types.VerificationMethod{&verificationMethod}
	verificationRelationship := types.NewVerificationRelationship(verificationMethods[0].Id)
	authentications := []types.VerificationRelationship{verificationRelationship}
	verificationRelationshipDedicated := types.NewVerificationRelationshipDedicated(
		types.NewVerificationMethod(
			types.NewVerificationMethodID(did, "key2"),
			types.ES256K_2019, did, secp256k1util.PubKeyBytes(pubKey),
		),
	)
	assertionMethods := []types.VerificationRelationship{verificationRelationshipDedicated}
	service := types.NewService("service1", "LinkedDomains", "https://example.org")
	services := []*types.Service{&service}

	return types.NewDIDDocument(
		did,
		types.WithVerificationMethods(verificationMethods),
		types.WithAuthentications(authentications),
		types.WithAssertionMethods(assertionMethods),
		types.WithServices(services),
	)
}
