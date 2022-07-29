package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// DONTCOVER

var (
	ErrOracleRegistrationVote     = sdkerrors.Register(ModuleName, 1, "Error while voting for OracleRegistration")
	ErrDetectionMaliciousBehavior = sdkerrors.Register(ModuleName, 2, "Errors caused by malicious actions")
)
