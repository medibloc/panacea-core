import { Reader, Writer as Writer1 } from "protobufjs/minimal";
import { Owner } from "../aol/owner";
import { PageRequest, PageResponse } from "../cosmos/base/query/v1beta1/pagination";
import { Record } from "../aol/record";
import { Writer } from "../aol/writer";
import { Topic } from "../aol/topic";
export declare const protobufPackage = "medibloc.panaceacore.aol";
/** this line is used by starport scaffolding # 3 */
export interface QueryGetOwnerRequest {
    id: number;
}
export interface QueryGetOwnerResponse {
    Owner: Owner | undefined;
}
export interface QueryAllOwnerRequest {
    pagination: PageRequest | undefined;
}
export interface QueryAllOwnerResponse {
    Owner: Owner[];
    pagination: PageResponse | undefined;
}
export interface QueryGetRecordRequest {
    id: number;
}
export interface QueryGetRecordResponse {
    Record: Record | undefined;
}
export interface QueryAllRecordRequest {
    pagination: PageRequest | undefined;
}
export interface QueryAllRecordResponse {
    Record: Record[];
    pagination: PageResponse | undefined;
}
export interface QueryGetWriterRequest {
    id: number;
}
export interface QueryGetWriterResponse {
    Writer: Writer | undefined;
}
export interface QueryAllWriterRequest {
    pagination: PageRequest | undefined;
}
export interface QueryAllWriterResponse {
    Writer: Writer[];
    pagination: PageResponse | undefined;
}
export interface QueryGetTopicRequest {
    id: number;
}
export interface QueryGetTopicResponse {
    Topic: Topic | undefined;
}
export interface QueryAllTopicRequest {
    pagination: PageRequest | undefined;
}
export interface QueryAllTopicResponse {
    Topic: Topic[];
    pagination: PageResponse | undefined;
}
export declare const QueryGetOwnerRequest: {
    encode(message: QueryGetOwnerRequest, writer?: Writer1): Writer1;
    decode(input: Reader | Uint8Array, length?: number): QueryGetOwnerRequest;
    fromJSON(object: any): QueryGetOwnerRequest;
    toJSON(message: QueryGetOwnerRequest): unknown;
    fromPartial(object: DeepPartial<QueryGetOwnerRequest>): QueryGetOwnerRequest;
};
export declare const QueryGetOwnerResponse: {
    encode(message: QueryGetOwnerResponse, writer?: Writer1): Writer1;
    decode(input: Reader | Uint8Array, length?: number): QueryGetOwnerResponse;
    fromJSON(object: any): QueryGetOwnerResponse;
    toJSON(message: QueryGetOwnerResponse): unknown;
    fromPartial(object: DeepPartial<QueryGetOwnerResponse>): QueryGetOwnerResponse;
};
export declare const QueryAllOwnerRequest: {
    encode(message: QueryAllOwnerRequest, writer?: Writer1): Writer1;
    decode(input: Reader | Uint8Array, length?: number): QueryAllOwnerRequest;
    fromJSON(object: any): QueryAllOwnerRequest;
    toJSON(message: QueryAllOwnerRequest): unknown;
    fromPartial(object: DeepPartial<QueryAllOwnerRequest>): QueryAllOwnerRequest;
};
export declare const QueryAllOwnerResponse: {
    encode(message: QueryAllOwnerResponse, writer?: Writer1): Writer1;
    decode(input: Reader | Uint8Array, length?: number): QueryAllOwnerResponse;
    fromJSON(object: any): QueryAllOwnerResponse;
    toJSON(message: QueryAllOwnerResponse): unknown;
    fromPartial(object: DeepPartial<QueryAllOwnerResponse>): QueryAllOwnerResponse;
};
export declare const QueryGetRecordRequest: {
    encode(message: QueryGetRecordRequest, writer?: Writer1): Writer1;
    decode(input: Reader | Uint8Array, length?: number): QueryGetRecordRequest;
    fromJSON(object: any): QueryGetRecordRequest;
    toJSON(message: QueryGetRecordRequest): unknown;
    fromPartial(object: DeepPartial<QueryGetRecordRequest>): QueryGetRecordRequest;
};
export declare const QueryGetRecordResponse: {
    encode(message: QueryGetRecordResponse, writer?: Writer1): Writer1;
    decode(input: Reader | Uint8Array, length?: number): QueryGetRecordResponse;
    fromJSON(object: any): QueryGetRecordResponse;
    toJSON(message: QueryGetRecordResponse): unknown;
    fromPartial(object: DeepPartial<QueryGetRecordResponse>): QueryGetRecordResponse;
};
export declare const QueryAllRecordRequest: {
    encode(message: QueryAllRecordRequest, writer?: Writer1): Writer1;
    decode(input: Reader | Uint8Array, length?: number): QueryAllRecordRequest;
    fromJSON(object: any): QueryAllRecordRequest;
    toJSON(message: QueryAllRecordRequest): unknown;
    fromPartial(object: DeepPartial<QueryAllRecordRequest>): QueryAllRecordRequest;
};
export declare const QueryAllRecordResponse: {
    encode(message: QueryAllRecordResponse, writer?: Writer1): Writer1;
    decode(input: Reader | Uint8Array, length?: number): QueryAllRecordResponse;
    fromJSON(object: any): QueryAllRecordResponse;
    toJSON(message: QueryAllRecordResponse): unknown;
    fromPartial(object: DeepPartial<QueryAllRecordResponse>): QueryAllRecordResponse;
};
export declare const QueryGetWriterRequest: {
    encode(message: QueryGetWriterRequest, writer?: Writer1): Writer1;
    decode(input: Reader | Uint8Array, length?: number): QueryGetWriterRequest;
    fromJSON(object: any): QueryGetWriterRequest;
    toJSON(message: QueryGetWriterRequest): unknown;
    fromPartial(object: DeepPartial<QueryGetWriterRequest>): QueryGetWriterRequest;
};
export declare const QueryGetWriterResponse: {
    encode(message: QueryGetWriterResponse, writer?: Writer1): Writer1;
    decode(input: Reader | Uint8Array, length?: number): QueryGetWriterResponse;
    fromJSON(object: any): QueryGetWriterResponse;
    toJSON(message: QueryGetWriterResponse): unknown;
    fromPartial(object: DeepPartial<QueryGetWriterResponse>): QueryGetWriterResponse;
};
export declare const QueryAllWriterRequest: {
    encode(message: QueryAllWriterRequest, writer?: Writer1): Writer1;
    decode(input: Reader | Uint8Array, length?: number): QueryAllWriterRequest;
    fromJSON(object: any): QueryAllWriterRequest;
    toJSON(message: QueryAllWriterRequest): unknown;
    fromPartial(object: DeepPartial<QueryAllWriterRequest>): QueryAllWriterRequest;
};
export declare const QueryAllWriterResponse: {
    encode(message: QueryAllWriterResponse, writer?: Writer1): Writer1;
    decode(input: Reader | Uint8Array, length?: number): QueryAllWriterResponse;
    fromJSON(object: any): QueryAllWriterResponse;
    toJSON(message: QueryAllWriterResponse): unknown;
    fromPartial(object: DeepPartial<QueryAllWriterResponse>): QueryAllWriterResponse;
};
export declare const QueryGetTopicRequest: {
    encode(message: QueryGetTopicRequest, writer?: Writer1): Writer1;
    decode(input: Reader | Uint8Array, length?: number): QueryGetTopicRequest;
    fromJSON(object: any): QueryGetTopicRequest;
    toJSON(message: QueryGetTopicRequest): unknown;
    fromPartial(object: DeepPartial<QueryGetTopicRequest>): QueryGetTopicRequest;
};
export declare const QueryGetTopicResponse: {
    encode(message: QueryGetTopicResponse, writer?: Writer1): Writer1;
    decode(input: Reader | Uint8Array, length?: number): QueryGetTopicResponse;
    fromJSON(object: any): QueryGetTopicResponse;
    toJSON(message: QueryGetTopicResponse): unknown;
    fromPartial(object: DeepPartial<QueryGetTopicResponse>): QueryGetTopicResponse;
};
export declare const QueryAllTopicRequest: {
    encode(message: QueryAllTopicRequest, writer?: Writer1): Writer1;
    decode(input: Reader | Uint8Array, length?: number): QueryAllTopicRequest;
    fromJSON(object: any): QueryAllTopicRequest;
    toJSON(message: QueryAllTopicRequest): unknown;
    fromPartial(object: DeepPartial<QueryAllTopicRequest>): QueryAllTopicRequest;
};
export declare const QueryAllTopicResponse: {
    encode(message: QueryAllTopicResponse, writer?: Writer1): Writer1;
    decode(input: Reader | Uint8Array, length?: number): QueryAllTopicResponse;
    fromJSON(object: any): QueryAllTopicResponse;
    toJSON(message: QueryAllTopicResponse): unknown;
    fromPartial(object: DeepPartial<QueryAllTopicResponse>): QueryAllTopicResponse;
};
/** Query defines the gRPC querier service. */
export interface Query {
    /** this line is used by starport scaffolding # 2 */
    Owner(request: QueryGetOwnerRequest): Promise<QueryGetOwnerResponse>;
    OwnerAll(request: QueryAllOwnerRequest): Promise<QueryAllOwnerResponse>;
    Record(request: QueryGetRecordRequest): Promise<QueryGetRecordResponse>;
    RecordAll(request: QueryAllRecordRequest): Promise<QueryAllRecordResponse>;
    Writer(request: QueryGetWriterRequest): Promise<QueryGetWriterResponse>;
    WriterAll(request: QueryAllWriterRequest): Promise<QueryAllWriterResponse>;
    Topic(request: QueryGetTopicRequest): Promise<QueryGetTopicResponse>;
    TopicAll(request: QueryAllTopicRequest): Promise<QueryAllTopicResponse>;
}
export declare class QueryClientImpl implements Query {
    private readonly rpc;
    constructor(rpc: Rpc);
    Owner(request: QueryGetOwnerRequest): Promise<QueryGetOwnerResponse>;
    OwnerAll(request: QueryAllOwnerRequest): Promise<QueryAllOwnerResponse>;
    Record(request: QueryGetRecordRequest): Promise<QueryGetRecordResponse>;
    RecordAll(request: QueryAllRecordRequest): Promise<QueryAllRecordResponse>;
    Writer(request: QueryGetWriterRequest): Promise<QueryGetWriterResponse>;
    WriterAll(request: QueryAllWriterRequest): Promise<QueryAllWriterResponse>;
    Topic(request: QueryGetTopicRequest): Promise<QueryGetTopicResponse>;
    TopicAll(request: QueryAllTopicRequest): Promise<QueryAllTopicResponse>;
}
interface Rpc {
    request(service: string, method: string, data: Uint8Array): Promise<Uint8Array>;
}
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
