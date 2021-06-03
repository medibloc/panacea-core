import { Owner } from "../aol/owner";
import { Record } from "../aol/record";
import { Writer } from "../aol/writer";
import { Topic } from "../aol/topic";
import { Writer as Writer1, Reader } from "protobufjs/minimal";
export declare const protobufPackage = "medibloc.panaceacore.aol";
/** GenesisState defines the aol module's genesis state. */
export interface GenesisState {
    /** this line is used by starport scaffolding # genesis/proto/state */
    ownerList: Owner[];
    /** this line is used by starport scaffolding # genesis/proto/stateField */
    recordList: Record[];
    /** this line is used by starport scaffolding # genesis/proto/stateField */
    writerList: Writer[];
    /** this line is used by starport scaffolding # genesis/proto/stateField */
    topicList: Topic[];
}
export declare const GenesisState: {
    encode(message: GenesisState, writer?: Writer1): Writer1;
    decode(input: Reader | Uint8Array, length?: number): GenesisState;
    fromJSON(object: any): GenesisState;
    toJSON(message: GenesisState): unknown;
    fromPartial(object: DeepPartial<GenesisState>): GenesisState;
};
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
