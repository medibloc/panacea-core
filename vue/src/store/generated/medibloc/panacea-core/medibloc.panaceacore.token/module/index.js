// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.
import { SigningStargateClient } from "@cosmjs/stargate";
import { Registry } from "@cosmjs/proto-signing";
import { Api } from "./rest";
import { MsgDeleteToken } from "./types/token/tx";
import { MsgCreateToken } from "./types/token/tx";
import { MsgUpdateToken } from "./types/token/tx";
const types = [
    ["/medibloc.panaceacore.token.MsgDeleteToken", MsgDeleteToken],
    ["/medibloc.panaceacore.token.MsgCreateToken", MsgCreateToken],
    ["/medibloc.panaceacore.token.MsgUpdateToken", MsgUpdateToken],
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
        msgDeleteToken: (data) => ({ typeUrl: "/medibloc.panaceacore.token.MsgDeleteToken", value: data }),
        msgCreateToken: (data) => ({ typeUrl: "/medibloc.panaceacore.token.MsgCreateToken", value: data }),
        msgUpdateToken: (data) => ({ typeUrl: "/medibloc.panaceacore.token.MsgUpdateToken", value: data }),
    };
};
const queryClient = async ({ addr: addr } = { addr: "http://localhost:1317" }) => {
    return new Api({ baseUrl: addr });
};
export { txClient, queryClient, };
