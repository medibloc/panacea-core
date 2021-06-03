/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer as Writer1, Reader } from "protobufjs/minimal";
export const protobufPackage = "medibloc.panaceacore.aol";
const baseWriter = {
    creator: "",
    id: 0,
    moniker: "",
    description: "",
    nanoTimestamp: 0,
};
export const Writer = {
    encode(message, writer = Writer1.create()) {
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
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseWriter };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.creator = reader.string();
                    break;
                case 2:
                    message.id = longToNumber(reader.uint64());
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
    fromJSON(object) {
        const message = { ...baseWriter };
        if (object.creator !== undefined && object.creator !== null) {
            message.creator = String(object.creator);
        }
        else {
            message.creator = "";
        }
        if (object.id !== undefined && object.id !== null) {
            message.id = Number(object.id);
        }
        else {
            message.id = 0;
        }
        if (object.moniker !== undefined && object.moniker !== null) {
            message.moniker = String(object.moniker);
        }
        else {
            message.moniker = "";
        }
        if (object.description !== undefined && object.description !== null) {
            message.description = String(object.description);
        }
        else {
            message.description = "";
        }
        if (object.nanoTimestamp !== undefined && object.nanoTimestamp !== null) {
            message.nanoTimestamp = Number(object.nanoTimestamp);
        }
        else {
            message.nanoTimestamp = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.creator !== undefined && (obj.creator = message.creator);
        message.id !== undefined && (obj.id = message.id);
        message.moniker !== undefined && (obj.moniker = message.moniker);
        message.description !== undefined &&
            (obj.description = message.description);
        message.nanoTimestamp !== undefined &&
            (obj.nanoTimestamp = message.nanoTimestamp);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseWriter };
        if (object.creator !== undefined && object.creator !== null) {
            message.creator = object.creator;
        }
        else {
            message.creator = "";
        }
        if (object.id !== undefined && object.id !== null) {
            message.id = object.id;
        }
        else {
            message.id = 0;
        }
        if (object.moniker !== undefined && object.moniker !== null) {
            message.moniker = object.moniker;
        }
        else {
            message.moniker = "";
        }
        if (object.description !== undefined && object.description !== null) {
            message.description = object.description;
        }
        else {
            message.description = "";
        }
        if (object.nanoTimestamp !== undefined && object.nanoTimestamp !== null) {
            message.nanoTimestamp = object.nanoTimestamp;
        }
        else {
            message.nanoTimestamp = 0;
        }
        return message;
    },
};
var globalThis = (() => {
    if (typeof globalThis !== "undefined")
        return globalThis;
    if (typeof self !== "undefined")
        return self;
    if (typeof window !== "undefined")
        return window;
    if (typeof global !== "undefined")
        return global;
    throw "Unable to locate global object";
})();
function longToNumber(long) {
    if (long.gt(Number.MAX_SAFE_INTEGER)) {
        throw new globalThis.Error("Value is larger than Number.MAX_SAFE_INTEGER");
    }
    return long.toNumber();
}
if (util.Long !== Long) {
    util.Long = Long;
    configure();
}
