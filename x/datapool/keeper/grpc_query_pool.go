package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/types/query"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/medibloc/panacea-core/v2/x/datapool/types"
)

func (k Keeper) Pool(goCtx context.Context, req *types.QueryPoolRequest) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pool, err := k.GetPool(ctx, req.PoolId)
	if err != nil {
		return nil, err
	}

	return &types.QueryPoolResponse{
		Pool: pool,
	}, nil
}

func (k Keeper) DataValidator(goCtx context.Context, req *types.QueryDataValidatorRequest) (*types.QueryDataValidatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	accValidatorAddress, err := sdk.AccAddressFromBech32(req.GetAddress())
	if err != nil {
		return nil, err
	}

	dataValidator, err := k.GetDataValidator(ctx, accValidatorAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryDataValidatorResponse{DataValidator: &dataValidator}, nil
}

func (k Keeper) DataPoolParams(goCtx context.Context, req *types.QueryDataPoolParamsRequest) (*types.QueryDataPoolParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)

	return &types.QueryDataPoolParamsResponse{Params: &params}, nil
}

func (k Keeper) DataPoolModuleAddr(goCtx context.Context, req *types.QueryDataPoolModuleAddrRequest) (*types.QueryDataPoolModuleAddrResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	moduleAddr := types.GetModuleAddress()

	return &types.QueryDataPoolModuleAddrResponse{Address: moduleAddr.String()}, nil
}

func (k Keeper) DataValidationCertificates(goCtx context.Context, req *types.QueryDataValidationCertificatesRequest) (*types.QueryDataValidationCertificatesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	store := ctx.KVStore(k.storeKey)
	certsStore := prefix.NewStore(store, types.GetKeyPrefixDataValidateCerts(req.PoolId, req.Round))

	var certs []types.DataValidationCertificate
	pageRes, err := query.Paginate(certsStore, req.Pagination, func(_ []byte, value []byte) error {
		var cert types.DataValidationCertificate
		err := k.cdc.UnmarshalBinaryLengthPrefixed(value, &cert)
		if err != nil {
			return err
		}
		certs = append(certs, cert)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDataValidationCertificatesResponse{
		DataValidationCertificates: certs,
		Pagination:                 pageRes,
	}, nil
}

func (k Keeper) DataPassRedeemReceipt(goCtx context.Context, req *types.QueryDataPassRedeemReceiptRequest) (*types.QueryDataPassRedeemReceiptResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	redeemReceipt, err := k.GetDataPassRedeemReceipt(ctx, req.PoolId, req.Round, req.NftId)
	if err != nil {
		return nil, err
	}

	return &types.QueryDataPassRedeemReceiptResponse{
		DataPassRedeemReceipt: redeemReceipt,
	}, nil

}

func (k Keeper) DataPassRedeemReceipts(goCtx context.Context, req *types.QueryDataPassRedeemReceiptsRequest) (*types.QueryDataPassRedeemReceiptsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	store := ctx.KVStore(k.storeKey)
	redeemReceiptStore := prefix.NewStore(store, types.GetKeyPrefixNFTRedeemReceiptByPoolID(req.PoolId))

	var redeemReceipts []types.DataPassRedeemReceipt
	pageRes, err := query.Paginate(redeemReceiptStore, req.Pagination, func(_, value []byte) error {
		var redeemReceipt types.DataPassRedeemReceipt
		err := k.cdc.UnmarshalBinaryLengthPrefixed(value, &redeemReceipt)
		if err != nil {
			return err
		}
		redeemReceipts = append(redeemReceipts, redeemReceipt)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDataPassRedeemReceiptsResponse{
		DataPassRedeemReceipts: redeemReceipts,
		Pagination:             pageRes,
	}, nil
}
