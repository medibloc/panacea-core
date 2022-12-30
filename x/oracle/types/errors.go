package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrOracleRegistration        = sdkerrors.Register(ModuleName, 1, "error while registering a oracle")
	ErrGetOracleRegistration     = sdkerrors.Register(ModuleName, 2, "error while getting a oracle registration")
	ErrGetOracle                 = sdkerrors.Register(ModuleName, 3, "error while getting a oracle")
	ErrOracleNotFound            = sdkerrors.Register(ModuleName, 4, "oracle not found")
	ErrInvalidUniqueID           = sdkerrors.Register(ModuleName, 5, "invalid unique id")
	ErrApproveOracleRegistration = sdkerrors.Register(ModuleName, 6, "error while approving an oracle registration")
	ErrUpdateOracle              = sdkerrors.Register(ModuleName, 7, "error while updating oracle info")
	ErrCommissionNegative        = sdkerrors.Register(ModuleName, 8, "commission must be positive")
	ErrCommissionGTMaxRate       = sdkerrors.Register(ModuleName, 9, "commission cannot be more than the max rate")
	ErrCommissionUpdateTime      = sdkerrors.Register(ModuleName, 10, "commission cannot be changed more than once in 24h")
	ErrCommissionGTMaxChangeRate = sdkerrors.Register(ModuleName, 11, "commission cannot be changed more than max change rate")
	ErrOracleUpgradeInfoNotFound = sdkerrors.Register(ModuleName, 12, "oracle upgrade information not found")
	ErrGetOracleUpgradeInfo      = sdkerrors.Register(ModuleName, 13, "error while get oracleUpgradeInfo")
	ErrUpgradeOracle             = sdkerrors.Register(ModuleName, 14, "error while upgrading a oracle")
	ErrGetOracleUpgrade          = sdkerrors.Register(ModuleName, 15, "error while get oracleUpgrade")
	ErrOracleUpgradeNotFound     = sdkerrors.Register(ModuleName, 16, "oracle upgrade not found")
)
