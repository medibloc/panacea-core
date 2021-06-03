import { Writer, Reader } from "protobufjs/minimal";
export declare const protobufPackage = "medibloc.panaceacore.aol";
export interface Owner {
    creator: string;
    id: number;
    totalTopics: number;
}
export declare const Owner: {
    encode(message: Owner, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): Owner;
    fromJSON(object: any): Owner;
    toJSON(message: Owner): unknown;
    fromPartial(object: DeepPartial<Owner>): Owner;
};
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
