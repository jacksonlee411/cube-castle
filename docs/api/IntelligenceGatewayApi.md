# IntelligenceGatewayApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**interpretQuery**](#interpretquery) | **POST** /api/v1/interpret | Interpret Natural Language Query|

# **interpretQuery**
> GeneralResponse interpretQuery(interpretRequest)


### Example

```typescript
import {
    IntelligenceGatewayApi,
    Configuration,
    InterpretRequest
} from 'cube-castle-api';

const configuration = new Configuration();
const apiInstance = new IntelligenceGatewayApi(configuration);

let interpretRequest: InterpretRequest; //

const { status, data } = await apiInstance.interpretQuery(
    interpretRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **interpretRequest** | **InterpretRequest**|  | |


### Return type

**GeneralResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Successful interpretation |  -  |
|**400** | Bad request |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

