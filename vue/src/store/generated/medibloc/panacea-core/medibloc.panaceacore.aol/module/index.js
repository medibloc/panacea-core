// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.
import { SigningStargateClient } from "@cosmjs/stargate";
import { Registry } from "@cosmjs/proto-signing";
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
const registry = new Registry(types);
const defaultFee = {
    amount: [],
    gas: "200000",
};
const txClient = async (wallet, { addr: addr } = { addr: "http://localhost:26657" }) => {
    if (!wallet)
        throw new Error("wallet is required");
    const client = await SigningStargateClient.connectWithSigner(addr, wallet, { registry });
    const { address } = (await wallet.getAccounts())[0];
    return {
        signAndBroadcast: (msgs, { fee = defaultFee, memo = null }) => memo ? client.signAndBroadcast(address, msgs, fee, memo) : client.signAndBroadcast(address, msgs, fee),
        msgDeleteWriter: (data) => ({ typeUrl: "/medibloc.panaceacore.aol.MsgDeleteWriter", value: data }),
        msgDeleteTopic: (data) => ({ typeUrl: "/medibloc.panaceacore.aol.MsgDeleteTopic", value: data }),
        msgDeleteOwner: (data) => ({ typeUrl: "/medibloc.panaceacore.aol.MsgDeleteOwner", value: data }),
        msgCreateRecord: (data) => ({ typeUrl: "/medibloc.panaceacore.aol.MsgCreateRecord", value: data }),
        msgUpdateRecord: (data) => ({ typeUrl: "/medibloc.panaceacore.aol.MsgUpdateRecord", value: data }),
        msgCreateWriter: (data) => ({ typeUrl: "/medibloc.panaceacore.aol.MsgCreateWriter", value: data }),
        msgCreateTopic: (data) => ({ typeUrl: "/medibloc.panaceacore.aol.MsgCreateTopic", value: data }),
        msgDeleteRecord: (data) => ({ typeUrl: "/medibloc.panaceacore.aol.MsgDeleteRecord", value: data }),
        msgUpdateTopic: (data) => ({ typeUrl: "/medibloc.panaceacore.aol.MsgUpdateTopic", value: data }),
        msgCreateOwner: (data) => ({ typeUrl: "/medibloc.panaceacore.aol.MsgCreateOwner", value: data }),
        msgUpdateOwner: (data) => ({ typeUrl: "/medibloc.panaceacore.aol.MsgUpdateOwner", value: data }),
        msgUpdateWriter: (data) => ({ typeUrl: "/medibloc.panaceacore.aol.MsgUpdateWriter", value: data }),
    };
};
const queryClient = async ({ addr: addr } = { addr: "http://localhost:1317" }) => {
    return new Api({ baseUrl: addr });
};
export { txClient, queryClient, };
