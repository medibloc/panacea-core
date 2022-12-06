package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Oracles(goCtx context.Context, request *types.QueryOraclesRequest) (*types.QueryOraclesResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	store := ctx.KVStore(k.storeKey)
	oracleStore := prefix.NewStore(store, types.OraclesKey)

	var oracles []*types.Oracle
	pageRes, err := query.Paginate(oracleStore, request.Pagination, func(_, value []byte) error {
		var oracle types.Oracle
		if err := k.cdc.UnmarshalLengthPrefixed(value, &oracle); err != nil {
			return err
		}

		oracles = append(oracles, &oracle)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryOraclesResponse{
		Oracles:    oracles,
		Pagination: pageRes,
	}, nil
}

func (k Keeper) Oracle(goCtx context.Context, request *types.QueryOracleRequest) (*types.QueryOracleResponse, error) {
	oracle, err := k.GetOracle(sdk.UnwrapSDKContext(goCtx), request.OracleAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryOracleResponse{
		Oracle: oracle,
	}, nil
}

func (k Keeper) OracleRegistrations(goCtx context.Context, request *types.QueryOracleRegistrationsRequest) (*types.QueryOracleRegistrationsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	store := ctx.KVStore(k.storeKey)
	oracleRegistrationStore := prefix.NewStore(store, types.OracleRegistrationKey)

	var oracleRegistrations []*types.OracleRegistration
	pageRes, err := query.Paginate(oracleRegistrationStore, request.Pagination, func(_, value []byte) error {
		var oracleRegistration types.OracleRegistration
		if err := k.cdc.UnmarshalLengthPrefixed(value, &oracleRegistration); err != nil {
			return err
		}

		oracleRegistrations = append(oracleRegistrations, &oracleRegistration)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryOracleRegistrationsResponse{
		OracleRegistrations: oracleRegistrations,
		Pagination:          pageRes,
	}, nil
}

func (k Keeper) OracleRegistration(goCtx context.Context, request *types.QueryOracleRegistrationRequest) (*types.QueryOracleRegistrationResponse, error) {
	oracleRegistration, err := k.GetOracleRegistration(sdk.UnwrapSDKContext(goCtx), request.UniqueId, request.OracleAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryOracleRegistrationResponse{OracleRegistration: oracleRegistration}, nil
}

func (k Keeper) Params(ctx context.Context, request *types.QueryOracleParamsRequest) (*types.QueryParamsResponse, error) {
	//TODO implement me
	panic("implement me")
}
