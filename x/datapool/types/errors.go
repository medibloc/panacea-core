package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DONTCOVER

// x/datapool module sentinel errors
var (
	ErrNotEnoughPoolDeposit      = sdkerrors.Register(ModuleName, 1, "the deposit is not enough to make a data pool")
	ErrNotRegisteredOracle       = sdkerrors.Register(ModuleName, 2, "oracle is not registered")
	ErrNoRegisteredNFTContract   = sdkerrors.Register(ModuleName, 3, "no NFT contract is registered")
	ErrPoolNotFound              = sdkerrors.Register(ModuleName, 4, "pool not found")
	ErrNotEqualsSeller           = sdkerrors.Register(ModuleName, 5, "the requester does not matched certificate information")
	ErrInvalidSignature          = sdkerrors.Register(ModuleName, 6, "failed to signature verify.")
	ErrInvalidDataValidationCert = sdkerrors.Register(ModuleName, 7, "certificate is not valid")
	ErrExistSameDataHash         = sdkerrors.Register(ModuleName, 8, "data already exists in the pool")
	ErrGetDataValidationCert     = sdkerrors.Register(ModuleName, 9, "failed get certificate.")
	ErrRevenueDistribution       = sdkerrors.Register(ModuleName, 10, "failed to revenue distribution")
	ErrNFTAllIssued              = sdkerrors.Register(ModuleName, 11, "all NFTs issued")
	ErrRoundNotMatched           = sdkerrors.Register(ModuleName, 12, "data pool sales round not matched")
	ErrPaymentNotMatched         = sdkerrors.Register(ModuleName, 13, "payment not matched")
	ErrCreatePool                = sdkerrors.Register(ModuleName, 14, "failed to create data pool")
	ErrMintNFT                   = sdkerrors.Register(ModuleName, 15, "failed to mint NFT")
	ErrInstantiateContract       = sdkerrors.Register(ModuleName, 16, "failed to instantiate contract")
	ErrBuyDataPass               = sdkerrors.Register(ModuleName, 17, "failed to buy data pass")
	ErrGetDataPassRedeemReceipt  = sdkerrors.Register(ModuleName, 18, "failed to get data pass receipt")
	ErrRedeemDataPass            = sdkerrors.Register(ModuleName, 19, "failed to redeem data pass")
	ErrNotOwnedRedeemerNft       = sdkerrors.Register(ModuleName, 20, "invalid nft id")
	ErrRedeemDataPassNotFound    = sdkerrors.Register(ModuleName, 21, "redeem data pass not found")
	ErrRedeemHistoryNotFound     = sdkerrors.Register(ModuleName, 22, "redeem history not found")
)
