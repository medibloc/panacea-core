package types

import (
	"os"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("panacea", "panaceapub")
	config.Seal()

	os.Exit(m.Run())
}

func TestNewSymbol(t *testing.T) {
	symbol := NewSymbol("lov", "1234567890ABCDEFGHIJKLMNOPQRSTUXYZ")
	require.Equal(t, "LOV-123", symbol)
}

func TestValidateName(t *testing.T) {
	require.NoError(t, validateName("my name"))
	require.ErrorIs(t, validateName(""), ErrInvalidName)

	maxLenName := "12345678901234567890123456789012"
	require.NoError(t, validateName(maxLenName))

	require.ErrorIs(t, validateName(maxLenName+"X"), ErrInvalidName)
}

func TestValidateShortSymbol(t *testing.T) {
	require.NoError(t, validateShortSymbol("lov"))

	// if it's the same as MED (stake coin)
	require.ErrorIs(t, validateShortSymbol("med"), ErrSymbolNotAllowed)
	require.ErrorIs(t, validateShortSymbol("MED"), ErrSymbolNotAllowed)

	require.ErrorIs(t, validateShortSymbol("2ov"), ErrInvalidSymbol)
	require.ErrorIs(t, validateShortSymbol("Lo"), ErrInvalidSymbol)

	maxLenSymbol := "A123456789012"
	require.NoError(t, validateShortSymbol(maxLenSymbol))
	require.ErrorIs(t, validateShortSymbol(maxLenSymbol+"X"), ErrInvalidSymbol)
}

func TestValidateSymbol(t *testing.T) {
	require.NoError(t, validateSymbol("LoV-0eA"))
	require.ErrorIs(t, validateSymbol("LoV0eA"), ErrInvalidSymbol)
	require.ErrorIs(t, validateSymbol("LoV-0eAB"), ErrInvalidSymbol)
}

func TestValidateTotalSupplyAmount(t *testing.T) {
	require.NoError(t, validateTotalSupplyAmount(sdk.NewInt(int64(100))))

	// if not positive
	require.ErrorIs(t, validateTotalSupplyAmount(sdk.NewInt(int64(0))), ErrInvalidTotalSupply)
	require.ErrorIs(t, validateTotalSupplyAmount(sdk.NewInt(int64(-1))), ErrInvalidTotalSupply)

	// regards to the max length
	require.NoError(t, validateTotalSupplyAmount(sdk.NewInt(int64(90000000000000000))))
	require.ErrorIs(t, validateTotalSupplyAmount(sdk.NewInt(int64(90000000000000000+1))), ErrInvalidTotalSupply)
}

func TestValidateToken(t *testing.T) {
	require.NoError(t, validateToken(&Token{
		Name:         "this is a name",
		Symbol:       "LOV-0EA",
		TotalSupply:  sdk.NewCoin("ulov0ea", sdk.NewInt(int64(100))),
		OwnerAddress: "panacea1qzl3j4srlymax4urwxhfwv50y07jh3txtml7gk",
	}))

	require.ErrorIs(t, validateToken(&Token{
		Name:         "", // invalid name
		Symbol:       "LOV-0EA",
		TotalSupply:  sdk.NewCoin("ulov0ea", sdk.NewInt(int64(100))),
		OwnerAddress: "panacea1qzl3j4srlymax4urwxhfwv50y07jh3txtml7gk",
	}), ErrInvalidName)

	require.ErrorIs(t, validateToken(&Token{
		Name:         "this is a name",
		Symbol:       "", // invalid symbol
		TotalSupply:  sdk.NewCoin("ulov0ea", sdk.NewInt(int64(100))),
		OwnerAddress: "panacea1qzl3j4srlymax4urwxhfwv50y07jh3txtml7gk",
	}), ErrInvalidSymbol)

	require.ErrorIs(t, validateToken(&Token{
		Name:         "this is a name",
		Symbol:       "LOV-0EA",
		TotalSupply:  sdk.NewCoin("ulov0ea", sdk.NewInt(int64(0))), // invalid total supply amount
		OwnerAddress: "panacea1qzl3j4srlymax4urwxhfwv50y07jh3txtml7gk",
	}), ErrInvalidTotalSupply)

	require.ErrorIs(t, validateToken(&Token{
		Name:         "this is a name",
		Symbol:       "LOV-0EA",
		TotalSupply:  sdk.NewCoin("ulov1ea", sdk.NewInt(int64(100))), // denom doesn't match with symbol
		OwnerAddress: "panacea1qzl3j4srlymax4urwxhfwv50y07jh3txtml7gk",
	}), ErrInvalidDenom)

	require.ErrorIs(t, validateToken(&Token{
		Name:         "this is a name",
		Symbol:       "LOV-0EA",
		TotalSupply:  sdk.NewCoin("ulov0ea", sdk.NewInt(int64(100))),
		OwnerAddress: "INVALID", // invalid address
	}), sdkerrors.ErrInvalidAddress)
}

func TestGetMicroDenom(t *testing.T) {
	require.Equal(t, "ulov0ea", GetMicroDenom("LOV-0EA"))
}
