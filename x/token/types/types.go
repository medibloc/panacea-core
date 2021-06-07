package types

import (
	"fmt"
	"regexp"
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/types/assets"
)

const (
	maxNameLen = 32

	// Cosmos SDK restricts the length of denominations to [3,16].
	// Also, we need to reserve 3 chars for the random suffix. So, [3,13].
	regexShortSymbolStr = `[A-Z][A-Z0-9]{2,12}`
	symbolSuffixLen     = 3
)

var (
	regexShortSymbol = regexp.MustCompile(fmt.Sprintf("^%s$", regexShortSymbolStr))
	regexSymbol      = regexp.MustCompile(fmt.Sprintf(`^%s-[A-Z0-9]{3}$`, regexShortSymbolStr))
	maxTotalSupply   = sdk.NewInt(90000000000000000)
)

func NewSymbol(shortSymbol string, txHash string) string {
	return fmt.Sprintf(
		"%s-%s",
		strings.ToUpper(string(shortSymbol)),
		strings.ToUpper(txHash[:symbolSuffixLen]),
	)
}

func validateName(name string) error {
	if len(name) <= 0 || len(name) > maxNameLen {
		return ErrInvalidName
	}
	return nil
}

func validateShortSymbol(sym string) error {
	if strings.EqualFold(sym, assets.MicroMedDenom[1:]) {
		return ErrSymbolNotAllowed
	}
	if !regexShortSymbol.MatchString(strings.ToUpper(sym)) {
		return ErrInvalidSymbol
	}
	return nil
}

func validateSymbol(sym string) error {
	if !regexSymbol.MatchString(strings.ToUpper(sym)) {
		return ErrInvalidSymbol
	}

	shortSymbol := strings.Split(sym, "-")[0]
	return validateShortSymbol(shortSymbol)
}

func validateTotalSupplyAmount(amount sdk.Int) error {
	if !amount.IsPositive() || !amount.LTE(maxTotalSupply) {
		return ErrInvalidTotalSupply
	}
	return nil
}

func validateToken(t *Token) error {
	if err := validateName(t.Name); err != nil {
		return err
	}
	if err := validateSymbol(t.Symbol); err != nil {
		return err
	}
	if err := validateTotalSupplyAmount(t.TotalSupply.Amount); err != nil {
		return err
	}
	if GetMicroDenom(t.Symbol) != t.TotalSupply.Denom {
		return ErrInvalidDenom
	}
	if _, err := sdk.AccAddressFromBech32(t.OwnerAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address (%s)", err)
	}
	return nil
}

func GetMicroDenom(symbol string) string {
	withoutDash := strings.ReplaceAll(symbol, "-", "")
	return fmt.Sprintf("u%s", strings.ToLower(withoutDash))
}
