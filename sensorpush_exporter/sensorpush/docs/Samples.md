# Samples

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LastTime** | **float32** | Timestamp of the last sample returned. Use this as the start_ts to query for the next page of samples. | [optional] 
**Sensors** | [**map[string][]Sample**](array.md) | Map of sensors and the associated samples. | [optional] 
**Status** | **string** | Message describing state of the api call. | [optional] 
**TotalSamples** | **float32** | Total number of samples across all sensors | [optional] 
**TotalSensors** | **float32** | Total number of sensors returned | [optional] 
**Truncated** | **bool** | The query returned too many results, causing the sample list to be truncated. Consider adjusting the limit or startTime parameters. | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


