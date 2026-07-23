package nft

import autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

// AutoCLIOptions combines generated commands with the custom NFT roots.
func (AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: standardQueryCommandDescriptor(),
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

// standardQueryCommandDescriptor adds the standard NFT Query service to the
// custom query root, which already contains Panacea record queries.
func standardQueryCommandDescriptor() *autocliv1.ServiceCommandDescriptor {
	return &autocliv1.ServiceCommandDescriptor{
		Service:              "cosmos.nft.v1beta1.Query",
		EnhanceCustomCommand: true,
		RpcCommandOptions: []*autocliv1.RpcCommandOptions{
			{
				RpcMethod: "Balance",
				Use:       "balance [owner] [class-id]",
				Short:     "Query an owner's NFT balance in a class",
				PositionalArgs: []*autocliv1.PositionalArgDescriptor{
					{ProtoField: "owner"},
					{ProtoField: "class_id"},
				},
			},
			{
				RpcMethod: "Owner",
				Use:       "owner [class-id] [nft-id]",
				Short:     "Query the owner of an NFT",
				PositionalArgs: []*autocliv1.PositionalArgDescriptor{
					{ProtoField: "class_id"},
					{ProtoField: "id"},
				},
			},
			{
				RpcMethod: "Supply",
				Use:       "supply [class-id]",
				Short:     "Query the live NFT supply of a class",
				PositionalArgs: []*autocliv1.PositionalArgDescriptor{
					{ProtoField: "class_id"},
				},
			},
			{
				RpcMethod: "NFTs",
				Use:       "nfts [class-id]",
				Short:     "Query live NFTs by class, owner, or both",
				PositionalArgs: []*autocliv1.PositionalArgDescriptor{
					{ProtoField: "class_id", Optional: true},
				},
			},
			{
				RpcMethod: "NFT",
				Use:       "nft [class-id] [nft-id]",
				Short:     "Query a live NFT",
				PositionalArgs: []*autocliv1.PositionalArgDescriptor{
					{ProtoField: "class_id"},
					{ProtoField: "id"},
				},
			},
			{
				RpcMethod: "Class",
				Use:       "class [class-id]",
				Short:     "Query a standard NFT class",
				PositionalArgs: []*autocliv1.PositionalArgDescriptor{
					{ProtoField: "class_id"},
				},
			},
			{
				RpcMethod: "Classes",
				Use:       "classes",
				Short:     "Query all standard NFT classes",
			},
		},
	}
}
