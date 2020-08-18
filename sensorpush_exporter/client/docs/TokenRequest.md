# TokenRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClientId** | **string** | Client Id assigned to 3rd party applications. Contact support@sensorpush.com to register you app. | [optional] 
**ClientSecret** | **string** | Password associated with the client_id | [optional] 
**Code** | **string** | This can be an authorization, access, or refresh token. Depending on which grant_type you are using. | [optional] 
**GrantType** | **string** | Accepted values are password, refresh_token, and access_token | [optional] 
**Password** | **string** | User&#39;s password | [optional] 
**RedirectUri** | **string** | Redirection url to the 3rd party application once the user has signed into the sensorpush logon. This value should be URL encoded. | [optional] 
**RefreshToken** | **string** | Refresh token used to request new access tokens. | [optional] 
**Username** | **string** | Email of the user to sign in. | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


