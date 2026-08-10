package keeper

import (
	"testing"

	"cosmossdk.io/collections"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Live lifecycle provenance must reject module accounts exactly like the burn
// tombstone path does. Both stores hold the same controller authority, so a
// future change to one path must not silently leave the other accepting a
// module account.
func TestNFTRecordRejectsModuleAccountLifecycleProvenance(t *testing.T) {
	for _, tc := range []struct {
		name          string
		revocable     bool
		expectedError string
		mutate        func(lifecycle *nfttypes.LifecycleRecord, moduleAccount string)
	}{
		{
			name:          "minted_by",
			expectedError: "lifecycle has invalid minter",
			mutate: func(lifecycle *nfttypes.LifecycleRecord, moduleAccount string) {
				lifecycle.Mint.MintedBy = moduleAccount
			},
		},
		{
			name:          "revoked_by",
			revocable:     true,
			expectedError: "lifecycle has invalid revoker",
			mutate: func(lifecycle *nfttypes.LifecycleRecord, moduleAccount string) {
				lifecycle.Revocation.RevokedBy = moduleAccount
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newKeeperFixture(t, true, true)
			classID, controller, _, _ := createNFTForRevokeTest(t, &fixture, tc.revocable)
			if tc.revocable {
				_, err := NewMsgServer(fixture.keeper).Revoke(
					fixture.ctx,
					&nfttypes.MsgRevokeRequest{
						ClassId:    classID,
						NftId:      "nft-1",
						Controller: controller,
					},
				)
				require.NoError(t, err)
			}

			key := collections.Join(classID, "nft-1")
			lifecycle, err := fixture.keeper.lifecycles.Get(fixture.ctx, key)
			require.NoError(t, err)
			moduleAccount := fixture.accountAddress(t, fixture.moduleAccountAddresses[0])
			tc.mutate(&lifecycle, moduleAccount)
			require.NoError(t, fixture.keeper.lifecycles.Set(fixture.ctx, key, lifecycle))

			_, err = NewQueryServer(fixture.keeper).NFTRecord(
				fixture.ctx,
				&nfttypes.QueryNFTRecordRequest{ClassId: classID, NftId: "nft-1"},
			)
			require.Equal(t, codes.Internal, status.Code(err))
			require.ErrorContains(t, err, tc.expectedError)
			require.ErrorContains(t, err, "must not be a module account")
		})
	}
}
