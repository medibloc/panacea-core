/* eslint-disable */
import { Reader, util, configure, Writer as Writer1 } from "protobufjs/minimal";
import * as Long from "long";
import { Owner } from "../aol/owner";
import {
  PageRequest,
  PageResponse,
} from "../cosmos/base/query/v1beta1/pagination";
import { Record } from "../aol/record";
import { Writer } from "../aol/writer";
import { Topic } from "../aol/topic";

export const protobufPackage = "medibloc.panaceacore.aol";

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

const baseQueryGetOwnerRequest: object = { id: 0 };

export const QueryGetOwnerRequest = {
  encode(
    message: QueryGetOwnerRequest,
    writer: Writer1 = Writer1.create()
  ): Writer1 {
    if (message.id !== 0) {
      writer.uint32(8).uint64(message.id);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryGetOwnerRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryGetOwnerRequest } as QueryGetOwnerRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetOwnerRequest {
    const message = { ...baseQueryGetOwnerRequest } as QueryGetOwnerRequest;
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    return message;
  },

  toJSON(message: QueryGetOwnerRequest): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(object: DeepPartial<QueryGetOwnerRequest>): QueryGetOwnerRequest {
    const message = { ...baseQueryGetOwnerRequest } as QueryGetOwnerRequest;
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    return message;
  },
};

const baseQueryGetOwnerResponse: object = {};

export const QueryGetOwnerResponse = {
  encode(
    message: QueryGetOwnerResponse,
    writer: Writer1 = Writer1.create()
  ): Writer1 {
    if (message.Owner !== undefined) {
      Owner.encode(message.Owner, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryGetOwnerResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryGetOwnerResponse } as QueryGetOwnerResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.Owner = Owner.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetOwnerResponse {
    const message = { ...baseQueryGetOwnerResponse } as QueryGetOwnerResponse;
    if (object.Owner !== undefined && object.Owner !== null) {
      message.Owner = Owner.fromJSON(object.Owner);
    } else {
      message.Owner = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetOwnerResponse): unknown {
    const obj: any = {};
    message.Owner !== undefined &&
      (obj.Owner = message.Owner ? Owner.toJSON(message.Owner) : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetOwnerResponse>
  ): QueryGetOwnerResponse {
    const message = { ...baseQueryGetOwnerResponse } as QueryGetOwnerResponse;
    if (object.Owner !== undefined && object.Owner !== null) {
      message.Owner = Owner.fromPartial(object.Owner);
    } else {
      message.Owner = undefined;
    }
    return message;
  },
};

const baseQueryAllOwnerRequest: object = {};

export const QueryAllOwnerRequest = {
  encode(
    message: QueryAllOwnerRequest,
    writer: Writer1 = Writer1.create()
  ): Writer1 {
    if (message.pagination !== undefined) {
      PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryAllOwnerRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryAllOwnerRequest } as QueryAllOwnerRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.pagination = PageRequest.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllOwnerRequest {
    const message = { ...baseQueryAllOwnerRequest } as QueryAllOwnerRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllOwnerRequest): unknown {
    const obj: any = {};
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageRequest.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<QueryAllOwnerRequest>): QueryAllOwnerRequest {
    const message = { ...baseQueryAllOwnerRequest } as QueryAllOwnerRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryAllOwnerResponse: object = {};

export const QueryAllOwnerResponse = {
  encode(
    message: QueryAllOwnerResponse,
    writer: Writer1 = Writer1.create()
  ): Writer1 {
    for (const v of message.Owner) {
      Owner.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.pagination !== undefined) {
      PageResponse.encode(
        message.pagination,
        writer.uint32(18).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryAllOwnerResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryAllOwnerResponse } as QueryAllOwnerResponse;
    message.Owner = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.Owner.push(Owner.decode(reader, reader.uint32()));
          break;
        case 2:
          message.pagination = PageResponse.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllOwnerResponse {
    const message = { ...baseQueryAllOwnerResponse } as QueryAllOwnerResponse;
    message.Owner = [];
    if (object.Owner !== undefined && object.Owner !== null) {
      for (const e of object.Owner) {
        message.Owner.push(Owner.fromJSON(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllOwnerResponse): unknown {
    const obj: any = {};
    if (message.Owner) {
      obj.Owner = message.Owner.map((e) => (e ? Owner.toJSON(e) : undefined));
    } else {
      obj.Owner = [];
    }
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageResponse.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllOwnerResponse>
  ): QueryAllOwnerResponse {
    const message = { ...baseQueryAllOwnerResponse } as QueryAllOwnerResponse;
    message.Owner = [];
    if (object.Owner !== undefined && object.Owner !== null) {
      for (const e of object.Owner) {
        message.Owner.push(Owner.fromPartial(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryGetRecordRequest: object = { id: 0 };

export const QueryGetRecordRequest = {
  encode(
    message: QueryGetRecordRequest,
    writer: Writer1 = Writer1.create()
  ): Writer1 {
    if (message.id !== 0) {
      writer.uint32(8).uint64(message.id);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryGetRecordRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryGetRecordRequest } as QueryGetRecordRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetRecordRequest {
    const message = { ...baseQueryGetRecordRequest } as QueryGetRecordRequest;
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    return message;
  },

  toJSON(message: QueryGetRecordRequest): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetRecordRequest>
  ): QueryGetRecordRequest {
    const message = { ...baseQueryGetRecordRequest } as QueryGetRecordRequest;
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    return message;
  },
};

const baseQueryGetRecordResponse: object = {};

export const QueryGetRecordResponse = {
  encode(
    message: QueryGetRecordResponse,
    writer: Writer1 = Writer1.create()
  ): Writer1 {
    if (message.Record !== undefined) {
      Record.encode(message.Record, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryGetRecordResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryGetRecordResponse } as QueryGetRecordResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.Record = Record.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetRecordResponse {
    const message = { ...baseQueryGetRecordResponse } as QueryGetRecordResponse;
    if (object.Record !== undefined && object.Record !== null) {
      message.Record = Record.fromJSON(object.Record);
    } else {
      message.Record = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetRecordResponse): unknown {
    const obj: any = {};
    message.Record !== undefined &&
      (obj.Record = message.Record ? Record.toJSON(message.Record) : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetRecordResponse>
  ): QueryGetRecordResponse {
    const message = { ...baseQueryGetRecordResponse } as QueryGetRecordResponse;
    if (object.Record !== undefined && object.Record !== null) {
      message.Record = Record.fromPartial(object.Record);
    } else {
      message.Record = undefined;
    }
    return message;
  },
};

const baseQueryAllRecordRequest: object = {};

export const QueryAllRecordRequest = {
  encode(
    message: QueryAllRecordRequest,
    writer: Writer1 = Writer1.create()
  ): Writer1 {
    if (message.pagination !== undefined) {
      PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryAllRecordRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryAllRecordRequest } as QueryAllRecordRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.pagination = PageRequest.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllRecordRequest {
    const message = { ...baseQueryAllRecordRequest } as QueryAllRecordRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllRecordRequest): unknown {
    const obj: any = {};
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageRequest.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllRecordRequest>
  ): QueryAllRecordRequest {
    const message = { ...baseQueryAllRecordRequest } as QueryAllRecordRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryAllRecordResponse: object = {};

export const QueryAllRecordResponse = {
  encode(
    message: QueryAllRecordResponse,
    writer: Writer1 = Writer1.create()
  ): Writer1 {
    for (const v of message.Record) {
      Record.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.pagination !== undefined) {
      PageResponse.encode(
        message.pagination,
        writer.uint32(18).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryAllRecordResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryAllRecordResponse } as QueryAllRecordResponse;
    message.Record = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.Record.push(Record.decode(reader, reader.uint32()));
          break;
        case 2:
          message.pagination = PageResponse.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllRecordResponse {
    const message = { ...baseQueryAllRecordResponse } as QueryAllRecordResponse;
    message.Record = [];
    if (object.Record !== undefined && object.Record !== null) {
      for (const e of object.Record) {
        message.Record.push(Record.fromJSON(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllRecordResponse): unknown {
    const obj: any = {};
    if (message.Record) {
      obj.Record = message.Record.map((e) =>
        e ? Record.toJSON(e) : undefined
      );
    } else {
      obj.Record = [];
    }
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageResponse.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllRecordResponse>
  ): QueryAllRecordResponse {
    const message = { ...baseQueryAllRecordResponse } as QueryAllRecordResponse;
    message.Record = [];
    if (object.Record !== undefined && object.Record !== null) {
      for (const e of object.Record) {
        message.Record.push(Record.fromPartial(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryGetWriterRequest: object = { id: 0 };

export const QueryGetWriterRequest = {
  encode(
    message: QueryGetWriterRequest,
    writer: Writer1 = Writer1.create()
  ): Writer1 {
    if (message.id !== 0) {
      writer.uint32(8).uint64(message.id);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryGetWriterRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryGetWriterRequest } as QueryGetWriterRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetWriterRequest {
    const message = { ...baseQueryGetWriterRequest } as QueryGetWriterRequest;
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    return message;
  },

  toJSON(message: QueryGetWriterRequest): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetWriterRequest>
  ): QueryGetWriterRequest {
    const message = { ...baseQueryGetWriterRequest } as QueryGetWriterRequest;
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    return message;
  },
};

const baseQueryGetWriterResponse: object = {};

export const QueryGetWriterResponse = {
  encode(
    message: QueryGetWriterResponse,
    writer: Writer1 = Writer1.create()
  ): Writer1 {
    if (message.Writer !== undefined) {
      Writer.encode(message.Writer, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryGetWriterResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryGetWriterResponse } as QueryGetWriterResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.Writer = Writer.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetWriterResponse {
    const message = { ...baseQueryGetWriterResponse } as QueryGetWriterResponse;
    if (object.Writer !== undefined && object.Writer !== null) {
      message.Writer = Writer.fromJSON(object.Writer);
    } else {
      message.Writer = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetWriterResponse): unknown {
    const obj: any = {};
    message.Writer !== undefined &&
      (obj.Writer = message.Writer ? Writer.toJSON(message.Writer) : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetWriterResponse>
  ): QueryGetWriterResponse {
    const message = { ...baseQueryGetWriterResponse } as QueryGetWriterResponse;
    if (object.Writer !== undefined && object.Writer !== null) {
      message.Writer = Writer.fromPartial(object.Writer);
    } else {
      message.Writer = undefined;
    }
    return message;
  },
};

const baseQueryAllWriterRequest: object = {};

export const QueryAllWriterRequest = {
  encode(
    message: QueryAllWriterRequest,
    writer: Writer1 = Writer1.create()
  ): Writer1 {
    if (message.pagination !== undefined) {
      PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryAllWriterRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryAllWriterRequest } as QueryAllWriterRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.pagination = PageRequest.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllWriterRequest {
    const message = { ...baseQueryAllWriterRequest } as QueryAllWriterRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllWriterRequest): unknown {
    const obj: any = {};
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageRequest.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllWriterRequest>
  ): QueryAllWriterRequest {
    const message = { ...baseQueryAllWriterRequest } as QueryAllWriterRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryAllWriterResponse: object = {};

export const QueryAllWriterResponse = {
  encode(
    message: QueryAllWriterResponse,
    writer: Writer1 = Writer1.create()
  ): Writer1 {
    for (const v of message.Writer) {
      Writer.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.pagination !== undefined) {
      PageResponse.encode(
        message.pagination,
        writer.uint32(18).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryAllWriterResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryAllWriterResponse } as QueryAllWriterResponse;
    message.Writer = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.Writer.push(Writer.decode(reader, reader.uint32()));
          break;
        case 2:
          message.pagination = PageResponse.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllWriterResponse {
    const message = { ...baseQueryAllWriterResponse } as QueryAllWriterResponse;
    message.Writer = [];
    if (object.Writer !== undefined && object.Writer !== null) {
      for (const e of object.Writer) {
        message.Writer.push(Writer.fromJSON(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllWriterResponse): unknown {
    const obj: any = {};
    if (message.Writer) {
      obj.Writer = message.Writer.map((e) =>
        e ? Writer.toJSON(e) : undefined
      );
    } else {
      obj.Writer = [];
    }
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageResponse.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllWriterResponse>
  ): QueryAllWriterResponse {
    const message = { ...baseQueryAllWriterResponse } as QueryAllWriterResponse;
    message.Writer = [];
    if (object.Writer !== undefined && object.Writer !== null) {
      for (const e of object.Writer) {
        message.Writer.push(Writer.fromPartial(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryGetTopicRequest: object = { id: 0 };

export const QueryGetTopicRequest = {
  encode(
    message: QueryGetTopicRequest,
    writer: Writer1 = Writer1.create()
  ): Writer1 {
    if (message.id !== 0) {
      writer.uint32(8).uint64(message.id);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryGetTopicRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryGetTopicRequest } as QueryGetTopicRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetTopicRequest {
    const message = { ...baseQueryGetTopicRequest } as QueryGetTopicRequest;
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    return message;
  },

  toJSON(message: QueryGetTopicRequest): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(object: DeepPartial<QueryGetTopicRequest>): QueryGetTopicRequest {
    const message = { ...baseQueryGetTopicRequest } as QueryGetTopicRequest;
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    return message;
  },
};

const baseQueryGetTopicResponse: object = {};

export const QueryGetTopicResponse = {
  encode(
    message: QueryGetTopicResponse,
    writer: Writer1 = Writer1.create()
  ): Writer1 {
    if (message.Topic !== undefined) {
      Topic.encode(message.Topic, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryGetTopicResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryGetTopicResponse } as QueryGetTopicResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.Topic = Topic.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetTopicResponse {
    const message = { ...baseQueryGetTopicResponse } as QueryGetTopicResponse;
    if (object.Topic !== undefined && object.Topic !== null) {
      message.Topic = Topic.fromJSON(object.Topic);
    } else {
      message.Topic = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetTopicResponse): unknown {
    const obj: any = {};
    message.Topic !== undefined &&
      (obj.Topic = message.Topic ? Topic.toJSON(message.Topic) : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetTopicResponse>
  ): QueryGetTopicResponse {
    const message = { ...baseQueryGetTopicResponse } as QueryGetTopicResponse;
    if (object.Topic !== undefined && object.Topic !== null) {
      message.Topic = Topic.fromPartial(object.Topic);
    } else {
      message.Topic = undefined;
    }
    return message;
  },
};

const baseQueryAllTopicRequest: object = {};

export const QueryAllTopicRequest = {
  encode(
    message: QueryAllTopicRequest,
    writer: Writer1 = Writer1.create()
  ): Writer1 {
    if (message.pagination !== undefined) {
      PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryAllTopicRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryAllTopicRequest } as QueryAllTopicRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.pagination = PageRequest.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllTopicRequest {
    const message = { ...baseQueryAllTopicRequest } as QueryAllTopicRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllTopicRequest): unknown {
    const obj: any = {};
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageRequest.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<QueryAllTopicRequest>): QueryAllTopicRequest {
    const message = { ...baseQueryAllTopicRequest } as QueryAllTopicRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryAllTopicResponse: object = {};

export const QueryAllTopicResponse = {
  encode(
    message: QueryAllTopicResponse,
    writer: Writer1 = Writer1.create()
  ): Writer1 {
    for (const v of message.Topic) {
      Topic.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.pagination !== undefined) {
      PageResponse.encode(
        message.pagination,
        writer.uint32(18).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryAllTopicResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryAllTopicResponse } as QueryAllTopicResponse;
    message.Topic = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.Topic.push(Topic.decode(reader, reader.uint32()));
          break;
        case 2:
          message.pagination = PageResponse.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllTopicResponse {
    const message = { ...baseQueryAllTopicResponse } as QueryAllTopicResponse;
    message.Topic = [];
    if (object.Topic !== undefined && object.Topic !== null) {
      for (const e of object.Topic) {
        message.Topic.push(Topic.fromJSON(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllTopicResponse): unknown {
    const obj: any = {};
    if (message.Topic) {
      obj.Topic = message.Topic.map((e) => (e ? Topic.toJSON(e) : undefined));
    } else {
      obj.Topic = [];
    }
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageResponse.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllTopicResponse>
  ): QueryAllTopicResponse {
    const message = { ...baseQueryAllTopicResponse } as QueryAllTopicResponse;
    message.Topic = [];
    if (object.Topic !== undefined && object.Topic !== null) {
      for (const e of object.Topic) {
        message.Topic.push(Topic.fromPartial(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
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

export class QueryClientImpl implements Query {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  Owner(request: QueryGetOwnerRequest): Promise<QueryGetOwnerResponse> {
    const data = QueryGetOwnerRequest.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Query",
      "Owner",
      data
    );
    return promise.then((data) =>
      QueryGetOwnerResponse.decode(new Reader(data))
    );
  }

  OwnerAll(request: QueryAllOwnerRequest): Promise<QueryAllOwnerResponse> {
    const data = QueryAllOwnerRequest.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Query",
      "OwnerAll",
      data
    );
    return promise.then((data) =>
      QueryAllOwnerResponse.decode(new Reader(data))
    );
  }

  Record(request: QueryGetRecordRequest): Promise<QueryGetRecordResponse> {
    const data = QueryGetRecordRequest.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Query",
      "Record",
      data
    );
    return promise.then((data) =>
      QueryGetRecordResponse.decode(new Reader(data))
    );
  }

  RecordAll(request: QueryAllRecordRequest): Promise<QueryAllRecordResponse> {
    const data = QueryAllRecordRequest.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Query",
      "RecordAll",
      data
    );
    return promise.then((data) =>
      QueryAllRecordResponse.decode(new Reader(data))
    );
  }

  Writer(request: QueryGetWriterRequest): Promise<QueryGetWriterResponse> {
    const data = QueryGetWriterRequest.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Query",
      "Writer",
      data
    );
    return promise.then((data) =>
      QueryGetWriterResponse.decode(new Reader(data))
    );
  }

  WriterAll(request: QueryAllWriterRequest): Promise<QueryAllWriterResponse> {
    const data = QueryAllWriterRequest.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Query",
      "WriterAll",
      data
    );
    return promise.then((data) =>
      QueryAllWriterResponse.decode(new Reader(data))
    );
  }

  Topic(request: QueryGetTopicRequest): Promise<QueryGetTopicResponse> {
    const data = QueryGetTopicRequest.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Query",
      "Topic",
      data
    );
    return promise.then((data) =>
      QueryGetTopicResponse.decode(new Reader(data))
    );
  }

  TopicAll(request: QueryAllTopicRequest): Promise<QueryAllTopicResponse> {
    const data = QueryAllTopicRequest.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Query",
      "TopicAll",
      data
    );
    return promise.then((data) =>
      QueryAllTopicResponse.decode(new Reader(data))
    );
  }
}

interface Rpc {
  request(
    service: string,
    method: string,
    data: Uint8Array
  ): Promise<Uint8Array>;
}

declare var self: any | undefined;
declare var window: any | undefined;
var globalThis: any = (() => {
  if (typeof globalThis !== "undefined") return globalThis;
  if (typeof self !== "undefined") return self;
  if (typeof window !== "undefined") return window;
  if (typeof global !== "undefined") return global;
  throw "Unable to locate global object";
})();

type Builtin = Date | Function | Uint8Array | string | number | undefined;
export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

function longToNumber(long: Long): number {
  if (long.gt(Number.MAX_SAFE_INTEGER)) {
    throw new globalThis.Error("Value is larger than Number.MAX_SAFE_INTEGER");
  }
  return long.toNumber();
}

if (util.Long !== Long) {
  util.Long = Long as any;
  configure();
}
