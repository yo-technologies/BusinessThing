/* eslint-disable */
/* tslint:disable */
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

export interface AgentChat {
  id?: string;
  organizationId?: string;
  userId?: string;
  agentKey?: string;
  title?: string;
  status?: AgentChatStatus;
  /** для субагентов */
  parentChatId?: string;
  /** связь с tool call родительского агента */
  parentToolCallId?: string;
  /** @format date-time */
  createdAt?: string;
  /** @format date-time */
  updatedAt?: string;
}

/** Отправляется при создании чата и при изменении имени */
export interface AgentChatEvent {
  chatId?: string;
  chatName?: string;
}

/** @default "CHAT_STATUS_UNSPECIFIED" */
export enum AgentChatStatus {
  CHAT_STATUS_UNSPECIFIED = "CHAT_STATUS_UNSPECIFIED",
  CHAT_STATUS_ACTIVE = "CHAT_STATUS_ACTIVE",
  CHAT_STATUS_COMPLETED = "CHAT_STATUS_COMPLETED",
  CHAT_STATUS_FAILED = "CHAT_STATUS_FAILED",
  CHAT_STATUS_ARCHIVED = "CHAT_STATUS_ARCHIVED",
}

export interface AgentChatUsage {
  /** @format int32 */
  promptTokens?: number;
  /** @format int32 */
  completionTokens?: number;
  /** @format int32 */
  totalTokens?: number;
}

export interface AgentCreateMemoryFactRequest {
  orgId?: string;
  content?: string;
}

export interface AgentCreateMemoryFactResponse {
  fact?: AgentMemoryFact;
}

export interface AgentErrorEvent {
  code?: string;
  message?: string;
}

/** Отправляется в конце стрима с полным состоянием чата */
export interface AgentFinalEvent {
  chat?: AgentChat;
  messages?: AgentMessage[];
  /** @format int32 */
  totalMessages?: number;
}

export interface AgentGetChatResponse {
  chat?: AgentChat;
}

export interface AgentGetLLMLimitsResponse {
  /** @format int32 */
  dailyLimit?: number;
  /** @format int32 */
  used?: number;
  /** @format int32 */
  remaining?: number;
}

export interface AgentGetMessagesResponse {
  messages?: AgentMessage[];
  /** @format int32 */
  total?: number;
}

export interface AgentListChatsResponse {
  chats?: AgentChat[];
  /** @format int32 */
  total?: number;
  /** @format int32 */
  page?: number;
  /** @format int32 */
  pageSize?: number;
}

export interface AgentListMemoryFactsResponse {
  facts?: AgentMemoryFact[];
}

export interface AgentMemoryFact {
  id?: string;
  content?: string;
  /** @format date-time */
  createdAt?: string;
}

export interface AgentMessage {
  id?: string;
  chatId?: string;
  role?: AgentMessageRole;
  content?: string;
  /** null для user, agent_key для агентов */
  sender?: string;
  toolCalls?: AgentToolCall[];
  /** для tool результатов */
  toolCallId?: string;
  /** @format date-time */
  createdAt?: string;
}

export interface AgentMessageChunk {
  content?: string;
}

/** @default "MESSAGE_ROLE_UNSPECIFIED" */
export enum AgentMessageRole {
  MESSAGE_ROLE_UNSPECIFIED = "MESSAGE_ROLE_UNSPECIFIED",
  MESSAGE_ROLE_SYSTEM = "MESSAGE_ROLE_SYSTEM",
  MESSAGE_ROLE_USER = "MESSAGE_ROLE_USER",
  MESSAGE_ROLE_ASSISTANT = "MESSAGE_ROLE_ASSISTANT",
  MESSAGE_ROLE_TOOL = "MESSAGE_ROLE_TOOL",
}

export interface AgentNewMessagePayload {
  chatId?: string;
  orgId?: string;
  content?: string;
}

export interface AgentStreamMessageResponse {
  chunk?: AgentMessageChunk;
  message?: AgentMessage;
  toolCall?: AgentToolCallEvent;
  usage?: AgentUsageEvent;
  error?: AgentErrorEvent;
  chat?: AgentChatEvent;
  final?: AgentFinalEvent;
}

export interface AgentTestGenerateContractRequest {
  orgId?: string;
  templateId?: string;
  contractName?: string;
  /** JSON объект с заполненными данными (ключ - имя поля БЕЗ скобок, значение - данные) */
  filledData?: Record<string, string>;
}

export interface AgentTestGenerateContractResponse {
  contractId?: string;
  contractName?: string;
  downloadUrl?: string;
  s3Key?: string;
  templateName?: string;
  /** @format date-time */
  createdAt?: string;
}

export interface AgentToolCall {
  id?: string;
  name?: string;
  /** JSON string */
  arguments?: string;
  /** JSON string (после выполнения) */
  result?: string;
  status?: AgentToolCallStatus;
  /** @format date-time */
  createdAt?: string;
  /** @format date-time */
  completedAt?: string;
}

export interface AgentToolCallEvent {
  toolCallId?: string;
  toolName?: string;
  arguments?: string;
  status?: string;
}

/** @default "TOOL_CALL_STATUS_UNSPECIFIED" */
export enum AgentToolCallStatus {
  TOOL_CALL_STATUS_UNSPECIFIED = "TOOL_CALL_STATUS_UNSPECIFIED",
  TOOL_CALL_STATUS_PENDING = "TOOL_CALL_STATUS_PENDING",
  TOOL_CALL_STATUS_EXECUTING = "TOOL_CALL_STATUS_EXECUTING",
  TOOL_CALL_STATUS_COMPLETED = "TOOL_CALL_STATUS_COMPLETED",
  TOOL_CALL_STATUS_FAILED = "TOOL_CALL_STATUS_FAILED",
}

export interface AgentUsageEvent {
  usage?: AgentChatUsage;
}

export interface ProtobufAny {
  "@type"?: string;
  [key: string]: any;
}

export interface RpcStatus {
  /** @format int32 */
  code?: number;
  message?: string;
  details?: ProtobufAny[];
}

import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, HeadersDefaults, ResponseType } from "axios";

export type QueryParamsType = Record<string | number, any>;

export interface FullRequestParams extends Omit<AxiosRequestConfig, "data" | "params" | "url" | "responseType"> {
  /** set parameter to `true` for call `securityWorker` for this request */
  secure?: boolean;
  /** request path */
  path: string;
  /** content type of request body */
  type?: ContentType;
  /** query params */
  query?: QueryParamsType;
  /** format of response (i.e. response.json() -> format: "json") */
  format?: ResponseType;
  /** request body */
  body?: unknown;
}

export type RequestParams = Omit<FullRequestParams, "body" | "method" | "query" | "path">;

export interface ApiConfig<SecurityDataType = unknown> extends Omit<AxiosRequestConfig, "data" | "cancelToken"> {
  securityWorker?: (
    securityData: SecurityDataType | null,
  ) => Promise<AxiosRequestConfig | void> | AxiosRequestConfig | void;
  secure?: boolean;
  format?: ResponseType;
}

export enum ContentType {
  Json = "application/json",
  FormData = "multipart/form-data",
  UrlEncoded = "application/x-www-form-urlencoded",
  Text = "text/plain",
}

export class HttpClient<SecurityDataType = unknown> {
  public instance: AxiosInstance;
  private securityData: SecurityDataType | null = null;
  private securityWorker?: ApiConfig<SecurityDataType>["securityWorker"];
  private secure?: boolean;
  private format?: ResponseType;

  constructor({ securityWorker, secure, format, ...axiosConfig }: ApiConfig<SecurityDataType> = {}) {
    this.instance = axios.create({ ...axiosConfig, baseURL: axiosConfig.baseURL || "/api" });
    this.secure = secure;
    this.format = format;
    this.securityWorker = securityWorker;
  }

  public setSecurityData = (data: SecurityDataType | null) => {
    this.securityData = data;
  };

  protected mergeRequestParams(params1: AxiosRequestConfig, params2?: AxiosRequestConfig): AxiosRequestConfig {
    const method = params1.method || (params2 && params2.method);

    return {
      ...this.instance.defaults,
      ...params1,
      ...(params2 || {}),
      headers: {
        ...((method && this.instance.defaults.headers[method.toLowerCase() as keyof HeadersDefaults]) || {}),
        ...(params1.headers || {}),
        ...((params2 && params2.headers) || {}),
      },
    };
  }

  protected stringifyFormItem(formItem: unknown) {
    if (typeof formItem === "object" && formItem !== null) {
      return JSON.stringify(formItem);
    } else {
      return `${formItem}`;
    }
  }

  protected createFormData(input: Record<string, unknown>): FormData {
    return Object.keys(input || {}).reduce((formData, key) => {
      const property = input[key];
      const propertyContent: any[] = property instanceof Array ? property : [property];

      for (const formItem of propertyContent) {
        const isFileType = formItem instanceof Blob || formItem instanceof File;
        formData.append(key, isFileType ? formItem : this.stringifyFormItem(formItem));
      }

      return formData;
    }, new FormData());
  }

  public request = async <T = any, _E = any>({
    secure,
    path,
    type,
    query,
    format,
    body,
    ...params
  }: FullRequestParams): Promise<AxiosResponse<T>> => {
    const secureParams =
      ((typeof secure === "boolean" ? secure : this.secure) &&
        this.securityWorker &&
        (await this.securityWorker(this.securityData))) ||
      {};
    const requestParams = this.mergeRequestParams(params, secureParams);
    const responseFormat = format || this.format || undefined;

    if (type === ContentType.FormData && body && body !== null && typeof body === "object") {
      body = this.createFormData(body as Record<string, unknown>);
    }

    if (type === ContentType.Text && body && body !== null && typeof body !== "string") {
      body = JSON.stringify(body);
    }

    return this.instance.request({
      ...requestParams,
      headers: {
        ...(requestParams.headers || {}),
        ...(type && type !== ContentType.FormData ? { "Content-Type": type } : {}),
      },
      params: query,
      responseType: responseFormat,
      data: body,
      url: path,
    });
  };
}

/**
 * @title LLM Service API
 * @version 1.0.0
 * @baseUrl /api
 *
 * API for LLM Service of BusinessThing
 */
export class Api<SecurityDataType extends unknown> extends HttpClient<SecurityDataType> {
  v1 = {
    /**
     * No description
     *
     * @tags AgentService
     * @name AgentServiceListChats
     * @summary Получить список чатов
     * @request GET:/v1/chats
     * @secure
     */
    agentServiceListChats: (
      query?: {
        orgId?: string;
        /** @format int32 */
        page?: number;
        /** @format int32 */
        pageSize?: number;
      },
      params: RequestParams = {},
    ) =>
      this.request<AgentListChatsResponse, RpcStatus>({
        path: `/v1/chats`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags AgentService
     * @name AgentServiceGetChat
     * @summary Получить чат по ID
     * @request GET:/v1/chats/{chatId}
     * @secure
     */
    agentServiceGetChat: (
      chatId: string,
      query?: {
        orgId?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<AgentGetChatResponse, RpcStatus>({
        path: `/v1/chats/${chatId}`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags AgentService
     * @name AgentServiceDeleteChat
     * @summary Удалить чат
     * @request DELETE:/v1/chats/{chatId}
     * @secure
     */
    agentServiceDeleteChat: (
      chatId: string,
      query?: {
        orgId?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<object, RpcStatus>({
        path: `/v1/chats/${chatId}`,
        method: "DELETE",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags AgentService
     * @name AgentServiceGetMessages
     * @summary Получить историю сообщений чата
     * @request GET:/v1/chats/{chatId}/messages
     * @secure
     */
    agentServiceGetMessages: (
      chatId: string,
      query?: {
        orgId?: string;
        /** @format int32 */
        limit?: number;
        /** @format int32 */
        offset?: number;
      },
      params: RequestParams = {},
    ) =>
      this.request<AgentGetMessagesResponse, RpcStatus>({
        path: `/v1/chats/${chatId}/messages`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags ContractsService
     * @name ContractsServiceTestGenerateContract
     * @summary Тестовая генерация контракта из шаблона
     * @request POST:/v1/contracts/test-generate
     * @secure
     */
    contractsServiceTestGenerateContract: (body: AgentTestGenerateContractRequest, params: RequestParams = {}) =>
      this.request<AgentTestGenerateContractResponse, RpcStatus>({
        path: `/v1/contracts/test-generate`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags AgentService
     * @name AgentServiceGetLlmLimits
     * @summary Получить лимиты использования LLM
     * @request GET:/v1/llm/limits
     * @secure
     */
    agentServiceGetLlmLimits: (params: RequestParams = {}) =>
      this.request<AgentGetLLMLimitsResponse, RpcStatus>({
        path: `/v1/llm/limits`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags MemoryService
     * @name MemoryServiceListMemoryFacts
     * @summary Список всех фактов об организации
     * @request GET:/v1/memory/facts
     * @secure
     */
    memoryServiceListMemoryFacts: (
      query?: {
        orgId?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<AgentListMemoryFactsResponse, RpcStatus>({
        path: `/v1/memory/facts`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags MemoryService
     * @name MemoryServiceCreateMemoryFact
     * @summary Создать новый факт
     * @request POST:/v1/memory/facts
     * @secure
     */
    memoryServiceCreateMemoryFact: (body: AgentCreateMemoryFactRequest, params: RequestParams = {}) =>
      this.request<AgentCreateMemoryFactResponse, RpcStatus>({
        path: `/v1/memory/facts`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags MemoryService
     * @name MemoryServiceDeleteMemoryFact
     * @summary Удалить факт по id
     * @request DELETE:/v1/memory/facts/{id}
     * @secure
     */
    memoryServiceDeleteMemoryFact: (
      id: string,
      query?: {
        orgId?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<object, RpcStatus>({
        path: `/v1/memory/facts/${id}`,
        method: "DELETE",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),
  };
}
