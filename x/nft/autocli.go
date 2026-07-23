package nft

import autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

// AutoCLIOptions adds the Panacea Msg service commands to the custom NFT
// transaction root, which already contains the policy-aware standard Send.
func (AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              "panacea.nft.v1.Msg",
			EnhanceCustomCommand: true,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "CreateClass",
					Use:       "create-class [local-class-id] [name] [symbol] [transfer-policy] [revocable] [max-supply]",
					Short:     "Create an NFT class and its immutable policy",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "local_class_id"},
						{ProtoField: "name"},
						{ProtoField: "symbol"},
						{ProtoField: "transfer_policy"},
						{ProtoField: "revocable"},
						{ProtoField: "max_supply"},
					},
				},
				{
					RpcMethod: "UpdateController",
					Use:       "update-controller [class-id] [new-controller]",
					Short:     "Transfer NFT class control to another account",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "new_controller"},
					},
				},
				{
					RpcMethod: "Mint",
					Use:       "mint [class-id] [nft-id] [recipient]",
					Short:     "Mint an NFT to a recipient",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "nft_id"},
						{ProtoField: "recipient"},
					},
				},
				{
					RpcMethod: "Revoke",
					Use:       "revoke [class-id] [nft-id]",
					Short:     "Irreversibly revoke an NFT",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "nft_id"},
					},
				},
				{
					RpcMethod: "Burn",
					Use:       "burn [class-id] [nft-id]",
					Short:     "Permanently burn an owned NFT",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "nft_id"},
					},
				},
			},
		},
	}
}
