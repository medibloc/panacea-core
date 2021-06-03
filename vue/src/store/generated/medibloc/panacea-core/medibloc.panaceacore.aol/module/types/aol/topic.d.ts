import { Writer, Reader } from "protobufjs/minimal";
export declare const protobufPackage = "medibloc.panaceacore.aol";
export interface Topic {
    creator: string;
    id: number;
    description: string;
    totalRecords: number;
    totalWriter: number;
}
export declare const Topic: {
    encode(message: Topic, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): Topic;
    fromJSON(object: any): Topic;
    toJSON(message: Topic): unknown;
    fromPartial(object: DeepPartial<Topic>): Topic;
};
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
