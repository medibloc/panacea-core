package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/medibloc/panacea-core/v2/types/assets"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
	"github.com/stretchr/testify/suite"
)

type queryPoolTestSuite struct {
	testsuite.TestSuite
}

func TestQueryPoolTest(t *testing.T) {
	suite.Run(t, new(queryPoolTestSuite))
}

var (
	nftContractAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
)

func (suite queryPoolTestSuite) TestQueryDataPoolParams() {
	// set datapool module params
	params := &types.Params{
		DataPoolDepositRate:        types.DefaultDataPoolDepositRate,
		DataPoolNftContractAddress: nftContractAddr.String(),
		DataPoolCodeId:             2,
	}

	suite.DataPoolKeeper.SetParams(suite.Ctx, *params)

	res, err := suite.DataPoolKeeper.DataPoolParams(sdk.WrapSDKContext(suite.Ctx), &types.QueryDataPoolParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(params, res.GetParams())
}

func (suite *queryPoolTestSuite) TestQueryDataValidator() {
	dataValidator := types.DataValidator{
		Address:  dataVal1.String(),
		Endpoint: "https://my-validator-url.org",
	}
	err := suite.DataPoolKeeper.SetDataValidator(suite.Ctx, dataValidator)
	suite.Require().NoError(err)

	req := types.QueryDataValidatorRequest{
		Address: dataVal1.String(),
	}

	res, err := suite.DataPoolKeeper.DataValidator(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(dataValidator, *res.DataValidator)
}

func (suite *queryPoolTestSuite) TestQueryPool() {
	pool := suite.setPool()

	req := types.QueryPoolRequest{
		PoolId: pool.GetPoolId(),
	}
	res, err := suite.DataPoolKeeper.Pool(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	resultPool := res.Pool
	suite.Require().Equal(pool.PoolId, resultPool.PoolId)
	suite.Require().Equal(pool.PoolAddress, resultPool.PoolAddress)
	suite.Require().Equal(pool.Round, uint64(1))
	suite.Require().Equal(pool.PoolParams, resultPool.PoolParams)
	suite.Require().Equal(uint64(0), resultPool.CurNumData)
	suite.Require().Equal(pool.NumIssuedNfts, resultPool.NumIssuedNfts)
	suite.Require().Equal(types.PENDING, resultPool.Status)
	suite.Require().Equal(pool.Curator, resultPool.Curator)
}

func (suite queryPoolTestSuite) TestQueryDataValidationCertificates() {
	suite.setDataValidatorAccount()
	pool := suite.setPool()

	req := types.QueryDataValidationCertificatesRequest{
		PoolId: pool.GetPoolId(),
		Round:  pool.GetRound(),
	}

	res, err := suite.DataPoolKeeper.DataValidationCertificates(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Len(res.DataValidationCertificates, 0)

	dataHash1 := []byte("data1")
	cert1, err := makeTestDataCertificate(suite.Cdc.Marshaler, pool.GetPoolId(), pool.GetRound(), dataHash1, requesterAddr.String())
	suite.Require().NoError(err)

	suite.DataPoolKeeper.SetDataValidationCertificate(suite.Ctx, *cert1)

	dataHash2 := []byte("data2")
	cert2, err := makeTestDataCertificate(suite.Cdc.Marshaler, pool.GetPoolId(), pool.GetRound(), dataHash2, requesterAddr.String())
	suite.Require().NoError(err)

	suite.DataPoolKeeper.SetDataValidationCertificate(suite.Ctx, *cert2)

	res, err = suite.DataPoolKeeper.DataValidationCertificates(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Len(res.DataValidationCertificates, 2)
	suite.Require().Contains(res.DataValidationCertificates, *cert1)
	suite.Require().Contains(res.DataValidationCertificates, *cert2)
}

func (suite *queryPoolTestSuite) TestQueryDataPassRedeemReceipt() {
	pool := suite.setPool()

	dataPassRedeemReceipt := makeTestDataPassRedeemReceipt(pool.GetPoolId(), pool.GetRound(), uint64(suite.Ctx.BlockHeight()), dataVal1.String())

	err := suite.DataPoolKeeper.SetDataPassRedeemReceipt(suite.Ctx, *dataPassRedeemReceipt)
	suite.Require().NoError(err)

	req := &types.QueryDataPassRedeemReceiptRequest{
		PoolId: pool.GetPoolId(),
		Round:  pool.GetRound(),
		NftId:  1,
	}

	res, err := suite.DataPoolKeeper.DataPassRedeemReceipt(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(*dataPassRedeemReceipt, res.DataPassRedeemReceipt)
}

func (suite *queryPoolTestSuite) setDataValidatorAccount() {
	validatorAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, dataVal1)
	err := validatorAccount.SetPubKey(dataValPubKey)
	suite.Require().NoError(err)
	suite.AccountKeeper.SetAccount(suite.Ctx, validatorAccount)
}

func (suite *queryPoolTestSuite) setPool() *types.Pool {
	poolID := uint64(1)
	nftPrice := sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(1000000))
	downloadPeriod := time.Hour
	poolParams := types.PoolParams{
		DataSchema:            []string{"https://json.schemastore.org/github-issue-forms.json"},
		TargetNumData:         100,
		MaxNftSupply:          10,
		NftPrice:              &nftPrice,
		TrustedDataValidators: []string{dataVal1.String()},
		DownloadPeriod:        &downloadPeriod,
	}

	pool := types.NewPool(poolID, curatorAddr, poolParams)
	suite.DataPoolKeeper.SetPool(suite.Ctx, pool)

	return pool
}
