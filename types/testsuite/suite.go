package testsuite

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v2/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	ibchost "github.com/cosmos/ibc-go/v2/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v2/modules/core/keeper"
	aolkeeper "github.com/medibloc/panacea-core/v2/x/aol/keeper"
	aoltypes "github.com/medibloc/panacea-core/v2/x/aol/types"
	burnkeeper "github.com/medibloc/panacea-core/v2/x/burn/keeper"
	burntypes "github.com/medibloc/panacea-core/v2/x/burn/types"
	oraclekeeper "github.com/medibloc/panacea-core/v2/x/oracle/keeper"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types/time"
	dbm "github.com/tendermint/tm-db"

	"github.com/medibloc/panacea-core/v2/app/params"
	datadealkeeper "github.com/medibloc/panacea-core/v2/x/datadeal/keeper"
	datadealtypes "github.com/medibloc/panacea-core/v2/x/datadeal/types"
	didkeeper "github.com/medibloc/panacea-core/v2/x/did/keeper"
	didtypes "github.com/medibloc/panacea-core/v2/x/did/types"
)

type TestProtocolVersionSetter struct{}

type TestSuite struct {
	suite.Suite

	Ctx sdk.Context

	AccountKeeper     authkeeper.AccountKeeper
	StakingKeeper     stakingkeeper.Keeper
	AolKeeper         aolkeeper.Keeper
	AolMsgServer      aoltypes.MsgServer
	BankKeeper        bankkeeper.Keeper
	BurnKeeper        burnkeeper.Keeper
	CapabilityKeeper  *capabilitykeeper.Keeper
	DistrKeeper       distrkeeper.Keeper
	IBCKeeper         *ibckeeper.Keeper
	TransferKeeper    ibctransferkeeper.Keeper
	DIDMsgServer      didtypes.MsgServer
	DIDKeeper         didkeeper.Keeper
	DataDealKeeper    datadealkeeper.Keeper
	DataDealMsgServer datadealtypes.MsgServer
	OracleKeeper      oraclekeeper.Keeper
	OracleMsgServer   oracletypes.MsgServer
	WasmKeeper        wasm.Keeper
	UpgradeKeeper     upgradekeeper.Keeper
}

func (suite *TestSuite) SetupTest() {
	keyParams := sdk.NewKVStoreKeys(
		aoltypes.StoreKey,
		authtypes.StoreKey,
		banktypes.StoreKey,
		paramstypes.StoreKey,
		didtypes.StoreKey,
		datadealtypes.StoreKey,
		oracletypes.StoreKey,
		wasm.StoreKey,
		ibchost.StoreKey,
		capabilitytypes.StoreKey,
		ibctransfertypes.StoreKey,
		upgradetypes.StoreKey,
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
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		wasm.ModuleName:                {authtypes.Burner},
		burntypes.ModuleName:           {authtypes.Burner},
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

	suite.CapabilityKeeper = capabilitykeeper.NewKeeper(cdc.Marshaler, keyParams[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])

	scopedIBCKeeper := suite.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)

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
	suite.DistrKeeper = distrkeeper.NewKeeper(
		cdc.Marshaler, keyParams[distrtypes.StoreKey], paramsKeeper.Subspace(distrtypes.ModuleName), suite.AccountKeeper, suite.BankKeeper, &suite.StakingKeeper, "test_fee_collector", modAccAddrs,
	)
	suite.UpgradeKeeper = upgradekeeper.NewKeeper(map[int64]bool{}, keyParams[upgradetypes.StoreKey], cdc.Marshaler, suite.T().TempDir(), NewTestProtocolVersionSetter())
	suite.IBCKeeper = ibckeeper.NewKeeper(
		cdc.Marshaler, keyParams[ibchost.StoreKey], paramsKeeper.Subspace(ibchost.ModuleName), suite.StakingKeeper, suite.UpgradeKeeper, scopedIBCKeeper,
	)
	suite.TransferKeeper = ibctransferkeeper.NewKeeper(
		cdc.Marshaler, keyParams[ibctransfertypes.StoreKey], paramsKeeper.Subspace(ibctransfertypes.ModuleName),
		suite.IBCKeeper.ChannelKeeper, &suite.IBCKeeper.PortKeeper,
		suite.AccountKeeper, suite.BankKeeper, scopedIBCKeeper,
	)

	msgRouter := baseapp.NewMsgServiceRouter()

	querier := baseapp.NewGRPCQueryRouter()

	supportedFeatures := "iterator,staking,stargate"
	suite.WasmKeeper = wasmkeeper.NewKeeper(
		cdc.Marshaler,
		keyParams[wasm.StoreKey],
		paramsKeeper.Subspace(wasm.ModuleName),
		suite.AccountKeeper,
		suite.BankKeeper,
		suite.StakingKeeper,
		suite.DistrKeeper,
		suite.IBCKeeper.ChannelKeeper,
		&suite.IBCKeeper.PortKeeper,
		scopedIBCKeeper,
		suite.TransferKeeper,
		msgRouter,
		querier,
		suite.T().TempDir(),
		wasmtypes.DefaultWasmConfig(),
		supportedFeatures,
		[]wasm.Option{}...,
	)

	wasmGenState := wasmtypes.GenesisState{
		Codes:     []wasmtypes.Code{},
		Sequences: []wasmtypes.Sequence{},
		Params:    wasmtypes.DefaultParams(),
	}
	_, err := wasmkeeper.InitGenesis(suite.Ctx, &suite.WasmKeeper, wasmGenState, suite.StakingKeeper, wasmkeeper.TestHandler(wasmkeeper.NewDefaultPermissionKeeper(suite.WasmKeeper)))
	suite.Require().NoError(err)

	suite.DIDKeeper = *didkeeper.NewKeeper(
		cdc.Marshaler,
		keyParams[didtypes.StoreKey],
		memKeys[didtypes.MemStoreKey],
	)
	suite.DIDMsgServer = didkeeper.NewMsgServerImpl(suite.DIDKeeper)

	suite.DataDealKeeper = *datadealkeeper.NewKeeper(
		cdc.Marshaler,
		keyParams[datadealtypes.StoreKey],
		memKeys[datadealtypes.MemStoreKey],
		suite.AccountKeeper,
		suite.BankKeeper,
	)
	suite.DataDealMsgServer = datadealkeeper.NewMsgServerImpl(suite.DataDealKeeper)

	suite.OracleKeeper = *oraclekeeper.NewKeeper(
		cdc.Marshaler,
		keyParams[oracletypes.StoreKey],
		memKeys[oracletypes.MemStoreKey],
		paramsKeeper.Subspace(oracletypes.ModuleName),
	)

	suite.OracleMsgServer = oraclekeeper.NewMsgServerImpl(suite.OracleKeeper)
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
	//TODO: When other Msgs are implemented, I'll remove this comment.
	//datadealtypes.RegisterInterfaces(interfaceRegistry)
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

func (suite *TestSuite) FundAccount(ctx sdk.Context, addr sdk.AccAddress, amounts sdk.Coins) error {
	if err := suite.BankKeeper.MintCoins(ctx, minttypes.ModuleName, amounts); err != nil {
		return err
	}

	return suite.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, amounts)
}

func NewTestProtocolVersionSetter() TestProtocolVersionSetter {
	return TestProtocolVersionSetter{}
}

func (vs TestProtocolVersionSetter) SetProtocolVersion(v uint64) {}
