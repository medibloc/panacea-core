package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
)

func TestLegacyMsgNestedContainerRoundTrip(t *testing.T) {
	cdc := newLegacyContainerTestCodec()

	testCases := []struct {
		name          string
		newOriginal   func(*testing.T) proto.Message
		newDecoded    func() proto.Message
		assertDecoded func(*testing.T, proto.Message)
	}{
		{
			name: "stored gov proposal",
			newOriginal: func(t *testing.T) proto.Message {
				return &govv1.Proposal{
					Id:       1,
					Messages: []*cdctypes.Any{newRawLegacyMintAny(t)},
					Status:   govv1.StatusVotingPeriod,
					Title:    "legacy PNFT proposal",
					Summary:  "execute a legacy PNFT message",
				}
			},
			newDecoded: func() proto.Message { return &govv1.Proposal{} },
			assertDecoded: func(t *testing.T, decoded proto.Message) {
				proposal, ok := decoded.(*govv1.Proposal)
				require.True(t, ok)
				require.Len(t, proposal.Messages, 1)
				assertCachedLegacyMint(t, proposal.Messages[0])
			},
		},
		{
			name: "stored group proposal",
			newOriginal: func(t *testing.T) proto.Message {
				return &group.Proposal{
					Id:                 1,
					GroupPolicyAddress: "panacea1grouppolicy",
					Proposers:          []string{"panacea1proposer"},
					Status:             group.PROPOSAL_STATUS_SUBMITTED,
					Messages:           []*cdctypes.Any{newRawLegacyMintAny(t)},
					Title:              "legacy PNFT proposal",
					Summary:            "execute a legacy PNFT message",
				}
			},
			newDecoded: func() proto.Message { return &group.Proposal{} },
			assertDecoded: func(t *testing.T, decoded proto.Message) {
				proposal, ok := decoded.(*group.Proposal)
				require.True(t, ok)
				require.Len(t, proposal.Messages, 1)
				assertCachedLegacyMint(t, proposal.Messages[0])
			},
		},
		{
			name: "gov proposal submission",
			newOriginal: func(t *testing.T) proto.Message {
				return &govv1.MsgSubmitProposal{
					Messages: []*cdctypes.Any{newRawLegacyMintAny(t)},
					Proposer: "panacea1proposer",
					Title:    "legacy PNFT proposal",
					Summary:  "execute a legacy PNFT message",
				}
			},
			newDecoded: func() proto.Message { return &govv1.MsgSubmitProposal{} },
			assertDecoded: func(t *testing.T, decoded proto.Message) {
				proposal, ok := decoded.(*govv1.MsgSubmitProposal)
				require.True(t, ok)
				require.Len(t, proposal.Messages, 1)
				assertCachedLegacyMint(t, proposal.Messages[0])
			},
		},
		{
			name: "group proposal submission",
			newOriginal: func(t *testing.T) proto.Message {
				return &group.MsgSubmitProposal{
					GroupPolicyAddress: "panacea1grouppolicy",
					Proposers:          []string{"panacea1proposer"},
					Messages:           []*cdctypes.Any{newRawLegacyMintAny(t)},
					Title:              "legacy PNFT proposal",
					Summary:            "execute a legacy PNFT message",
				}
			},
			newDecoded: func() proto.Message { return &group.MsgSubmitProposal{} },
			assertDecoded: func(t *testing.T, decoded proto.Message) {
				proposal, ok := decoded.(*group.MsgSubmitProposal)
				require.True(t, ok)
				require.Len(t, proposal.Messages, 1)
				assertCachedLegacyMint(t, proposal.Messages[0])
			},
		},
		{
			name: "authz exec",
			newOriginal: func(t *testing.T) proto.Message {
				return newLegacyAuthzExec(t)
			},
			newDecoded: func() proto.Message { return &authz.MsgExec{} },
			assertDecoded: func(t *testing.T, decoded proto.Message) {
				exec, ok := decoded.(*authz.MsgExec)
				require.True(t, ok)
				assertAuthzCachesLegacyMint(t, exec)
			},
		},
		{
			name: "stored gov proposal with authz exec",
			newOriginal: func(t *testing.T) proto.Message {
				return &govv1.Proposal{
					Id:       1,
					Messages: []*cdctypes.Any{newLegacyAuthzExecAny(t)},
					Status:   govv1.StatusVotingPeriod,
					Title:    "nested legacy PNFT proposal",
					Summary:  "execute authz containing a legacy PNFT message",
				}
			},
			newDecoded: func() proto.Message { return &govv1.Proposal{} },
			assertDecoded: func(t *testing.T, decoded proto.Message) {
				proposal, ok := decoded.(*govv1.Proposal)
				require.True(t, ok)
				require.Len(t, proposal.Messages, 1)
				assertCachedAuthzWithLegacyMint(t, proposal.Messages[0])
			},
		},
		{
			name: "stored group proposal with authz exec",
			newOriginal: func(t *testing.T) proto.Message {
				return &group.Proposal{
					Id:                 1,
					GroupPolicyAddress: "panacea1grouppolicy",
					Proposers:          []string{"panacea1proposer"},
					Status:             group.PROPOSAL_STATUS_SUBMITTED,
					Messages:           []*cdctypes.Any{newLegacyAuthzExecAny(t)},
					Title:              "nested legacy PNFT proposal",
					Summary:            "execute authz containing a legacy PNFT message",
				}
			},
			newDecoded: func() proto.Message { return &group.Proposal{} },
			assertDecoded: func(t *testing.T, decoded proto.Message) {
				proposal, ok := decoded.(*group.Proposal)
				require.True(t, ok)
				require.Len(t, proposal.Messages, 1)
				assertCachedAuthzWithLegacyMint(t, proposal.Messages[0])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			original := tc.newOriginal(t)

			t.Run("binary", func(t *testing.T) {
				encoded, err := cdc.Marshal(original)
				require.NoError(t, err)

				decoded := tc.newDecoded()
				require.NoError(t, cdc.Unmarshal(encoded, decoded))
				tc.assertDecoded(t, decoded)
			})

			t.Run("json", func(t *testing.T) {
				encoded, err := cdc.MarshalJSON(original)
				require.NoError(t, err)

				decoded := tc.newDecoded()
				require.NoError(t, cdc.UnmarshalJSON(encoded, decoded))
				tc.assertDecoded(t, decoded)
			})
		})
	}
}

func newLegacyContainerTestCodec() *codec.ProtoCodec {
	registry := cdctypes.NewInterfaceRegistry()
	sdk.RegisterInterfaces(registry)
	authz.RegisterInterfaces(registry)
	govv1.RegisterInterfaces(registry)
	group.RegisterInterfaces(registry)
	RegisterInterfaces(registry)

	return codec.NewProtoCodec(registry)
}

func newRawLegacyMintAny(t *testing.T) *cdctypes.Any {
	t.Helper()

	value, err := proto.Marshal(&MsgMintPNFTRequest{
		DenomId: "denom",
		Id:      "pnft",
		Creator: "panacea1creator",
	})
	require.NoError(t, err)

	return &cdctypes.Any{
		TypeUrl: "/panacea.pnft.v2.MsgMintPNFTRequest",
		Value:   value,
	}
}

func newLegacyAuthzExec(t *testing.T) *authz.MsgExec {
	t.Helper()

	return &authz.MsgExec{
		Grantee: "panacea1grantee",
		Msgs:    []*cdctypes.Any{newRawLegacyMintAny(t)},
	}
}

func newLegacyAuthzExecAny(t *testing.T) *cdctypes.Any {
	t.Helper()

	packed, err := cdctypes.NewAnyWithValue(newLegacyAuthzExec(t))
	require.NoError(t, err)

	return packed
}

func assertCachedLegacyMint(t *testing.T, packed *cdctypes.Any) {
	t.Helper()

	msg, ok := packed.GetCachedValue().(*MsgMintPNFTRequest)
	require.True(t, ok)
	require.Equal(t, "denom", msg.DenomId)
	require.Equal(t, "pnft", msg.Id)
	require.Equal(t, "panacea1creator", msg.Creator)
}

func assertCachedAuthzWithLegacyMint(t *testing.T, packed *cdctypes.Any) {
	t.Helper()

	exec, ok := packed.GetCachedValue().(*authz.MsgExec)
	require.True(t, ok)
	assertAuthzCachesLegacyMint(t, exec)
}

func assertAuthzCachesLegacyMint(t *testing.T, exec *authz.MsgExec) {
	t.Helper()

	require.Len(t, exec.Msgs, 1)
	assertCachedLegacyMint(t, exec.Msgs[0])
}
