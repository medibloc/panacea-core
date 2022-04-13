package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DONTCOVER

// x/datapool module sentinel errors
var (
	ErrDataValidatorNotFound      = sdkerrors.Register(ModuleName, 1, "data validator not found")
	ErrDataValidatorAlreadyExist  = sdkerrors.Register(ModuleName, 2, "data validator already exists")
	ErrNotEnoughPoolDeposit       = sdkerrors.Register(ModuleName, 3, "the balance is not enough to make a data pool")
	ErrNotRegisteredDataValidator = sdkerrors.Register(ModuleName, 4, "data validator is not registered")
	ErrNoRegisteredNFTContract    = sdkerrors.Register(ModuleName, 5, "no NFT contract is registered")
	ErrPoolNotFound               = sdkerrors.Register(ModuleName, 6, "pool not found")
	ErrNFTAllIssued               = sdkerrors.Register(ModuleName, 12, "all NFTs issued")
	ErrRoundNotMatched            = sdkerrors.Register(ModuleName, 13, "data pool sales round not matched")
	ErrPaymentNotMatched          = sdkerrors.Register(ModuleName, 14, "payment not matched")
)
