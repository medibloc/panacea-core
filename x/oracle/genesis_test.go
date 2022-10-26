package oracle_test

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/oracle"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/suite"
)

var (
	uniqueID = "uniqueID"

	oraclePrivKey        = secp256k1.GenPrivKey()
	oraclePubKey         = oraclePrivKey.PubKey()
	oracleAcc            = sdk.AccAddress(oraclePubKey.Address())
	oracleNodePrivKey, _ = btcec.NewPrivateKey(btcec.S256())
	oracleNodePubKey     = oracleNodePrivKey.PubKey()

	oracle2PrivKey = secp256k1.GenPrivKey()
	oracle2PubKey  = oracle2PrivKey.PubKey()
	oracle2Acc     = sdk.AccAddress(oracle2PubKey.Address())
)

type genesisTestSuite struct {
	testsuite.TestSuite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}

func makeSampleDate() (types.Oracle, types.OracleRegistration, types.OracleRegistrationVote, types.OracleUpgradeInfo) {
	return types.Oracle{
			Address:  oracleAcc.String(),
			Status:   types.ORACLE_STATUS_ACTIVE,
			Uptime:   100,
			JailedAt: nil,
		},
		types.OracleRegistration{
			UniqueId:               uniqueID,
			Address:                oracle2Acc.String(),
			NodePubKey:             oracleNodePubKey.SerializeCompressed(),
			NodePubKeyRemoteReport: nil,
			TrustedBlockHeight:     100,
			TrustedBlockHash:       nil,
			EncryptedOraclePrivKey: nil,
			Status:                 types.ORACLE_REGISTRATION_STATUS_PASSED,
			VotingPeriod: &types.VotingPeriod{
				VotingStartTime: time.Now(),
				VotingEndTime:   time.Now().Add(5 * time.Second),
			},
			TallyResult: &types.TallyResult{
				Yes: sdk.NewInt(5),
				No:  sdk.NewInt(1),
				InvalidYes: []*types.ConsensusTally{
					{
						ConsensusValue: []byte("invalidConsensusValue"),
						VotingAmount:   sdk.NewInt(1),
					},
				},
				ConsensusValue: []byte("encryptedOraclePrivKey"),
				ValidVoters: []*types.VoterInfo{
					{
						VoterAddress: oracleAcc.String(),
						VotingPower:  sdk.NewInt(5),
					},
				},
			},
		},
		types.OracleRegistrationVote{
			UniqueId:               uniqueID,
			VoterAddress:           oracleAcc.String(),
			VotingTargetAddress:    oracle2Acc.String(),
			VoteOption:             types.VOTE_OPTION_YES,
			EncryptedOraclePrivKey: []byte("encryptedOraclePrivKey"),
		},
		types.OracleUpgradeInfo{
			UniqueId: "UpgradeUniqueID",
			Height:   10000,
		}
}

func (m *genesisTestSuite) TestInitGenesis() {
	oracle1, oracleRegistration, oracleRegistrationVote, upgradeInfo := makeSampleDate()

	genesis := types.GenesisState{
		Oracles: []types.Oracle{
			oracle1,
		},
		OracleRegistrations: []types.OracleRegistration{
			oracleRegistration,
		},
		OracleRegistrationVotes: []types.OracleRegistrationVote{
			oracleRegistrationVote,
		},
		Params:            types.DefaultParams(),
		OracleUpgradeInfo: &upgradeInfo,
	}

	oracle.InitGenesis(m.Ctx, m.OracleKeeper, genesis)

	getOracle, err := m.OracleKeeper.GetOracle(m.Ctx, oracleAcc.String())
	m.Require().NoError(err)
	m.Require().Equal(genesis.Oracles[0], *getOracle)

	oracleUpgradeInfo, err := m.OracleKeeper.GetOracleUpgradeInfo(m.Ctx)
	m.Require().NoError(err)
	m.Require().Equal(&genesis.OracleUpgradeInfo, oracleUpgradeInfo)
}

func (m *genesisTestSuite) TestExportGenesis() {
	oracle1, oracleRegistration, oracleRegistrationVote, upgradeInfo := makeSampleDate()

	err := m.OracleKeeper.SetOracle(m.Ctx, &oracle1)
	m.Require().NoError(err)

	err = m.OracleKeeper.SetOracleRegistration(m.Ctx, &oracleRegistration)
	m.Require().NoError(err)

	err = m.OracleKeeper.SetOracleRegistrationVote(m.Ctx, &oracleRegistrationVote)
	m.Require().NoError(err)

	params := types.DefaultParams()
	m.OracleKeeper.SetParams(m.Ctx, params)

	err = m.OracleKeeper.SetOracleUpgradeInfo(m.Ctx, &upgradeInfo)
	m.Require().NoError(err)

	genesisStatus := oracle.ExportGenesis(m.Ctx, m.OracleKeeper)
	m.Require().Equal(oracle1, genesisStatus.Oracles[0])
	m.Require().Equal(oracleRegistration.UniqueId, genesisStatus.OracleRegistrations[0].UniqueId)
	m.Require().Equal(oracleRegistration.Address, genesisStatus.OracleRegistrations[0].Address)
	m.Require().Equal(oracleRegistration.NodePubKey, genesisStatus.OracleRegistrations[0].NodePubKey)
	m.Require().Equal(oracleRegistration.NodePubKeyRemoteReport, genesisStatus.OracleRegistrations[0].NodePubKeyRemoteReport)
	m.Require().Equal(oracleRegistration.TrustedBlockHeight, genesisStatus.OracleRegistrations[0].TrustedBlockHeight)
	m.Require().Equal(oracleRegistration.TrustedBlockHash, genesisStatus.OracleRegistrations[0].TrustedBlockHash)
	m.Require().Equal(oracleRegistration.EncryptedOraclePrivKey, genesisStatus.OracleRegistrations[0].EncryptedOraclePrivKey)
	m.Require().Equal(oracleRegistration.Status, genesisStatus.OracleRegistrations[0].Status)
	m.Require().Equal(oracleRegistration.VotingPeriod.VotingStartTime.Local(), genesisStatus.OracleRegistrations[0].VotingPeriod.VotingStartTime.Local())
	m.Require().Equal(oracleRegistration.VotingPeriod.VotingEndTime.Local(), genesisStatus.OracleRegistrations[0].VotingPeriod.VotingEndTime.Local())
	m.Require().Equal(oracleRegistration.TallyResult, genesisStatus.OracleRegistrations[0].TallyResult)
	m.Require().Equal(oracleRegistrationVote, genesisStatus.OracleRegistrationVotes[0])
	m.Require().Equal(params, genesisStatus.Params)
	m.Require().Equal(upgradeInfo, genesisStatus.OracleUpgradeInfo)
}
