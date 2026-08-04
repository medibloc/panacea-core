package app_test

import (
	"encoding/json"
	"testing"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	panaceaapp "github.com/medibloc/panacea-core/v2/app"
)

func TestZeroHeightExportDecodesLengthPrefixedValidatorStoreKeys(t *testing.T) {
	testApp, _, ctx := newPNFTProposalTestApp(t)

	validatorsBefore, err := testApp.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, validatorsBefore)

	delegationsBefore, err := testApp.StakingKeeper.GetAllDelegations(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, delegationsBefore)

	store := ctx.KVStore(testApp.GetKey(stakingtypes.StoreKey))
	iter := storetypes.KVStorePrefixIterator(store, stakingtypes.ValidatorsKey)
	require.True(t, iter.Valid())

	// Validator keys include an address-length byte after the store prefix. The
	// former key[1:] decoding therefore identifies no validator.
	legacyDecodedAddr := sdk.ValAddress(iter.Key()[1:])
	_, err = testApp.StakingKeeper.GetValidator(ctx, legacyDecodedAddr)
	require.ErrorContains(t, err, "validator does not exist")

	currentDecodedAddr := sdk.ValAddress(stakingtypes.AddressFromValidatorsKey(iter.Key()))
	_, err = testApp.StakingKeeper.GetValidator(ctx, currentDecodedAddr)
	require.NoError(t, err)
	require.NoError(t, iter.Close())

	exported, err := testApp.ExportAppStateAndValidators(true, nil, nil)
	require.NoError(t, err)
	require.Equal(t, int64(0), exported.Height)
	require.NotEmpty(t, exported.Validators)

	var appState panaceaapp.GenesisState
	require.NoError(t, json.Unmarshal(exported.AppState, &appState))

	var stakingState stakingtypes.GenesisState
	require.NoError(t, testApp.AppCodec().UnmarshalJSON(appState[stakingtypes.ModuleName], &stakingState))
	require.Len(t, stakingState.Validators, len(validatorsBefore))
	require.Len(t, stakingState.Delegations, len(delegationsBefore))
	require.ElementsMatch(t, delegationsBefore, stakingState.Delegations)

	validatorAddrsBefore := make([]string, 0, len(validatorsBefore))
	for _, validator := range validatorsBefore {
		validatorAddrsBefore = append(validatorAddrsBefore, validator.OperatorAddress)
	}
	validatorAddrsAfter := make([]string, 0, len(stakingState.Validators))
	for _, validator := range stakingState.Validators {
		validatorAddrsAfter = append(validatorAddrsAfter, validator.OperatorAddress)
		require.Zero(t, validator.UnbondingHeight)
	}
	require.ElementsMatch(t, validatorAddrsBefore, validatorAddrsAfter)
}
