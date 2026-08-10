package keeper

import (
	"errors"

	upstreamnft "cosmossdk.io/x/nft"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapQueryStateError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, upstreamnft.ErrClassNotExists),
		errors.Is(err, upstreamnft.ErrNFTNotExists):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
