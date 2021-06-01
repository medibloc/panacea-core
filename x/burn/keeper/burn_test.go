package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/medibloc/panacea-core/x/burn/keeper"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types/time"
	dbm "github.com/tendermint/tm-db"
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/medibloc/panacea-core/app/params"
	"github.com/medibloc/panacea-core/types/assets"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

var (
	pubKeys = []crypto.PubKey{
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
	}

	address = []sdk.AccAddress{
		sdk.AccAddress(pubKeys[0].Address()),
		sdk.AccAddress(pubKeys[1].Address()),
		sdk.AccAddress(pubKeys[2].Address()),
	}

	initTokens = sdk.TokensFromConsensusPower(200)
	initCoin = sdk.NewCoin(assets.MicroMedDenom, initTokens)
	initCoins  = sdk.NewCoins(initCoin)
)

func newTestCodec() params.EncodingConfig {
	cdc := codec.NewLegacyAmino()
	interfaceRegistry := types.NewInterfaceRegistry()
	banktypes.RegisterInterfaces(interfaceRegistry)
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	return params.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          tx.NewTxConfig(marshaler, tx.DefaultSignModes),
		Amino:             cdc,
	}

}

type TestInput struct {
	suite.Suite

	Ctx           sdk.Context

	AccountKeeper authkeeper.AccountKeeper
	BankKeeper    bankkeeper.Keeper
	Keeper        keeper.Keeper
}

func (suite *TestInput) SetupTest() {
	keyBank := sdk.NewKVStoreKey(banktypes.StoreKey)
	keyParams := sdk.NewKVStoreKey(paramstypes.StoreKey)
	tKeyParams := sdk.NewTransientStoreKey(paramstypes.TStoreKey)

	cdc := newTestCodec()
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ctx := sdk.NewContext(ms, tmproto.Header{Time: time.Now()}, false, log.NewNopLogger())

	ms.MountStoreWithDB(keyBank, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tKeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)

	suite.Require().NoError(ms.LoadLatestVersion())

	maccPerms := map[string][]string{
		authtypes.FeeCollectorName: nil,
	}

	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	paramsKeeper := paramskeeper.NewKeeper(
		cdc.Marshaler,
		cdc.Amino,
		keyParams,
		tKeyParams)
	accountKeeper := authkeeper.NewAccountKeeper(
		cdc.Marshaler,
		keyBank,
		paramsKeeper.Subspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount,
		maccPerms,
	)
	accountKeeper.SetParams(ctx, authtypes.DefaultParams())
	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc.Marshaler,
		keyBank,
		accountKeeper,
		paramsKeeper.Subspace(banktypes.ModuleName),
		modAccAddrs,
	)
	bankKeeper.SetParams(ctx, banktypes.DefaultParams())
	burnKeeper := keeper.NewKeeper(bankKeeper)

	for _, addr := range address {
		err := bankKeeper.AddCoins(ctx, addr, initCoins)
		suite.Require().NoError(err)
	}

	bankKeeper.SetSupply(ctx, banktypes.DefaultSupply())
	supply := bankKeeper.GetSupply(ctx)
	supply.SetTotal(sdk.NewCoins(sdk.NewCoin("umed", sdk.TokensFromConsensusPower(600))))
	bankKeeper.SetSupply(ctx, supply)

	totalCoins := bankKeeper.GetSupply(ctx).GetTotal()

	suite.Require().Equal(1, len(totalCoins))
	suite.Require().Equal(assets.MicroMedDenom, totalCoins[0].Denom)
	suite.Require().Equal(sdk.TokensFromConsensusPower(600), totalCoins[0].Amount)

	suite.Ctx = ctx
	suite.AccountKeeper = accountKeeper
	suite.BankKeeper = bankKeeper
	suite.Keeper = *burnKeeper


}

func (suite *TestInput) TestBurnCoins() {
	ctx := suite.Ctx
	bankKeeper := suite.BankKeeper
	burnKeeper := suite.Keeper

	beforeBurnCoin1 := bankKeeper.GetBalance(ctx, address[0], assets.MicroMedDenom)
	beforeBurnCoin2 := bankKeeper.GetBalance(ctx, address[1], assets.MicroMedDenom)
	beforeBurnCoin3 := bankKeeper.GetBalance(ctx, address[2], assets.MicroMedDenom)
	beforeBurnTotal := bankKeeper.GetSupply(ctx).GetTotal()

	suite.Require().Equal(initCoin, beforeBurnCoin1)
	suite.Require().Equal(initCoin, beforeBurnCoin2)
	suite.Require().Equal(initCoin, beforeBurnCoin3)
	suite.Require().Equal(assets.MicroMedDenom, beforeBurnTotal[0].GetDenom())
	suite.Require().Equal(sdk.TokensFromConsensusPower(600), beforeBurnTotal[0].Amount)

	err := burnKeeper.BurnCoins(ctx, address[1].String())

	suite.Require().NoError(err)

	afterBurnCoin1 := bankKeeper.GetBalance(ctx, address[0], assets.MicroMedDenom)
	afterBurnCoin2 := bankKeeper.GetBalance(ctx, address[1], assets.MicroMedDenom)
	afterBurnCoin3 := bankKeeper.GetBalance(ctx, address[2], assets.MicroMedDenom)
	afterBurnTotal := bankKeeper.GetSupply(ctx).GetTotal()

	suite.Require().True(afterBurnCoin2.IsZero())
	suite.Require().Equal(initCoin, afterBurnCoin1)
	suite.Require().Equal(initCoin, afterBurnCoin3)
	suite.Require().Equal(assets.MicroMedDenom, afterBurnTotal[0].GetDenom())
	suite.Require().Equal(sdk.TokensFromConsensusPower(400), afterBurnTotal[0].Amount)
}

func (suite *TestInput)TestGetAccount_NotExistAddress() {
	ctx := suite.Ctx
	bankKeeper := suite.BankKeeper
	burnKeeper := suite.Keeper

	beforeBurnCoin1 := bankKeeper.GetBalance(ctx, address[0], assets.MicroMedDenom)
	beforeBurnCoin2 := bankKeeper.GetBalance(ctx, address[1], assets.MicroMedDenom)
	beforeBurnCoin3 := bankKeeper.GetBalance(ctx, address[2], assets.MicroMedDenom)
	beforeBurnTotal := bankKeeper.GetSupply(ctx).GetTotal()

	suite.Require().Equal(initCoin, beforeBurnCoin1)
	suite.Require().Equal(initCoin, beforeBurnCoin2)
	suite.Require().Equal(initCoin, beforeBurnCoin3)
	suite.Require().Equal(assets.MicroMedDenom, beforeBurnTotal[0].GetDenom())
	suite.Require().Equal(sdk.TokensFromConsensusPower(600), beforeBurnTotal[0].Amount)

	err := burnKeeper.BurnCoins(ctx, "invalid address")

	suite.Require().Error(err)

	afterBurnCoin1 := bankKeeper.GetBalance(ctx, address[0], assets.MicroMedDenom)
	afterBurnCoin2 := bankKeeper.GetBalance(ctx, address[1], assets.MicroMedDenom)
	afterBurnCoin3 := bankKeeper.GetBalance(ctx, address[2], assets.MicroMedDenom)
	afterBurnTotal := bankKeeper.GetSupply(ctx).GetTotal()

	suite.Require().Equal(initCoin, afterBurnCoin1)
	suite.Require().Equal(initCoin, afterBurnCoin2)
	suite.Require().Equal(initCoin, afterBurnCoin3)
	suite.Require().Equal(assets.MicroMedDenom, afterBurnTotal[0].GetDenom())
	suite.Require().Equal(sdk.TokensFromConsensusPower(600), afterBurnTotal[0].Amount)

}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestInput))
}
