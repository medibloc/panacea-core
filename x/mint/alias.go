package mint

import (
	"github.com/cosmos/cosmos-sdk/x/mint"
)

const (
	ModuleName            = mint.ModuleName
	DefaultParamspace     = mint.DefaultParamspace
	StoreKey              = mint.StoreKey
	QuerierRoute          = mint.QuerierRoute
	QueryParameters       = mint.QueryParameters
	QueryInflation        = mint.QueryInflation
	QueryAnnualProvisions = mint.QueryAnnualProvisions
)

var (
	// functions aliases
	NewKeeper            = mint.NewKeeper
	NewQuerier           = mint.NewQuerier
	NewMinter            = mint.NewMinter
	InitialMinter        = mint.InitialMinter
	DefaultInitialMinter = mint.DefaultInitialMinter
	ValidateMinter       = mint.ValidateMinter
	ParamKeyTable        = mint.ParamKeyTable
	NewParams            = mint.NewParams
	DefaultParams        = mint.DefaultParams
	ValidateParams       = mint.ValidateParams

	// variable aliases
	ModuleCdc              = mint.ModuleCdc
	MinterKey              = mint.MinterKey
	KeyMintDenom           = mint.KeyMintDenom
	KeyInflationRateChange = mint.KeyInflationRateChange
	KeyInflationMax        = mint.KeyInflationMax
	KeyInflationMin        = mint.KeyInflationMin
	KeyGoalBonded          = mint.KeyGoalBonded
	KeyBlocksPerYear       = mint.KeyBlocksPerYear
)

type (
	Keeper = mint.Keeper
	Minter = mint.Minter
	Params = mint.Params
)
