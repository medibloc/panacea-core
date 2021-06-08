export interface AolMsgCreateOwnerResponse {
    /** @format uint64 */
    id?: string;
}
export interface AolMsgCreateRecordResponse {
    /** @format uint64 */
    id?: string;
}
export interface AolMsgCreateTopicResponse {
    /** @format uint64 */
    id?: string;
}
export interface AolMsgCreateWriterResponse {
    /** @format uint64 */
    id?: string;
}
export declare type AolMsgDeleteOwnerResponse = object;
export declare type AolMsgDeleteRecordResponse = object;
export declare type AolMsgDeleteTopicResponse = object;
export declare type AolMsgDeleteWriterResponse = object;
export declare type AolMsgUpdateOwnerResponse = object;
export declare type AolMsgUpdateRecordResponse = object;
export declare type AolMsgUpdateTopicResponse = object;
export declare type AolMsgUpdateWriterResponse = object;
export interface AolOwner {
    creator?: string;
    /** @format uint64 */
    id?: string;
    /** @format int32 */
    totalTopics?: number;
}
export interface AolQueryAllOwnerResponse {
    Owner?: AolOwner[];
    /**
     * PageResponse is to be embedded in gRPC response messages where the
     * corresponding request message has used PageRequest.
     *
     *  message SomeResponse {
     *          repeated Bar results = 1;
     *          PageResponse page = 2;
     *  }
     */
    pagination?: V1Beta1PageResponse;
}
export interface AolQueryAllRecordResponse {
    Record?: AolRecord[];
    /**
     * PageResponse is to be embedded in gRPC response messages where the
     * corresponding request message has used PageRequest.
     *
     *  message SomeResponse {
     *          repeated Bar results = 1;
     *          PageResponse page = 2;
     *  }
     */
    pagination?: V1Beta1PageResponse;
}
export interface AolQueryAllTopicResponse {
    Topic?: AolTopic[];
    /**
     * PageResponse is to be embedded in gRPC response messages where the
     * corresponding request message has used PageRequest.
     *
     *  message SomeResponse {
     *          repeated Bar results = 1;
     *          PageResponse page = 2;
     *  }
     */
    pagination?: V1Beta1PageResponse;
}
export interface AolQueryAllWriterResponse {
    Writer?: AolWriter[];
    /**
     * PageResponse is to be embedded in gRPC response messages where the
     * corresponding request message has used PageRequest.
     *
     *  message SomeResponse {
     *          repeated Bar results = 1;
     *          PageResponse page = 2;
     *  }
     */
    pagination?: V1Beta1PageResponse;
}
export interface AolQueryGetOwnerResponse {
    Owner?: AolOwner;
}
export interface AolQueryGetRecordResponse {
    Record?: AolRecord;
}
export interface AolQueryGetTopicResponse {
    Topic?: AolTopic;
}
export interface AolQueryGetWriterResponse {
    Writer?: AolWriter;
}
export interface AolRecord {
    creator?: string;
    /** @format uint64 */
    id?: string;
    key?: string;
    value?: string;
    /** @format int32 */
    nanoTimestamp?: number;
    writerAddress?: string;
}
export interface AolTopic {
    creator?: string;
    /** @format uint64 */
    id?: string;
    description?: string;
    /** @format int32 */
    totalRecords?: number;
    /** @format int32 */
    totalWriter?: number;
}
export interface AolWriter {
    creator?: string;
    /** @format uint64 */
    id?: string;
    moniker?: string;
    description?: string;
    /** @format int32 */
    nanoTimestamp?: number;
}
export interface ProtobufAny {
    typeUrl?: string;
    /** @format byte */
    value?: string;
}
export interface RpcStatus {
    /** @format int32 */
    code?: number;
    message?: string;
    details?: ProtobufAny[];
}
/**
* message SomeRequest {
         Foo some_parameter = 1;
         PageRequest pagination = 2;
 }
*/
export interface V1Beta1PageRequest {
    /**
     * key is a value returned in PageResponse.next_key to begin
     * querying the next page most efficiently. Only one of offset or key
     * should be set.
     * @format byte
     */
    key?: string;
    /**
     * offset is a numeric offset that can be used when key is unavailable.
     * It is less efficient than using key. Only one of offset or key should
     * be set.
     * @format uint64
     */
    offset?: string;
    /**
     * limit is the total number of results to be returned in the result page.
     * If left empty it will default to a value to be set by each app.
     * @format uint64
     */
    limit?: string;
    /**
     * count_total is set to true  to indicate that the result set should include
     * a count of the total number of items available for pagination in UIs.
     * count_total is only respected when offset is used. It is ignored when key
     * is set.
     */
    countTotal?: boolean;
}
/**
* PageResponse is to be embedded in gRPC response messages where the
corresponding request message has used PageRequest.

 message SomeResponse {
         repeated Bar results = 1;
         PageResponse page = 2;
 }
*/
export interface V1Beta1PageResponse {
    /** @format byte */
    nextKey?: string;
    /** @format uint64 */
    total?: string;
}
export declare type QueryParamsType = Record<string | number, any>;
export declare type ResponseFormat = keyof Omit<Body, "body" | "bodyUsed">;
export interface FullRequestParams extends Omit<RequestInit, "body"> {
    /** set parameter to `true` for call `securityWorker` for this request */
    secure?: boolean;
    /** request path */
    path: string;
    /** content type of request body */
    type?: ContentType;
    /** query params */
    query?: QueryParamsType;
    /** format of response (i.e. response.json() -> format: "json") */
    format?: keyof Omit<Body, "body" | "bodyUsed">;
    /** request body */
    body?: unknown;
    /** base url */
    baseUrl?: string;
    /** request cancellation token */
    cancelToken?: CancelToken;
}
export declare type RequestParams = Omit<FullRequestParams, "body" | "method" | "query" | "path">;
export interface ApiConfig<SecurityDataType = unknown> {
    baseUrl?: string;
    baseApiParams?: Omit<RequestParams, "baseUrl" | "cancelToken" | "signal">;
    securityWorker?: (securityData: SecurityDataType) => RequestParams | void;
}
export interface HttpResponse<D extends unknown, E extends unknown = unknown> extends Response {
    data: D;
    error: E;
}
declare type CancelToken = Symbol | string | number;
export declare enum ContentType {
    Json = "application/json",
    FormData = "multipart/form-data",
    UrlEncoded = "application/x-www-form-urlencoded"
}
export declare class HttpClient<SecurityDataType = unknown> {
    baseUrl: string;
    private securityData;
    private securityWorker;
    private abortControllers;
    private baseApiParams;
    constructor(apiConfig?: ApiConfig<SecurityDataType>);
    setSecurityData: (data: SecurityDataType) => void;
    private addQueryParam;
    protected toQueryString(rawQuery?: QueryParamsType): string;
    protected addQueryParams(rawQuery?: QueryParamsType): string;
    private contentFormatters;
    private mergeRequestParams;
    private createAbortSignal;
    abortRequest: (cancelToken: CancelToken) => void;
    request: <T = any, E = any>({ body, secure, path, type, query, format, baseUrl, cancelToken, ...params }: FullRequestParams) => Promise<HttpResponse<T, E>>;
}
/**
 * @title aol/tx.proto
 * @version version not set
 */
export declare class Api<SecurityDataType extends unknown> extends HttpClient<SecurityDataType> {
    /**
     * No description
     *
     * @tags Query
     * @name QueryOwnerAll
     * @request GET:/medibloc/panaceacore/aol/owner
     */
    queryOwnerAll: (query?: {
        "pagination.key"?: string;
        "pagination.offset"?: string;
        "pagination.limit"?: string;
        "pagination.countTotal"?: boolean;
    }, params?: RequestParams) => Promise<HttpResponse<AolQueryAllOwnerResponse, RpcStatus>>;
    /**
     * No description
     *
     * @tags Query
     * @name QueryOwner
     * @summary this line is used by starport scaffolding # 2
     * @request GET:/medibloc/panaceacore/aol/owner/{id}
     */
    queryOwner: (id: string, params?: RequestParams) => Promise<HttpResponse<AolQueryGetOwnerResponse, RpcStatus>>;
    /**
     * No description
     *
     * @tags Query
     * @name QueryRecordAll
     * @request GET:/medibloc/panaceacore/aol/record
     */
    queryRecordAll: (query?: {
        "pagination.key"?: string;
        "pagination.offset"?: string;
        "pagination.limit"?: string;
        "pagination.countTotal"?: boolean;
    }, params?: RequestParams) => Promise<HttpResponse<AolQueryAllRecordResponse, RpcStatus>>;
    /**
     * No description
     *
     * @tags Query
     * @name QueryRecord
     * @request GET:/medibloc/panaceacore/aol/record/{id}
     */
    queryRecord: (id: string, params?: RequestParams) => Promise<HttpResponse<AolQueryGetRecordResponse, RpcStatus>>;
    /**
     * No description
     *
     * @tags Query
     * @name QueryTopicAll
     * @request GET:/medibloc/panaceacore/aol/topic
     */
    queryTopicAll: (query?: {
        "pagination.key"?: string;
        "pagination.offset"?: string;
        "pagination.limit"?: string;
        "pagination.countTotal"?: boolean;
    }, params?: RequestParams) => Promise<HttpResponse<AolQueryAllTopicResponse, RpcStatus>>;
    /**
     * No description
     *
     * @tags Query
     * @name QueryTopic
     * @request GET:/medibloc/panaceacore/aol/topic/{id}
     */
    queryTopic: (id: string, params?: RequestParams) => Promise<HttpResponse<AolQueryGetTopicResponse, RpcStatus>>;
    /**
     * No description
     *
     * @tags Query
     * @name QueryWriterAll
     * @request GET:/medibloc/panaceacore/aol/writer
     */
    queryWriterAll: (query?: {
        "pagination.key"?: string;
        "pagination.offset"?: string;
        "pagination.limit"?: string;
        "pagination.countTotal"?: boolean;
    }, params?: RequestParams) => Promise<HttpResponse<AolQueryAllWriterResponse, RpcStatus>>;
    /**
     * No description
     *
     * @tags Query
     * @name QueryWriter
     * @request GET:/medibloc/panaceacore/aol/writer/{id}
     */
    queryWriter: (id: string, params?: RequestParams) => Promise<HttpResponse<AolQueryGetWriterResponse, RpcStatus>>;
}
export {};
