package types_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ariesdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/medibloc/panacea-core/v2/x/did/internal/secp256k1util"
	"github.com/medibloc/panacea-core/v2/x/did/types"
)

func TestMain(m *testing.M) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("panacea", "panaceapub")
	config.Seal()

	os.Exit(m.Run())
}

func TestNewDID(t *testing.T) {
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))

	did := types.NewDID(pubKey)
	regex := fmt.Sprintf("^did:panacea:[%s]{32,44}$", types.Base58Charset)
	require.Regexp(t, regex, did)
}

func TestValidateDID(t *testing.T) {
	str := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	did, err := types.ValidateDID(str)
	require.NoError(t, err)
	require.EqualValues(t, str, did)

	str = "did:panacea:"
	_, err = types.ValidateDID(str)
	require.ErrorIs(t, types.ErrInvalidDID, err)

	str = "did:panacea"
	_, err = types.ValidateDID(str)
	require.ErrorIs(t, types.ErrInvalidDID, err)

	str = "invalid:panacea:abcdefg123"
	_, err = types.ValidateDID(str)
	require.ErrorIs(t, types.ErrInvalidDID, err)
}

func TestNewDIDDocument(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")

	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)
	authentication := types.NewVerification(verificationMethod, ariesdid.Authentication)
	service := types.NewService("service1", "LinkedDomains", "https://service.org")
	cratedTime := time.Now()

	document := types.NewDocument(did,
		ariesdid.WithVerificationMethod([]ariesdid.VerificationMethod{verificationMethod}),
		ariesdid.WithAuthentication([]ariesdid.Verification{authentication}),
		ariesdid.WithService([]ariesdid.Service{service}),
		ariesdid.WithCreatedTime(cratedTime))

	require.Equal(t, did, document.ID)
	require.EqualValues(t, verificationMethod, document.VerificationMethod[0])
	require.EqualValues(t, authentication, document.Authentication[0])
	require.Empty(t, document.AssertionMethod)
	require.Empty(t, document.KeyAgreement)
	require.Empty(t, document.CapabilityInvocation)
	require.Empty(t, document.CapabilityDelegation)
	require.EqualValues(t, service, document.Service[0])
}

func TestDIDDocumentEmpty(t *testing.T) {
	require.False(t, getValidDIDDocument().Empty())
	require.True(t, types.DIDDocument{}.Empty())
}

func TestDIDDocumentEmptyDID(t *testing.T) {
	document := types.NewDocument("")

	_, err := types.NewDIDDocument(document, "test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "id is required")
}

func TestDIDDocumentWithSeqEmpty(t *testing.T) {
	document := getValidDIDDocument()
	require.False(t, types.NewDIDDocumentWithSeq(&document, types.InitialSequence).Empty())
	require.True(t, types.DIDDocumentWithSeq{}.Empty())
}

func TestDIDDocumentWithSeqDeactivate(t *testing.T) {
	document := getValidDIDDocument()
	docWithSeq := types.NewDIDDocumentWithSeq(&document, types.InitialSequence)
	deactivated := docWithSeq.Deactivate(types.InitialSequence + 1)
	require.True(t, deactivated.Deactivated())
	require.False(t, deactivated.Empty())
	require.True(t, deactivated.Valid())
}

func getValidDIDDocument() types.DIDDocument {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)
	authentication := types.NewVerification(verificationMethod, ariesdid.Authentication)

	document := types.NewDocument(did,
		ariesdid.WithVerificationMethod([]ariesdid.VerificationMethod{verificationMethod}),
		ariesdid.WithAuthentication([]ariesdid.Verification{authentication}))

	didDocument, _ := types.NewDIDDocument(document, "aries-framework-go@v0.1.8")
	return didDocument
}
