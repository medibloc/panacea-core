/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";
export const protobufPackage = "medibloc.panaceacore.aol";
const baseMsgCreateOwner = { creator: "", totalTopics: 0 };
export const MsgCreateOwner = {
    encode(message, writer = Writer.create()) {
        if (message.creator !== "") {
            writer.uint32(10).string(message.creator);
        }
        if (message.totalTopics !== 0) {
            writer.uint32(16).int32(message.totalTopics);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgCreateOwner };
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
    fromJSON(object) {
        const message = { ...baseMsgCreateOwner };
        if (object.creator !== undefined && object.creator !== null) {
            message.creator = String(object.creator);
        }
        else {
            message.creator = "";
        }
        if (object.totalTopics !== undefined && object.totalTopics !== null) {
            message.totalTopics = Number(object.totalTopics);
        }
        else {
            message.totalTopics = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.creator !== undefined && (obj.creator = message.creator);
        message.totalTopics !== undefined &&
            (obj.totalTopics = message.totalTopics);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgCreateOwner };
        if (object.creator !== undefined && object.creator !== null) {
            message.creator = object.creator;
        }
        else {
            message.creator = "";
        }
        if (object.totalTopics !== undefined && object.totalTopics !== null) {
            message.totalTopics = object.totalTopics;
        }
        else {
            message.totalTopics = 0;
        }
        return message;
    },
};
const baseMsgCreateOwnerResponse = { id: 0 };
export const MsgCreateOwnerResponse = {
    encode(message, writer = Writer.create()) {
        if (message.id !== 0) {
            writer.uint32(8).uint64(message.id);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgCreateOwnerResponse };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.id = longToNumber(reader.uint64());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseMsgCreateOwnerResponse };
        if (object.id !== undefined && object.id !== null) {
            message.id = Number(object.id);
        }
        else {
            message.id = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.id !== undefined && (obj.id = message.id);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgCreateOwnerResponse };
        if (object.id !== undefined && object.id !== null) {
            message.id = object.id;
        }
        else {
            message.id = 0;
        }
        return message;
    },
};
const baseMsgUpdateOwner = { creator: "", id: 0, totalTopics: 0 };
export const MsgUpdateOwner = {
    encode(message, writer = Writer.create()) {
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
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgUpdateOwner };
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
                    message.totalTopics = reader.int32();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseMsgUpdateOwner };
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
        if (object.totalTopics !== undefined && object.totalTopics !== null) {
            message.totalTopics = Number(object.totalTopics);
        }
        else {
            message.totalTopics = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.creator !== undefined && (obj.creator = message.creator);
        message.id !== undefined && (obj.id = message.id);
        message.totalTopics !== undefined &&
            (obj.totalTopics = message.totalTopics);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgUpdateOwner };
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
        if (object.totalTopics !== undefined && object.totalTopics !== null) {
            message.totalTopics = object.totalTopics;
        }
        else {
            message.totalTopics = 0;
        }
        return message;
    },
};
const baseMsgUpdateOwnerResponse = {};
export const MsgUpdateOwnerResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgUpdateOwnerResponse };
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
    fromJSON(_) {
        const message = { ...baseMsgUpdateOwnerResponse };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = { ...baseMsgUpdateOwnerResponse };
        return message;
    },
};
const baseMsgDeleteOwner = { creator: "", id: 0 };
export const MsgDeleteOwner = {
    encode(message, writer = Writer.create()) {
        if (message.creator !== "") {
            writer.uint32(10).string(message.creator);
        }
        if (message.id !== 0) {
            writer.uint32(16).uint64(message.id);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgDeleteOwner };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.creator = reader.string();
                    break;
                case 2:
                    message.id = longToNumber(reader.uint64());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseMsgDeleteOwner };
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
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.creator !== undefined && (obj.creator = message.creator);
        message.id !== undefined && (obj.id = message.id);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgDeleteOwner };
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
        return message;
    },
};
const baseMsgDeleteOwnerResponse = {};
export const MsgDeleteOwnerResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgDeleteOwnerResponse };
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
    fromJSON(_) {
        const message = { ...baseMsgDeleteOwnerResponse };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = { ...baseMsgDeleteOwnerResponse };
        return message;
    },
};
const baseMsgCreateRecord = {
    creator: "",
    key: "",
    value: "",
    nanoTimestamp: 0,
    writerAddress: "",
};
export const MsgCreateRecord = {
    encode(message, writer = Writer.create()) {
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
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgCreateRecord };
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
    fromJSON(object) {
        const message = { ...baseMsgCreateRecord };
        if (object.creator !== undefined && object.creator !== null) {
            message.creator = String(object.creator);
        }
        else {
            message.creator = "";
        }
        if (object.key !== undefined && object.key !== null) {
            message.key = String(object.key);
        }
        else {
            message.key = "";
        }
        if (object.value !== undefined && object.value !== null) {
            message.value = String(object.value);
        }
        else {
            message.value = "";
        }
        if (object.nanoTimestamp !== undefined && object.nanoTimestamp !== null) {
            message.nanoTimestamp = Number(object.nanoTimestamp);
        }
        else {
            message.nanoTimestamp = 0;
        }
        if (object.writerAddress !== undefined && object.writerAddress !== null) {
            message.writerAddress = String(object.writerAddress);
        }
        else {
            message.writerAddress = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.creator !== undefined && (obj.creator = message.creator);
        message.key !== undefined && (obj.key = message.key);
        message.value !== undefined && (obj.value = message.value);
        message.nanoTimestamp !== undefined &&
            (obj.nanoTimestamp = message.nanoTimestamp);
        message.writerAddress !== undefined &&
            (obj.writerAddress = message.writerAddress);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgCreateRecord };
        if (object.creator !== undefined && object.creator !== null) {
            message.creator = object.creator;
        }
        else {
            message.creator = "";
        }
        if (object.key !== undefined && object.key !== null) {
            message.key = object.key;
        }
        else {
            message.key = "";
        }
        if (object.value !== undefined && object.value !== null) {
            message.value = object.value;
        }
        else {
            message.value = "";
        }
        if (object.nanoTimestamp !== undefined && object.nanoTimestamp !== null) {
            message.nanoTimestamp = object.nanoTimestamp;
        }
        else {
            message.nanoTimestamp = 0;
        }
        if (object.writerAddress !== undefined && object.writerAddress !== null) {
            message.writerAddress = object.writerAddress;
        }
        else {
            message.writerAddress = "";
        }
        return message;
    },
};
const baseMsgCreateRecordResponse = { id: 0 };
export const MsgCreateRecordResponse = {
    encode(message, writer = Writer.create()) {
        if (message.id !== 0) {
            writer.uint32(8).uint64(message.id);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseMsgCreateRecordResponse,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.id = longToNumber(reader.uint64());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = {
            ...baseMsgCreateRecordResponse,
        };
        if (object.id !== undefined && object.id !== null) {
            message.id = Number(object.id);
        }
        else {
            message.id = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.id !== undefined && (obj.id = message.id);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseMsgCreateRecordResponse,
        };
        if (object.id !== undefined && object.id !== null) {
            message.id = object.id;
        }
        else {
            message.id = 0;
        }
        return message;
    },
};
const baseMsgUpdateRecord = {
    creator: "",
    id: 0,
    key: "",
    value: "",
    nanoTimestamp: 0,
    writerAddress: "",
};
export const MsgUpdateRecord = {
    encode(message, writer = Writer.create()) {
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
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgUpdateRecord };
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
    fromJSON(object) {
        const message = { ...baseMsgUpdateRecord };
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
        if (object.key !== undefined && object.key !== null) {
            message.key = String(object.key);
        }
        else {
            message.key = "";
        }
        if (object.value !== undefined && object.value !== null) {
            message.value = String(object.value);
        }
        else {
            message.value = "";
        }
        if (object.nanoTimestamp !== undefined && object.nanoTimestamp !== null) {
            message.nanoTimestamp = Number(object.nanoTimestamp);
        }
        else {
            message.nanoTimestamp = 0;
        }
        if (object.writerAddress !== undefined && object.writerAddress !== null) {
            message.writerAddress = String(object.writerAddress);
        }
        else {
            message.writerAddress = "";
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
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
    fromPartial(object) {
        const message = { ...baseMsgUpdateRecord };
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
        if (object.key !== undefined && object.key !== null) {
            message.key = object.key;
        }
        else {
            message.key = "";
        }
        if (object.value !== undefined && object.value !== null) {
            message.value = object.value;
        }
        else {
            message.value = "";
        }
        if (object.nanoTimestamp !== undefined && object.nanoTimestamp !== null) {
            message.nanoTimestamp = object.nanoTimestamp;
        }
        else {
            message.nanoTimestamp = 0;
        }
        if (object.writerAddress !== undefined && object.writerAddress !== null) {
            message.writerAddress = object.writerAddress;
        }
        else {
            message.writerAddress = "";
        }
        return message;
    },
};
const baseMsgUpdateRecordResponse = {};
export const MsgUpdateRecordResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseMsgUpdateRecordResponse,
        };
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
    fromJSON(_) {
        const message = {
            ...baseMsgUpdateRecordResponse,
        };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = {
            ...baseMsgUpdateRecordResponse,
        };
        return message;
    },
};
const baseMsgDeleteRecord = { creator: "", id: 0 };
export const MsgDeleteRecord = {
    encode(message, writer = Writer.create()) {
        if (message.creator !== "") {
            writer.uint32(10).string(message.creator);
        }
        if (message.id !== 0) {
            writer.uint32(16).uint64(message.id);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgDeleteRecord };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.creator = reader.string();
                    break;
                case 2:
                    message.id = longToNumber(reader.uint64());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseMsgDeleteRecord };
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
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.creator !== undefined && (obj.creator = message.creator);
        message.id !== undefined && (obj.id = message.id);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgDeleteRecord };
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
        return message;
    },
};
const baseMsgDeleteRecordResponse = {};
export const MsgDeleteRecordResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseMsgDeleteRecordResponse,
        };
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
    fromJSON(_) {
        const message = {
            ...baseMsgDeleteRecordResponse,
        };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = {
            ...baseMsgDeleteRecordResponse,
        };
        return message;
    },
};
const baseMsgCreateWriter = {
    creator: "",
    moniker: "",
    description: "",
    nanoTimestamp: 0,
};
export const MsgCreateWriter = {
    encode(message, writer = Writer.create()) {
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
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgCreateWriter };
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
    fromJSON(object) {
        const message = { ...baseMsgCreateWriter };
        if (object.creator !== undefined && object.creator !== null) {
            message.creator = String(object.creator);
        }
        else {
            message.creator = "";
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
        message.moniker !== undefined && (obj.moniker = message.moniker);
        message.description !== undefined &&
            (obj.description = message.description);
        message.nanoTimestamp !== undefined &&
            (obj.nanoTimestamp = message.nanoTimestamp);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgCreateWriter };
        if (object.creator !== undefined && object.creator !== null) {
            message.creator = object.creator;
        }
        else {
            message.creator = "";
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
const baseMsgCreateWriterResponse = { id: 0 };
export const MsgCreateWriterResponse = {
    encode(message, writer = Writer.create()) {
        if (message.id !== 0) {
            writer.uint32(8).uint64(message.id);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseMsgCreateWriterResponse,
        };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.id = longToNumber(reader.uint64());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = {
            ...baseMsgCreateWriterResponse,
        };
        if (object.id !== undefined && object.id !== null) {
            message.id = Number(object.id);
        }
        else {
            message.id = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.id !== undefined && (obj.id = message.id);
        return obj;
    },
    fromPartial(object) {
        const message = {
            ...baseMsgCreateWriterResponse,
        };
        if (object.id !== undefined && object.id !== null) {
            message.id = object.id;
        }
        else {
            message.id = 0;
        }
        return message;
    },
};
const baseMsgUpdateWriter = {
    creator: "",
    id: 0,
    moniker: "",
    description: "",
    nanoTimestamp: 0,
};
export const MsgUpdateWriter = {
    encode(message, writer = Writer.create()) {
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
        const message = { ...baseMsgUpdateWriter };
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
        const message = { ...baseMsgUpdateWriter };
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
        const message = { ...baseMsgUpdateWriter };
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
const baseMsgUpdateWriterResponse = {};
export const MsgUpdateWriterResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseMsgUpdateWriterResponse,
        };
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
    fromJSON(_) {
        const message = {
            ...baseMsgUpdateWriterResponse,
        };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = {
            ...baseMsgUpdateWriterResponse,
        };
        return message;
    },
};
const baseMsgDeleteWriter = { creator: "", id: 0 };
export const MsgDeleteWriter = {
    encode(message, writer = Writer.create()) {
        if (message.creator !== "") {
            writer.uint32(10).string(message.creator);
        }
        if (message.id !== 0) {
            writer.uint32(16).uint64(message.id);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgDeleteWriter };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.creator = reader.string();
                    break;
                case 2:
                    message.id = longToNumber(reader.uint64());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseMsgDeleteWriter };
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
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.creator !== undefined && (obj.creator = message.creator);
        message.id !== undefined && (obj.id = message.id);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgDeleteWriter };
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
        return message;
    },
};
const baseMsgDeleteWriterResponse = {};
export const MsgDeleteWriterResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = {
            ...baseMsgDeleteWriterResponse,
        };
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
    fromJSON(_) {
        const message = {
            ...baseMsgDeleteWriterResponse,
        };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = {
            ...baseMsgDeleteWriterResponse,
        };
        return message;
    },
};
const baseMsgCreateTopic = {
    creator: "",
    description: "",
    totalRecords: 0,
    totalWriter: 0,
};
export const MsgCreateTopic = {
    encode(message, writer = Writer.create()) {
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
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgCreateTopic };
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
    fromJSON(object) {
        const message = { ...baseMsgCreateTopic };
        if (object.creator !== undefined && object.creator !== null) {
            message.creator = String(object.creator);
        }
        else {
            message.creator = "";
        }
        if (object.description !== undefined && object.description !== null) {
            message.description = String(object.description);
        }
        else {
            message.description = "";
        }
        if (object.totalRecords !== undefined && object.totalRecords !== null) {
            message.totalRecords = Number(object.totalRecords);
        }
        else {
            message.totalRecords = 0;
        }
        if (object.totalWriter !== undefined && object.totalWriter !== null) {
            message.totalWriter = Number(object.totalWriter);
        }
        else {
            message.totalWriter = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.creator !== undefined && (obj.creator = message.creator);
        message.description !== undefined &&
            (obj.description = message.description);
        message.totalRecords !== undefined &&
            (obj.totalRecords = message.totalRecords);
        message.totalWriter !== undefined &&
            (obj.totalWriter = message.totalWriter);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgCreateTopic };
        if (object.creator !== undefined && object.creator !== null) {
            message.creator = object.creator;
        }
        else {
            message.creator = "";
        }
        if (object.description !== undefined && object.description !== null) {
            message.description = object.description;
        }
        else {
            message.description = "";
        }
        if (object.totalRecords !== undefined && object.totalRecords !== null) {
            message.totalRecords = object.totalRecords;
        }
        else {
            message.totalRecords = 0;
        }
        if (object.totalWriter !== undefined && object.totalWriter !== null) {
            message.totalWriter = object.totalWriter;
        }
        else {
            message.totalWriter = 0;
        }
        return message;
    },
};
const baseMsgCreateTopicResponse = { id: 0 };
export const MsgCreateTopicResponse = {
    encode(message, writer = Writer.create()) {
        if (message.id !== 0) {
            writer.uint32(8).uint64(message.id);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgCreateTopicResponse };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.id = longToNumber(reader.uint64());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseMsgCreateTopicResponse };
        if (object.id !== undefined && object.id !== null) {
            message.id = Number(object.id);
        }
        else {
            message.id = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.id !== undefined && (obj.id = message.id);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgCreateTopicResponse };
        if (object.id !== undefined && object.id !== null) {
            message.id = object.id;
        }
        else {
            message.id = 0;
        }
        return message;
    },
};
const baseMsgUpdateTopic = {
    creator: "",
    id: 0,
    description: "",
    totalRecords: 0,
    totalWriter: 0,
};
export const MsgUpdateTopic = {
    encode(message, writer = Writer.create()) {
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
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgUpdateTopic };
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
    fromJSON(object) {
        const message = { ...baseMsgUpdateTopic };
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
        if (object.description !== undefined && object.description !== null) {
            message.description = String(object.description);
        }
        else {
            message.description = "";
        }
        if (object.totalRecords !== undefined && object.totalRecords !== null) {
            message.totalRecords = Number(object.totalRecords);
        }
        else {
            message.totalRecords = 0;
        }
        if (object.totalWriter !== undefined && object.totalWriter !== null) {
            message.totalWriter = Number(object.totalWriter);
        }
        else {
            message.totalWriter = 0;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
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
    fromPartial(object) {
        const message = { ...baseMsgUpdateTopic };
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
        if (object.description !== undefined && object.description !== null) {
            message.description = object.description;
        }
        else {
            message.description = "";
        }
        if (object.totalRecords !== undefined && object.totalRecords !== null) {
            message.totalRecords = object.totalRecords;
        }
        else {
            message.totalRecords = 0;
        }
        if (object.totalWriter !== undefined && object.totalWriter !== null) {
            message.totalWriter = object.totalWriter;
        }
        else {
            message.totalWriter = 0;
        }
        return message;
    },
};
const baseMsgUpdateTopicResponse = {};
export const MsgUpdateTopicResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgUpdateTopicResponse };
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
    fromJSON(_) {
        const message = { ...baseMsgUpdateTopicResponse };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = { ...baseMsgUpdateTopicResponse };
        return message;
    },
};
const baseMsgDeleteTopic = { creator: "", id: 0 };
export const MsgDeleteTopic = {
    encode(message, writer = Writer.create()) {
        if (message.creator !== "") {
            writer.uint32(10).string(message.creator);
        }
        if (message.id !== 0) {
            writer.uint32(16).uint64(message.id);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgDeleteTopic };
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.creator = reader.string();
                    break;
                case 2:
                    message.id = longToNumber(reader.uint64());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseMsgDeleteTopic };
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
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.creator !== undefined && (obj.creator = message.creator);
        message.id !== undefined && (obj.id = message.id);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseMsgDeleteTopic };
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
        return message;
    },
};
const baseMsgDeleteTopicResponse = {};
export const MsgDeleteTopicResponse = {
    encode(_, writer = Writer.create()) {
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseMsgDeleteTopicResponse };
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
    fromJSON(_) {
        const message = { ...baseMsgDeleteTopicResponse };
        return message;
    },
    toJSON(_) {
        const obj = {};
        return obj;
    },
    fromPartial(_) {
        const message = { ...baseMsgDeleteTopicResponse };
        return message;
    },
};
export class MsgClientImpl {
    constructor(rpc) {
        this.rpc = rpc;
    }
    CreateOwner(request) {
        const data = MsgCreateOwner.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Msg", "CreateOwner", data);
        return promise.then((data) => MsgCreateOwnerResponse.decode(new Reader(data)));
    }
    UpdateOwner(request) {
        const data = MsgUpdateOwner.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Msg", "UpdateOwner", data);
        return promise.then((data) => MsgUpdateOwnerResponse.decode(new Reader(data)));
    }
    DeleteOwner(request) {
        const data = MsgDeleteOwner.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Msg", "DeleteOwner", data);
        return promise.then((data) => MsgDeleteOwnerResponse.decode(new Reader(data)));
    }
    CreateRecord(request) {
        const data = MsgCreateRecord.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Msg", "CreateRecord", data);
        return promise.then((data) => MsgCreateRecordResponse.decode(new Reader(data)));
    }
    UpdateRecord(request) {
        const data = MsgUpdateRecord.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Msg", "UpdateRecord", data);
        return promise.then((data) => MsgUpdateRecordResponse.decode(new Reader(data)));
    }
    DeleteRecord(request) {
        const data = MsgDeleteRecord.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Msg", "DeleteRecord", data);
        return promise.then((data) => MsgDeleteRecordResponse.decode(new Reader(data)));
    }
    CreateWriter(request) {
        const data = MsgCreateWriter.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Msg", "CreateWriter", data);
        return promise.then((data) => MsgCreateWriterResponse.decode(new Reader(data)));
    }
    UpdateWriter(request) {
        const data = MsgUpdateWriter.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Msg", "UpdateWriter", data);
        return promise.then((data) => MsgUpdateWriterResponse.decode(new Reader(data)));
    }
    DeleteWriter(request) {
        const data = MsgDeleteWriter.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Msg", "DeleteWriter", data);
        return promise.then((data) => MsgDeleteWriterResponse.decode(new Reader(data)));
    }
    CreateTopic(request) {
        const data = MsgCreateTopic.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Msg", "CreateTopic", data);
        return promise.then((data) => MsgCreateTopicResponse.decode(new Reader(data)));
    }
    UpdateTopic(request) {
        const data = MsgUpdateTopic.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Msg", "UpdateTopic", data);
        return promise.then((data) => MsgUpdateTopicResponse.decode(new Reader(data)));
    }
    DeleteTopic(request) {
        const data = MsgDeleteTopic.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Msg", "DeleteTopic", data);
        return promise.then((data) => MsgDeleteTopicResponse.decode(new Reader(data)));
    }
}
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
