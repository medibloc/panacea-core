import { Reader, Writer } from "protobufjs/minimal";
export declare const protobufPackage = "medibloc.panaceacore.token";
/** this line is used by starport scaffolding # proto/tx/message */
export interface MsgCreateToken {
    creator: string;
    name: string;
    symbol: string;
    totalSupply: number;
    mintable: boolean;
    ownerAddress: string;
}
export interface MsgCreateTokenResponse {
    id: number;
}
export interface MsgUpdateToken {
    creator: string;
    id: number;
    name: string;
    symbol: string;
    totalSupply: number;
    mintable: boolean;
    ownerAddress: string;
}
export interface MsgUpdateTokenResponse {
}
export interface MsgDeleteToken {
    creator: string;
    id: number;
}
export interface MsgDeleteTokenResponse {
}
export declare const MsgCreateToken: {
    encode(message: MsgCreateToken, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgCreateToken;
    fromJSON(object: any): MsgCreateToken;
    toJSON(message: MsgCreateToken): unknown;
    fromPartial(object: DeepPartial<MsgCreateToken>): MsgCreateToken;
};
export declare const MsgCreateTokenResponse: {
    encode(message: MsgCreateTokenResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgCreateTokenResponse;
    fromJSON(object: any): MsgCreateTokenResponse;
    toJSON(message: MsgCreateTokenResponse): unknown;
    fromPartial(object: DeepPartial<MsgCreateTokenResponse>): MsgCreateTokenResponse;
};
export declare const MsgUpdateToken: {
    encode(message: MsgUpdateToken, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgUpdateToken;
    fromJSON(object: any): MsgUpdateToken;
    toJSON(message: MsgUpdateToken): unknown;
    fromPartial(object: DeepPartial<MsgUpdateToken>): MsgUpdateToken;
};
export declare const MsgUpdateTokenResponse: {
    encode(_: MsgUpdateTokenResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgUpdateTokenResponse;
    fromJSON(_: any): MsgUpdateTokenResponse;
    toJSON(_: MsgUpdateTokenResponse): unknown;
    fromPartial(_: DeepPartial<MsgUpdateTokenResponse>): MsgUpdateTokenResponse;
};
export declare const MsgDeleteToken: {
    encode(message: MsgDeleteToken, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgDeleteToken;
    fromJSON(object: any): MsgDeleteToken;
    toJSON(message: MsgDeleteToken): unknown;
    fromPartial(object: DeepPartial<MsgDeleteToken>): MsgDeleteToken;
};
export declare const MsgDeleteTokenResponse: {
    encode(_: MsgDeleteTokenResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgDeleteTokenResponse;
    fromJSON(_: any): MsgDeleteTokenResponse;
    toJSON(_: MsgDeleteTokenResponse): unknown;
    fromPartial(_: DeepPartial<MsgDeleteTokenResponse>): MsgDeleteTokenResponse;
};
/** Msg defines the Msg service. */
export interface Msg {
    /** this line is used by starport scaffolding # proto/tx/rpc */
    CreateToken(request: MsgCreateToken): Promise<MsgCreateTokenResponse>;
    UpdateToken(request: MsgUpdateToken): Promise<MsgUpdateTokenResponse>;
    DeleteToken(request: MsgDeleteToken): Promise<MsgDeleteTokenResponse>;
}
export declare class MsgClientImpl implements Msg {
    private readonly rpc;
    constructor(rpc: Rpc);
    CreateToken(request: MsgCreateToken): Promise<MsgCreateTokenResponse>;
    UpdateToken(request: MsgUpdateToken): Promise<MsgUpdateTokenResponse>;
    DeleteToken(request: MsgDeleteToken): Promise<MsgDeleteTokenResponse>;
}
interface Rpc {
    request(service: string, method: string, data: Uint8Array): Promise<Uint8Array>;
}
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
