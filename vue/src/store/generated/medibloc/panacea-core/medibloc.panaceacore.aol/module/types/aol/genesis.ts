/* eslint-disable */
import { Owner } from "../aol/owner";
import { Record } from "../aol/record";
import { Writer } from "../aol/writer";
import { Topic } from "../aol/topic";
import { Writer as Writer1, Reader } from "protobufjs/minimal";

export const protobufPackage = "medibloc.panaceacore.aol";

/** GenesisState defines the aol module's genesis state. */
export interface GenesisState {
  /** this line is used by starport scaffolding # genesis/proto/state */
  ownerList: Owner[];
  /** this line is used by starport scaffolding # genesis/proto/stateField */
  recordList: Record[];
  /** this line is used by starport scaffolding # genesis/proto/stateField */
  writerList: Writer[];
  /** this line is used by starport scaffolding # genesis/proto/stateField */
  topicList: Topic[];
}

const baseGenesisState: object = {};

export const GenesisState = {
  encode(message: GenesisState, writer: Writer1 = Writer1.create()): Writer1 {
    for (const v of message.ownerList) {
      Owner.encode(v!, writer.uint32(34).fork()).ldelim();
    }
    for (const v of message.recordList) {
      Record.encode(v!, writer.uint32(26).fork()).ldelim();
    }
    for (const v of message.writerList) {
      Writer.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    for (const v of message.topicList) {
      Topic.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): GenesisState {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseGenesisState } as GenesisState;
    message.ownerList = [];
    message.recordList = [];
    message.writerList = [];
    message.topicList = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 4:
          message.ownerList.push(Owner.decode(reader, reader.uint32()));
          break;
        case 3:
          message.recordList.push(Record.decode(reader, reader.uint32()));
          break;
        case 2:
          message.writerList.push(Writer.decode(reader, reader.uint32()));
          break;
        case 1:
          message.topicList.push(Topic.decode(reader, reader.uint32()));
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GenesisState {
    const message = { ...baseGenesisState } as GenesisState;
    message.ownerList = [];
    message.recordList = [];
    message.writerList = [];
    message.topicList = [];
    if (object.ownerList !== undefined && object.ownerList !== null) {
      for (const e of object.ownerList) {
        message.ownerList.push(Owner.fromJSON(e));
      }
    }
    if (object.recordList !== undefined && object.recordList !== null) {
      for (const e of object.recordList) {
        message.recordList.push(Record.fromJSON(e));
      }
    }
    if (object.writerList !== undefined && object.writerList !== null) {
      for (const e of object.writerList) {
        message.writerList.push(Writer.fromJSON(e));
      }
    }
    if (object.topicList !== undefined && object.topicList !== null) {
      for (const e of object.topicList) {
        message.topicList.push(Topic.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: GenesisState): unknown {
    const obj: any = {};
    if (message.ownerList) {
      obj.ownerList = message.ownerList.map((e) =>
        e ? Owner.toJSON(e) : undefined
      );
    } else {
      obj.ownerList = [];
    }
    if (message.recordList) {
      obj.recordList = message.recordList.map((e) =>
        e ? Record.toJSON(e) : undefined
      );
    } else {
      obj.recordList = [];
    }
    if (message.writerList) {
      obj.writerList = message.writerList.map((e) =>
        e ? Writer.toJSON(e) : undefined
      );
    } else {
      obj.writerList = [];
    }
    if (message.topicList) {
      obj.topicList = message.topicList.map((e) =>
        e ? Topic.toJSON(e) : undefined
      );
    } else {
      obj.topicList = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<GenesisState>): GenesisState {
    const message = { ...baseGenesisState } as GenesisState;
    message.ownerList = [];
    message.recordList = [];
    message.writerList = [];
    message.topicList = [];
    if (object.ownerList !== undefined && object.ownerList !== null) {
      for (const e of object.ownerList) {
        message.ownerList.push(Owner.fromPartial(e));
      }
    }
    if (object.recordList !== undefined && object.recordList !== null) {
      for (const e of object.recordList) {
        message.recordList.push(Record.fromPartial(e));
      }
    }
    if (object.writerList !== undefined && object.writerList !== null) {
      for (const e of object.writerList) {
        message.writerList.push(Writer.fromPartial(e));
      }
    }
    if (object.topicList !== undefined && object.topicList !== null) {
      for (const e of object.topicList) {
        message.topicList.push(Topic.fromPartial(e));
      }
    }
    return message;
  },
};

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
