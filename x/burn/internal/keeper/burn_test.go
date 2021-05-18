package keeper_test

import (
	"time"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"testing"

	"github.com/medibloc/panacea-core/types/assets"
	"github.com/medibloc/panacea-core/x/burn/internal/keeper"
)

var (
	pubKeys = []crypto.PubKey{
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
	}

	addrs = []sdk.AccAddress{
		sdk.AccAddress(pubKeys[0].Address()),
		sdk.AccAddress(pubKeys[1].Address()),
		sdk.AccAddress(pubKeys[2].Address()),
	}

	initTokens = sdk.TokensFromConsensusPower(200)
	initCoins  = sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, initTokens))
)

// TestInput nolint
type TestInput struct {
	Ctx           sdk.Context
	Cdc           *codec.Codec
	AccountKeeper auth.AccountKeeper
	BankKeeper    bank.Keeper
	SupplyKeeper  supply.Keeper
	Keeper        keeper.Keeper
}

func newTestCodec() *codec.Codec {
	cdc := codec.New()

	auth.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	supply.RegisterCodec(cdc)
	params.RegisterCodec(cdc)
	bank.RegisterCodec(cdc)

	return cdc
}

func createTestInput(t *testing.T) TestInput {
	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tKeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	keySupply := sdk.NewKVStoreKey(supply.StoreKey)

	cdc := newTestCodec()
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ctx := sdk.NewContext(ms, abci.Header{Time: time.Now().UTC()}, false, log.NewNopLogger())

	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tKeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySupply, sdk.StoreTypeIAVL, db)

	require.NoError(t, ms.LoadLatestVersion())

	blackListAddrs := map[string]bool{
		auth.FeeCollectorName: true,
	}

	paramsKeeper := params.NewKeeper(cdc, keyParams, tKeyParams, params.DefaultCodespace)
	accountKeeper := auth.NewAccountKeeper(cdc, keyAcc, paramsKeeper.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	bankKeeper := bank.NewBaseKeeper(
		accountKeeper,
		paramsKeeper.Subspace(bank.DefaultParamspace),
		bank.DefaultCodespace,
		blackListAddrs)
	bankKeeper.SetSendEnabled(ctx, true)

	maccPerms := map[string][]string{
		auth.FeeCollectorName: nil,
	}
	supplyKeeper := supply.NewKeeper(
		cdc,
		keySupply,
		accountKeeper,
		bankKeeper, maccPerms)

	totalSupply := sdk.NewCoins(
		sdk.NewCoin(assets.MicroMedDenom, initTokens.MulRaw(int64(len(addrs)))))
	supplyKeeper.SetSupply(ctx, supply.NewSupply(totalSupply))

	for _, addr := range addrs {
		_, err := bankKeeper.AddCoins(ctx, addr, initCoins)
		require.NoError(t, err)
	}

	supply2 := supplyKeeper.GetSupply(ctx)
	supply2 = supply2.SetTotal(sdk.NewCoins(sdk.NewCoin(assets.MicroMedDenom, initTokens.MulRaw(int64(len(addrs))))))
	supplyKeeper.SetSupply(ctx, supply2)

	burnKeeper := keeper.NewKeeper(accountKeeper, bankKeeper, supplyKeeper)

	return TestInput{ctx, cdc, accountKeeper, bankKeeper, supplyKeeper, burnKeeper}
}

func TestGetAccount(t *testing.T) {
	input := createTestInput(t)

	beforeBurnCoin1 := input.AccountKeeper.GetAccount(input.Ctx, addrs[0]).GetCoins()
	beforeBurnCoin2 := input.AccountKeeper.GetAccount(input.Ctx, addrs[1]).GetCoins()
	beforeBurnCoin3 := input.AccountKeeper.GetAccount(input.Ctx, addrs[2]).GetCoins()
	beforeBurnTotal := input.SupplyKeeper.GetSupply(input.Ctx).GetTotal()

	require.Equal(t, initCoins, beforeBurnCoin1)
	require.Equal(t, initCoins, beforeBurnCoin2)
	require.Equal(t, initCoins, beforeBurnCoin3)
	// required 600 umed
	require.Equal(t, beforeBurnTotal, initCoins.Add(initCoins).Add(initCoins))

	err := input.Keeper.BurnCoins(input.Ctx, addrs[1].String())

	require.NoError(t, err)

	afterBurnCoin1 := input.AccountKeeper.GetAccount(input.Ctx, addrs[0]).GetCoins()
	afterBurnCoin2 := input.AccountKeeper.GetAccount(input.Ctx, addrs[1]).GetCoins()
	afterBurnCoin3 := input.AccountKeeper.GetAccount(input.Ctx, addrs[2]).GetCoins()
	afterBurnTotal := input.SupplyKeeper.GetSupply(input.Ctx).GetTotal()

	require.True(t, afterBurnCoin2.IsZero())
	require.Equal(t, initCoins, afterBurnCoin1)
	require.Equal(t, initCoins, afterBurnCoin3)
	// required 400 umed
	require.Equal(t, afterBurnTotal, initCoins.Add(initCoins))

}
