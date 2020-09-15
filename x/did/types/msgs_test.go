package types_test

import (
	"os"
	"testing"

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

	msg := types.NewMsgCreateDID(doc.ID, doc, doc.VeriMethods[0].ID, sig, fromAddr)
	require.Equal(t, doc.ID, msg.DID)
	require.Equal(t, doc, msg.Document)
	require.Equal(t, doc.VeriMethods[0].ID, msg.VeriMethodID)
	require.Equal(t, sig, msg.Signature)
	require.Equal(t, fromAddr, msg.FromAddress)

	require.Equal(t, types.RouterKey, msg.Route())
	require.Equal(t, "create_did", msg.Type())
	require.Nil(t, msg.ValidateBasic())
	require.Equal(t, 1, len(msg.GetSigners()))
	require.Equal(t, fromAddr, msg.GetSigners()[0])

	require.Equal(t,
		`{"type":"did/MsgCreateDID","value":{"did":"did:panacea:testnet:KS5zGZt66Me8MCctZBYrP","document":{"@context":"https://www.w3.org/ns/did/v1","authentication":["did:panacea:testnet:KS5zGZt66Me8MCctZBYrP#key1"],"id":"did:panacea:testnet:KS5zGZt66Me8MCctZBYrP","verificationMethod":[{"controller":"did:panacea:testnet:KS5zGZt66Me8MCctZBYrP","id":"did:panacea:testnet:KS5zGZt66Me8MCctZBYrP#key1","publicKeyBase58":"qoRmLNBEXoaKDE8dKffMq2DBNxacTEfvbKRuFrccYW1b","type":"Secp256k1VerificationKey2018"}]},"from_address":"panacea154p6kyu9kqgvcmq63w3vpn893ssy6anpu8ykfq","signature":"bXktc2ln","verification_method_id":"did:panacea:testnet:KS5zGZt66Me8MCctZBYrP#key1"}}`,
		string(msg.GetSignBytes()),
	)
}

func TestMsgUpdateDID(t *testing.T) {
	doc := newDIDDocument()
	sig := []byte("my-sig")
	fromAddr := getFromAddress(t)

	msg := types.NewMsgUpdateDID(doc.ID, doc, doc.VeriMethods[0].ID, sig, fromAddr)
	require.Equal(t, doc.ID, msg.DID)
	require.Equal(t, doc, msg.Document)
	require.Equal(t, doc.VeriMethods[0].ID, msg.VeriMethodID)
	require.Equal(t, sig, msg.Signature)
	require.Equal(t, fromAddr, msg.FromAddress)

	require.Equal(t, types.RouterKey, msg.Route())
	require.Equal(t, "update_did", msg.Type())
	require.Nil(t, msg.ValidateBasic())
	require.Equal(t, 1, len(msg.GetSigners()))
	require.Equal(t, fromAddr, msg.GetSigners()[0])

	require.Equal(t,
		`{"type":"did/MsgUpdateDID","value":{"did":"did:panacea:testnet:KS5zGZt66Me8MCctZBYrP","document":{"@context":"https://www.w3.org/ns/did/v1","authentication":["did:panacea:testnet:KS5zGZt66Me8MCctZBYrP#key1"],"id":"did:panacea:testnet:KS5zGZt66Me8MCctZBYrP","verificationMethod":[{"controller":"did:panacea:testnet:KS5zGZt66Me8MCctZBYrP","id":"did:panacea:testnet:KS5zGZt66Me8MCctZBYrP#key1","publicKeyBase58":"qoRmLNBEXoaKDE8dKffMq2DBNxacTEfvbKRuFrccYW1b","type":"Secp256k1VerificationKey2018"}]},"from_address":"panacea154p6kyu9kqgvcmq63w3vpn893ssy6anpu8ykfq","signature":"bXktc2ln","verification_method_id":"did:panacea:testnet:KS5zGZt66Me8MCctZBYrP#key1"}}`,
		string(msg.GetSignBytes()),
	)
}

func TestDeactivateDID(t *testing.T) {
	doc := newDIDDocument()
	sig := []byte("my-sig")
	fromAddr := getFromAddress(t)

	msg := types.NewMsgDeactivateDID(doc.ID, doc.VeriMethods[0].ID, sig, fromAddr)
	require.Equal(t, doc.ID, msg.DID)
	require.Equal(t, doc.VeriMethods[0].ID, msg.VeriMethodID)
	require.Equal(t, sig, msg.Signature)
	require.Equal(t, fromAddr, msg.FromAddress)

	require.Equal(t, types.RouterKey, msg.Route())
	require.Equal(t, "deactivate_did", msg.Type())
	require.Nil(t, msg.ValidateBasic())
	require.Equal(t, 1, len(msg.GetSigners()))
	require.Equal(t, fromAddr, msg.GetSigners()[0])

	require.Equal(t,
		`{"type":"did/MsgDeactivateDID","value":{"did":"did:panacea:testnet:KS5zGZt66Me8MCctZBYrP","from_address":"panacea154p6kyu9kqgvcmq63w3vpn893ssy6anpu8ykfq","signature":"bXktc2ln","verification_method_id":"did:panacea:testnet:KS5zGZt66Me8MCctZBYrP#key1"}}`,
		string(msg.GetSignBytes()),
	)
}

func getFromAddress(t *testing.T) sdk.AccAddress {
	fromAddr, err := sdk.AccAddressFromBech32("panacea154p6kyu9kqgvcmq63w3vpn893ssy6anpu8ykfq")
	require.NoError(t, err)
	return fromAddr
}

func newDIDDocument() types.DIDDocument {
	did, _ := types.ParseDID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	veriMethodID := types.NewVeriMethodID(did, "key1")
	pubKeyBase58, _ := types.NewPubKeyFromBase58("qoRmLNBEXoaKDE8dKffMq2DBNxacTEfvbKRuFrccYW1b")
	veriMethod := types.NewVeriMethod(veriMethodID, types.ES256K, did, pubKeyBase58)
	return types.NewDIDDocument(did, veriMethod)
}
