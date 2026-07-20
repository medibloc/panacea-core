package keeper

import "cosmossdk.io/collections"

var (
	classPoliciesPrefix = collections.NewPrefix(0)
	mintedCountsPrefix  = collections.NewPrefix(1)
	lifecyclesPrefix    = collections.NewPrefix(2)
	tombstonesPrefix    = collections.NewPrefix(3)

	nftKeyCodec = collections.PairKeyCodec(collections.StringKey, collections.StringKey)
)
