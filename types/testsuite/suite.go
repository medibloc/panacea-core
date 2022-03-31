package testsuite

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
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
	aolkeeper "github.com/medibloc/panacea-core/v2/x/aol/keeper"
	aoltypes "github.com/medibloc/panacea-core/v2/x/aol/types"
	burnkeeper "github.com/medibloc/panacea-core/v2/x/burn/keeper"
	datadealkeeper "github.com/medibloc/panacea-core/v2/x/datadeal/keeper"
	datadealtypes "github.com/medibloc/panacea-core/v2/x/datadeal/types"
	datapoolkeeper "github.com/medibloc/panacea-core/v2/x/datapool/keeper"
	datapooltypes "github.com/medibloc/panacea-core/v2/x/datapool/types"
	tokenkeeper "github.com/medibloc/panacea-core/v2/x/token/keeper"
	tokentypes "github.com/medibloc/panacea-core/v2/x/token/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types/time"
	dbm "github.com/tendermint/tm-db"

	"github.com/medibloc/panacea-core/v2/app/params"
	didkeeper "github.com/medibloc/panacea-core/v2/x/did/keeper"
	didtypes "github.com/medibloc/panacea-core/v2/x/did/types"
)

type TestSuite struct {
	suite.Suite

	Ctx sdk.Context

	AccountKeeper     authkeeper.AccountKeeper
	AolKeeper         aolkeeper.Keeper
	AolMsgServer      aoltypes.MsgServer
	BankKeeper        bankkeeper.Keeper
	BurnKeeper        burnkeeper.Keeper
	DIDMsgServer      didtypes.MsgServer
	DIDKeeper         didkeeper.Keeper
	TokenKeeper       tokenkeeper.Keeper
	DataDealKeeper    datadealkeeper.Keeper
	DataDealMsgServer datadealtypes.MsgServer
	DataPoolKeeper    datapoolkeeper.Keeper
	DataPoolMsgServer datapooltypes.MsgServer
	WasmKeeper        wasmkeeper.Keeper
}

func (suite *TestSuite) SetupTest() {
	keyParams := sdk.NewKVStoreKeys(
		aoltypes.StoreKey,
		authtypes.StoreKey,
		banktypes.StoreKey,
		paramstypes.StoreKey,
		didtypes.StoreKey,
		tokentypes.StoreKey,
		datadealtypes.StoreKey,
		datapooltypes.StoreKey,
	)
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
	suite.AolKeeper = *aolkeeper.NewKeeper(
		cdc.Marshaler,
		keyParams[aoltypes.StoreKey],
		memKeys[aoltypes.MemStoreKey],
	)
	suite.AolMsgServer = aolkeeper.NewMsgServerImpl(suite.AolKeeper)
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
	suite.TokenKeeper = *tokenkeeper.NewKeeper(
		cdc.Marshaler,
		keyParams[tokentypes.StoreKey],
		memKeys[tokentypes.MemStoreKey],
		suite.BankKeeper,
	)

	suite.DataDealKeeper = *datadealkeeper.NewKeeper(
		cdc.Marshaler,
		keyParams[datadealtypes.StoreKey],
		memKeys[datadealtypes.MemStoreKey],
		suite.BankKeeper,
		suite.AccountKeeper)
	suite.DataDealMsgServer = datadealkeeper.NewMsgServerImpl(suite.DataDealKeeper)

	suite.DataPoolKeeper = *datapoolkeeper.NewKeeper(
		cdc.Marshaler,
		keyParams[datapooltypes.StoreKey],
		memKeys[datapooltypes.MemStoreKey],
		paramsKeeper.Subspace(datapooltypes.ModuleName),
		suite.BankKeeper,
		suite.AccountKeeper,
		suite.WasmKeeper,
	)
	suite.DataPoolMsgServer = datapoolkeeper.NewMsgServerImpl(suite.DataPoolKeeper)
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
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	didtypes.RegisterInterfaces(interfaceRegistry)
	datadealtypes.RegisterInterfaces(interfaceRegistry)
	datapooltypes.RegisterInterfaces(interfaceRegistry)
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	return params.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          tx.NewTxConfig(marshaler, tx.DefaultSignModes),
		Amino:             cdc,
	}

}

func (suite *TestSuite) GetAccAddress() sdk.AccAddress {
	return sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
}
