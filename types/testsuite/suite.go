package testsuite

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	burnkeeper "github.com/medibloc/panacea-core/x/burn/keeper"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types/time"
	dbm "github.com/tendermint/tm-db"

	"github.com/medibloc/panacea-core/app/params"
	didkeeper "github.com/medibloc/panacea-core/x/did/keeper"
	didtypes "github.com/medibloc/panacea-core/x/did/types"
)

type TestSuite struct {
	suite.Suite

	Ctx sdk.Context

	AccountKeeper authkeeper.AccountKeeper
	BankKeeper    bankkeeper.Keeper
	BurnKeeper    burnkeeper.Keeper
	DIDMsgServer  didtypes.MsgServer
	DIDKeeper     didkeeper.Keeper
}

func (suite *TestSuite) SetupTest() {
	keyParams := sdk.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		paramstypes.StoreKey,
		didtypes.StoreKey)
	/*didKeyParam := sdk.NewKVStoreKey(didtypes.StoreKey)*/
	tKeyParams := sdk.NewTransientStoreKey(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	cdc := newTestCodec()
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ctx := sdk.NewContext(ms, tmproto.Header{Time: time.Now()}, false, log.NewNopLogger())

	ms.MountStoreWithDB(tKeyParams, sdk.StoreTypeTransient, db)
	for _, v := range keyParams {
		ms.MountStoreWithDB(v, sdk.StoreTypeIAVL, db)
	}

	sdk.GetConfig().SetBech32PrefixForAccount("panacea", "panaceapub")

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
		keyParams[paramstypes.StoreKey],
		tKeyParams)

	suite.Ctx = ctx
	suite.AccountKeeper = authkeeper.NewAccountKeeper(
		cdc.Marshaler,
		keyParams[authtypes.StoreKey],
		paramsKeeper.Subspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount,
		maccPerms,
		)
	suite.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	suite.BankKeeper = bankkeeper.NewBaseKeeper(
		cdc.Marshaler,
		keyParams[banktypes.StoreKey],
		suite.AccountKeeper,
		paramsKeeper.Subspace(banktypes.ModuleName),
		modAccAddrs,
		)
	suite.BankKeeper.SetParams(ctx, banktypes.DefaultParams())
	suite.BurnKeeper = *burnkeeper.NewKeeper(suite.BankKeeper)
	suite.DIDKeeper = *didkeeper.NewKeeper(
		cdc.Marshaler,
		keyParams[didtypes.StoreKey],
		memKeys[didtypes.MemStoreKey],
	)
	suite.DIDMsgServer = didkeeper.NewMsgServerImpl(suite.DIDKeeper)
}

func (suite *TestSuite) BeforeTest(suiteName, testName string) {
	log.NewNopLogger().Info("Pass BeforeTest. suiteName: %s, testName: %s", suiteName, testName)
}

func (suite *TestSuite) AfterTest(suiteName, testName string) {
	log.NewNopLogger().Info("Pass AfterTest. suiteName: %s, testName: %s", suiteName, testName)
}

func newTestCodec() params.EncodingConfig {
	cdc := codec.NewLegacyAmino()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	authtypes.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	didtypes.RegisterInterfaces(interfaceRegistry)
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	return params.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          tx.NewTxConfig(marshaler, tx.DefaultSignModes),
		Amino:             cdc,
	}

}
