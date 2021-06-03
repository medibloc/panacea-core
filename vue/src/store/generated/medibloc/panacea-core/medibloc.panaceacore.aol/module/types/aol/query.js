/* eslint-disable */
import { Reader, util, configure, Writer as Writer1 } from "protobufjs/minimal";
import * as Long from "long";
import { Owner } from "../aol/owner";
import { PageRequest, PageResponse, } from "../cosmos/base/query/v1beta1/pagination";
import { Record } from "../aol/record";
import { Writer } from "../aol/writer";
import { Topic } from "../aol/topic";
export const protobufPackage = "medibloc.panaceacore.aol";
const baseQueryGetOwnerRequest = { id: 0 };
export const QueryGetOwnerRequest = {
    encode(message, writer = Writer1.create()) {
        if (message.id !== 0) {
            writer.uint32(8).uint64(message.id);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryGetOwnerRequest };
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
        const message = { ...baseQueryGetOwnerRequest };
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
        const message = { ...baseQueryGetOwnerRequest };
        if (object.id !== undefined && object.id !== null) {
            message.id = object.id;
        }
        else {
            message.id = 0;
        }
        return message;
    },
};
const baseQueryGetOwnerResponse = {};
export const QueryGetOwnerResponse = {
    encode(message, writer = Writer1.create()) {
        if (message.Owner !== undefined) {
            Owner.encode(message.Owner, writer.uint32(10).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryGetOwnerResponse };
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
    fromJSON(object) {
        const message = { ...baseQueryGetOwnerResponse };
        if (object.Owner !== undefined && object.Owner !== null) {
            message.Owner = Owner.fromJSON(object.Owner);
        }
        else {
            message.Owner = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.Owner !== undefined &&
            (obj.Owner = message.Owner ? Owner.toJSON(message.Owner) : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryGetOwnerResponse };
        if (object.Owner !== undefined && object.Owner !== null) {
            message.Owner = Owner.fromPartial(object.Owner);
        }
        else {
            message.Owner = undefined;
        }
        return message;
    },
};
const baseQueryAllOwnerRequest = {};
export const QueryAllOwnerRequest = {
    encode(message, writer = Writer1.create()) {
        if (message.pagination !== undefined) {
            PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryAllOwnerRequest };
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
    fromJSON(object) {
        const message = { ...baseQueryAllOwnerRequest };
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageRequest.fromJSON(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.pagination !== undefined &&
            (obj.pagination = message.pagination
                ? PageRequest.toJSON(message.pagination)
                : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryAllOwnerRequest };
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageRequest.fromPartial(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
};
const baseQueryAllOwnerResponse = {};
export const QueryAllOwnerResponse = {
    encode(message, writer = Writer1.create()) {
        for (const v of message.Owner) {
            Owner.encode(v, writer.uint32(10).fork()).ldelim();
        }
        if (message.pagination !== undefined) {
            PageResponse.encode(message.pagination, writer.uint32(18).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryAllOwnerResponse };
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
    fromJSON(object) {
        const message = { ...baseQueryAllOwnerResponse };
        message.Owner = [];
        if (object.Owner !== undefined && object.Owner !== null) {
            for (const e of object.Owner) {
                message.Owner.push(Owner.fromJSON(e));
            }
        }
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageResponse.fromJSON(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        if (message.Owner) {
            obj.Owner = message.Owner.map((e) => (e ? Owner.toJSON(e) : undefined));
        }
        else {
            obj.Owner = [];
        }
        message.pagination !== undefined &&
            (obj.pagination = message.pagination
                ? PageResponse.toJSON(message.pagination)
                : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryAllOwnerResponse };
        message.Owner = [];
        if (object.Owner !== undefined && object.Owner !== null) {
            for (const e of object.Owner) {
                message.Owner.push(Owner.fromPartial(e));
            }
        }
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageResponse.fromPartial(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
};
const baseQueryGetRecordRequest = { id: 0 };
export const QueryGetRecordRequest = {
    encode(message, writer = Writer1.create()) {
        if (message.id !== 0) {
            writer.uint32(8).uint64(message.id);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryGetRecordRequest };
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
        const message = { ...baseQueryGetRecordRequest };
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
        const message = { ...baseQueryGetRecordRequest };
        if (object.id !== undefined && object.id !== null) {
            message.id = object.id;
        }
        else {
            message.id = 0;
        }
        return message;
    },
};
const baseQueryGetRecordResponse = {};
export const QueryGetRecordResponse = {
    encode(message, writer = Writer1.create()) {
        if (message.Record !== undefined) {
            Record.encode(message.Record, writer.uint32(10).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryGetRecordResponse };
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
    fromJSON(object) {
        const message = { ...baseQueryGetRecordResponse };
        if (object.Record !== undefined && object.Record !== null) {
            message.Record = Record.fromJSON(object.Record);
        }
        else {
            message.Record = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.Record !== undefined &&
            (obj.Record = message.Record ? Record.toJSON(message.Record) : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryGetRecordResponse };
        if (object.Record !== undefined && object.Record !== null) {
            message.Record = Record.fromPartial(object.Record);
        }
        else {
            message.Record = undefined;
        }
        return message;
    },
};
const baseQueryAllRecordRequest = {};
export const QueryAllRecordRequest = {
    encode(message, writer = Writer1.create()) {
        if (message.pagination !== undefined) {
            PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryAllRecordRequest };
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
    fromJSON(object) {
        const message = { ...baseQueryAllRecordRequest };
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageRequest.fromJSON(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.pagination !== undefined &&
            (obj.pagination = message.pagination
                ? PageRequest.toJSON(message.pagination)
                : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryAllRecordRequest };
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageRequest.fromPartial(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
};
const baseQueryAllRecordResponse = {};
export const QueryAllRecordResponse = {
    encode(message, writer = Writer1.create()) {
        for (const v of message.Record) {
            Record.encode(v, writer.uint32(10).fork()).ldelim();
        }
        if (message.pagination !== undefined) {
            PageResponse.encode(message.pagination, writer.uint32(18).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryAllRecordResponse };
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
    fromJSON(object) {
        const message = { ...baseQueryAllRecordResponse };
        message.Record = [];
        if (object.Record !== undefined && object.Record !== null) {
            for (const e of object.Record) {
                message.Record.push(Record.fromJSON(e));
            }
        }
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageResponse.fromJSON(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        if (message.Record) {
            obj.Record = message.Record.map((e) => e ? Record.toJSON(e) : undefined);
        }
        else {
            obj.Record = [];
        }
        message.pagination !== undefined &&
            (obj.pagination = message.pagination
                ? PageResponse.toJSON(message.pagination)
                : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryAllRecordResponse };
        message.Record = [];
        if (object.Record !== undefined && object.Record !== null) {
            for (const e of object.Record) {
                message.Record.push(Record.fromPartial(e));
            }
        }
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageResponse.fromPartial(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
};
const baseQueryGetWriterRequest = { id: 0 };
export const QueryGetWriterRequest = {
    encode(message, writer = Writer1.create()) {
        if (message.id !== 0) {
            writer.uint32(8).uint64(message.id);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryGetWriterRequest };
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
        const message = { ...baseQueryGetWriterRequest };
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
        const message = { ...baseQueryGetWriterRequest };
        if (object.id !== undefined && object.id !== null) {
            message.id = object.id;
        }
        else {
            message.id = 0;
        }
        return message;
    },
};
const baseQueryGetWriterResponse = {};
export const QueryGetWriterResponse = {
    encode(message, writer = Writer1.create()) {
        if (message.Writer !== undefined) {
            Writer.encode(message.Writer, writer.uint32(10).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryGetWriterResponse };
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
    fromJSON(object) {
        const message = { ...baseQueryGetWriterResponse };
        if (object.Writer !== undefined && object.Writer !== null) {
            message.Writer = Writer.fromJSON(object.Writer);
        }
        else {
            message.Writer = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.Writer !== undefined &&
            (obj.Writer = message.Writer ? Writer.toJSON(message.Writer) : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryGetWriterResponse };
        if (object.Writer !== undefined && object.Writer !== null) {
            message.Writer = Writer.fromPartial(object.Writer);
        }
        else {
            message.Writer = undefined;
        }
        return message;
    },
};
const baseQueryAllWriterRequest = {};
export const QueryAllWriterRequest = {
    encode(message, writer = Writer1.create()) {
        if (message.pagination !== undefined) {
            PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryAllWriterRequest };
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
    fromJSON(object) {
        const message = { ...baseQueryAllWriterRequest };
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageRequest.fromJSON(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.pagination !== undefined &&
            (obj.pagination = message.pagination
                ? PageRequest.toJSON(message.pagination)
                : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryAllWriterRequest };
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageRequest.fromPartial(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
};
const baseQueryAllWriterResponse = {};
export const QueryAllWriterResponse = {
    encode(message, writer = Writer1.create()) {
        for (const v of message.Writer) {
            Writer.encode(v, writer.uint32(10).fork()).ldelim();
        }
        if (message.pagination !== undefined) {
            PageResponse.encode(message.pagination, writer.uint32(18).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryAllWriterResponse };
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
    fromJSON(object) {
        const message = { ...baseQueryAllWriterResponse };
        message.Writer = [];
        if (object.Writer !== undefined && object.Writer !== null) {
            for (const e of object.Writer) {
                message.Writer.push(Writer.fromJSON(e));
            }
        }
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageResponse.fromJSON(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        if (message.Writer) {
            obj.Writer = message.Writer.map((e) => e ? Writer.toJSON(e) : undefined);
        }
        else {
            obj.Writer = [];
        }
        message.pagination !== undefined &&
            (obj.pagination = message.pagination
                ? PageResponse.toJSON(message.pagination)
                : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryAllWriterResponse };
        message.Writer = [];
        if (object.Writer !== undefined && object.Writer !== null) {
            for (const e of object.Writer) {
                message.Writer.push(Writer.fromPartial(e));
            }
        }
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageResponse.fromPartial(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
};
const baseQueryGetTopicRequest = { id: 0 };
export const QueryGetTopicRequest = {
    encode(message, writer = Writer1.create()) {
        if (message.id !== 0) {
            writer.uint32(8).uint64(message.id);
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryGetTopicRequest };
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
        const message = { ...baseQueryGetTopicRequest };
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
        const message = { ...baseQueryGetTopicRequest };
        if (object.id !== undefined && object.id !== null) {
            message.id = object.id;
        }
        else {
            message.id = 0;
        }
        return message;
    },
};
const baseQueryGetTopicResponse = {};
export const QueryGetTopicResponse = {
    encode(message, writer = Writer1.create()) {
        if (message.Topic !== undefined) {
            Topic.encode(message.Topic, writer.uint32(10).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryGetTopicResponse };
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
    fromJSON(object) {
        const message = { ...baseQueryGetTopicResponse };
        if (object.Topic !== undefined && object.Topic !== null) {
            message.Topic = Topic.fromJSON(object.Topic);
        }
        else {
            message.Topic = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.Topic !== undefined &&
            (obj.Topic = message.Topic ? Topic.toJSON(message.Topic) : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryGetTopicResponse };
        if (object.Topic !== undefined && object.Topic !== null) {
            message.Topic = Topic.fromPartial(object.Topic);
        }
        else {
            message.Topic = undefined;
        }
        return message;
    },
};
const baseQueryAllTopicRequest = {};
export const QueryAllTopicRequest = {
    encode(message, writer = Writer1.create()) {
        if (message.pagination !== undefined) {
            PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryAllTopicRequest };
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
    fromJSON(object) {
        const message = { ...baseQueryAllTopicRequest };
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageRequest.fromJSON(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        message.pagination !== undefined &&
            (obj.pagination = message.pagination
                ? PageRequest.toJSON(message.pagination)
                : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryAllTopicRequest };
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageRequest.fromPartial(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
};
const baseQueryAllTopicResponse = {};
export const QueryAllTopicResponse = {
    encode(message, writer = Writer1.create()) {
        for (const v of message.Topic) {
            Topic.encode(v, writer.uint32(10).fork()).ldelim();
        }
        if (message.pagination !== undefined) {
            PageResponse.encode(message.pagination, writer.uint32(18).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseQueryAllTopicResponse };
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
    fromJSON(object) {
        const message = { ...baseQueryAllTopicResponse };
        message.Topic = [];
        if (object.Topic !== undefined && object.Topic !== null) {
            for (const e of object.Topic) {
                message.Topic.push(Topic.fromJSON(e));
            }
        }
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageResponse.fromJSON(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        if (message.Topic) {
            obj.Topic = message.Topic.map((e) => (e ? Topic.toJSON(e) : undefined));
        }
        else {
            obj.Topic = [];
        }
        message.pagination !== undefined &&
            (obj.pagination = message.pagination
                ? PageResponse.toJSON(message.pagination)
                : undefined);
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseQueryAllTopicResponse };
        message.Topic = [];
        if (object.Topic !== undefined && object.Topic !== null) {
            for (const e of object.Topic) {
                message.Topic.push(Topic.fromPartial(e));
            }
        }
        if (object.pagination !== undefined && object.pagination !== null) {
            message.pagination = PageResponse.fromPartial(object.pagination);
        }
        else {
            message.pagination = undefined;
        }
        return message;
    },
};
export class QueryClientImpl {
    constructor(rpc) {
        this.rpc = rpc;
    }
    Owner(request) {
        const data = QueryGetOwnerRequest.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Query", "Owner", data);
        return promise.then((data) => QueryGetOwnerResponse.decode(new Reader(data)));
    }
    OwnerAll(request) {
        const data = QueryAllOwnerRequest.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Query", "OwnerAll", data);
        return promise.then((data) => QueryAllOwnerResponse.decode(new Reader(data)));
    }
    Record(request) {
        const data = QueryGetRecordRequest.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Query", "Record", data);
        return promise.then((data) => QueryGetRecordResponse.decode(new Reader(data)));
    }
    RecordAll(request) {
        const data = QueryAllRecordRequest.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Query", "RecordAll", data);
        return promise.then((data) => QueryAllRecordResponse.decode(new Reader(data)));
    }
    Writer(request) {
        const data = QueryGetWriterRequest.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Query", "Writer", data);
        return promise.then((data) => QueryGetWriterResponse.decode(new Reader(data)));
    }
    WriterAll(request) {
        const data = QueryAllWriterRequest.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Query", "WriterAll", data);
        return promise.then((data) => QueryAllWriterResponse.decode(new Reader(data)));
    }
    Topic(request) {
        const data = QueryGetTopicRequest.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Query", "Topic", data);
        return promise.then((data) => QueryGetTopicResponse.decode(new Reader(data)));
    }
    TopicAll(request) {
        const data = QueryAllTopicRequest.encode(request).finish();
        const promise = this.rpc.request("medibloc.panaceacore.aol.Query", "TopicAll", data);
        return promise.then((data) => QueryAllTopicResponse.decode(new Reader(data)));
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
