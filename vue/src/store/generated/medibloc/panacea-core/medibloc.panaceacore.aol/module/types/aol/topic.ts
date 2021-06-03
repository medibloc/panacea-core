/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "medibloc.panaceacore.aol";

export interface Topic {
  creator: string;
  id: number;
  description: string;
  totalRecords: number;
  totalWriter: number;
}

const baseTopic: object = {
  creator: "",
  id: 0,
  description: "",
  totalRecords: 0,
  totalWriter: 0,
};

export const Topic = {
  encode(message: Topic, writer: Writer = Writer.create()): Writer {
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

  decode(input: Reader | Uint8Array, length?: number): Topic {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseTopic } as Topic;
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

  fromJSON(object: any): Topic {
    const message = { ...baseTopic } as Topic;
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

  toJSON(message: Topic): unknown {
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

  fromPartial(object: DeepPartial<Topic>): Topic {
    const message = { ...baseTopic } as Topic;
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
