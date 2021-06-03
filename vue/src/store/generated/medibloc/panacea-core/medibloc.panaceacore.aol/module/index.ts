// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.

import { StdFee } from "@cosmjs/launchpad";
import { SigningStargateClient } from "@cosmjs/stargate";
import { Registry, OfflineSigner, EncodeObject, DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import { Api } from "./rest";
import { MsgDeleteWriter } from "./types/aol/tx";
import { MsgDeleteTopic } from "./types/aol/tx";
import { MsgDeleteOwner } from "./types/aol/tx";
import { MsgCreateRecord } from "./types/aol/tx";
import { MsgUpdateRecord } from "./types/aol/tx";
import { MsgCreateWriter } from "./types/aol/tx";
import { MsgCreateTopic } from "./types/aol/tx";
import { MsgDeleteRecord } from "./types/aol/tx";
import { MsgUpdateTopic } from "./types/aol/tx";
import { MsgCreateOwner } from "./types/aol/tx";
import { MsgUpdateOwner } from "./types/aol/tx";
import { MsgUpdateWriter } from "./types/aol/tx";


const types = [
  ["/medibloc.panaceacore.aol.MsgDeleteWriter", MsgDeleteWriter],
  ["/medibloc.panaceacore.aol.MsgDeleteTopic", MsgDeleteTopic],
  ["/medibloc.panaceacore.aol.MsgDeleteOwner", MsgDeleteOwner],
  ["/medibloc.panaceacore.aol.MsgCreateRecord", MsgCreateRecord],
  ["/medibloc.panaceacore.aol.MsgUpdateRecord", MsgUpdateRecord],
  ["/medibloc.panaceacore.aol.MsgCreateWriter", MsgCreateWriter],
  ["/medibloc.panaceacore.aol.MsgCreateTopic", MsgCreateTopic],
  ["/medibloc.panaceacore.aol.MsgDeleteRecord", MsgDeleteRecord],
  ["/medibloc.panaceacore.aol.MsgUpdateTopic", MsgUpdateTopic],
  ["/medibloc.panaceacore.aol.MsgCreateOwner", MsgCreateOwner],
  ["/medibloc.panaceacore.aol.MsgUpdateOwner", MsgUpdateOwner],
  ["/medibloc.panaceacore.aol.MsgUpdateWriter", MsgUpdateWriter],
  
];

const registry = new Registry(<any>types);

const defaultFee = {
  amount: [],
  gas: "200000",
};

interface TxClientOptions {
  addr: string
}

interface SignAndBroadcastOptions {
  fee: StdFee,
  memo?: string
}

const txClient = async (wallet: OfflineSigner, { addr: addr }: TxClientOptions = { addr: "http://localhost:26657" }) => {
  if (!wallet) throw new Error("wallet is required");

  const client = await SigningStargateClient.connectWithSigner(addr, wallet, { registry });
  const { address } = (await wallet.getAccounts())[0];

  return {
    signAndBroadcast: (msgs: EncodeObject[], { fee=defaultFee, memo=null }: SignAndBroadcastOptions) => memo?client.signAndBroadcast(address, msgs, fee,memo):client.signAndBroadcast(address, msgs, fee),
    msgDeleteWriter: (data: MsgDeleteWriter): EncodeObject => ({ typeUrl: "/medibloc.panaceacore.aol.MsgDeleteWriter", value: data }),
    msgDeleteTopic: (data: MsgDeleteTopic): EncodeObject => ({ typeUrl: "/medibloc.panaceacore.aol.MsgDeleteTopic", value: data }),
    msgDeleteOwner: (data: MsgDeleteOwner): EncodeObject => ({ typeUrl: "/medibloc.panaceacore.aol.MsgDeleteOwner", value: data }),
    msgCreateRecord: (data: MsgCreateRecord): EncodeObject => ({ typeUrl: "/medibloc.panaceacore.aol.MsgCreateRecord", value: data }),
    msgUpdateRecord: (data: MsgUpdateRecord): EncodeObject => ({ typeUrl: "/medibloc.panaceacore.aol.MsgUpdateRecord", value: data }),
    msgCreateWriter: (data: MsgCreateWriter): EncodeObject => ({ typeUrl: "/medibloc.panaceacore.aol.MsgCreateWriter", value: data }),
    msgCreateTopic: (data: MsgCreateTopic): EncodeObject => ({ typeUrl: "/medibloc.panaceacore.aol.MsgCreateTopic", value: data }),
    msgDeleteRecord: (data: MsgDeleteRecord): EncodeObject => ({ typeUrl: "/medibloc.panaceacore.aol.MsgDeleteRecord", value: data }),
    msgUpdateTopic: (data: MsgUpdateTopic): EncodeObject => ({ typeUrl: "/medibloc.panaceacore.aol.MsgUpdateTopic", value: data }),
    msgCreateOwner: (data: MsgCreateOwner): EncodeObject => ({ typeUrl: "/medibloc.panaceacore.aol.MsgCreateOwner", value: data }),
    msgUpdateOwner: (data: MsgUpdateOwner): EncodeObject => ({ typeUrl: "/medibloc.panaceacore.aol.MsgUpdateOwner", value: data }),
    msgUpdateWriter: (data: MsgUpdateWriter): EncodeObject => ({ typeUrl: "/medibloc.panaceacore.aol.MsgUpdateWriter", value: data }),
    
  };
};

interface QueryClientOptions {
  addr: string
}

const queryClient = async ({ addr: addr }: QueryClientOptions = { addr: "http://localhost:1317" }) => {
  return new Api({ baseUrl: addr });
};

export {
  txClient,
  queryClient,
};
