package types_test

import (
	"os"
	"testing"

	"github.com/medibloc/panacea-core/x/did/internal/secp256k1util"

	"github.com/medibloc/panacea-core/x/did/types"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMain(m *testing.M) {
	sdk.GetConfig().SetBech32PrefixForAccount("panacea", "panaceapub")
	os.Exit(m.Run())
}

func TestMsgCreateDID(t *testing.T) {
	doc := newDIDDocument()
	sig := []byte("my-sig")
	fromAddr := getFromAddress(t)

	msg := types.NewMsgCreateDID(doc.ID, doc, doc.VerificationMethods[0].ID, sig, fromAddr)
	require.Equal(t, doc.ID, msg.DID)
	require.Equal(t, doc, msg.Document)
	require.Equal(t, doc.VerificationMethods[0].ID, msg.VerificationMethodID)
	require.Equal(t, sig, msg.Signature)
	require.Equal(t, fromAddr, msg.FromAddress)

	require.Equal(t, types.RouterKey, msg.Route())
	require.Equal(t, "create_did", msg.Type())
	require.Nil(t, msg.ValidateBasic())
	require.Equal(t, 1, len(msg.GetSigners()))
	require.Equal(t, fromAddr, msg.GetSigners()[0])

	require.Equal(t,
		`{"type":"did/MsgCreateDID","value":{"did":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm","document":{"@context":"https://www.w3.org/ns/did/v1","assertionMethod":[{"controller":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm","id":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key2","publicKeyBase58":"qoRmLNBEXoaKDE8dKffMq2DBNxacTEfvbKRuFrccYW1b","type":"EcdsaSecp256k1VerificationKey2019"}],"authentication":["did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key1"],"id":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm","service":[{"id":"service1","serviceEndpoint":"https://example.org","type":"LinkedDomains"}],"verificationMethod":[{"controller":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm","id":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key1","publicKeyBase58":"qoRmLNBEXoaKDE8dKffMq2DBNxacTEfvbKRuFrccYW1b","type":"EcdsaSecp256k1VerificationKey2019"}]},"from_address":"panacea154p6kyu9kqgvcmq63w3vpn893ssy6anpu8ykfq","signature":"bXktc2ln","verification_method_id":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key1"}}`,
		string(msg.GetSignBytes()),
	)
}

func TestMsgUpdateDID(t *testing.T) {
	doc := newDIDDocument()
	sig := []byte("my-sig")
	fromAddr := getFromAddress(t)

	msg := types.NewMsgUpdateDID(doc.ID, doc, doc.VerificationMethods[0].ID, sig, fromAddr)
	require.Equal(t, doc.ID, msg.DID)
	require.Equal(t, doc, msg.Document)
	require.Equal(t, doc.VerificationMethods[0].ID, msg.VerificationMethodID)
	require.Equal(t, sig, msg.Signature)
	require.Equal(t, fromAddr, msg.FromAddress)

	require.Equal(t, types.RouterKey, msg.Route())
	require.Equal(t, "update_did", msg.Type())
	require.Nil(t, msg.ValidateBasic())
	require.Equal(t, 1, len(msg.GetSigners()))
	require.Equal(t, fromAddr, msg.GetSigners()[0])

	require.Equal(t,
		`{"type":"did/MsgUpdateDID","value":{"did":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm","document":{"@context":"https://www.w3.org/ns/did/v1","assertionMethod":[{"controller":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm","id":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key2","publicKeyBase58":"qoRmLNBEXoaKDE8dKffMq2DBNxacTEfvbKRuFrccYW1b","type":"EcdsaSecp256k1VerificationKey2019"}],"authentication":["did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key1"],"id":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm","service":[{"id":"service1","serviceEndpoint":"https://example.org","type":"LinkedDomains"}],"verificationMethod":[{"controller":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm","id":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key1","publicKeyBase58":"qoRmLNBEXoaKDE8dKffMq2DBNxacTEfvbKRuFrccYW1b","type":"EcdsaSecp256k1VerificationKey2019"}]},"from_address":"panacea154p6kyu9kqgvcmq63w3vpn893ssy6anpu8ykfq","signature":"bXktc2ln","verification_method_id":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key1"}}`,
		string(msg.GetSignBytes()),
	)
}

func TestDeactivateDID(t *testing.T) {
	doc := newDIDDocument()
	sig := []byte("my-sig")
	fromAddr := getFromAddress(t)

	msg := types.NewMsgDeactivateDID(doc.ID, doc.VerificationMethods[0].ID, sig, fromAddr)
	require.Equal(t, doc.ID, msg.DID)
	require.Equal(t, doc.VerificationMethods[0].ID, msg.VerificationMethodID)
	require.Equal(t, sig, msg.Signature)
	require.Equal(t, fromAddr, msg.FromAddress)

	require.Equal(t, types.RouterKey, msg.Route())
	require.Equal(t, "deactivate_did", msg.Type())
	require.Nil(t, msg.ValidateBasic())
	require.Equal(t, 1, len(msg.GetSigners()))
	require.Equal(t, fromAddr, msg.GetSigners()[0])

	require.Equal(t,
		`{"type":"did/MsgDeactivateDID","value":{"did":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm","from_address":"panacea154p6kyu9kqgvcmq63w3vpn893ssy6anpu8ykfq","signature":"bXktc2ln","verification_method_id":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key1"}}`,
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
	verificationMethods := []types.VerificationMethod{
		types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, secp256k1util.PubKeyBytes(pubKey)),
	}
	authentications := []types.VerificationRelationship{
		types.NewVerificationRelationship(verificationMethods[0].ID),
	}
	assertionMethods := []types.VerificationRelationship{
		types.NewVerificationRelationshipDedicated(
			types.NewVerificationMethod(
				types.NewVerificationMethodID(did, "key2"),
				types.ES256K_2019, did, secp256k1util.PubKeyBytes(pubKey),
			),
		),
	}
	services := []types.Service{
		types.NewService("service1", "LinkedDomains", "https://example.org"),
	}

	return types.NewDIDDocument(
		did,
		types.WithVerificationMethods(verificationMethods),
		types.WithAuthentications(authentications),
		types.WithAssertionMethods(assertionMethods),
		types.WithServices(services),
	)
}
