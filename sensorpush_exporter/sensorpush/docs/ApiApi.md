# \ApiApi

All URIs are relative to *https://api.sensorpush.com/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AccessToken**](ApiApi.md#AccessToken) | **Post** /oauth/accesstoken | Request a temporary oAuth access code.
[**Download**](ApiApi.md#Download) | **Post** /reports/download | Download bulk reports.
[**Gateways**](ApiApi.md#Gateways) | **Post** /devices/gateways | Lists all gateways.
[**List**](ApiApi.md#List) | **Post** /reports/list | Lists reports available for download.
[**OauthAuthorizePost**](ApiApi.md#OauthAuthorizePost) | **Post** /oauth/authorize | Sign in and request an authorization code
[**RootPost**](ApiApi.md#RootPost) | **Post** / | SensorPush API status
[**Samples**](ApiApi.md#Samples) | **Post** /samples | Queries for sensor samples.
[**Sensors**](ApiApi.md#Sensors) | **Post** /devices/sensors | Lists all sensors.
[**Token**](ApiApi.md#Token) | **Post** /oauth/token | oAuth 2.0 for authorization, access, and refresh tokens



## AccessToken

> AccessTokenResponse AccessToken(ctx, accessTokenRequest)

Request a temporary oAuth access code.

This is a simplified version of oAuth in that it only supports accesstokens and does not require a client_id. See the endpoint '/api/v1/oauth/token' for the more advanced oAuth endpoint. Once a user has been authorized, the client app will call this service to receive the access token. The access token will be used to grant permissions to data stores. An access token expires every hour. After that, request a new access token.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**accessTokenRequest** | [**AccessTokenRequest**](AccessTokenRequest.md)|  | 

### Return type

[**AccessTokenResponse**](AccessTokenResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Download

> Download(ctx, reportsRequest)

Download bulk reports.

This service will download bulk generated reports.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**reportsRequest** | [**ReportsRequest**](ReportsRequest.md)|  | 

### Return type

 (empty response body)

### Authorization

[oauth](../README.md#oauth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Gateways

> map[string]Gateway Gateways(ctx, gatewaysRequest)

Lists all gateways.

This service will return an inventory of all registered gateways for this account.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**gatewaysRequest** | [**GatewaysRequest**](GatewaysRequest.md)|  | 

### Return type

[**map[string]Gateway**](Gateway.md)

### Authorization

[oauth](../README.md#oauth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## List

> ListResponse List(ctx, reportsRequest)

Lists reports available for download.

This service will list all bulk generated reports available to download.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**reportsRequest** | [**ReportsRequest**](ReportsRequest.md)|  | 

### Return type

[**ListResponse**](ListResponse.md)

### Authorization

[oauth](../README.md#oauth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## OauthAuthorizePost

> AuthorizeResponse OauthAuthorizePost(ctx, authorizeRequest)

Sign in and request an authorization code

Sign into the SensorPush API via redirect to SensorPush logon. Then signin using email/password, or an api id. This service will return an oAuth authorization code that can be exchanged for an oAuth access token using the accesstoken service.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**authorizeRequest** | [**AuthorizeRequest**](AuthorizeRequest.md)|  | 

### Return type

[**AuthorizeResponse**](AuthorizeResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RootPost

> Status RootPost(ctx, )

SensorPush API status

This service is used as a simple method for clients to verify they can connect to the API.

### Required Parameters

This endpoint does not need any parameter.

### Return type

[**Status**](Status.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Samples

> Samples Samples(ctx, samplesRequest)

Queries for sensor samples.

This service is used to query for samples persisted by the sensors. The service will return all samples after the given parameter {startTime}. Queries that produce greater than ~5MB of data will be truncated. If results return truncated, consider using the sensors parameter list. This will allow you to retrieve more data per sensor. For example, a query that does not provide a sensor list, will attempt to return equal amounts of data for all sensors (i.e. ~5MB divided by N sensors). However, if one sensor is specified, than all ~5MB will be filled for that one sensor (i.e. ~5MB divided by 1). Another option is to paginate through results by time, using {startTime} as the last date in your previous page of results.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**samplesRequest** | [**SamplesRequest**](SamplesRequest.md)|  | 

### Return type

[**Samples**](Samples.md)

### Authorization

[oauth](../README.md#oauth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Sensors

> map[string]Sensor Sensors(ctx, sensorsRequest)

Lists all sensors.

This service will return an inventory of all registered sensors for this account.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sensorsRequest** | [**SensorsRequest**](SensorsRequest.md)|  | 

### Return type

[**map[string]Sensor**](Sensor.md)

### Authorization

[oauth](../README.md#oauth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Token

> TokenResponse Token(ctx, tokenRequest)

oAuth 2.0 for authorization, access, and refresh tokens

This is a more advanced endpoint that implements the oAuth 2.0 specification. Supports grant_types: password, refresh_token, and access_token. If grant_type is null an access_token will be returned. (see <a href=\"https://oauth.net/2/grant-types/\">oAuth Grant Types</a>). A client_id is required for this endpoint. Contact support@sensorpush.com to register your application and recieve a client_id.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tokenRequest** | [**TokenRequest**](TokenRequest.md)|  | 

### Return type

[**TokenResponse**](TokenResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

