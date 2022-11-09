package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DONTCOVER

var (
	ErrOracleRegistrationVote      = sdkerrors.Register(ModuleName, 1, "error while voting for OracleRegistration")
	ErrGetOracle                   = sdkerrors.Register(ModuleName, 2, "error while get oracle")
	ErrGetOracleRegistration       = sdkerrors.Register(ModuleName, 3, "error while get oracleRegistration")
	ErrEmptyNodePubKey             = sdkerrors.Register(ModuleName, 4, "node public key is empty")
	ErrInvalidNodePubKey           = sdkerrors.Register(ModuleName, 5, "invalid node public key")
	ErrEmptyNodePubKeyRemoteReport = sdkerrors.Register(ModuleName, 6, "remote report of node public key is empty")
	ErrInvalidTrustedBlockHeight   = sdkerrors.Register(ModuleName, 7, "trusted block height must be greater than zero")
	ErrTrustedBlockHashNil         = sdkerrors.Register(ModuleName, 8, "trusted block hash should not be nil")
	ErrRegisterOracle              = sdkerrors.Register(ModuleName, 9, "error while registering a oracle")
	ErrValidatorNotFound           = sdkerrors.Register(ModuleName, 10, "validator not found")
	ErrJailedValidator             = sdkerrors.Register(ModuleName, 11, "jailed validator cannot be a oracle")
	ErrOracleRegistrationNotFound  = sdkerrors.Register(ModuleName, 12, "oracle registration not found")
	ErrOracleUpgradeInfoNotFound   = sdkerrors.Register(ModuleName, 13, "oracle upgrade information not found")
	ErrGetOracleUpgradeInfo        = sdkerrors.Register(ModuleName, 14, "error while get oracleUpgradeInfo")
	ErrUpgradeOracle               = sdkerrors.Register(ModuleName, 15, "error while upgrading a oracle")
)
