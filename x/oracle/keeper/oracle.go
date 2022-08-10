package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

// VoteOracleRegistration defines to vote for the new oracle's verification results.
func (k Keeper) VoteOracleRegistration(ctx sdk.Context, vote *types.OracleRegistrationVote, signature []byte) error {
	if err := vote.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(types.ErrOracleRegistrationVote, err.Error())
	}

	if k.verifyVoteSignature(ctx, vote, signature) {
		// TODO implements request slashing
		return sdkerrors.Wrap(types.ErrDetectionMaliciousBehavior, "")
	}

	// Check if the oracle can be voted to be registered.
	if err := k.validateOracleRegistrationVote(ctx, vote); err != nil {
		return sdkerrors.Wrap(types.ErrOracleRegistrationVote, err.Error())
	}

	// When all validations pass, the vote is saved.
	// If it is an oracle that has already voted, it will be overwritten.
	if err := k.SetOracleRegistrationVote(ctx, vote); err != nil {
		return sdkerrors.Wrap(types.ErrOracleRegistrationVote, err.Error())
	}

	return nil
}

// verifyVoteSignature defines to check for malicious requests.
func (k Keeper) verifyVoteSignature(ctx sdk.Context, vote *types.OracleRegistrationVote, signature []byte) bool {
	voteBz := k.cdc.MustMarshal(vote)
	// Verifies that voting requests are signed with oraclePrivKey.
	ok := secp256k1.PubKey(k.GetParams(ctx).OraclePublicKey).VerifySignature(voteBz, signature)

	return !ok
}

// validateOracleRegistrationVote checks the oracle/registration status in the Panacea to ensure that the oracle can be voted to be registered.
func (k Keeper) validateOracleRegistrationVote(ctx sdk.Context, vote *types.OracleRegistrationVote) error {
	params := k.GetParams(ctx)

	if params.UniqueId != vote.UniqueId {
		return fmt.Errorf("not matched with the currently active uniqueID. expected %s, got %s", params.UniqueId, vote.UniqueId)
	}

	oracle, err := k.GetOracle(ctx, vote.VoterAddress)
	if err != nil {
		return err
	}
	if oracle.Status != types.ORACLE_STATUS_ACTIVE {
		return fmt.Errorf("this oracle is not in 'ACTIVE' state")
	}

	oracleRegistration, err := k.GetOracleRegistration(ctx, vote.UniqueId, vote.VotingTargetAddress)
	if err != nil {
		return err
	}

	if oracleRegistration.Status != types.ORACLE_REGISTRATION_STATUS_VOTING_PERIOD {
		return fmt.Errorf("the currently voted oracle's status is not 'VOTING_PERIOD'")
	}

	return nil
}

func (k Keeper) GetAllOracleList(ctx sdk.Context) ([]types.Oracle, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.OraclesKey)
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
	accAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}
	key := types.GetOracleKey(accAddr)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf(fmt.Sprintf("'%s' is not exist oracle", address))
	}

	oracle := &types.Oracle{}

	err = k.cdc.UnmarshalLengthPrefixed(bz, oracle)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrGetOracle, err.Error())
	}

	return oracle, nil
}

func (k Keeper) SetOracle(ctx sdk.Context, oracle *types.Oracle) error {
	store := ctx.KVStore(k.storeKey)
	accAddr, err := sdk.AccAddressFromBech32(oracle.Address)
	if err != nil {
		return err
	}
	key := types.GetOracleKey(accAddr)
	bz, err := k.cdc.MarshalLengthPrefixed(oracle)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}

func (k Keeper) GetAllOracleRegistrationList(ctx sdk.Context) ([]types.OracleRegistration, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.OracleRegistrationsKey)
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

func (k Keeper) GetOracleRegistration(ctx sdk.Context, uniqueID, address string) (*types.OracleRegistration, error) {
	store := ctx.KVStore(k.storeKey)
	accAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}
	key := types.GetOracleRegistrationKey(uniqueID, accAddr)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("is not exist oracleRegistration with the address of '%s'", address)
	}

	oracleRegistration := &types.OracleRegistration{}

	err = k.cdc.UnmarshalLengthPrefixed(bz, oracleRegistration)
	if err != nil {
		return nil, err
	}

	return oracleRegistration, nil
}

func (k Keeper) SetOracleRegistration(ctx sdk.Context, regOracle *types.OracleRegistration) error {
	store := ctx.KVStore(k.storeKey)
	accAddr, err := sdk.AccAddressFromBech32(regOracle.Address)
	if err != nil {
		return err
	}
	key := types.GetOracleRegistrationKey(regOracle.UniqueId, accAddr)
	bz, err := k.cdc.MarshalLengthPrefixed(regOracle)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}

func (k Keeper) GetAllOracleRegistrationVoteList(ctx sdk.Context) ([]types.OracleRegistrationVote, error) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.OracleRegistrationVotesKey)
	defer iterator.Close()

	oracleRegistrationVotes := make([]types.OracleRegistrationVote, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var oracleRegistrationVote types.OracleRegistrationVote
		err := k.cdc.UnmarshalLengthPrefixed(bz, &oracleRegistrationVote)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrOracleRegistrationVote, err.Error())
		}

		oracleRegistrationVotes = append(oracleRegistrationVotes, oracleRegistrationVote)
	}

	return oracleRegistrationVotes, nil
}

func (k Keeper) GetOracleRegistrationVoteIterator(ctx sdk.Context, uniqueID, voteTargetAddress string) sdk.Iterator {
	accAddr, err := sdk.AccAddressFromBech32(voteTargetAddress)
	if err != nil {
		panic("")
	}
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, types.GetOracleRegistrationVotesKey(uniqueID, accAddr))
}

func (k Keeper) GetOracleRegistrationVote(ctx sdk.Context, uniqueId, votingTargetAddress, voterAddress string) (*types.OracleRegistrationVote, error) {
	store := ctx.KVStore(k.storeKey)
	votingTargetAccAddr, err := sdk.AccAddressFromBech32(votingTargetAddress)
	if err != nil {
		return nil, err
	}
	voterAccAddr, err := sdk.AccAddressFromBech32(voterAddress)
	if err != nil {
		return nil, err
	}
	key := types.GetOracleRegistrationVoteKey(uniqueId, votingTargetAccAddr, voterAccAddr)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("oracle does not exist. uniqueID: %s, votingTargetAddress: %s, voterAddress: %s", uniqueId, votingTargetAddress, voterAddress)
	}

	vote := &types.OracleRegistrationVote{}
	err = k.cdc.UnmarshalLengthPrefixed(bz, vote)
	if err != nil {
		return nil, err
	}

	return vote, nil
}

func (k Keeper) SetOracleRegistrationVote(ctx sdk.Context, vote *types.OracleRegistrationVote) error {
	store := ctx.KVStore(k.storeKey)
	votingTargetAccAddr, err := sdk.AccAddressFromBech32(vote.VotingTargetAddress)
	if err != nil {
		return err
	}
	voterAccAddr, err := sdk.AccAddressFromBech32(vote.VoterAddress)
	if err != nil {
		return err
	}
	key := types.GetOracleRegistrationVoteKey(vote.UniqueId, votingTargetAccAddr, voterAccAddr)
	bz, err := k.cdc.MarshalLengthPrefixed(vote)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}
