package types

const (
	QueryToken      = "token"
	QueryListTokens = "listTokens"
)

type QueryTokenParams struct {
	Symbol Symbol
}

func NewQueryTokenParams(symbol Symbol) *QueryTokenParams {
	return &QueryTokenParams{Symbol: symbol}
}
