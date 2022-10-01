# Enrollment Callback API

The Transaction Controls Client Host Callback API will be hosted external to Visa. Card Enrollment callback will be
hosted by the authorizing agent.

| RPC | HTTP Method + Path | Description | Scopes |
|-----|--------------------|-------------|--------|
| Enroll | `POST /visa.service.enrollmentcallback.v1.EnrollmentCallbackAPI/Enroll` | Use this service to notify the host that a cardholder has enrolled for a service. |  |
| Disenroll | `POST /visa.service.enrollmentcallback.v1.EnrollmentCallbackAPI/Disenroll` | Use this service to notify the host that a cardholder has de-enrolled for a service. |  |

## rpc Enroll ([Request](#message-request)) returns ([Response](#message-response))

Use this service to notify the host that a cardholder has enrolled for a service.

### REST Gateway

Method: `POST`
Path: `/webhook/Visa/AccountServices/v3/Enrollment/Notification`

```http request
POST https://gw.fabric.anz.com/Visa/AccountServices/v3/Enrollment/Notification
```

---

## rpc Disenroll ([Request](#message-request)) returns ([Request](#message-response)

Use this service to notify the host that a cardholder has de-enrolled for a service.

### REST Gateway

Method: `DELETE`
Path: `/webhook/Visa/AccountServices/v3/Enrollment/Notification`

```http request
DELETE https://gw.fabric.anz.com/Visa/AccountServices/v3/Enrollment/Notification
```

---

### message BulkEnrollmentObjectList

```protobuf
message BulkEnrollmentObjectList {
  //  Primary account number to enroll in the given service type
  string primary_account_number = 1 [json_name = "primaryAccountNumber"];
  // Service type this card is being enrolled in.
  // "CTC", "CTC_ALERTS", "ALERTS", "MLC". A PAN can only be enrolled in either CTC or CTC_ALERTS.
  repeated string service_types = 2 [json_name = "serviceTypes"];
  // The token for which the enrollment notification request is being made
  string token = 3 [json_name = "token"];
}
```

### message Request

```protobuf
message Request {
  // List of primary account numbers or tokens to bulk enroll, either a PAN or token is required.
  repeated BulkEnrollmentObjectList bulk_enrollment_object_list = 1 [json_name = "bulkEnrollmentObjectList"];
}
```

### message Response

```protobuf
message Response {}
```
