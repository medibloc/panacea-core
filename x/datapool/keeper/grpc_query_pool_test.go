package keeper_test

import (
	"testing"

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
		DataPoolCommissionRate:     types.DefaultDataPoolCommissionRate,
		DataPoolNftContractAddress: nftContractAddr.String(),
		DataPoolCodeId:             2,
	}

	suite.DataPoolKeeper.SetParams(suite.Ctx, *params)

	res, err := suite.DataPoolKeeper.DataPoolParams(sdk.WrapSDKContext(suite.Ctx), &types.QueryDataPoolParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(params, res.GetParams())
}

func (suite *queryPoolTestSuite) TestQueryOracle() {
	oracle := types.Oracle{
		Address:  oracle1.String(),
		Endpoint: "https://my-oracle-url.org",
	}
	err := suite.DataPoolKeeper.SetOracle(suite.Ctx, oracle)
	suite.Require().NoError(err)

	req := types.QueryOracleRequest{
		Address: oracle1.String(),
	}

	res, err := suite.DataPoolKeeper.Oracle(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(oracle, *res.Oracle)
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
	suite.setOracleAccount()
	pool := suite.setPool()

	req := types.QueryDataCertsRequest{
		PoolId: pool.GetPoolId(),
		Round:  pool.GetRound(),
	}

	res, err := suite.DataPoolKeeper.DataCerts(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Len(res.DataCerts, 0)

	dataHash1 := []byte("data1")
	cert1, err := makeTestDataCert(suite.Cdc.Marshaler, pool.GetPoolId(), pool.GetRound(), dataHash1, requesterAddr.String())
	suite.Require().NoError(err)

	suite.DataPoolKeeper.SetDataCert(suite.Ctx, *cert1)

	dataHash2 := []byte("data2")
	cert2, err := makeTestDataCert(suite.Cdc.Marshaler, pool.GetPoolId(), pool.GetRound(), dataHash2, requesterAddr.String())
	suite.Require().NoError(err)

	suite.DataPoolKeeper.SetDataCert(suite.Ctx, *cert2)

	res, err = suite.DataPoolKeeper.DataCerts(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().Len(res.DataCerts, 2)
	suite.Require().Contains(res.DataCerts, *cert1)
	suite.Require().Contains(res.DataCerts, *cert2)
}

func (suite *queryPoolTestSuite) TestQueryDataPassRedeemReceipt() {
	pool := suite.setPool()

	dataPassRedeemReceipt := makeTestDataPassRedeemReceipt(pool.GetPoolId(), pool.GetRound(), oracle1.String())

	suite.DataPoolKeeper.SetDataPassRedeemReceipt(suite.Ctx, *dataPassRedeemReceipt)

	req := &types.QueryDataPassRedeemReceiptRequest{
		PoolId:     pool.GetPoolId(),
		Round:      pool.GetRound(),
		DataPassId: 1,
	}

	res, err := suite.DataPoolKeeper.DataPassRedeemReceipt(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(*dataPassRedeemReceipt, res.DataPassRedeemReceipt)
}

func (suite *queryPoolTestSuite) TestQueryRedeemReceiptsList() {
	pool := suite.setPool()

	redeemer := oracle1.String()
	poolID := pool.GetPoolId()

	redeemHistory := makeTestDataPassRedeemHistory(poolID, pool.GetRound(), redeemer)

	suite.DataPoolKeeper.SetDataPassRedeemHistory(suite.Ctx, *redeemHistory)

	req := &types.QueryDataPassRedeemHistoryRequest{
		Redeemer: redeemer,
		PoolId:   poolID,
	}

	res, err := suite.DataPoolKeeper.DataPassRedeemHistory(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(*redeemHistory, res.DataPassRedeemHistories)
}

func (suite *queryPoolTestSuite) setOracleAccount() {
	oracleAccount := suite.AccountKeeper.NewAccountWithAddress(suite.Ctx, oracle1)
	err := oracleAccount.SetPubKey(oraclePubKey)
	suite.Require().NoError(err)
	suite.AccountKeeper.SetAccount(suite.Ctx, oracleAccount)
}

func (suite *queryPoolTestSuite) setPool() *types.Pool {
	poolID := uint64(1)
	nftPrice := sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(1000000))
	poolParams := types.PoolParams{
		DataSchema:     []string{"https://json.schemastore.org/github-issue-forms.json"},
		TargetNumData:  100,
		MaxNftSupply:   10,
		NftPrice:       &nftPrice,
		TrustedOracles: []string{oracle1.String()},
	}

	pool := types.NewPool(poolID, curatorAddr, poolParams)
	suite.DataPoolKeeper.SetPool(suite.Ctx, pool)

	return pool
}
