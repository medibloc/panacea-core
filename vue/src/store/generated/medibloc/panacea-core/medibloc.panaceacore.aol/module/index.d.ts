import { StdFee } from "@cosmjs/launchpad";
import { OfflineSigner, EncodeObject } from "@cosmjs/proto-signing";
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
interface TxClientOptions {
    addr: string;
}
interface SignAndBroadcastOptions {
    fee: StdFee;
    memo?: string;
}
declare const txClient: (wallet: OfflineSigner, { addr: addr }?: TxClientOptions) => Promise<{
    signAndBroadcast: (msgs: EncodeObject[], { fee, memo }: SignAndBroadcastOptions) => Promise<import("@cosmjs/stargate").BroadcastTxResponse>;
    msgDeleteWriter: (data: MsgDeleteWriter) => EncodeObject;
    msgDeleteTopic: (data: MsgDeleteTopic) => EncodeObject;
    msgDeleteOwner: (data: MsgDeleteOwner) => EncodeObject;
    msgCreateRecord: (data: MsgCreateRecord) => EncodeObject;
    msgUpdateRecord: (data: MsgUpdateRecord) => EncodeObject;
    msgCreateWriter: (data: MsgCreateWriter) => EncodeObject;
    msgCreateTopic: (data: MsgCreateTopic) => EncodeObject;
    msgDeleteRecord: (data: MsgDeleteRecord) => EncodeObject;
    msgUpdateTopic: (data: MsgUpdateTopic) => EncodeObject;
    msgCreateOwner: (data: MsgCreateOwner) => EncodeObject;
    msgUpdateOwner: (data: MsgUpdateOwner) => EncodeObject;
    msgUpdateWriter: (data: MsgUpdateWriter) => EncodeObject;
}>;
interface QueryClientOptions {
    addr: string;
}
declare const queryClient: ({ addr: addr }?: QueryClientOptions) => Promise<Api<unknown>>;
export { txClient, queryClient, };
