package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/stretchr/testify/require"
)

func TestMsgCreateDealValidateBasic(t *testing.T) {
	consumerAddress := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()
	budget := sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000))

	msg := &MsgCreateDeal{
		DataSchema:      []string{"https://jsonld.com"},
		Budget:          &budget,
		MaxNumData:      10,
		ConsumerAddress: consumerAddress,
	}

	err := msg.ValidateBasic()
	require.NoError(t, err)
}

func TestMsgCreateDealValidateBasicEmptyValue(t *testing.T) {
	consumerAddress := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()
	budget := sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(10000))

	msg := &MsgCreateDeal{
		DataSchema:      []string{},
		Budget:          &budget,
		MaxNumData:      10,
		ConsumerAddress: consumerAddress,
	}

	err := msg.ValidateBasic()
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.ErrorContains(t, err, "no data schema")

	msg.DataSchema = []string{"https://jsonld.com"}
	msg.MaxNumData = 0
	err = msg.ValidateBasic()
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.ErrorContains(t, err, "max num of data is negative number")

	msg.MaxNumData = 10
	msg.Budget = nil
	err = msg.ValidateBasic()
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.ErrorContains(t, err, "budget is empty")

	msg.Budget = &sdk.Coin{
		Denom: assets.MicroMedDenom, Amount: sdk.NewInt(-1),
	}
	err = msg.ValidateBasic()
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.ErrorContains(t, err, "budget is not a valid Coin object")

	msg.Budget = &budget
	msg.ConsumerAddress = ""
	err = msg.ValidateBasic()
	require.ErrorIs(t, err, sdkerrors.ErrInvalidAddress)
	require.ErrorContains(t, err, "invalid consumer address")
}

func TestMsgSubmitConsentValidateBasic(t *testing.T) {
	oracleAddress := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()
	providerAddress := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()
	msg := &MsgSubmitConsent{
		Certificate: &Certificate{
			UnsignedCertificate: &UnsignedCertificate{
				Cid:             "cid",
				UniqueId:        "uniqueID",
				OracleAddress:   oracleAddress,
				DealId:          1,
				ProviderAddress: providerAddress,
				DataHash:        "dataHash",
			},
			Signature: []byte("signature"),
		},
	}

	err := msg.ValidateBasic()
	require.NoError(t, err)
}

func TestMsgSubmitConsentValidateBasicEmptyValue(t *testing.T) {
	oracleAddress := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()
	providerAddress := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()

	msg := &MsgSubmitConsent{}

	err := msg.ValidateBasic()
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.ErrorContains(t, err, "certificate is empty")

	msg.Certificate = &Certificate{}
	err = msg.ValidateBasic()
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.ErrorContains(t, err, "unsignedCertificate is empty")

	msg.Certificate.UnsignedCertificate = &UnsignedCertificate{}
	err = msg.ValidateBasic()
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.ErrorContains(t, err, "failed to validation certificate")
	require.ErrorContains(t, err, "cid is empty")

	msg.Certificate.UnsignedCertificate.Cid = "cid"
	err = msg.ValidateBasic()
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.ErrorContains(t, err, "failed to validation certificate")
	require.ErrorContains(t, err, "uniqueId is empty")

	msg.Certificate.UnsignedCertificate.UniqueId = "uniqueID"
	err = msg.ValidateBasic()
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.ErrorContains(t, err, "failed to validation certificate")
	require.ErrorContains(t, err, "oracleAddress is invalid")

	msg.Certificate.UnsignedCertificate.OracleAddress = oracleAddress
	err = msg.ValidateBasic()
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.ErrorContains(t, err, "failed to validation certificate")
	require.ErrorContains(t, err, "dealId is greater than 0")

	msg.Certificate.UnsignedCertificate.DealId = 1
	err = msg.ValidateBasic()
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.ErrorContains(t, err, "failed to validation certificate")
	require.ErrorContains(t, err, "providerAddress is invalid")

	msg.Certificate.UnsignedCertificate.ProviderAddress = providerAddress
	err = msg.ValidateBasic()
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.ErrorContains(t, err, "failed to validation certificate")
	require.ErrorContains(t, err, "dataHash is empty")
}

func TestMsgDeactivateDeal(t *testing.T) {
	requesterAddress := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()

	msg := &MsgDeactivateDeal{
		DealId:           1,
		RequesterAddress: requesterAddress,
	}

	err := msg.ValidateBasic()
	require.NoError(t, err)
}

func TestMsgDeactivateDealEmptyValue(t *testing.T) {
	requesterAddress := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()

	msg := &MsgDeactivateDeal{}
	err := msg.ValidateBasic()
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.ErrorContains(t, err, "requesterAddress is invalid")

	msg.RequesterAddress = requesterAddress
	err = msg.ValidateBasic()
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.ErrorContains(t, err, "dealId is greater than 0")
}