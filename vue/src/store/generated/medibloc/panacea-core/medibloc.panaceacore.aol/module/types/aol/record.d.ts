import { Writer, Reader } from "protobufjs/minimal";
export declare const protobufPackage = "medibloc.panaceacore.aol";
export interface Record {
    creator: string;
    id: number;
    key: string;
    value: string;
    nanoTimestamp: number;
    writerAddress: string;
}
export declare const Record: {
    encode(message: Record, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): Record;
    fromJSON(object: any): Record;
    toJSON(message: Record): unknown;
    fromPartial(object: DeepPartial<Record>): Record;
};
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
