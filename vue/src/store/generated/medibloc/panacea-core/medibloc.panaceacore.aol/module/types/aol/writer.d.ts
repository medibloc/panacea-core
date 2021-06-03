import { Writer as Writer1, Reader } from "protobufjs/minimal";
export declare const protobufPackage = "medibloc.panaceacore.aol";
export interface Writer {
    creator: string;
    id: number;
    moniker: string;
    description: string;
    nanoTimestamp: number;
}
export declare const Writer: {
    encode(message: Writer, writer?: Writer1): Writer1;
    decode(input: Reader | Uint8Array, length?: number): Writer;
    fromJSON(object: any): Writer;
    toJSON(message: Writer): unknown;
    fromPartial(object: DeepPartial<Writer>): Writer;
};
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
