package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Deal(goCtx context.Context, req *types.QueryDealRequest) (*types.QueryDealResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	deal, err := k.GetDeal(sdk.UnwrapSDKContext(goCtx), req.DealId)
	if err != nil {
		return nil, err
	}

	return &types.QueryDealResponse{
		Deal: deal,
	}, nil
}

func (k Keeper) Deals(goCtx context.Context, req *types.QueryDealsRequest) (*types.QueryDealsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	store := ctx.KVStore(k.storeKey)
	dealsStore := prefix.NewStore(store, types.DealKey)

	var deals []*types.Deal
	pageRes, err := query.Paginate(dealsStore, req.Pagination, func(_ []byte, value []byte) error {
		var deal types.Deal
		err := k.cdc.UnmarshalLengthPrefixed(value, &deal)
		if err != nil {
			return err
		}
		deals = append(deals, &deal)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDealsResponse{
		Deals:      deals,
		Pagination: pageRes,
	}, nil
}

func (k Keeper) Consents(goCtx context.Context, req *types.QueryConsents) (*types.QueryConsentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	store := ctx.KVStore(k.storeKey)

	consentsStore := prefix.NewStore(store, append(types.ConsentKey, sdk.Uint64ToBigEndian(req.DealId)...))

	var consents []*types.Consent
	pageRes, err := query.Paginate(consentsStore, req.Pagination, func(_ []byte, value []byte) error {
		var consent types.Consent
		err := k.cdc.UnmarshalLengthPrefixed(value, &consent)
		if err != nil {
			return err
		}
		consents = append(consents, &consent)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryConsentsResponse{
		Consents:   consents,
		Pagination: pageRes,
	}, nil
}

func (k Keeper) Consent(goCtx context.Context, req *types.QueryConsent) (*types.QueryConsentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	consent, err := k.GetConsent(sdk.UnwrapSDKContext(goCtx), req.DealId, req.DataHash)
	if err != nil {
		return nil, err
	}

	return &types.QueryConsentResponse{
		Consent: consent,
	}, nil
}
