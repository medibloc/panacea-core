package keeper

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
)

const (
	defaultQueryPageLimit uint64 = 100
	maximumQueryPageLimit uint64 = 100
)

func normalizeQueryPageRequest(request *query.PageRequest) (*query.PageRequest, error) {
	if request == nil {
		return &query.PageRequest{Limit: defaultQueryPageLimit}, nil
	}
	if request.Offset > 0 {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("pagination offset is not supported")
	}
	if request.CountTotal {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("pagination count_total is not supported")
	}
	if request.Limit > maximumQueryPageLimit {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf(
			"pagination limit must not exceed %d",
			maximumQueryPageLimit,
		)
	}

	normalized := *request
	normalized.Key = append([]byte(nil), request.Key...)
	if normalized.Limit == 0 {
		normalized.Limit = defaultQueryPageLimit
	}
	return &normalized, nil
}
