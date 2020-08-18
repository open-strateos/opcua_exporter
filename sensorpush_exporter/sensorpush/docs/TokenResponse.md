# TokenResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AccessToken** | **string** | JWT oAuth access token. Pass this token to the data services via the http header &#39;Authorization&#39;. Example &#39;Authorization&#39; : &#39;Bearer &lt;access token&gt;&#39; | [optional] 
**ExpiresIn** | **float32** | TTL of the token in seconds | [optional] 
**RefreshToken** | **string** | JWT oAuth refresh token. Pass this token to the token service to retrieve a new access token. | [optional] 
**TokenType** | **string** | Type of token returned | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


