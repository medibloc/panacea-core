package v2_3_0

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/medibloc/panacea-core/v2/app/keepers"
)

func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator, keepers *keepers.AppKeepersWithKey) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		toVM, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return fromVM, err
		}

		if fromVersion, exists := fromVM[govtypes.ModuleName]; !exists || fromVersion >= toVM[govtypes.ModuleName] {
			return toVM, nil
		}

		params, err := keepers.GovKeeper.Params.Get(ctx)
		if err != nil {
			return fromVM, err
		}
		params.ExpeditedMinDeposit = sdk.Coins(params.MinDeposit).MulInt(
			math.NewInt(govv1.DefaultMinExpeditedDepositTokensRatio),
		)
		if err := normalizeMigratedExpeditedVotingPeriod(&params); err != nil {
			return fromVM, err
		}
		if err := params.ValidateBasic(); err != nil {
			return fromVM, fmt.Errorf("validate migrated governance params: %w", err)
		}
		if err := keepers.GovKeeper.Params.Set(ctx, params); err != nil {
			return fromVM, err
		}

		return toVM, nil
	}
}

func normalizeMigratedExpeditedVotingPeriod(params *govv1.Params) error {
	if params == nil {
		return fmt.Errorf("cannot normalize expedited voting period from nil governance params")
	}
	if params.VotingPeriod == nil {
		return fmt.Errorf("cannot normalize expedited voting period without regular voting period")
	}
	if params.ExpeditedVotingPeriod == nil {
		return fmt.Errorf("cannot normalize nil expedited voting period")
	}
	if *params.VotingPeriod <= time.Nanosecond {
		return fmt.Errorf(
			"cannot derive a positive expedited voting period below regular period %s",
			*params.VotingPeriod,
		)
	}
	if *params.ExpeditedVotingPeriod < *params.VotingPeriod {
		return nil
	}

	clamped := *params.VotingPeriod / 2
	if clamped <= 0 || clamped >= *params.VotingPeriod {
		return fmt.Errorf(
			"cannot derive a positive expedited voting period below regular period %s",
			*params.VotingPeriod,
		)
	}
	params.ExpeditedVotingPeriod = &clamped
	return nil
}
