/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "medibloc.panaceacore.aol";

export interface Record {
  creator: string;
  id: number;
  key: string;
  value: string;
  nanoTimestamp: number;
  writerAddress: string;
}

const baseRecord: object = {
  creator: "",
  id: 0,
  key: "",
  value: "",
  nanoTimestamp: 0,
  writerAddress: "",
};

export const Record = {
  encode(message: Record, writer: Writer = Writer.create()): Writer {
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

  decode(input: Reader | Uint8Array, length?: number): Record {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseRecord } as Record;
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

  fromJSON(object: any): Record {
    const message = { ...baseRecord } as Record;
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

  toJSON(message: Record): unknown {
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

  fromPartial(object: DeepPartial<Record>): Record {
    const message = { ...baseRecord } as Record;
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
