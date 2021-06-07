/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";

export const protobufPackage = "medibloc.panaceacore.token";

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

export interface MsgUpdateTokenResponse {}

export interface MsgDeleteToken {
  creator: string;
  id: number;
}

export interface MsgDeleteTokenResponse {}

const baseMsgCreateToken: object = {
  creator: "",
  name: "",
  symbol: "",
  totalSupply: 0,
  mintable: false,
  ownerAddress: "",
};

export const MsgCreateToken = {
  encode(message: MsgCreateToken, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.name !== "") {
      writer.uint32(18).string(message.name);
    }
    if (message.symbol !== "") {
      writer.uint32(26).string(message.symbol);
    }
    if (message.totalSupply !== 0) {
      writer.uint32(32).int32(message.totalSupply);
    }
    if (message.mintable === true) {
      writer.uint32(40).bool(message.mintable);
    }
    if (message.ownerAddress !== "") {
      writer.uint32(50).string(message.ownerAddress);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgCreateToken {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgCreateToken } as MsgCreateToken;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.name = reader.string();
          break;
        case 3:
          message.symbol = reader.string();
          break;
        case 4:
          message.totalSupply = reader.int32();
          break;
        case 5:
          message.mintable = reader.bool();
          break;
        case 6:
          message.ownerAddress = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgCreateToken {
    const message = { ...baseMsgCreateToken } as MsgCreateToken;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.symbol !== undefined && object.symbol !== null) {
      message.symbol = String(object.symbol);
    } else {
      message.symbol = "";
    }
    if (object.totalSupply !== undefined && object.totalSupply !== null) {
      message.totalSupply = Number(object.totalSupply);
    } else {
      message.totalSupply = 0;
    }
    if (object.mintable !== undefined && object.mintable !== null) {
      message.mintable = Boolean(object.mintable);
    } else {
      message.mintable = false;
    }
    if (object.ownerAddress !== undefined && object.ownerAddress !== null) {
      message.ownerAddress = String(object.ownerAddress);
    } else {
      message.ownerAddress = "";
    }
    return message;
  },

  toJSON(message: MsgCreateToken): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.name !== undefined && (obj.name = message.name);
    message.symbol !== undefined && (obj.symbol = message.symbol);
    message.totalSupply !== undefined &&
      (obj.totalSupply = message.totalSupply);
    message.mintable !== undefined && (obj.mintable = message.mintable);
    message.ownerAddress !== undefined &&
      (obj.ownerAddress = message.ownerAddress);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgCreateToken>): MsgCreateToken {
    const message = { ...baseMsgCreateToken } as MsgCreateToken;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.symbol !== undefined && object.symbol !== null) {
      message.symbol = object.symbol;
    } else {
      message.symbol = "";
    }
    if (object.totalSupply !== undefined && object.totalSupply !== null) {
      message.totalSupply = object.totalSupply;
    } else {
      message.totalSupply = 0;
    }
    if (object.mintable !== undefined && object.mintable !== null) {
      message.mintable = object.mintable;
    } else {
      message.mintable = false;
    }
    if (object.ownerAddress !== undefined && object.ownerAddress !== null) {
      message.ownerAddress = object.ownerAddress;
    } else {
      message.ownerAddress = "";
    }
    return message;
  },
};

const baseMsgCreateTokenResponse: object = { id: 0 };

export const MsgCreateTokenResponse = {
  encode(
    message: MsgCreateTokenResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.id !== 0) {
      writer.uint32(8).uint64(message.id);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgCreateTokenResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgCreateTokenResponse } as MsgCreateTokenResponse;
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

  fromJSON(object: any): MsgCreateTokenResponse {
    const message = { ...baseMsgCreateTokenResponse } as MsgCreateTokenResponse;
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    return message;
  },

  toJSON(message: MsgCreateTokenResponse): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgCreateTokenResponse>
  ): MsgCreateTokenResponse {
    const message = { ...baseMsgCreateTokenResponse } as MsgCreateTokenResponse;
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    return message;
  },
};

const baseMsgUpdateToken: object = {
  creator: "",
  id: 0,
  name: "",
  symbol: "",
  totalSupply: 0,
  mintable: false,
  ownerAddress: "",
};

export const MsgUpdateToken = {
  encode(message: MsgUpdateToken, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.id !== 0) {
      writer.uint32(16).uint64(message.id);
    }
    if (message.name !== "") {
      writer.uint32(26).string(message.name);
    }
    if (message.symbol !== "") {
      writer.uint32(34).string(message.symbol);
    }
    if (message.totalSupply !== 0) {
      writer.uint32(40).int32(message.totalSupply);
    }
    if (message.mintable === true) {
      writer.uint32(48).bool(message.mintable);
    }
    if (message.ownerAddress !== "") {
      writer.uint32(58).string(message.ownerAddress);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgUpdateToken {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgUpdateToken } as MsgUpdateToken;
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
          message.name = reader.string();
          break;
        case 4:
          message.symbol = reader.string();
          break;
        case 5:
          message.totalSupply = reader.int32();
          break;
        case 6:
          message.mintable = reader.bool();
          break;
        case 7:
          message.ownerAddress = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgUpdateToken {
    const message = { ...baseMsgUpdateToken } as MsgUpdateToken;
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
    if (object.name !== undefined && object.name !== null) {
      message.name = String(object.name);
    } else {
      message.name = "";
    }
    if (object.symbol !== undefined && object.symbol !== null) {
      message.symbol = String(object.symbol);
    } else {
      message.symbol = "";
    }
    if (object.totalSupply !== undefined && object.totalSupply !== null) {
      message.totalSupply = Number(object.totalSupply);
    } else {
      message.totalSupply = 0;
    }
    if (object.mintable !== undefined && object.mintable !== null) {
      message.mintable = Boolean(object.mintable);
    } else {
      message.mintable = false;
    }
    if (object.ownerAddress !== undefined && object.ownerAddress !== null) {
      message.ownerAddress = String(object.ownerAddress);
    } else {
      message.ownerAddress = "";
    }
    return message;
  },

  toJSON(message: MsgUpdateToken): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.id !== undefined && (obj.id = message.id);
    message.name !== undefined && (obj.name = message.name);
    message.symbol !== undefined && (obj.symbol = message.symbol);
    message.totalSupply !== undefined &&
      (obj.totalSupply = message.totalSupply);
    message.mintable !== undefined && (obj.mintable = message.mintable);
    message.ownerAddress !== undefined &&
      (obj.ownerAddress = message.ownerAddress);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgUpdateToken>): MsgUpdateToken {
    const message = { ...baseMsgUpdateToken } as MsgUpdateToken;
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
    if (object.name !== undefined && object.name !== null) {
      message.name = object.name;
    } else {
      message.name = "";
    }
    if (object.symbol !== undefined && object.symbol !== null) {
      message.symbol = object.symbol;
    } else {
      message.symbol = "";
    }
    if (object.totalSupply !== undefined && object.totalSupply !== null) {
      message.totalSupply = object.totalSupply;
    } else {
      message.totalSupply = 0;
    }
    if (object.mintable !== undefined && object.mintable !== null) {
      message.mintable = object.mintable;
    } else {
      message.mintable = false;
    }
    if (object.ownerAddress !== undefined && object.ownerAddress !== null) {
      message.ownerAddress = object.ownerAddress;
    } else {
      message.ownerAddress = "";
    }
    return message;
  },
};

const baseMsgUpdateTokenResponse: object = {};

export const MsgUpdateTokenResponse = {
  encode(_: MsgUpdateTokenResponse, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgUpdateTokenResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgUpdateTokenResponse } as MsgUpdateTokenResponse;
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

  fromJSON(_: any): MsgUpdateTokenResponse {
    const message = { ...baseMsgUpdateTokenResponse } as MsgUpdateTokenResponse;
    return message;
  },

  toJSON(_: MsgUpdateTokenResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<MsgUpdateTokenResponse>): MsgUpdateTokenResponse {
    const message = { ...baseMsgUpdateTokenResponse } as MsgUpdateTokenResponse;
    return message;
  },
};

const baseMsgDeleteToken: object = { creator: "", id: 0 };

export const MsgDeleteToken = {
  encode(message: MsgDeleteToken, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.id !== 0) {
      writer.uint32(16).uint64(message.id);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgDeleteToken {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgDeleteToken } as MsgDeleteToken;
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

  fromJSON(object: any): MsgDeleteToken {
    const message = { ...baseMsgDeleteToken } as MsgDeleteToken;
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

  toJSON(message: MsgDeleteToken): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgDeleteToken>): MsgDeleteToken {
    const message = { ...baseMsgDeleteToken } as MsgDeleteToken;
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

const baseMsgDeleteTokenResponse: object = {};

export const MsgDeleteTokenResponse = {
  encode(_: MsgDeleteTokenResponse, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgDeleteTokenResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgDeleteTokenResponse } as MsgDeleteTokenResponse;
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

  fromJSON(_: any): MsgDeleteTokenResponse {
    const message = { ...baseMsgDeleteTokenResponse } as MsgDeleteTokenResponse;
    return message;
  },

  toJSON(_: MsgDeleteTokenResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<MsgDeleteTokenResponse>): MsgDeleteTokenResponse {
    const message = { ...baseMsgDeleteTokenResponse } as MsgDeleteTokenResponse;
    return message;
  },
};

/** Msg defines the Msg service. */
export interface Msg {
  /** this line is used by starport scaffolding # proto/tx/rpc */
  CreateToken(request: MsgCreateToken): Promise<MsgCreateTokenResponse>;
  UpdateToken(request: MsgUpdateToken): Promise<MsgUpdateTokenResponse>;
  DeleteToken(request: MsgDeleteToken): Promise<MsgDeleteTokenResponse>;
}

export class MsgClientImpl implements Msg {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  CreateToken(request: MsgCreateToken): Promise<MsgCreateTokenResponse> {
    const data = MsgCreateToken.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.token.Msg",
      "CreateToken",
      data
    );
    return promise.then((data) =>
      MsgCreateTokenResponse.decode(new Reader(data))
    );
  }

  UpdateToken(request: MsgUpdateToken): Promise<MsgUpdateTokenResponse> {
    const data = MsgUpdateToken.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.token.Msg",
      "UpdateToken",
      data
    );
    return promise.then((data) =>
      MsgUpdateTokenResponse.decode(new Reader(data))
    );
  }

  DeleteToken(request: MsgDeleteToken): Promise<MsgDeleteTokenResponse> {
    const data = MsgDeleteToken.encode(request).finish();
    const promise = this.rpc.request(
      "medibloc.panaceacore.token.Msg",
      "DeleteToken",
      data
    );
    return promise.then((data) =>
      MsgDeleteTokenResponse.decode(new Reader(data))
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
