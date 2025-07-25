# CoreHRApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createEmployee**](#createemployee) | **POST** /api/v1/corehr/employees | Create a new employee|
|[**deleteEmployee**](#deleteemployee) | **DELETE** /api/v1/corehr/employees/{employee_id} | Delete employee|
|[**getEmployee**](#getemployee) | **GET** /api/v1/corehr/employees/{employee_id} | Get employee by ID|
|[**getOrganizationTree**](#getorganizationtree) | **GET** /api/v1/corehr/organizations/tree | Get organization tree|
|[**listEmployees**](#listemployees) | **GET** /api/v1/corehr/employees | List employees with pagination|
|[**listOrganizations**](#listorganizations) | **GET** /api/v1/corehr/organizations | List organizations|
|[**postPhoneNumberUpdateEvent**](#postphonenumberupdateevent) | **POST** /api/v1/internal/corehr/employee-events/phone-number-update | Create a phone number update event|
|[**updateEmployee**](#updateemployee) | **PUT** /api/v1/corehr/employees/{employee_id} | Update employee|

# **createEmployee**
> Employee createEmployee(createEmployeeRequest)


### Example

```typescript
import {
    CoreHRApi,
    Configuration,
    CreateEmployeeRequest
} from 'cube-castle-api';

const configuration = new Configuration();
const apiInstance = new CoreHRApi(configuration);

let createEmployeeRequest: CreateEmployeeRequest; //

const { status, data } = await apiInstance.createEmployee(
    createEmployeeRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createEmployeeRequest** | **CreateEmployeeRequest**|  | |


### Return type

**Employee**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Employee created successfully |  -  |
|**400** | Bad request |  -  |
|**409** | Employee already exists |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteEmployee**
> deleteEmployee()


### Example

```typescript
import {
    CoreHRApi,
    Configuration
} from 'cube-castle-api';

const configuration = new Configuration();
const apiInstance = new CoreHRApi(configuration);

let employeeId: string; // (default to undefined)

const { status, data } = await apiInstance.deleteEmployee(
    employeeId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **employeeId** | [**string**] |  | defaults to undefined|


### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**204** | Employee deleted successfully |  -  |
|**404** | Employee not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getEmployee**
> Employee getEmployee()


### Example

```typescript
import {
    CoreHRApi,
    Configuration
} from 'cube-castle-api';

const configuration = new Configuration();
const apiInstance = new CoreHRApi(configuration);

let employeeId: string; // (default to undefined)

const { status, data } = await apiInstance.getEmployee(
    employeeId
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **employeeId** | [**string**] |  | defaults to undefined|


### Return type

**Employee**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Employee details |  -  |
|**404** | Employee not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getOrganizationTree**
> OrganizationTreeResponse getOrganizationTree()


### Example

```typescript
import {
    CoreHRApi,
    Configuration
} from 'cube-castle-api';

const configuration = new Configuration();
const apiInstance = new CoreHRApi(configuration);

const { status, data } = await apiInstance.getOrganizationTree();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**OrganizationTreeResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Organization tree structure |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listEmployees**
> EmployeeListResponse listEmployees()


### Example

```typescript
import {
    CoreHRApi,
    Configuration
} from 'cube-castle-api';

const configuration = new Configuration();
const apiInstance = new CoreHRApi(configuration);

let page: number; //Page number (1-based) (optional) (default to 1)
let pageSize: number; //Number of items per page (optional) (default to 20)
let search: string; //Search term for employee name or email (optional) (default to undefined)

const { status, data } = await apiInstance.listEmployees(
    page,
    pageSize,
    search
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number (1-based) | (optional) defaults to 1|
| **pageSize** | [**number**] | Number of items per page | (optional) defaults to 20|
| **search** | [**string**] | Search term for employee name or email | (optional) defaults to undefined|


### Return type

**EmployeeListResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | List of employees |  -  |
|**400** | Bad request |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listOrganizations**
> OrganizationListResponse listOrganizations()


### Example

```typescript
import {
    CoreHRApi,
    Configuration
} from 'cube-castle-api';

const configuration = new Configuration();
const apiInstance = new CoreHRApi(configuration);

const { status, data } = await apiInstance.listOrganizations();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**OrganizationListResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | List of organizations |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **postPhoneNumberUpdateEvent**
> GeneralResponse postPhoneNumberUpdateEvent(phoneNumberUpdateEventRequest)


### Example

```typescript
import {
    CoreHRApi,
    Configuration,
    PhoneNumberUpdateEventRequest
} from 'cube-castle-api';

const configuration = new Configuration();
const apiInstance = new CoreHRApi(configuration);

let phoneNumberUpdateEventRequest: PhoneNumberUpdateEventRequest; //

const { status, data } = await apiInstance.postPhoneNumberUpdateEvent(
    phoneNumberUpdateEventRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **phoneNumberUpdateEventRequest** | **PhoneNumberUpdateEventRequest**|  | |


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
|**202** | Event accepted for processing. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateEmployee**
> Employee updateEmployee(updateEmployeeRequest)


### Example

```typescript
import {
    CoreHRApi,
    Configuration,
    UpdateEmployeeRequest
} from 'cube-castle-api';

const configuration = new Configuration();
const apiInstance = new CoreHRApi(configuration);

let employeeId: string; // (default to undefined)
let updateEmployeeRequest: UpdateEmployeeRequest; //

const { status, data } = await apiInstance.updateEmployee(
    employeeId,
    updateEmployeeRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateEmployeeRequest** | **UpdateEmployeeRequest**|  | |
| **employeeId** | [**string**] |  | defaults to undefined|


### Return type

**Employee**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Employee updated successfully |  -  |
|**400** | Bad request |  -  |
|**404** | Employee not found |  -  |
|**500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

