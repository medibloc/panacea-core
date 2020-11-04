package types

const (
	QueryIssuance = "issuance"
)

type QueryIssuanceParams struct {
	Denom string
}

func NewQueryIssuanceParams(denom string) *QueryIssuanceParams {
	return &QueryIssuanceParams{Denom: denom}
}
