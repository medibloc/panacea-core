/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";

export const protobufPackage = "medibloc.panaceacore.aol";

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

export interface MsgUpdateOwnerResponse {}

export interface MsgDeleteOwner {
  creator: string;
  id: number;
}

export interface MsgDeleteOwnerResponse {}

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

export interface MsgUpdateRecordResponse {}

export interface MsgDeleteRecord {
  creator: string;
  id: number;
}

export interface MsgDeleteRecordResponse {}

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

export interface MsgUpdateWriterResponse {}

export interface MsgDeleteWriter {
  creator: string;
  id: number;
}

export interface MsgDeleteWriterResponse {}

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

export interface MsgUpdateTopicResponse {}

export interface MsgDeleteTopic {
  creator: string;
  id: number;
}

export interface MsgDeleteTopicResponse {}

const baseMsgCreateOwner: object = { creator: "", totalTopics: 0 };

export const MsgCreateOwner = {
  encode(message: MsgCreateOwner, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.totalTopics !== 0) {
      writer.uint32(16).int32(message.totalTopics);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgCreateOwner {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgCreateOwner } as MsgCreateOwner;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.totalTopics = reader.int32();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgCreateOwner {
    const message = { ...baseMsgCreateOwner } as MsgCreateOwner;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.totalTopics !== undefined && object.totalTopics !== null) {
      message.totalTopics = Number(object.totalTopics);
    } else {
      message.totalTopics = 0;
    }
    return message;
  },

  toJSON(message: MsgCreateOwner): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.totalTopics !== undefined &&
      (obj.totalTopics = message.totalTopics);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgCreateOwner>): MsgCreateOwner {
    const message = { ...baseMsgCreateOwner } as MsgCreateOwner;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.totalTopics !== undefined && object.totalTopics !== null) {
      message.totalTopics = object.totalTopics;
    } else {
      message.totalTopics = 0;
    }
    return message;
  },
};

const baseMsgCreateOwnerResponse: object = { id: 0 };

export const MsgCreateOwnerResponse = {
  encode(
    message: MsgCreateOwnerResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.id !== 0) {
      writer.uint32(8).uint64(message.id);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgCreateOwnerResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgCreateOwnerResponse } as MsgCreateOwnerResponse;
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

  fromJSON(object: any): MsgCreateOwnerResponse {
    const message = { ...baseMsgCreateOwnerResponse } as MsgCreateOwnerResponse;
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    return message;
  },

  toJSON(message: MsgCreateOwnerResponse): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgCreateOwnerResponse>
  ): MsgCreateOwnerResponse {
    const message = { ...baseMsgCreateOwnerResponse } as MsgCreateOwnerResponse;
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    return message;
  },
};

const baseMsgUpdateOwner: object = { creator: "", id: 0, totalTopics: 0 };

export const MsgUpdateOwner = {
  encode(message: MsgUpdateOwner, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.id !== 0) {
      writer.uint32(16).uint64(message.id);
    }
    if (message.totalTopics !== 0) {
      writer.uint32(24).int32(message.totalTopics);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgUpdateOwner {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgUpdateOwner } as MsgUpdateOwner;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        case 3:
          message.totalTopics = reader.int32();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgUpdateOwner {
    const message = { ...baseMsgUpdateOwner } as MsgUpdateOwner;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    if (object.totalTopics !== undefined && object.totalTopics !== null) {
      message.totalTopics = Number(object.totalTopics);
    } else {
      message.totalTopics = 0;
    }
    return message;
  },

  toJSON(message: MsgUpdateOwner): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.id !== undefined && (obj.id = message.id);
    message.totalTopics !== undefined &&
      (obj.totalTopics = message.totalTopics);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgUpdateOwner>): MsgUpdateOwner {
    const message = { ...baseMsgUpdateOwner } as MsgUpdateOwner;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    if (object.totalTopics !== undefined && object.totalTopics !== null) {
      message.totalTopics = object.totalTopics;
    } else {
      message.totalTopics = 0;
    }
    return message;
  },
};

const baseMsgUpdateOwnerResponse: object = {};

export const MsgUpdateOwnerResponse = {
  encode(_: MsgUpdateOwnerResponse, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgUpdateOwnerResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgUpdateOwnerResponse } as MsgUpdateOwnerResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgUpdateOwnerResponse {
    const message = { ...baseMsgUpdateOwnerResponse } as MsgUpdateOwnerResponse;
    return message;
  },

  toJSON(_: MsgUpdateOwnerResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<MsgUpdateOwnerResponse>): MsgUpdateOwnerResponse {
    const message = { ...baseMsgUpdateOwnerResponse } as MsgUpdateOwnerResponse;
    return message;
  },
};

const baseMsgDeleteOwner: object = { creator: "", id: 0 };

export const MsgDeleteOwner = {
  encode(message: MsgDeleteOwner, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.id !== 0) {
      writer.uint32(16).uint64(message.id);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgDeleteOwner {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgDeleteOwner } as MsgDeleteOwner;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgDeleteOwner {
    const message = { ...baseMsgDeleteOwner } as MsgDeleteOwner;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    return message;
  },

  toJSON(message: MsgDeleteOwner): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgDeleteOwner>): MsgDeleteOwner {
    const message = { ...baseMsgDeleteOwner } as MsgDeleteOwner;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    return message;
  },
};

const baseMsgDeleteOwnerResponse: object = {};

export const MsgDeleteOwnerResponse = {
  encode(_: MsgDeleteOwnerResponse, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgDeleteOwnerResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgDeleteOwnerResponse } as MsgDeleteOwnerResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgDeleteOwnerResponse {
    const message = { ...baseMsgDeleteOwnerResponse } as MsgDeleteOwnerResponse;
    return message;
  },

  toJSON(_: MsgDeleteOwnerResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<MsgDeleteOwnerResponse>): MsgDeleteOwnerResponse {
    const message = { ...baseMsgDeleteOwnerResponse } as MsgDeleteOwnerResponse;
    return message;
  },
};

const baseMsgCreateRecord: object = {
  creator: "",
  key: "",
  value: "",
  nanoTimestamp: 0,
  writerAddress: "",
};

export const MsgCreateRecord = {
  encode(message: MsgCreateRecord, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.key !== "") {
      writer.uint32(18).string(message.key);
    }
    if (message.value !== "") {
      writer.uint32(26).string(message.value);
    }
    if (message.nanoTimestamp !== 0) {
      writer.uint32(32).int32(message.nanoTimestamp);
    }
    if (message.writerAddress !== "") {
      writer.uint32(42).string(message.writerAddress);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgCreateRecord {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgCreateRecord } as MsgCreateRecord;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.key = reader.string();
          break;
        case 3:
          message.value = reader.string();
          break;
        case 4:
          message.nanoTimestamp = reader.int32();
          break;
        case 5:
          message.writerAddress = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgCreateRecord {
    const message = { ...baseMsgCreateRecord } as MsgCreateRecord;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.key !== undefined && object.key !== null) {
      message.key = String(object.key);
    } else {
      message.key = "";
    }
    if (object.value !== undefined && object.value !== null) {
      message.value = String(object.value);
    } else {
      message.value = "";
    }
    if (object.nanoTimestamp !== undefined && object.nanoTimestamp !== null) {
      message.nanoTimestamp = Number(object.nanoTimestamp);
    } else {
      message.nanoTimestamp = 0;
    }
    if (object.writerAddress !== undefined && object.writerAddress !== null) {
      message.writerAddress = String(object.writerAddress);
    } else {
      message.writerAddress = "";
    }
    return message;
  },

  toJSON(message: MsgCreateRecord): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.key !== undefined && (obj.key = message.key);
    message.value !== undefined && (obj.value = message.value);
    message.nanoTimestamp !== undefined &&
      (obj.nanoTimestamp = message.nanoTimestamp);
    message.writerAddress !== undefined &&
      (obj.writerAddress = message.writerAddress);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgCreateRecord>): MsgCreateRecord {
    const message = { ...baseMsgCreateRecord } as MsgCreateRecord;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.key !== undefined && object.key !== null) {
      message.key = object.key;
    } else {
      message.key = "";
    }
    if (object.value !== undefined && object.value !== null) {
      message.value = object.value;
    } else {
      message.value = "";
    }
    if (object.nanoTimestamp !== undefined && object.nanoTimestamp !== null) {
      message.nanoTimestamp = object.nanoTimestamp;
    } else {
      message.nanoTimestamp = 0;
    }
    if (object.writerAddress !== undefined && object.writerAddress !== null) {
      message.writerAddress = object.writerAddress;
    } else {
      message.writerAddress = "";
    }
    return message;
  },
};

const baseMsgCreateRecordResponse: object = { id: 0 };

export const MsgCreateRecordResponse = {
  encode(
    message: MsgCreateRecordResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.id !== 0) {
      writer.uint32(8).uint64(message.id);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgCreateRecordResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgCreateRecordResponse,
    } as MsgCreateRecordResponse;
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

  fromJSON(object: any): MsgCreateRecordResponse {
    const message = {
      ...baseMsgCreateRecordResponse,
    } as MsgCreateRecordResponse;
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    return message;
  },

  toJSON(message: MsgCreateRecordResponse): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgCreateRecordResponse>
  ): MsgCreateRecordResponse {
    const message = {
      ...baseMsgCreateRecordResponse,
    } as MsgCreateRecordResponse;
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    return message;
  },
};

const baseMsgUpdateRecord: object = {
  creator: "",
  id: 0,
  key: "",
  value: "",
  nanoTimestamp: 0,
  writerAddress: "",
};

export const MsgUpdateRecord = {
  encode(message: MsgUpdateRecord, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.id !== 0) {
      writer.uint32(16).uint64(message.id);
    }
    if (message.key !== "") {
      writer.uint32(26).string(message.key);
    }
    if (message.value !== "") {
      writer.uint32(34).string(message.value);
    }
    if (message.nanoTimestamp !== 0) {
      writer.uint32(40).int32(message.nanoTimestamp);
    }
    if (message.writerAddress !== "") {
      writer.uint32(50).string(message.writerAddress);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgUpdateRecord {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgUpdateRecord } as MsgUpdateRecord;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        case 3:
          message.key = reader.string();
          break;
        case 4:
          message.value = reader.string();
          break;
        case 5:
          message.nanoTimestamp = reader.int32();
          break;
        case 6:
          message.writerAddress = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgUpdateRecord {
    const message = { ...baseMsgUpdateRecord } as MsgUpdateRecord;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    if (object.key !== undefined && object.key !== null) {
      message.key = String(object.key);
    } else {
      message.key = "";
    }
    if (object.value !== undefined && object.value !== null) {
      message.value = String(object.value);
    } else {
      message.value = "";
    }
    if (object.nanoTimestamp !== undefined && object.nanoTimestamp !== null) {
      message.nanoTimestamp = Number(object.nanoTimestamp);
    } else {
      message.nanoTimestamp = 0;
    }
    if (object.writerAddress !== undefined && object.writerAddress !== null) {
      message.writerAddress = String(object.writerAddress);
    } else {
      message.writerAddress = "";
    }
    return message;
  },

  toJSON(message: MsgUpdateRecord): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.id !== undefined && (obj.id = message.id);
    message.key !== undefined && (obj.key = message.key);
    message.value !== undefined && (obj.value = message.value);
    message.nanoTimestamp !== undefined &&
      (obj.nanoTimestamp = message.nanoTimestamp);
    message.writerAddress !== undefined &&
      (obj.writerAddress = message.writerAddress);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgUpdateRecord>): MsgUpdateRecord {
    const message = { ...baseMsgUpdateRecord } as MsgUpdateRecord;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    if (object.key !== undefined && object.key !== null) {
      message.key = object.key;
    } else {
      message.key = "";
    }
    if (object.value !== undefined && object.value !== null) {
      message.value = object.value;
    } else {
      message.value = "";
    }
    if (object.nanoTimestamp !== undefined && object.nanoTimestamp !== null) {
      message.nanoTimestamp = object.nanoTimestamp;
    } else {
      message.nanoTimestamp = 0;
    }
    if (object.writerAddress !== undefined && object.writerAddress !== null) {
      message.writerAddress = object.writerAddress;
    } else {
      message.writerAddress = "";
    }
    return message;
  },
};

const baseMsgUpdateRecordResponse: object = {};

export const MsgUpdateRecordResponse = {
  encode(_: MsgUpdateRecordResponse, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgUpdateRecordResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgUpdateRecordResponse,
    } as MsgUpdateRecordResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgUpdateRecordResponse {
    const message = {
      ...baseMsgUpdateRecordResponse,
    } as MsgUpdateRecordResponse;
    return message;
  },

  toJSON(_: MsgUpdateRecordResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgUpdateRecordResponse>
  ): MsgUpdateRecordResponse {
    const message = {
      ...baseMsgUpdateRecordResponse,
    } as MsgUpdateRecordResponse;
    return message;
  },
};

const baseMsgDeleteRecord: object = { creator: "", id: 0 };

export const MsgDeleteRecord = {
  encode(message: MsgDeleteRecord, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.id !== 0) {
      writer.uint32(16).uint64(message.id);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgDeleteRecord {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgDeleteRecord } as MsgDeleteRecord;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgDeleteRecord {
    const message = { ...baseMsgDeleteRecord } as MsgDeleteRecord;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    return message;
  },

  toJSON(message: MsgDeleteRecord): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgDeleteRecord>): MsgDeleteRecord {
    const message = { ...baseMsgDeleteRecord } as MsgDeleteRecord;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    return message;
  },
};

const baseMsgDeleteRecordResponse: object = {};

export const MsgDeleteRecordResponse = {
  encode(_: MsgDeleteRecordResponse, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgDeleteRecordResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgDeleteRecordResponse,
    } as MsgDeleteRecordResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgDeleteRecordResponse {
    const message = {
      ...baseMsgDeleteRecordResponse,
    } as MsgDeleteRecordResponse;
    return message;
  },

  toJSON(_: MsgDeleteRecordResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgDeleteRecordResponse>
  ): MsgDeleteRecordResponse {
    const message = {
      ...baseMsgDeleteRecordResponse,
    } as MsgDeleteRecordResponse;
    return message;
  },
};

const baseMsgCreateWriter: object = {
  creator: "",
  moniker: "",
  description: "",
  nanoTimestamp: 0,
};

export const MsgCreateWriter = {
  encode(message: MsgCreateWriter, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.moniker !== "") {
      writer.uint32(18).string(message.moniker);
    }
    if (message.description !== "") {
      writer.uint32(26).string(message.description);
    }
    if (message.nanoTimestamp !== 0) {
      writer.uint32(32).int32(message.nanoTimestamp);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgCreateWriter {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgCreateWriter } as MsgCreateWriter;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.moniker = reader.string();
          break;
        case 3:
          message.description = reader.string();
          break;
        case 4:
          message.nanoTimestamp = reader.int32();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgCreateWriter {
    const message = { ...baseMsgCreateWriter } as MsgCreateWriter;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.moniker !== undefined && object.moniker !== null) {
      message.moniker = String(object.moniker);
    } else {
      message.moniker = "";
    }
    if (object.description !== undefined && object.description !== null) {
      message.description = String(object.description);
    } else {
      message.description = "";
    }
    if (object.nanoTimestamp !== undefined && object.nanoTimestamp !== null) {
      message.nanoTimestamp = Number(object.nanoTimestamp);
    } else {
      message.nanoTimestamp = 0;
    }
    return message;
  },

  toJSON(message: MsgCreateWriter): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.moniker !== undefined && (obj.moniker = message.moniker);
    message.description !== undefined &&
      (obj.description = message.description);
    message.nanoTimestamp !== undefined &&
      (obj.nanoTimestamp = message.nanoTimestamp);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgCreateWriter>): MsgCreateWriter {
    const message = { ...baseMsgCreateWriter } as MsgCreateWriter;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.moniker !== undefined && object.moniker !== null) {
      message.moniker = object.moniker;
    } else {
      message.moniker = "";
    }
    if (object.description !== undefined && object.description !== null) {
      message.description = object.description;
    } else {
      message.description = "";
    }
    if (object.nanoTimestamp !== undefined && object.nanoTimestamp !== null) {
      message.nanoTimestamp = object.nanoTimestamp;
    } else {
      message.nanoTimestamp = 0;
    }
    return message;
  },
};

const baseMsgCreateWriterResponse: object = { id: 0 };

export const MsgCreateWriterResponse = {
  encode(
    message: MsgCreateWriterResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.id !== 0) {
      writer.uint32(8).uint64(message.id);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgCreateWriterResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgCreateWriterResponse,
    } as MsgCreateWriterResponse;
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

  fromJSON(object: any): MsgCreateWriterResponse {
    const message = {
      ...baseMsgCreateWriterResponse,
    } as MsgCreateWriterResponse;
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    return message;
  },

  toJSON(message: MsgCreateWriterResponse): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgCreateWriterResponse>
  ): MsgCreateWriterResponse {
    const message = {
      ...baseMsgCreateWriterResponse,
    } as MsgCreateWriterResponse;
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    return message;
  },
};

const baseMsgUpdateWriter: object = {
  creator: "",
  id: 0,
  moniker: "",
  description: "",
  nanoTimestamp: 0,
};

export const MsgUpdateWriter = {
  encode(message: MsgUpdateWriter, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.id !== 0) {
      writer.uint32(16).uint64(message.id);
    }
    if (message.moniker !== "") {
      writer.uint32(26).string(message.moniker);
    }
    if (message.description !== "") {
      writer.uint32(34).string(message.description);
    }
    if (message.nanoTimestamp !== 0) {
      writer.uint32(40).int32(message.nanoTimestamp);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgUpdateWriter {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgUpdateWriter } as MsgUpdateWriter;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        case 3:
          message.moniker = reader.string();
          break;
        case 4:
          message.description = reader.string();
          break;
        case 5:
          message.nanoTimestamp = reader.int32();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgUpdateWriter {
    const message = { ...baseMsgUpdateWriter } as MsgUpdateWriter;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    if (object.moniker !== undefined && object.moniker !== null) {
      message.moniker = String(object.moniker);
    } else {
      message.moniker = "";
    }
    if (object.description !== undefined && object.description !== null) {
      message.description = String(object.description);
    } else {
      message.description = "";
    }
    if (object.nanoTimestamp !== undefined && object.nanoTimestamp !== null) {
      message.nanoTimestamp = Number(object.nanoTimestamp);
    } else {
      message.nanoTimestamp = 0;
    }
    return message;
  },

  toJSON(message: MsgUpdateWriter): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.id !== undefined && (obj.id = message.id);
    message.moniker !== undefined && (obj.moniker = message.moniker);
    message.description !== undefined &&
      (obj.description = message.description);
    message.nanoTimestamp !== undefined &&
      (obj.nanoTimestamp = message.nanoTimestamp);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgUpdateWriter>): MsgUpdateWriter {
    const message = { ...baseMsgUpdateWriter } as MsgUpdateWriter;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    if (object.moniker !== undefined && object.moniker !== null) {
      message.moniker = object.moniker;
    } else {
      message.moniker = "";
    }
    if (object.description !== undefined && object.description !== null) {
      message.description = object.description;
    } else {
      message.description = "";
    }
    if (object.nanoTimestamp !== undefined && object.nanoTimestamp !== null) {
      message.nanoTimestamp = object.nanoTimestamp;
    } else {
      message.nanoTimestamp = 0;
    }
    return message;
  },
};

const baseMsgUpdateWriterResponse: object = {};

export const MsgUpdateWriterResponse = {
  encode(_: MsgUpdateWriterResponse, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgUpdateWriterResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgUpdateWriterResponse,
    } as MsgUpdateWriterResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgUpdateWriterResponse {
    const message = {
      ...baseMsgUpdateWriterResponse,
    } as MsgUpdateWriterResponse;
    return message;
  },

  toJSON(_: MsgUpdateWriterResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgUpdateWriterResponse>
  ): MsgUpdateWriterResponse {
    const message = {
      ...baseMsgUpdateWriterResponse,
    } as MsgUpdateWriterResponse;
    return message;
  },
};

const baseMsgDeleteWriter: object = { creator: "", id: 0 };

export const MsgDeleteWriter = {
  encode(message: MsgDeleteWriter, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.id !== 0) {
      writer.uint32(16).uint64(message.id);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgDeleteWriter {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgDeleteWriter } as MsgDeleteWriter;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgDeleteWriter {
    const message = { ...baseMsgDeleteWriter } as MsgDeleteWriter;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    return message;
  },

  toJSON(message: MsgDeleteWriter): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgDeleteWriter>): MsgDeleteWriter {
    const message = { ...baseMsgDeleteWriter } as MsgDeleteWriter;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    return message;
  },
};

const baseMsgDeleteWriterResponse: object = {};

export const MsgDeleteWriterResponse = {
  encode(_: MsgDeleteWriterResponse, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgDeleteWriterResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgDeleteWriterResponse,
    } as MsgDeleteWriterResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgDeleteWriterResponse {
    const message = {
      ...baseMsgDeleteWriterResponse,
    } as MsgDeleteWriterResponse;
    return message;
  },

  toJSON(_: MsgDeleteWriterResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgDeleteWriterResponse>
  ): MsgDeleteWriterResponse {
    const message = {
      ...baseMsgDeleteWriterResponse,
    } as MsgDeleteWriterResponse;
    return message;
  },
};

const baseMsgCreateTopic: object = {
  creator: "",
  description: "",
  totalRecords: 0,
  totalWriter: 0,
};

export const MsgCreateTopic = {
  encode(message: MsgCreateTopic, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.description !== "") {
      writer.uint32(18).string(message.description);
    }
    if (message.totalRecords !== 0) {
      writer.uint32(24).int32(message.totalRecords);
    }
    if (message.totalWriter !== 0) {
      writer.uint32(32).int32(message.totalWriter);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgCreateTopic {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgCreateTopic } as MsgCreateTopic;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.description = reader.string();
          break;
        case 3:
          message.totalRecords = reader.int32();
          break;
        case 4:
          message.totalWriter = reader.int32();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgCreateTopic {
    const message = { ...baseMsgCreateTopic } as MsgCreateTopic;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.description !== undefined && object.description !== null) {
      message.description = String(object.description);
    } else {
      message.description = "";
    }
    if (object.totalRecords !== undefined && object.totalRecords !== null) {
      message.totalRecords = Number(object.totalRecords);
    } else {
      message.totalRecords = 0;
    }
    if (object.totalWriter !== undefined && object.totalWriter !== null) {
      message.totalWriter = Number(object.totalWriter);
    } else {
      message.totalWriter = 0;
    }
    return message;
  },

  toJSON(message: MsgCreateTopic): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.description !== undefined &&
      (obj.description = message.description);
    message.totalRecords !== undefined &&
      (obj.totalRecords = message.totalRecords);
    message.totalWriter !== undefined &&
      (obj.totalWriter = message.totalWriter);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgCreateTopic>): MsgCreateTopic {
    const message = { ...baseMsgCreateTopic } as MsgCreateTopic;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.description !== undefined && object.description !== null) {
      message.description = object.description;
    } else {
      message.description = "";
    }
    if (object.totalRecords !== undefined && object.totalRecords !== null) {
      message.totalRecords = object.totalRecords;
    } else {
      message.totalRecords = 0;
    }
    if (object.totalWriter !== undefined && object.totalWriter !== null) {
      message.totalWriter = object.totalWriter;
    } else {
      message.totalWriter = 0;
    }
    return message;
  },
};

const baseMsgCreateTopicResponse: object = { id: 0 };

export const MsgCreateTopicResponse = {
  encode(
    message: MsgCreateTopicResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.id !== 0) {
      writer.uint32(8).uint64(message.id);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgCreateTopicResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgCreateTopicResponse } as MsgCreateTopicResponse;
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

  fromJSON(object: any): MsgCreateTopicResponse {
    const message = { ...baseMsgCreateTopicResponse } as MsgCreateTopicResponse;
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    return message;
  },

  toJSON(message: MsgCreateTopicResponse): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgCreateTopicResponse>
  ): MsgCreateTopicResponse {
    const message = { ...baseMsgCreateTopicResponse } as MsgCreateTopicResponse;
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    return message;
  },
};

const baseMsgUpdateTopic: object = {
  creator: "",
  id: 0,
  description: "",
  totalRecords: 0,
  totalWriter: 0,
};

export const MsgUpdateTopic = {
  encode(message: MsgUpdateTopic, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.id !== 0) {
      writer.uint32(16).uint64(message.id);
    }
    if (message.description !== "") {
      writer.uint32(26).string(message.description);
    }
    if (message.totalRecords !== 0) {
      writer.uint32(32).int32(message.totalRecords);
    }
    if (message.totalWriter !== 0) {
      writer.uint32(40).int32(message.totalWriter);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgUpdateTopic {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgUpdateTopic } as MsgUpdateTopic;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        case 3:
          message.description = reader.string();
          break;
        case 4:
          message.totalRecords = reader.int32();
          break;
        case 5:
          message.totalWriter = reader.int32();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgUpdateTopic {
    const message = { ...baseMsgUpdateTopic } as MsgUpdateTopic;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    if (object.description !== undefined && object.description !== null) {
      message.description = String(object.description);
    } else {
      message.description = "";
    }
    if (object.totalRecords !== undefined && object.totalRecords !== null) {
      message.totalRecords = Number(object.totalRecords);
    } else {
      message.totalRecords = 0;
    }
    if (object.totalWriter !== undefined && object.totalWriter !== null) {
      message.totalWriter = Number(object.totalWriter);
    } else {
      message.totalWriter = 0;
    }
    return message;
  },

  toJSON(message: MsgUpdateTopic): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.id !== undefined && (obj.id = message.id);
    message.description !== undefined &&
      (obj.description = message.description);
    message.totalRecords !== undefined &&
      (obj.totalRecords = message.totalRecords);
    message.totalWriter !== undefined &&
      (obj.totalWriter = message.totalWriter);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgUpdateTopic>): MsgUpdateTopic {
    const message = { ...baseMsgUpdateTopic } as MsgUpdateTopic;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    if (object.description !== undefined && object.description !== null) {
      message.description = object.description;
    } else {
      message.description = "";
    }
    if (object.totalRecords !== undefined && object.totalRecords !== null) {
      message.totalRecords = object.totalRecords;
    } else {
      message.totalRecords = 0;
    }
    if (object.totalWriter !== undefined && object.totalWriter !== null) {
      message.totalWriter = object.totalWriter;
    } else {
      message.totalWriter = 0;
    }
    return message;
  },
};

const baseMsgUpdateTopicResponse: object = {};

export const MsgUpdateTopicResponse = {
  encode(_: MsgUpdateTopicResponse, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgUpdateTopicResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgUpdateTopicResponse } as MsgUpdateTopicResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgUpdateTopicResponse {
    const message = { ...baseMsgUpdateTopicResponse } as MsgUpdateTopicResponse;
    return message;
  },

  toJSON(_: MsgUpdateTopicResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<MsgUpdateTopicResponse>): MsgUpdateTopicResponse {
    const message = { ...baseMsgUpdateTopicResponse } as MsgUpdateTopicResponse;
    return message;
  },
};

const baseMsgDeleteTopic: object = { creator: "", id: 0 };

export const MsgDeleteTopic = {
  encode(message: MsgDeleteTopic, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.id !== 0) {
      writer.uint32(16).uint64(message.id);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgDeleteTopic {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgDeleteTopic } as MsgDeleteTopic;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgDeleteTopic {
    const message = { ...baseMsgDeleteTopic } as MsgDeleteTopic;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    return message;
  },

  toJSON(message: MsgDeleteTopic): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgDeleteTopic>): MsgDeleteTopic {
    const message = { ...baseMsgDeleteTopic } as MsgDeleteTopic;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    return message;
  },
};

const baseMsgDeleteTopicResponse: object = {};

export const MsgDeleteTopicResponse = {
  encode(_: MsgDeleteTopicResponse, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgDeleteTopicResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgDeleteTopicResponse } as MsgDeleteTopicResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgDeleteTopicResponse {
    const message = { ...baseMsgDeleteTopicResponse } as MsgDeleteTopicResponse;
    return message;
  },

  toJSON(_: MsgDeleteTopicResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<MsgDeleteTopicResponse>): MsgDeleteTopicResponse {
    const message = { ...baseMsgDeleteTopicResponse } as MsgDeleteTopicResponse;
    return message;
  },
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

export class MsgClientImpl implements Msg {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  CreateOwner(request: MsgCreateOwner): Promise<MsgCreateOwnerResponse> {
    const data = MsgCreateOwner.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Msg",
      "CreateOwner",
      data
    );
    return promise.then((data) =>
      MsgCreateOwnerResponse.decode(new Reader(data))
    );
  }

  UpdateOwner(request: MsgUpdateOwner): Promise<MsgUpdateOwnerResponse> {
    const data = MsgUpdateOwner.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Msg",
      "UpdateOwner",
      data
    );
    return promise.then((data) =>
      MsgUpdateOwnerResponse.decode(new Reader(data))
    );
  }

  DeleteOwner(request: MsgDeleteOwner): Promise<MsgDeleteOwnerResponse> {
    const data = MsgDeleteOwner.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Msg",
      "DeleteOwner",
      data
    );
    return promise.then((data) =>
      MsgDeleteOwnerResponse.decode(new Reader(data))
    );
  }

  CreateRecord(request: MsgCreateRecord): Promise<MsgCreateRecordResponse> {
    const data = MsgCreateRecord.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Msg",
      "CreateRecord",
      data
    );
    return promise.then((data) =>
      MsgCreateRecordResponse.decode(new Reader(data))
    );
  }

  UpdateRecord(request: MsgUpdateRecord): Promise<MsgUpdateRecordResponse> {
    const data = MsgUpdateRecord.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Msg",
      "UpdateRecord",
      data
    );
    return promise.then((data) =>
      MsgUpdateRecordResponse.decode(new Reader(data))
    );
  }

  DeleteRecord(request: MsgDeleteRecord): Promise<MsgDeleteRecordResponse> {
    const data = MsgDeleteRecord.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Msg",
      "DeleteRecord",
      data
    );
    return promise.then((data) =>
      MsgDeleteRecordResponse.decode(new Reader(data))
    );
  }

  CreateWriter(request: MsgCreateWriter): Promise<MsgCreateWriterResponse> {
    const data = MsgCreateWriter.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Msg",
      "CreateWriter",
      data
    );
    return promise.then((data) =>
      MsgCreateWriterResponse.decode(new Reader(data))
    );
  }

  UpdateWriter(request: MsgUpdateWriter): Promise<MsgUpdateWriterResponse> {
    const data = MsgUpdateWriter.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Msg",
      "UpdateWriter",
      data
    );
    return promise.then((data) =>
      MsgUpdateWriterResponse.decode(new Reader(data))
    );
  }

  DeleteWriter(request: MsgDeleteWriter): Promise<MsgDeleteWriterResponse> {
    const data = MsgDeleteWriter.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Msg",
      "DeleteWriter",
      data
    );
    return promise.then((data) =>
      MsgDeleteWriterResponse.decode(new Reader(data))
    );
  }

  CreateTopic(request: MsgCreateTopic): Promise<MsgCreateTopicResponse> {
    const data = MsgCreateTopic.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Msg",
      "CreateTopic",
      data
    );
    return promise.then((data) =>
      MsgCreateTopicResponse.decode(new Reader(data))
    );
  }

  UpdateTopic(request: MsgUpdateTopic): Promise<MsgUpdateTopicResponse> {
    const data = MsgUpdateTopic.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Msg",
      "UpdateTopic",
      data
    );
    return promise.then((data) =>
      MsgUpdateTopicResponse.decode(new Reader(data))
    );
  }

  DeleteTopic(request: MsgDeleteTopic): Promise<MsgDeleteTopicResponse> {
    const data = MsgDeleteTopic.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.aol.Msg",
      "DeleteTopic",
      data
    );
    return promise.then((data) =>
      MsgDeleteTopicResponse.decode(new Reader(data))
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
