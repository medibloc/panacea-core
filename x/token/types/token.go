package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/medibloc/panacea-core/types/assets"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Token struct {
	Name         Name           `json:"name"`
	Symbol       Symbol         `json:"symbol"`
	TotalSupply  sdk.Coin       `json:"total_supply"`
	Mintable     bool           `json:"mintable"`
	OwnerAddress sdk.AccAddress `json:"owner_address"`
}

func NewToken(msg MsgIssueToken, txHash string) Token {
	symbol := newSymbol(msg.Symbol, txHash)
	return Token{
		Name:         msg.Name,
		Symbol:       symbol,
		TotalSupply:  sdk.NewCoin(symbol.MicroDenom(), msg.TotalSupply),
		Mintable:     msg.Mintable,
		OwnerAddress: msg.OwnerAddress,
	}
}

func (t Token) Empty() bool {
	return t.Name == "" && t.Symbol == "" && t.OwnerAddress.Empty()
}

func (t Token) validate() sdk.Error {
	if err := t.Name.validate(); err != nil {
		return err
	}
	if err := t.Symbol.validate(); err != nil {
		return err
	}
	if err := validateTotalSupply(t.TotalSupply); err != nil {
		return err
	}
	if t.OwnerAddress.Empty() {
		return sdk.ErrInvalidAddress(t.OwnerAddress.String())
	}
	return nil
}

func (t Token) String() string {
	bz, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	return string(bz)
}

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

type Name string

func (n Name) validate() sdk.Error {
	if len(n) <= 0 || len(n) > maxNameLen {
		return ErrInvalidName(string(n))
	}
	return nil
}

type ShortSymbol string

func (s ShortSymbol) validate() sdk.Error {
	str := string(s)
	if strings.ToLower(str) == strings.ToLower(assets.MicroMedDenom[1:]) {
		return ErrSymbolNotAllowed(str)
	}
	if !regexShortSymbol.MatchString(strings.ToUpper(str)) {
		return ErrInvalidSymbol(str)
	}
	return nil
}

type Symbol string

func newSymbol(short ShortSymbol, txHash string) Symbol {
	return Symbol(fmt.Sprintf(
		"%s-%s",
		strings.ToUpper(string(short)),
		strings.ToUpper(txHash[:symbolSuffixLen]),
	))
}

func (s Symbol) MicroDenom() string {
	withoutDash := strings.ReplaceAll(string(s), "-", "")
	return fmt.Sprintf("u%s", strings.ToLower(withoutDash))
}

// Symbol is case-insensitive. Although newSymbol converts all letters to uppercase,
// Symbol can contains lowercase letters if it's created by the simple Go type casting.
// So, when comparing Symbols to each other, they must be converted to uppercase first.
func (s Symbol) ToUpper() string {
	return strings.ToUpper(string(s))
}

func (s Symbol) validate() sdk.Error {
	str := string(s)
	if !regexSymbol.MatchString(strings.ToUpper(str)) {
		return ErrInvalidSymbol(str)
	}

	short := ShortSymbol(strings.Split(str, "-")[0])
	return short.validate()
}

func validateTotalSupply(totalSupply sdk.Coin) sdk.Error {
	if !totalSupply.IsValid() {
		return ErrInvalidTotalSupply("invalid total supply: " + totalSupply.String())
	}
	return validateTotalSupplyAmount(totalSupply.Amount)
}

func validateTotalSupplyAmount(amount sdk.Int) sdk.Error {
	if !amount.IsPositive() {
		return ErrInvalidTotalSupply("total supply must be positive")
	}
	if !amount.LTE(maxTotalSupply) {
		return ErrInvalidTotalSupply("too much total supply. max: " + maxTotalSupply.String())
	}
	return nil
}

type Tokens []string

func (s Tokens) String() string {
	var buf bytes.Buffer
	for i, symbol := range s {
		if i > 0 {
			buf.WriteRune('\n')
		}
		buf.WriteString(symbol)
	}
	return buf.String()
}
