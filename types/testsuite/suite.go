package testsuite

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
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
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	ibctransferkeeper "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/keeper"
	ibctransfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	ibchost "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	ibckeeper "github.com/cosmos/cosmos-sdk/x/ibc/core/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	aolkeeper "github.com/medibloc/panacea-core/v2/x/aol/keeper"
	aoltypes "github.com/medibloc/panacea-core/v2/x/aol/types"
	burnkeeper "github.com/medibloc/panacea-core/v2/x/burn/keeper"
	datadealkeeper "github.com/medibloc/panacea-core/v2/x/datadeal/keeper"
	datadealtypes "github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/medibloc/panacea-core/v2/x/datapool"
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
	Cdc params.EncodingConfig

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
	TokenKeeper       tokenkeeper.Keeper
	DataDealKeeper    datadealkeeper.Keeper
	DataDealMsgServer datadealtypes.MsgServer
	DataPoolKeeper    datapoolkeeper.Keeper
	DataPoolMsgServer datapooltypes.MsgServer
	WasmKeeper        wasm.Keeper
}

func (suite *TestSuite) SetupTest() {
	keyParams := sdk.NewKVStoreKeys(
		aoltypes.StoreKey,
		authtypes.StoreKey,
		banktypes.StoreKey,
		distrtypes.StoreKey,
		stakingtypes.StoreKey,
		paramstypes.StoreKey,
		didtypes.StoreKey,
		tokentypes.StoreKey,
		datadealtypes.StoreKey,
		datapooltypes.StoreKey,
		wasm.StoreKey,
		ibchost.StoreKey,
		capabilitytypes.StoreKey,
		ibctransfertypes.StoreKey,
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
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		wasm.ModuleName:                {authtypes.Burner},
		datapooltypes.ModuleName:       nil,
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
	suite.Cdc = cdc
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
	suite.StakingKeeper = stakingkeeper.NewKeeper(
		cdc.Marshaler, keyParams[stakingtypes.StoreKey], suite.AccountKeeper, suite.BankKeeper, paramsKeeper.Subspace(stakingtypes.ModuleName),
	)
	suite.BurnKeeper = *burnkeeper.NewKeeper(suite.BankKeeper)
	suite.DistrKeeper = distrkeeper.NewKeeper(
		cdc.Marshaler, keyParams[distrtypes.StoreKey], paramsKeeper.Subspace(distrtypes.ModuleName), suite.AccountKeeper, suite.BankKeeper, &suite.StakingKeeper, "test_fee_collector", modAccAddrs,
	)
	suite.IBCKeeper = ibckeeper.NewKeeper(
		cdc.Marshaler, keyParams[ibchost.StoreKey], paramsKeeper.Subspace(ibchost.ModuleName), suite.StakingKeeper, scopedIBCKeeper,
	)
	suite.TransferKeeper = ibctransferkeeper.NewKeeper(
		cdc.Marshaler, keyParams[ibctransfertypes.StoreKey], paramsKeeper.Subspace(ibctransfertypes.ModuleName),
		suite.IBCKeeper.ChannelKeeper, &suite.IBCKeeper.PortKeeper,
		suite.AccountKeeper, suite.BankKeeper, scopedIBCKeeper,
	)

	router := baseapp.NewRouter()

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
		router,
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

	dataPoolGenState := datapooltypes.GenesisState{
		DataValidators: []*datapooltypes.DataValidator{},
		NextPoolNumber: 1,
		Pools:          []*datapooltypes.Pool{},
		Params:         datapooltypes.Params{DataPoolDeposit: sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(1000000))},
	}
	datapool.InitGenesis(suite.Ctx, suite.DataPoolKeeper, dataPoolGenState)
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
