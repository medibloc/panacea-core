import { Reader, Writer } from "protobufjs/minimal";
export declare const protobufPackage = "medibloc.panaceacore.aol";
/** this line is used by starport scaffolding # proto/tx/message */
export interface MsgCreateOwner {
    creator: string;
    totalTopics: number;
}
export interface MsgCreateOwnerResponse {
    id: number;
}
export interface MsgUpdateOwner {
    creator: string;
    id: number;
    totalTopics: number;
}
export interface MsgUpdateOwnerResponse {
}
export interface MsgDeleteOwner {
    creator: string;
    id: number;
}
export interface MsgDeleteOwnerResponse {
}
export interface MsgCreateRecord {
    creator: string;
    key: string;
    value: string;
    nanoTimestamp: number;
    writerAddress: string;
}
export interface MsgCreateRecordResponse {
    id: number;
}
export interface MsgUpdateRecord {
    creator: string;
    id: number;
    key: string;
    value: string;
    nanoTimestamp: number;
    writerAddress: string;
}
export interface MsgUpdateRecordResponse {
}
export interface MsgDeleteRecord {
    creator: string;
    id: number;
}
export interface MsgDeleteRecordResponse {
}
export interface MsgCreateWriter {
    creator: string;
    moniker: string;
    description: string;
    nanoTimestamp: number;
}
export interface MsgCreateWriterResponse {
    id: number;
}
export interface MsgUpdateWriter {
    creator: string;
    id: number;
    moniker: string;
    description: string;
    nanoTimestamp: number;
}
export interface MsgUpdateWriterResponse {
}
export interface MsgDeleteWriter {
    creator: string;
    id: number;
}
export interface MsgDeleteWriterResponse {
}
export interface MsgCreateTopic {
    creator: string;
    description: string;
    totalRecords: number;
    totalWriter: number;
}
export interface MsgCreateTopicResponse {
    id: number;
}
export interface MsgUpdateTopic {
    creator: string;
    id: number;
    description: string;
    totalRecords: number;
    totalWriter: number;
}
export interface MsgUpdateTopicResponse {
}
export interface MsgDeleteTopic {
    creator: string;
    id: number;
}
export interface MsgDeleteTopicResponse {
}
export declare const MsgCreateOwner: {
    encode(message: MsgCreateOwner, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgCreateOwner;
    fromJSON(object: any): MsgCreateOwner;
    toJSON(message: MsgCreateOwner): unknown;
    fromPartial(object: DeepPartial<MsgCreateOwner>): MsgCreateOwner;
};
export declare const MsgCreateOwnerResponse: {
    encode(message: MsgCreateOwnerResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgCreateOwnerResponse;
    fromJSON(object: any): MsgCreateOwnerResponse;
    toJSON(message: MsgCreateOwnerResponse): unknown;
    fromPartial(object: DeepPartial<MsgCreateOwnerResponse>): MsgCreateOwnerResponse;
};
export declare const MsgUpdateOwner: {
    encode(message: MsgUpdateOwner, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgUpdateOwner;
    fromJSON(object: any): MsgUpdateOwner;
    toJSON(message: MsgUpdateOwner): unknown;
    fromPartial(object: DeepPartial<MsgUpdateOwner>): MsgUpdateOwner;
};
export declare const MsgUpdateOwnerResponse: {
    encode(_: MsgUpdateOwnerResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgUpdateOwnerResponse;
    fromJSON(_: any): MsgUpdateOwnerResponse;
    toJSON(_: MsgUpdateOwnerResponse): unknown;
    fromPartial(_: DeepPartial<MsgUpdateOwnerResponse>): MsgUpdateOwnerResponse;
};
export declare const MsgDeleteOwner: {
    encode(message: MsgDeleteOwner, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgDeleteOwner;
    fromJSON(object: any): MsgDeleteOwner;
    toJSON(message: MsgDeleteOwner): unknown;
    fromPartial(object: DeepPartial<MsgDeleteOwner>): MsgDeleteOwner;
};
export declare const MsgDeleteOwnerResponse: {
    encode(_: MsgDeleteOwnerResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgDeleteOwnerResponse;
    fromJSON(_: any): MsgDeleteOwnerResponse;
    toJSON(_: MsgDeleteOwnerResponse): unknown;
    fromPartial(_: DeepPartial<MsgDeleteOwnerResponse>): MsgDeleteOwnerResponse;
};
export declare const MsgCreateRecord: {
    encode(message: MsgCreateRecord, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgCreateRecord;
    fromJSON(object: any): MsgCreateRecord;
    toJSON(message: MsgCreateRecord): unknown;
    fromPartial(object: DeepPartial<MsgCreateRecord>): MsgCreateRecord;
};
export declare const MsgCreateRecordResponse: {
    encode(message: MsgCreateRecordResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgCreateRecordResponse;
    fromJSON(object: any): MsgCreateRecordResponse;
    toJSON(message: MsgCreateRecordResponse): unknown;
    fromPartial(object: DeepPartial<MsgCreateRecordResponse>): MsgCreateRecordResponse;
};
export declare const MsgUpdateRecord: {
    encode(message: MsgUpdateRecord, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgUpdateRecord;
    fromJSON(object: any): MsgUpdateRecord;
    toJSON(message: MsgUpdateRecord): unknown;
    fromPartial(object: DeepPartial<MsgUpdateRecord>): MsgUpdateRecord;
};
export declare const MsgUpdateRecordResponse: {
    encode(_: MsgUpdateRecordResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgUpdateRecordResponse;
    fromJSON(_: any): MsgUpdateRecordResponse;
    toJSON(_: MsgUpdateRecordResponse): unknown;
    fromPartial(_: DeepPartial<MsgUpdateRecordResponse>): MsgUpdateRecordResponse;
};
export declare const MsgDeleteRecord: {
    encode(message: MsgDeleteRecord, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgDeleteRecord;
    fromJSON(object: any): MsgDeleteRecord;
    toJSON(message: MsgDeleteRecord): unknown;
    fromPartial(object: DeepPartial<MsgDeleteRecord>): MsgDeleteRecord;
};
export declare const MsgDeleteRecordResponse: {
    encode(_: MsgDeleteRecordResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgDeleteRecordResponse;
    fromJSON(_: any): MsgDeleteRecordResponse;
    toJSON(_: MsgDeleteRecordResponse): unknown;
    fromPartial(_: DeepPartial<MsgDeleteRecordResponse>): MsgDeleteRecordResponse;
};
export declare const MsgCreateWriter: {
    encode(message: MsgCreateWriter, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgCreateWriter;
    fromJSON(object: any): MsgCreateWriter;
    toJSON(message: MsgCreateWriter): unknown;
    fromPartial(object: DeepPartial<MsgCreateWriter>): MsgCreateWriter;
};
export declare const MsgCreateWriterResponse: {
    encode(message: MsgCreateWriterResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgCreateWriterResponse;
    fromJSON(object: any): MsgCreateWriterResponse;
    toJSON(message: MsgCreateWriterResponse): unknown;
    fromPartial(object: DeepPartial<MsgCreateWriterResponse>): MsgCreateWriterResponse;
};
export declare const MsgUpdateWriter: {
    encode(message: MsgUpdateWriter, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgUpdateWriter;
    fromJSON(object: any): MsgUpdateWriter;
    toJSON(message: MsgUpdateWriter): unknown;
    fromPartial(object: DeepPartial<MsgUpdateWriter>): MsgUpdateWriter;
};
export declare const MsgUpdateWriterResponse: {
    encode(_: MsgUpdateWriterResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgUpdateWriterResponse;
    fromJSON(_: any): MsgUpdateWriterResponse;
    toJSON(_: MsgUpdateWriterResponse): unknown;
    fromPartial(_: DeepPartial<MsgUpdateWriterResponse>): MsgUpdateWriterResponse;
};
export declare const MsgDeleteWriter: {
    encode(message: MsgDeleteWriter, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgDeleteWriter;
    fromJSON(object: any): MsgDeleteWriter;
    toJSON(message: MsgDeleteWriter): unknown;
    fromPartial(object: DeepPartial<MsgDeleteWriter>): MsgDeleteWriter;
};
export declare const MsgDeleteWriterResponse: {
    encode(_: MsgDeleteWriterResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgDeleteWriterResponse;
    fromJSON(_: any): MsgDeleteWriterResponse;
    toJSON(_: MsgDeleteWriterResponse): unknown;
    fromPartial(_: DeepPartial<MsgDeleteWriterResponse>): MsgDeleteWriterResponse;
};
export declare const MsgCreateTopic: {
    encode(message: MsgCreateTopic, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgCreateTopic;
    fromJSON(object: any): MsgCreateTopic;
    toJSON(message: MsgCreateTopic): unknown;
    fromPartial(object: DeepPartial<MsgCreateTopic>): MsgCreateTopic;
};
export declare const MsgCreateTopicResponse: {
    encode(message: MsgCreateTopicResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgCreateTopicResponse;
    fromJSON(object: any): MsgCreateTopicResponse;
    toJSON(message: MsgCreateTopicResponse): unknown;
    fromPartial(object: DeepPartial<MsgCreateTopicResponse>): MsgCreateTopicResponse;
};
export declare const MsgUpdateTopic: {
    encode(message: MsgUpdateTopic, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgUpdateTopic;
    fromJSON(object: any): MsgUpdateTopic;
    toJSON(message: MsgUpdateTopic): unknown;
    fromPartial(object: DeepPartial<MsgUpdateTopic>): MsgUpdateTopic;
};
export declare const MsgUpdateTopicResponse: {
    encode(_: MsgUpdateTopicResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgUpdateTopicResponse;
    fromJSON(_: any): MsgUpdateTopicResponse;
    toJSON(_: MsgUpdateTopicResponse): unknown;
    fromPartial(_: DeepPartial<MsgUpdateTopicResponse>): MsgUpdateTopicResponse;
};
export declare const MsgDeleteTopic: {
    encode(message: MsgDeleteTopic, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgDeleteTopic;
    fromJSON(object: any): MsgDeleteTopic;
    toJSON(message: MsgDeleteTopic): unknown;
    fromPartial(object: DeepPartial<MsgDeleteTopic>): MsgDeleteTopic;
};
export declare const MsgDeleteTopicResponse: {
    encode(_: MsgDeleteTopicResponse, writer?: Writer): Writer;
    decode(input: Reader | Uint8Array, length?: number): MsgDeleteTopicResponse;
    fromJSON(_: any): MsgDeleteTopicResponse;
    toJSON(_: MsgDeleteTopicResponse): unknown;
    fromPartial(_: DeepPartial<MsgDeleteTopicResponse>): MsgDeleteTopicResponse;
};
/** Msg defines the Msg service. */
export interface Msg {
    /** this line is used by starport scaffolding # proto/tx/rpc */
    CreateOwner(request: MsgCreateOwner): Promise<MsgCreateOwnerResponse>;
    UpdateOwner(request: MsgUpdateOwner): Promise<MsgUpdateOwnerResponse>;
    DeleteOwner(request: MsgDeleteOwner): Promise<MsgDeleteOwnerResponse>;
    CreateRecord(request: MsgCreateRecord): Promise<MsgCreateRecordResponse>;
    UpdateRecord(request: MsgUpdateRecord): Promise<MsgUpdateRecordResponse>;
    DeleteRecord(request: MsgDeleteRecord): Promise<MsgDeleteRecordResponse>;
    CreateWriter(request: MsgCreateWriter): Promise<MsgCreateWriterResponse>;
    UpdateWriter(request: MsgUpdateWriter): Promise<MsgUpdateWriterResponse>;
    DeleteWriter(request: MsgDeleteWriter): Promise<MsgDeleteWriterResponse>;
    CreateTopic(request: MsgCreateTopic): Promise<MsgCreateTopicResponse>;
    UpdateTopic(request: MsgUpdateTopic): Promise<MsgUpdateTopicResponse>;
    DeleteTopic(request: MsgDeleteTopic): Promise<MsgDeleteTopicResponse>;
}
export declare class MsgClientImpl implements Msg {
    private readonly rpc;
    constructor(rpc: Rpc);
    CreateOwner(request: MsgCreateOwner): Promise<MsgCreateOwnerResponse>;
    UpdateOwner(request: MsgUpdateOwner): Promise<MsgUpdateOwnerResponse>;
    DeleteOwner(request: MsgDeleteOwner): Promise<MsgDeleteOwnerResponse>;
    CreateRecord(request: MsgCreateRecord): Promise<MsgCreateRecordResponse>;
    UpdateRecord(request: MsgUpdateRecord): Promise<MsgUpdateRecordResponse>;
    DeleteRecord(request: MsgDeleteRecord): Promise<MsgDeleteRecordResponse>;
    CreateWriter(request: MsgCreateWriter): Promise<MsgCreateWriterResponse>;
    UpdateWriter(request: MsgUpdateWriter): Promise<MsgUpdateWriterResponse>;
    DeleteWriter(request: MsgDeleteWriter): Promise<MsgDeleteWriterResponse>;
    CreateTopic(request: MsgCreateTopic): Promise<MsgCreateTopicResponse>;
    UpdateTopic(request: MsgUpdateTopic): Promise<MsgUpdateTopicResponse>;
    DeleteTopic(request: MsgDeleteTopic): Promise<MsgDeleteTopicResponse>;
}
interface Rpc {
    request(service: string, method: string, data: Uint8Array): Promise<Uint8Array>;
}
declare type Builtin = Date | Function | Uint8Array | string | number | undefined;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;
export {};
