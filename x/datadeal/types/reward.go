package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	// Oracle commission for data sale is divided into two parts: data verification and data delivery.

	// DataVerificationRewardFraction is a fraction for data verification rewards
	DataVerificationRewardFraction = sdk.NewDecWithPrec(5, 1) // 50%

	// DataDeliveryRewardFraction is a fraction for data delivery rewards
	DataDeliveryRewardFraction = sdk.OneDec().Sub(DataVerificationRewardFraction) // 50%
)
