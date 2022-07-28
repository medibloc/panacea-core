package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

// VoteOracleRegistration defines to vote for the new oracle's verification results.
func (k Keeper) VoteOracleRegistration(ctx sdk.Context, signedVote *types.SignedOracleRegistrationVote) error {
	if err := signedVote.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(types.ErrOracleRegistrationVote, err.Error())
	}

	if k.isMaliciousRequest(ctx, signedVote.OracleRegistrationVote, signedVote.Signature) {
		// TODO implements request slashing
		return sdkerrors.Wrap(types.ErrDetectionMaliciousBehavior, "")
	}

	// Validate the status of panacea to ensure that voting is possible.
	if err := k.validateOracleRegistrationVote(ctx, signedVote); err != nil {
		return sdkerrors.Wrap(types.ErrOracleRegistrationVote, err.Error())
	}

	// When all validations pass, the vote is saved.
	// If it is an oracle that has already voted, it will be overwritten.
	if err := k.SetOracleRegistrationVote(ctx, signedVote.OracleRegistrationVote); err != nil {
		return sdkerrors.Wrap(types.ErrOracleRegistrationVote, err.Error())
	}

	return nil
}

// isMaliciousRequest defines to check for malicious requests.
func (k Keeper) isMaliciousRequest(ctx sdk.Context, vote *types.OracleRegistrationVote, signature []byte) bool {
	voteBz := k.cdc.MustMarshal(vote)
	// Verifies that voting requests are signed with oraclePrivKey.
	ok := secp256k1.PubKey(k.GetParams(ctx).OraclePublicKey).VerifySignature(voteBz, signature)

	return !ok
}

// validateOracleRegistrationVote defines checking the status of a panacea to ensure that voting is possible.
func (k Keeper) validateOracleRegistrationVote(ctx sdk.Context, signedVote *types.SignedOracleRegistrationVote) error {
	vote := signedVote.OracleRegistrationVote

	params := k.GetParams(ctx)

	if params.UniqueId != vote.UniqueId {
		return fmt.Errorf("is not match the currently active uniqueID. expected %s, got %s", params.UniqueId, vote.UniqueId)
	}

	oracle, err := k.GetOracle(ctx, vote.VoterAddress)
	if err != nil {
		return err
	}
	if oracle.Status != types.ORACLE_STATUS_ACTIVE {
		return fmt.Errorf("this oracle is not in 'ACTIVE' state")
	}

	oracleRegistration, err := k.GetOracleRegistration(ctx, vote.VotingTargetAddress)
	if err != nil {
		return err
	}

	if oracleRegistration.Status != types.ORACLE_REGISTRATION_STATUS_VOTING_PERIOD {
		return fmt.Errorf("the currently voted oracle's status is not 'VOTING_PERIOD'")
	}

	return nil
}

func (k Keeper) GetAllOracles(ctx sdk.Context) ([]types.Oracle, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixOracle)
	defer iterator.Close()

	oracles := make([]types.Oracle, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var oracle types.Oracle

		err := k.cdc.UnmarshalLengthPrefixed(bz, &oracle)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrGetOracle, err.Error())
		}

		oracles = append(oracles, oracle)
	}

	return oracles, nil
}

func (k Keeper) GetOracle(ctx sdk.Context, address string) (*types.Oracle, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyPrefixOracle(address)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf(fmt.Sprintf("'%s' is not exist oracle", address))
	}

	oracle := &types.Oracle{}

	err := k.cdc.UnmarshalLengthPrefixed(bz, oracle)
	if err != nil {
		return nil, err
	}

	return oracle, nil
}

func (k Keeper) SetOracle(ctx sdk.Context, oracle *types.Oracle) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyPrefixOracle(oracle.Address)
	bz, err := k.cdc.MarshalLengthPrefixed(oracle)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}

func (k Keeper) GetAllOracleRegistration(ctx sdk.Context) ([]types.OracleRegistration, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixOracleRegistration)
	defer iterator.Close()

	oracleRegistrations := make([]types.OracleRegistration, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var oracleRegistration types.OracleRegistration
		err := k.cdc.UnmarshalLengthPrefixed(bz, &oracleRegistration)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrGetOracleRegistration, err.Error())
		}

		oracleRegistrations = append(oracleRegistrations, oracleRegistration)
	}

	return oracleRegistrations, nil

}

func (k Keeper) GetOracleRegistration(ctx sdk.Context, address string) (*types.OracleRegistration, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyPrefixOracleRegistration(address)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("is not exist oracleRegistration with the address of '%s'", address)
	}

	oracleRegistration := &types.OracleRegistration{}

	err := k.cdc.UnmarshalLengthPrefixed(bz, oracleRegistration)
	if err != nil {
		return nil, err
	}

	return oracleRegistration, nil
}

func (k Keeper) SetOracleRegistration(ctx sdk.Context, regOracle *types.OracleRegistration) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyPrefixOracleRegistration(regOracle.Address)
	bz, err := k.cdc.MarshalLengthPrefixed(regOracle)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}

func (k Keeper) GetAllOracleRegistrationVote(ctx sdk.Context) ([]types.OracleRegistrationVote, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixOracleRegistrationVote)
	defer iterator.Close()

	oracleRegistrationVotes := make([]types.OracleRegistrationVote, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var oracleRegistrationVote types.OracleRegistrationVote
		err := k.cdc.UnmarshalLengthPrefixed(bz, &oracleRegistrationVote)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrGetOracleRegistration, err.Error())
		}

		oracleRegistrationVotes = append(oracleRegistrationVotes, oracleRegistrationVote)
	}

	return oracleRegistrationVotes, nil
}

func (k Keeper) GetOracleRegistrationVote(ctx sdk.Context, uniqueId, votingTargetAddress, voterAddress string) (*types.OracleRegistrationVote, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyPrefixOracleRegistrationVote(uniqueId, votingTargetAddress, voterAddress)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("oracle does not exist. uniqueID: %s, votingTargetAddress: %s, voterAddress: %s", uniqueId, votingTargetAddress, voterAddress)
	}

	vote := &types.OracleRegistrationVote{}
	err := k.cdc.UnmarshalLengthPrefixed(bz, vote)
	if err != nil {
		return nil, err
	}

	return vote, nil
}

func (k Keeper) SetOracleRegistrationVote(ctx sdk.Context, vote *types.OracleRegistrationVote) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyPrefixOracleRegistrationVote(vote.UniqueId, vote.VotingTargetAddress, vote.VoterAddress)
	bz, err := k.cdc.MarshalLengthPrefixed(vote)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}
