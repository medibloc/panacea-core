package testsuite

import (
	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
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
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	aolkeeper "github.com/medibloc/panacea-core/v2/x/aol/keeper"
	aoltypes "github.com/medibloc/panacea-core/v2/x/aol/types"
	burnkeeper "github.com/medibloc/panacea-core/v2/x/burn/keeper"
	burntypes "github.com/medibloc/panacea-core/v2/x/burn/types"
	"github.com/stretchr/testify/suite"

	"github.com/medibloc/panacea-core/v2/app/params"
	didkeeper "github.com/medibloc/panacea-core/v2/x/did/keeper"
	didtypes "github.com/medibloc/panacea-core/v2/x/did/types"
)

type TestProtocolVersionSetter struct{}

type TestSuite struct {
	suite.Suite

	Ctx sdk.Context

	AccountKeeper    authkeeper.AccountKeeper
	StakingKeeper    *stakingkeeper.Keeper
	AolKeeper        aolkeeper.Keeper
	AolMsgServer     aoltypes.MsgServiceServer
	BankKeeper       bankkeeper.Keeper
	BurnKeeper       burnkeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	DistrKeeper      distrkeeper.Keeper
	IBCKeeper        *ibckeeper.Keeper
	TransferKeeper   ibctransferkeeper.Keeper
	DIDMsgServer     didtypes.MsgServiceServer
	DIDKeeper        didkeeper.Keeper
	UpgradeKeeper    *upgradekeeper.Keeper
}

func (suite *TestSuite) SetupTest() {
	keyParams := sdk.NewKVStoreKeys(
		aoltypes.StoreKey,
		authtypes.StoreKey,
		banktypes.StoreKey,
		paramstypes.StoreKey,
		didtypes.StoreKey,
		ibcexported.StoreKey,
		capabilitytypes.StoreKey,
		ibctransfertypes.StoreKey,
		upgradetypes.StoreKey,
	)
	tKeyParams := sdk.NewTransientStoreKey(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	encodingConfig := newTestCodec()
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ctx := sdk.NewContext(ms, tmproto.Header{Time: time.Now()}, false, log.NewNopLogger())

	ms.MountStoreWithDB(tKeyParams, storetypes.StoreTypeTransient, db)
	for _, v := range keyParams {
		ms.MountStoreWithDB(v, storetypes.StoreTypeIAVL, db)
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
		burntypes.ModuleName:           {authtypes.Burner},
	}

	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	appCodec := encodingConfig.Codec
	legacyAmino := encodingConfig.Amino
	paramsKeeper := paramskeeper.NewKeeper(
		appCodec,
		legacyAmino,
		keyParams[paramstypes.StoreKey],
		tKeyParams)

	suite.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keyParams[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])

	scopedIBCKeeper := suite.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)

	suite.Ctx = ctx
	suite.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		keyParams[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		maccPerms,
		"panacea",
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	suite.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	suite.AolKeeper = *aolkeeper.NewKeeper(
		appCodec,
		keyParams[aoltypes.StoreKey],
		memKeys[aoltypes.MemStoreKey],
	)
	suite.AolMsgServer = aolkeeper.NewMsgServerImpl(suite.AolKeeper)
	suite.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		keyParams[banktypes.StoreKey],
		suite.AccountKeeper,
		BlockedAddresses(maccPerms),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	suite.BankKeeper.SetParams(ctx, banktypes.DefaultParams())
	suite.BurnKeeper = *burnkeeper.NewKeeper(suite.BankKeeper)
	suite.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		keyParams[stakingtypes.StoreKey],
		suite.AccountKeeper,
		suite.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	suite.DistrKeeper = distrkeeper.NewKeeper(
		appCodec, keyParams[distrtypes.StoreKey], suite.AccountKeeper, suite.BankKeeper,
		suite.StakingKeeper, "test_fee_collector",
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	skipUpgradeHeights := map[int64]bool{}
	homePath := suite.T().TempDir()
	suite.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, keyParams[upgradetypes.StoreKey], appCodec, homePath, NewTestProtocolVersionSetter(), authtypes.NewModuleAddress(govtypes.ModuleName).String())
	suite.IBCKeeper = ibckeeper.NewKeeper(
		appCodec, keyParams[ibcexported.StoreKey], paramsKeeper.Subspace(ibcexported.ModuleName), suite.StakingKeeper, suite.UpgradeKeeper, scopedIBCKeeper,
	)
	suite.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec, keyParams[ibctransfertypes.StoreKey], paramsKeeper.Subspace(ibctransfertypes.ModuleName),
		suite.IBCKeeper.ChannelKeeper,
		suite.IBCKeeper.ChannelKeeper, &suite.IBCKeeper.PortKeeper,
		suite.AccountKeeper, suite.BankKeeper, scopedIBCKeeper,
	)

	suite.DIDKeeper = *didkeeper.NewKeeper(
		appCodec,
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
		Codec:             marshaler,
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

func BlockedAddresses(maccPerms map[string][]string) map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	// allow the following addresses to receive funds
	delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	return modAccAddrs
}
