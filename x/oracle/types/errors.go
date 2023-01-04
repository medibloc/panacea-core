package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrRegisterOracle             = sdkerrors.Register(ModuleName, 1, "error while registering a oracle")
	ErrGetOracleRegistration      = sdkerrors.Register(ModuleName, 2, "error while getting a oracle registration")
	ErrOracleRegistrationNotFound = sdkerrors.Register(ModuleName, 3, "oracle registration not found")
	ErrGetOracle                  = sdkerrors.Register(ModuleName, 4, "error while getting a oracle")
	ErrOracleNotFound             = sdkerrors.Register(ModuleName, 5, "oracle not found")
	ErrInvalidUniqueID            = sdkerrors.Register(ModuleName, 6, "invalid unique id")
	ErrApproveOracleRegistration  = sdkerrors.Register(ModuleName, 7, "error while approving an oracle registration")
	ErrUpdateOracle               = sdkerrors.Register(ModuleName, 8, "error while updating oracle info")
	ErrCommissionNegative         = sdkerrors.Register(ModuleName, 9, "commission must be positive")
	ErrCommissionGTMaxRate        = sdkerrors.Register(ModuleName, 10, "commission cannot be more than the max rate")
	ErrCommissionUpdateTime       = sdkerrors.Register(ModuleName, 11, "commission cannot be changed more than once in 24h")
	ErrCommissionGTMaxChangeRate  = sdkerrors.Register(ModuleName, 12, "commission cannot be changed more than max change rate")
	ErrOracleUpgradeInfoNotFound  = sdkerrors.Register(ModuleName, 13, "oracle upgrade information not found")
	ErrGetOracleUpgradeInfo       = sdkerrors.Register(ModuleName, 14, "error while get oracleUpgradeInfo")
	ErrUpgradeOracle              = sdkerrors.Register(ModuleName, 15, "error while upgrading a oracle")
	ErrGetOracleUpgrade           = sdkerrors.Register(ModuleName, 16, "error while get oracleUpgrade")
	ErrApproveOracleUpgrade       = sdkerrors.Register(ModuleName, 17, "error while approving oracle upgrade")
	ErrOracleUpgradeNotFound      = sdkerrors.Register(ModuleName, 18, "oracle upgrade not found")
)
