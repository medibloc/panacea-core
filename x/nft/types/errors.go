package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrTransferNotAllowed   = errorsmod.Register(PolicyStoreKey, 2, "nft transfer is not allowed")
	ErrRevocationNotAllowed = errorsmod.Register(PolicyStoreKey, 3, "nft revocation is not allowed")
	ErrNFTRevoked           = errorsmod.Register(PolicyStoreKey, 4, "nft is revoked")
	ErrMaxSupplyReached     = errorsmod.Register(PolicyStoreKey, 5, "maximum nft supply reached")
	ErrNFTIDPermanentlyUsed = errorsmod.Register(PolicyStoreKey, 6, "nft id is permanently used")
)
