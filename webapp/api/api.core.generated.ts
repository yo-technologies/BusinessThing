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

export interface ContractTemplateServiceUpdateTemplateBody {
  name?: string;
  description?: string;
  fieldsSchema?: string;
  s3TemplateKey?: string;
}

export interface DocumentServiceRegisterDocumentBody {
  name?: string;
  s3Key?: string;
  fileType?: string;
  /** @format int64 */
  fileSize?: string;
}

export interface DocumentServiceUpdateDocumentStatusBody {
  status?: CoreDocumentStatus;
  errorMessage?: string;
}

export interface GeneratedContractServiceRegisterContractBody {
  templateId?: string;
  name?: string;
  /** JSON с заполненными значениями полей */
  filledData?: string;
  s3Key?: string;
  fileType?: string;
}

export interface NoteServiceCreateNoteBody {
  content?: string;
}

export interface OrganizationServiceUpdateOrganizationBody {
  name?: string;
  industry?: string;
  region?: string;
  description?: string;
  profileData?: string;
}

export type UserServiceAcceptInvitationBody = object;

export type UserServiceDeactivateUserBody = object;

export interface UserServiceInviteUserBody {
  role?: CoreUserRole;
}

export interface UserServiceUpdateUserRoleBody {
  role?: CoreUserRole;
}

export interface CoreAcceptInvitationResponse {
  user?: CoreUser;
}

export interface CoreAuthenticateWithTelegramRequest {
  initData?: string;
}

export interface CoreAuthenticateWithTelegramResponse {
  accessToken?: string;
  user?: CoreUser;
  /** Флаг нового пользователя, требуется заполнение профиля */
  isNewUser?: boolean;
}

export interface CoreCompleteRegistrationRequest {
  userId?: string;
  firstName?: string;
  lastName?: string;
}

export interface CoreCompleteRegistrationResponse {
  user?: CoreUser;
}

export interface CoreContractTemplate {
  id?: string;
  name?: string;
  description?: string;
  templateType?: string;
  fieldsSchema?: string;
  s3TemplateKey?: string;
  /** @format date-time */
  createdAt?: string;
  /** @format date-time */
  updatedAt?: string;
}

export interface CoreCreateNoteResponse {
  note?: CoreNote;
}

export interface CoreCreateOrganizationRequest {
  name?: string;
  industry?: string;
  region?: string;
  description?: string;
  /** Расширяемая анкета в JSON */
  profileData?: string;
}

export interface CoreCreateOrganizationResponse {
  organization?: CoreOrganization;
}

export interface CoreCreateTemplateRequest {
  name?: string;
  description?: string;
  templateType?: string;
  /** JSON с полями шаблона */
  fieldsSchema?: string;
  s3TemplateKey?: string;
}

export interface CoreCreateTemplateResponse {
  template?: CoreContractTemplate;
}

export interface CoreDocument {
  id?: string;
  organizationId?: string;
  name?: string;
  s3Key?: string;
  fileType?: string;
  /** @format int64 */
  fileSize?: string;
  status?: CoreDocumentStatus;
  errorMessage?: string;
  /** @format date-time */
  createdAt?: string;
  /** @format date-time */
  updatedAt?: string;
}

/** @default "DOCUMENT_STATUS_UNSPECIFIED" */
export enum CoreDocumentStatus {
  DOCUMENT_STATUS_UNSPECIFIED = "DOCUMENT_STATUS_UNSPECIFIED",
  DOCUMENT_STATUS_PENDING = "DOCUMENT_STATUS_PENDING",
  DOCUMENT_STATUS_PROCESSING = "DOCUMENT_STATUS_PROCESSING",
  DOCUMENT_STATUS_INDEXED = "DOCUMENT_STATUS_INDEXED",
  DOCUMENT_STATUS_FAILED = "DOCUMENT_STATUS_FAILED",
}

export interface CoreGenerateDownloadURLRequest {
  s3Key?: string;
}

export interface CoreGenerateDownloadURLResponse {
  downloadUrl?: string;
  /** @format int64 */
  expiresInSeconds?: string;
}

export interface CoreGenerateUploadURLRequest {
  organizationId?: string;
  fileName?: string;
  contentType?: string;
}

export interface CoreGenerateUploadURLResponse {
  uploadUrl?: string;
  s3Key?: string;
  /** @format int64 */
  expiresInSeconds?: string;
}

export interface CoreGeneratedContract {
  id?: string;
  organizationId?: string;
  templateId?: string;
  name?: string;
  filledData?: string;
  s3Key?: string;
  fileType?: string;
  /** @format date-time */
  createdAt?: string;
}

export interface CoreGetContractResponse {
  contract?: CoreGeneratedContract;
}

export interface CoreGetDocumentResponse {
  document?: CoreDocument;
}

export interface CoreGetNoteResponse {
  note?: CoreNote;
}

export interface CoreGetOrganizationResponse {
  organization?: CoreOrganization;
}

export interface CoreGetTemplateResponse {
  template?: CoreContractTemplate;
}

export interface CoreGetUserResponse {
  user?: CoreUser;
}

export interface CoreInvitation {
  id?: string;
  organizationId?: string;
  token?: string;
  role?: CoreUserRole;
  /** @format date-time */
  expiresAt?: string;
  /** @format date-time */
  usedAt?: string;
  /** @format date-time */
  createdAt?: string;
}

export interface CoreInviteUserResponse {
  invitationToken?: string;
  invitationUrl?: string;
  /** @format date-time */
  expiresAt?: string;
}

export interface CoreListContractsResponse {
  contracts?: CoreGeneratedContract[];
  /** @format int32 */
  total?: number;
  /** @format int32 */
  page?: number;
  /** @format int32 */
  pageSize?: number;
}

export interface CoreListDocumentsResponse {
  documents?: CoreDocument[];
  /** @format int32 */
  total?: number;
  /** @format int32 */
  page?: number;
  /** @format int32 */
  pageSize?: number;
}

export interface CoreListInvitationsResponse {
  invitations?: CoreInvitation[];
  /** @format int32 */
  total?: number;
  /** @format int32 */
  page?: number;
}

export interface CoreListMyOrganizationsResponse {
  organizations?: CoreOrganization[];
}

export interface CoreListNotesResponse {
  notes?: CoreNote[];
}

export interface CoreListTemplatesResponse {
  templates?: CoreContractTemplate[];
  /** @format int32 */
  total?: number;
  /** @format int32 */
  page?: number;
  /** @format int32 */
  pageSize?: number;
}

export interface CoreListUsersResponse {
  users?: CoreUser[];
  /** @format int32 */
  total?: number;
  /** @format int32 */
  page?: number;
  /** @format int32 */
  pageSize?: number;
}

export interface CoreNote {
  id?: string;
  organizationId?: string;
  content?: string;
  /** @format date-time */
  createdAt?: string;
}

export interface CoreOrganization {
  id?: string;
  name?: string;
  industry?: string;
  region?: string;
  description?: string;
  profileData?: string;
  /** @format date-time */
  createdAt?: string;
  /** @format date-time */
  updatedAt?: string;
  /** @format date-time */
  deletedAt?: string;
}

export interface CoreRefreshTokenResponse {
  accessToken?: string;
}

export interface CoreRegisterContractResponse {
  contract?: CoreGeneratedContract;
}

export interface CoreRegisterDocumentResponse {
  document?: CoreDocument;
}

export interface CoreUpdateOrganizationResponse {
  organization?: CoreOrganization;
}

export interface CoreUpdateTemplateResponse {
  template?: CoreContractTemplate;
}

export interface CoreUpdateUserRoleResponse {
  user?: CoreUser;
}

export interface CoreUser {
  id?: string;
  organizationId?: string;
  email?: string;
  telegramId?: string;
  firstName?: string;
  lastName?: string;
  role?: CoreUserRole;
  status?: CoreUserStatus;
  /** @format date-time */
  createdAt?: string;
  /** @format date-time */
  updatedAt?: string;
}

/** @default "USER_ROLE_UNSPECIFIED" */
export enum CoreUserRole {
  USER_ROLE_UNSPECIFIED = "USER_ROLE_UNSPECIFIED",
  USER_ROLE_ADMIN = "USER_ROLE_ADMIN",
  USER_ROLE_EMPLOYEE = "USER_ROLE_EMPLOYEE",
}

/** @default "USER_STATUS_UNSPECIFIED" */
export enum CoreUserStatus {
  USER_STATUS_UNSPECIFIED = "USER_STATUS_UNSPECIFIED",
  USER_STATUS_PENDING = "USER_STATUS_PENDING",
  USER_STATUS_ACTIVE = "USER_STATUS_ACTIVE",
  USER_STATUS_INACTIVE = "USER_STATUS_INACTIVE",
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
 * @title Core Service API
 * @version 1.0.0
 * @baseUrl /api
 *
 * API for Core Service of BusinessThing - manages organizations, users, documents, notes, and contracts
 */
export class Api<SecurityDataType extends unknown> extends HttpClient<SecurityDataType> {
  v1 = {
    /**
 * No description
 *
 * @tags AuthService
 * @name AuthServiceCompleteRegistration
 * @summary Завершение регистрации нового пользователя (заполнение ФИО)
Не требует JWT токена, так как новый пользователь его еще не получил
 * @request POST:/v1/auth/complete-registration
 * @secure
 */
    authServiceCompleteRegistration: (body: CoreCompleteRegistrationRequest, params: RequestParams = {}) =>
      this.request<CoreCompleteRegistrationResponse, RpcStatus>({
        path: `/v1/auth/complete-registration`,
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
     * @tags AuthService
     * @name AuthServiceRefreshToken
     * @summary Обновить токен (получить новый токен с актуальным списком организаций)
     * @request POST:/v1/auth/refresh
     * @secure
     */
    authServiceRefreshToken: (params: RequestParams = {}) =>
      this.request<CoreRefreshTokenResponse, RpcStatus>({
        path: `/v1/auth/refresh`,
        method: "POST",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags AuthService
     * @name AuthServiceAuthenticateWithTelegram
     * @summary Авторизация через Telegram WebApp initData
     * @request POST:/v1/auth/telegram
     * @secure
     */
    authServiceAuthenticateWithTelegram: (body: CoreAuthenticateWithTelegramRequest, params: RequestParams = {}) =>
      this.request<CoreAuthenticateWithTelegramResponse, RpcStatus>({
        path: `/v1/auth/telegram`,
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
     * @tags GeneratedContractService
     * @name GeneratedContractServiceGetContract
     * @summary Получить договор
     * @request GET:/v1/contracts/{id}
     * @secure
     */
    generatedContractServiceGetContract: (id: string, params: RequestParams = {}) =>
      this.request<CoreGetContractResponse, RpcStatus>({
        path: `/v1/contracts/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags GeneratedContractService
     * @name GeneratedContractServiceDeleteContract
     * @summary Удалить договор
     * @request DELETE:/v1/contracts/{id}
     * @secure
     */
    generatedContractServiceDeleteContract: (id: string, params: RequestParams = {}) =>
      this.request<UserServiceAcceptInvitationBody, RpcStatus>({
        path: `/v1/contracts/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DocumentService
     * @name DocumentServiceGetDocument
     * @summary Получить документ
     * @request GET:/v1/documents/{id}
     * @secure
     */
    documentServiceGetDocument: (id: string, params: RequestParams = {}) =>
      this.request<CoreGetDocumentResponse, RpcStatus>({
        path: `/v1/documents/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DocumentService
     * @name DocumentServiceDeleteDocument
     * @summary Удалить документ
     * @request DELETE:/v1/documents/{id}
     * @secure
     */
    documentServiceDeleteDocument: (id: string, params: RequestParams = {}) =>
      this.request<UserServiceAcceptInvitationBody, RpcStatus>({
        path: `/v1/documents/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DocumentService
     * @name DocumentServiceUpdateDocumentStatus
     * @summary Обновить статус документа (вызов от Document Processing)
     * @request PATCH:/v1/documents/{id}/status
     * @secure
     */
    documentServiceUpdateDocumentStatus: (
      id: string,
      body: DocumentServiceUpdateDocumentStatusBody,
      params: RequestParams = {},
    ) =>
      this.request<UserServiceAcceptInvitationBody, RpcStatus>({
        path: `/v1/documents/${id}/status`,
        method: "PATCH",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags UserService
     * @name UserServiceDeleteInvitation
     * @summary Удалить/отозвать приглашение
     * @request DELETE:/v1/invitations/{id}
     * @secure
     */
    userServiceDeleteInvitation: (id: string, params: RequestParams = {}) =>
      this.request<UserServiceAcceptInvitationBody, RpcStatus>({
        path: `/v1/invitations/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
 * No description
 *
 * @tags UserService
 * @name UserServiceAcceptInvitation
 * @summary Принять приглашение (добавление пользователя в организацию)
Требует JWT токен пользователя, завершившего регистрацию
 * @request POST:/v1/invitations/{token}/accept
 * @secure
 */
    userServiceAcceptInvitation: (token: string, body: UserServiceAcceptInvitationBody, params: RequestParams = {}) =>
      this.request<CoreAcceptInvitationResponse, RpcStatus>({
        path: `/v1/invitations/${token}/accept`,
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
     * @tags NoteService
     * @name NoteServiceGetNote
     * @summary Получить заметку
     * @request GET:/v1/notes/{id}
     * @secure
     */
    noteServiceGetNote: (id: string, params: RequestParams = {}) =>
      this.request<CoreGetNoteResponse, RpcStatus>({
        path: `/v1/notes/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags NoteService
     * @name NoteServiceDeleteNote
     * @summary Удалить заметку
     * @request DELETE:/v1/notes/{id}
     * @secure
     */
    noteServiceDeleteNote: (id: string, params: RequestParams = {}) =>
      this.request<UserServiceAcceptInvitationBody, RpcStatus>({
        path: `/v1/notes/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OrganizationService
     * @name OrganizationServiceCreateOrganization
     * @summary Создать организацию
     * @request POST:/v1/organizations
     * @secure
     */
    organizationServiceCreateOrganization: (body: CoreCreateOrganizationRequest, params: RequestParams = {}) =>
      this.request<CoreCreateOrganizationResponse, RpcStatus>({
        path: `/v1/organizations`,
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
     * @tags OrganizationService
     * @name OrganizationServiceListMyOrganizations
     * @summary Получить список своих организаций
     * @request GET:/v1/organizations/me
     * @secure
     */
    organizationServiceListMyOrganizations: (params: RequestParams = {}) =>
      this.request<CoreListMyOrganizationsResponse, RpcStatus>({
        path: `/v1/organizations/me`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OrganizationService
     * @name OrganizationServiceGetOrganization
     * @summary Получить организацию по ID
     * @request GET:/v1/organizations/{id}
     * @secure
     */
    organizationServiceGetOrganization: (id: string, params: RequestParams = {}) =>
      this.request<CoreGetOrganizationResponse, RpcStatus>({
        path: `/v1/organizations/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OrganizationService
     * @name OrganizationServiceDeleteOrganization
     * @summary Удалить организацию (soft delete)
     * @request DELETE:/v1/organizations/{id}
     * @secure
     */
    organizationServiceDeleteOrganization: (id: string, params: RequestParams = {}) =>
      this.request<UserServiceAcceptInvitationBody, RpcStatus>({
        path: `/v1/organizations/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags OrganizationService
     * @name OrganizationServiceUpdateOrganization
     * @summary Обновить организацию
     * @request PUT:/v1/organizations/{id}
     * @secure
     */
    organizationServiceUpdateOrganization: (
      id: string,
      body: OrganizationServiceUpdateOrganizationBody,
      params: RequestParams = {},
    ) =>
      this.request<CoreUpdateOrganizationResponse, RpcStatus>({
        path: `/v1/organizations/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags GeneratedContractService
     * @name GeneratedContractServiceListContracts
     * @summary Список договоров организации
     * @request GET:/v1/organizations/{organizationId}/contracts
     * @secure
     */
    generatedContractServiceListContracts: (
      organizationId: string,
      query?: {
        /** @format int32 */
        page?: number;
        /** @format int32 */
        pageSize?: number;
      },
      params: RequestParams = {},
    ) =>
      this.request<CoreListContractsResponse, RpcStatus>({
        path: `/v1/organizations/${organizationId}/contracts`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags GeneratedContractService
     * @name GeneratedContractServiceRegisterContract
     * @summary Зарегистрировать сгенерированный договор (вызов от LLM Service)
     * @request POST:/v1/organizations/{organizationId}/contracts
     * @secure
     */
    generatedContractServiceRegisterContract: (
      organizationId: string,
      body: GeneratedContractServiceRegisterContractBody,
      params: RequestParams = {},
    ) =>
      this.request<CoreRegisterContractResponse, RpcStatus>({
        path: `/v1/organizations/${organizationId}/contracts`,
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
     * @tags DocumentService
     * @name DocumentServiceListDocuments
     * @summary Список документов организации
     * @request GET:/v1/organizations/{organizationId}/documents
     * @secure
     */
    documentServiceListDocuments: (
      organizationId: string,
      query?: {
        /** @format int32 */
        page?: number;
        /** @format int32 */
        pageSize?: number;
        /** @default "DOCUMENT_STATUS_UNSPECIFIED" */
        status?:
          | "DOCUMENT_STATUS_UNSPECIFIED"
          | "DOCUMENT_STATUS_PENDING"
          | "DOCUMENT_STATUS_PROCESSING"
          | "DOCUMENT_STATUS_INDEXED"
          | "DOCUMENT_STATUS_FAILED";
      },
      params: RequestParams = {},
    ) =>
      this.request<CoreListDocumentsResponse, RpcStatus>({
        path: `/v1/organizations/${organizationId}/documents`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DocumentService
     * @name DocumentServiceRegisterDocument
     * @summary Зарегистрировать документ после загрузки в S3
     * @request POST:/v1/organizations/{organizationId}/documents
     * @secure
     */
    documentServiceRegisterDocument: (
      organizationId: string,
      body: DocumentServiceRegisterDocumentBody,
      params: RequestParams = {},
    ) =>
      this.request<CoreRegisterDocumentResponse, RpcStatus>({
        path: `/v1/organizations/${organizationId}/documents`,
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
     * @tags UserService
     * @name UserServiceListInvitations
     * @summary Список приглашений организации
     * @request GET:/v1/organizations/{organizationId}/invitations
     * @secure
     */
    userServiceListInvitations: (
      organizationId: string,
      query?: {
        /** @format int32 */
        page?: number;
        /** @format int32 */
        pageSize?: number;
      },
      params: RequestParams = {},
    ) =>
      this.request<CoreListInvitationsResponse, RpcStatus>({
        path: `/v1/organizations/${organizationId}/invitations`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags NoteService
     * @name NoteServiceListNotes
     * @summary Список заметок организации
     * @request GET:/v1/organizations/{organizationId}/notes
     * @secure
     */
    noteServiceListNotes: (
      organizationId: string,
      query?: {
        /** @format int32 */
        limit?: number;
      },
      params: RequestParams = {},
    ) =>
      this.request<CoreListNotesResponse, RpcStatus>({
        path: `/v1/organizations/${organizationId}/notes`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags NoteService
     * @name NoteServiceCreateNote
     * @summary Создать заметку об организации
     * @request POST:/v1/organizations/{organizationId}/notes
     * @secure
     */
    noteServiceCreateNote: (organizationId: string, body: NoteServiceCreateNoteBody, params: RequestParams = {}) =>
      this.request<CoreCreateNoteResponse, RpcStatus>({
        path: `/v1/organizations/${organizationId}/notes`,
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
     * @tags UserService
     * @name UserServiceListUsers
     * @summary Список пользователей организации
     * @request GET:/v1/organizations/{organizationId}/users
     * @secure
     */
    userServiceListUsers: (
      organizationId: string,
      query?: {
        /** @format int32 */
        page?: number;
        /** @format int32 */
        pageSize?: number;
      },
      params: RequestParams = {},
    ) =>
      this.request<CoreListUsersResponse, RpcStatus>({
        path: `/v1/organizations/${organizationId}/users`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags UserService
     * @name UserServiceInviteUser
     * @summary Пригласить пользователя (создать одноразовую ссылку)
     * @request POST:/v1/organizations/{organizationId}/users/invite
     * @secure
     */
    userServiceInviteUser: (organizationId: string, body: UserServiceInviteUserBody, params: RequestParams = {}) =>
      this.request<CoreInviteUserResponse, RpcStatus>({
        path: `/v1/organizations/${organizationId}/users/invite`,
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
     * @tags StorageService
     * @name StorageServiceGenerateDownloadUrl
     * @summary Получить presigned URL для скачивания документа
     * @request POST:/v1/storage/download-url
     * @secure
     */
    storageServiceGenerateDownloadUrl: (body: CoreGenerateDownloadURLRequest, params: RequestParams = {}) =>
      this.request<CoreGenerateDownloadURLResponse, RpcStatus>({
        path: `/v1/storage/download-url`,
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
     * @tags StorageService
     * @name StorageServiceGenerateUploadUrl
     * @summary Получить presigned URL для загрузки документа
     * @request POST:/v1/storage/upload-url
     * @secure
     */
    storageServiceGenerateUploadUrl: (body: CoreGenerateUploadURLRequest, params: RequestParams = {}) =>
      this.request<CoreGenerateUploadURLResponse, RpcStatus>({
        path: `/v1/storage/upload-url`,
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
     * @tags ContractTemplateService
     * @name ContractTemplateServiceListTemplates
     * @summary Список шаблонов организации
     * @request GET:/v1/templates
     * @secure
     */
    contractTemplateServiceListTemplates: (
      query?: {
        /** @format int32 */
        page?: number;
        /** @format int32 */
        pageSize?: number;
      },
      params: RequestParams = {},
    ) =>
      this.request<CoreListTemplatesResponse, RpcStatus>({
        path: `/v1/templates`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags ContractTemplateService
     * @name ContractTemplateServiceCreateTemplate
     * @summary Создать шаблон договора
     * @request POST:/v1/templates
     * @secure
     */
    contractTemplateServiceCreateTemplate: (body: CoreCreateTemplateRequest, params: RequestParams = {}) =>
      this.request<CoreCreateTemplateResponse, RpcStatus>({
        path: `/v1/templates`,
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
     * @tags ContractTemplateService
     * @name ContractTemplateServiceGetTemplate
     * @summary Получить шаблон
     * @request GET:/v1/templates/{id}
     * @secure
     */
    contractTemplateServiceGetTemplate: (id: string, params: RequestParams = {}) =>
      this.request<CoreGetTemplateResponse, RpcStatus>({
        path: `/v1/templates/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags ContractTemplateService
     * @name ContractTemplateServiceDeleteTemplate
     * @summary Удалить шаблон
     * @request DELETE:/v1/templates/{id}
     * @secure
     */
    contractTemplateServiceDeleteTemplate: (id: string, params: RequestParams = {}) =>
      this.request<UserServiceAcceptInvitationBody, RpcStatus>({
        path: `/v1/templates/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags ContractTemplateService
     * @name ContractTemplateServiceUpdateTemplate
     * @summary Обновить шаблон
     * @request PUT:/v1/templates/{id}
     * @secure
     */
    contractTemplateServiceUpdateTemplate: (
      id: string,
      body: ContractTemplateServiceUpdateTemplateBody,
      params: RequestParams = {},
    ) =>
      this.request<CoreUpdateTemplateResponse, RpcStatus>({
        path: `/v1/templates/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags UserService
     * @name UserServiceGetUser
     * @summary Получить пользователя
     * @request GET:/v1/users/{id}
     * @secure
     */
    userServiceGetUser: (id: string, params: RequestParams = {}) =>
      this.request<CoreGetUserResponse, RpcStatus>({
        path: `/v1/users/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags UserService
     * @name UserServiceDeactivateUser
     * @summary Деактивировать пользователя
     * @request POST:/v1/users/{id}/deactivate
     * @secure
     */
    userServiceDeactivateUser: (id: string, body: UserServiceDeactivateUserBody, params: RequestParams = {}) =>
      this.request<UserServiceAcceptInvitationBody, RpcStatus>({
        path: `/v1/users/${id}/deactivate`,
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
     * @tags UserService
     * @name UserServiceUpdateUserRole
     * @summary Обновить роль пользователя
     * @request PATCH:/v1/users/{id}/role
     * @secure
     */
    userServiceUpdateUserRole: (id: string, body: UserServiceUpdateUserRoleBody, params: RequestParams = {}) =>
      this.request<CoreUpdateUserRoleResponse, RpcStatus>({
        path: `/v1/users/${id}/role`,
        method: "PATCH",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),
  };
}
