package did

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/did/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	QueryResolveDID = "resolveDid"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryResolveDID:
			return resolveDID(ctx, path[1:], req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown did query endpoint")
		}
	}
}

type ResolveDIDParams struct {
	DID types.DID
}

func resolveDID(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params ResolveDIDParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formated request data", err.Error()))
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, k.GetDID(ctx, params.DID))
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}
