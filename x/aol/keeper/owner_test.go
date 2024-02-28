package keeper_test

import (
	"testing"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	aoltypes "github.com/medibloc/panacea-core/v2/x/aol/types"
	"github.com/stretchr/testify/suite"
)

var (
	pubKeys = []crypto.PubKey{
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
	}

	addresses = []sdk.AccAddress{
		sdk.AccAddress(pubKeys[0].Address()),
		sdk.AccAddress(pubKeys[1].Address()),
	}
)

type ownerTestSuite struct {
	testsuite.TestSuite
}

func TestOwnerKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(ownerTestSuite))
}

func (suite *ownerTestSuite) TestOneOwner() {
	ctx := suite.Ctx
	aolKeeper := suite.AolKeeper

	key := aoltypes.OwnerCompositeKey{
		OwnerAddress: addresses[0],
	}
	owner := aoltypes.Owner{
		TotalTopics: 1,
	}
	aolKeeper.SetOwner(ctx, key, owner)

	// verify HasOwner
	suite.Require().True(aolKeeper.HasOwner(ctx, key))

	// verify GetOwner
	resultOwner := aolKeeper.GetOwner(ctx, key)
	suite.Require().Equal(owner, resultOwner)

	// verify GetAllOwner
	resultKeys, resultOwners := aolKeeper.GetAllOwners(ctx)
	suite.Require().Equal(1, len(resultKeys))
	suite.Require().Equal([]aoltypes.OwnerCompositeKey{key}, resultKeys)
	suite.Require().Equal(1, len(resultOwners))
	suite.Require().Equal([]aoltypes.Owner{owner}, resultOwners)
}

func (suite *ownerTestSuite) TestMultiOwner() {
	ctx := suite.Ctx
	aolKeeper := suite.AolKeeper

	key := aoltypes.OwnerCompositeKey{
		OwnerAddress: addresses[0],
	}
	key2 := aoltypes.OwnerCompositeKey{
		OwnerAddress: addresses[1],
	}
	owner := aoltypes.Owner{
		TotalTopics: 1,
	}
	owner2 := aoltypes.Owner{
		TotalTopics: 10,
	}
	aolKeeper.SetOwner(ctx, key, owner)
	aolKeeper.SetOwner(ctx, key2, owner2)

	// verify HasOwner
	suite.Require().True(aolKeeper.HasOwner(ctx, key))
	suite.Require().True(aolKeeper.HasOwner(ctx, key2))

	// verify GetOwner
	resultOwner := aolKeeper.GetOwner(ctx, key)
	suite.Require().Equal(owner, resultOwner)
	resultOwner2 := aolKeeper.GetOwner(ctx, key2)
	suite.Require().Equal(owner2, resultOwner2)

	// verify GetAllOwner
	resultKeys, resultOwners := aolKeeper.GetAllOwners(ctx)
	suite.Require().Equal(2, len(resultKeys))
	suite.Require().Contains(resultKeys, key)
	suite.Require().Contains(resultKeys, key2)
	suite.Require().Equal(2, len(resultOwners))
	suite.Require().Contains(resultOwners, owner)
	suite.Require().Contains(resultOwners, owner2)
}
