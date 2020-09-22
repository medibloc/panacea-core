package types

const (
	QueryDID = "did"
)

type QueryDIDParams struct {
	DID DID
}

func NewQueryDIDParams(DID DID) *QueryDIDParams {
	return &QueryDIDParams{DID: DID}
}
