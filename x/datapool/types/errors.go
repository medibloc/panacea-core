package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DONTCOVER

// x/datapool module sentinel errors
var (
	ErrNotEnoughPoolDeposit     = sdkerrors.Register(ModuleName, 1, "the deposit is not enough to make a data pool")
	ErrNoRegisteredNFTContract  = sdkerrors.Register(ModuleName, 2, "no NFT contract is registered")
	ErrPoolNotFound             = sdkerrors.Register(ModuleName, 3, "pool not found")
	ErrNotEqualsSeller          = sdkerrors.Register(ModuleName, 4, "the requester does not matched certificate information")
	ErrInvalidSignature         = sdkerrors.Register(ModuleName, 5, "failed to signature verify.")
	ErrInvalidDataCert          = sdkerrors.Register(ModuleName, 6, "certificate is not valid")
	ErrExistSameDataHash        = sdkerrors.Register(ModuleName, 7, "data already exists in the pool")
	ErrGetDataCert              = sdkerrors.Register(ModuleName, 8, "failed get certificate.")
	ErrRevenueDistribution      = sdkerrors.Register(ModuleName, 9, "failed to revenue distribution")
	ErrNFTAllIssued             = sdkerrors.Register(ModuleName, 10, "all NFTs issued")
	ErrRoundNotMatched          = sdkerrors.Register(ModuleName, 11, "data pool sales round not matched")
	ErrPaymentNotMatched        = sdkerrors.Register(ModuleName, 12, "payment not matched")
	ErrCreatePool               = sdkerrors.Register(ModuleName, 13, "failed to create data pool")
	ErrMintNFT                  = sdkerrors.Register(ModuleName, 14, "failed to mint NFT")
	ErrInstantiateContract      = sdkerrors.Register(ModuleName, 15, "failed to instantiate contract")
	ErrBuyDataPass              = sdkerrors.Register(ModuleName, 16, "failed to buy data pass")
	ErrGetDataPassRedeemReceipt = sdkerrors.Register(ModuleName, 17, "failed to get data pass receipt")
	ErrRedeemDataPass           = sdkerrors.Register(ModuleName, 18, "failed to redeem data pass")
	ErrNotOwnedRedeemerNft      = sdkerrors.Register(ModuleName, 19, "invalid nft id")
	ErrRedeemDataPassNotFound   = sdkerrors.Register(ModuleName, 20, "redeem data pass not found")
	ErrRedeemHistoryNotFound    = sdkerrors.Register(ModuleName, 21, "redeem history not found")
	ErrNoTrustedOracle          = sdkerrors.Register(ModuleName, 22, "no trusted oracle, but it is required")
)
