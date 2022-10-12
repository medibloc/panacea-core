package keeper

import (
	"context"
	"errors"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func (k Keeper) Oracle(goCtx context.Context, req *types.QueryOracleRequest) (*types.QueryOracleResponse, error) {
	oracle, err := k.GetOracle(sdk.UnwrapSDKContext(goCtx), req.GetAddress())
	if err != nil {
		return nil, err
	}

	return &types.QueryOracleResponse{
		Oracle: oracle,
	}, nil
}

// Oracles returns a list of oracles.
func (k Keeper) Oracles(goCtx context.Context, req *types.QueryOraclesRequest) (*types.QueryOraclesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	store := ctx.KVStore(k.storeKey)
	oracleStore := prefix.NewStore(store, types.OraclesKey)

	var oracles []*types.Oracle
	pageRes, err := query.Paginate(oracleStore, req.Pagination, func(_, value []byte) error {
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

// OracleRegistration returns a OracleRegistration details.
func (k Keeper) OracleRegistration(goCtx context.Context, req *types.QueryOracleRegistrationRequest) (*types.QueryOracleRegistrationResponse, error) {
	oracleRegistration, err := k.GetOracleRegistration(sdk.UnwrapSDKContext(goCtx), req.UniqueId, req.Address)
	if err != nil {
		return nil, err
	}

	return &types.QueryOracleRegistrationResponse{OracleRegistration: oracleRegistration}, nil
}

func (k Keeper) OracleRegistrationVote(goCtx context.Context, request *types.QueryOracleRegistrationVoteRequest) (*types.QueryOracleRegistrationVoteResponse, error) {
	vote, err := k.GetOracleRegistrationVote(
		sdk.UnwrapSDKContext(goCtx),
		request.UniqueId,
		request.VotingTargetAddress,
		request.VoterAddress,
	)

	if err != nil {
		return nil, err
	}

	return &types.QueryOracleRegistrationVoteResponse{
		OracleRegistrationVote: vote,
	}, nil
}

// Params returns params of oracle module.
func (k Keeper) Params(goCtx context.Context, req *types.QueryOracleParamsRequest) (*types.QueryParamsResponse, error) {
	params := k.GetParams(sdk.UnwrapSDKContext(goCtx))
	return &types.QueryParamsResponse{Params: &params}, nil
}

func (k Keeper) OracleUpgradeInfo(ctx context.Context, _ *types.QueryOracleUpgradeInfoRequest) (*types.QueryOracleUpgradeInfoResponse, error) {
	upgradeInfo, err := k.GetOracleUpgradeInfo(sdk.UnwrapSDKContext(ctx))
	if err != nil {
		if errors.Is(err, types.ErrOracleUpgradeInfoNotFound) {
			return &types.QueryOracleUpgradeInfoResponse{}, nil
		}
		return nil, err
	}
	return &types.QueryOracleUpgradeInfoResponse{
		OracleUpgradeInfo: upgradeInfo,
	}, nil
}
