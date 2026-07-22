package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrTransferNotAllowed   = errorsmod.Register(PolicyCodespace, 2, "nft transfer is not allowed")
	ErrRevocationNotAllowed = errorsmod.Register(PolicyCodespace, 3, "nft revocation is not allowed")
	ErrNFTRevoked           = errorsmod.Register(PolicyCodespace, 4, "nft is revoked")
	ErrMaxSupplyReached     = errorsmod.Register(PolicyCodespace, 5, "maximum nft supply reached")
	ErrNFTIDPermanentlyUsed = errorsmod.Register(PolicyCodespace, 6, "nft id is permanently used")
)
