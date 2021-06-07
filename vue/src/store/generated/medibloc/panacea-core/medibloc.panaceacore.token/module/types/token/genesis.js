/* eslint-disable */
import { Token } from "../token/token";
import { Writer, Reader } from "protobufjs/minimal";
export const protobufPackage = "medibloc.panaceacore.token";
const baseGenesisState = {};
export const GenesisState = {
    encode(message, writer = Writer.create()) {
        for (const v of message.tokenList) {
            Token.encode(v, writer.uint32(10).fork()).ldelim();
        }
        return writer;
    },
    decode(input, length) {
        const reader = input instanceof Uint8Array ? new Reader(input) : input;
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseGenesisState };
        message.tokenList = [];
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.tokenList.push(Token.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON(object) {
        const message = { ...baseGenesisState };
        message.tokenList = [];
        if (object.tokenList !== undefined && object.tokenList !== null) {
            for (const e of object.tokenList) {
                message.tokenList.push(Token.fromJSON(e));
            }
        }
        return message;
    },
    toJSON(message) {
        const obj = {};
        if (message.tokenList) {
            obj.tokenList = message.tokenList.map((e) => e ? Token.toJSON(e) : undefined);
        }
        else {
            obj.tokenList = [];
        }
        return obj;
    },
    fromPartial(object) {
        const message = { ...baseGenesisState };
        message.tokenList = [];
        if (object.tokenList !== undefined && object.tokenList !== null) {
            for (const e of object.tokenList) {
                message.tokenList.push(Token.fromPartial(e));
            }
        }
        return message;
    },
};
