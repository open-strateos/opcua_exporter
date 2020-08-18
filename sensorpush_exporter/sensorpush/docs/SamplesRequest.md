# SamplesRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Active** | **bool** | Filters sensors by active &#x3D; (true|false). Defaults to true | [optional] 
**Bulk** | **bool** | Queries that return large results are truncated (see comments on Samples endpoint). Set this flag to true for large reports. The report request will be queued and processed within 24 hours. Upon completion, the primary account holder will recieve an email with a link for download. | [optional] 
**Format** | **string** | Returns the results as the specified format (csv|json). Defaults to json | [optional] 
**Limit** | **int32** | Number of samples to return. | [optional] 
**Sensors** | **[]string** | Filters samples by sensor id. This will be the same id returned in the sensors api call. The parameter value must be a list of strings. Example: sensors: [\&quot;123.56789\&quot;]. | [optional] 
**StartTime** | **string** | Start time to find samples (example: 2019-04-07T00:00:00-0400). Leave blank or zero to get the most recent samples. | [optional] 
**StopTime** | **string** | Stop time to find samples (example: 2019-04-07T10:30:00-0400). Leave blank or zero to get the most recent samples. | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


