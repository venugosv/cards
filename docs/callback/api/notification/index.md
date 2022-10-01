# Notification Callback API

The Transaction Controls Client Host Callback API will be hosted external to Visa. Card Enrollment callback will be
hosted by the authorizing agent. Alert notification callback will be hosted by the application service responsible for
delivering notifications to the consumer.

| RPC | HTTP Method + Path | Description | Scopes |
|-----|--------------------|-------------|--------|
| Alert | `POST /visa.service.notificationcallback.v1.NotificationCallbackAPI/Alert` |  |  |

## rpc Alert ([Request](#message-request)) returns ([Response](#message-response))

### REST Gateway

Method: `POST`
Path: `/webhook/Visa/CardServices/v3/Notification/Alert`

```http request
POST https://gw.fabric.anz.com/Visa/CardServices/v3/Notification/Alert
```

---

### message Request

```protobuf
message Request {
  repeated string transaction_types = 1 [json_name = "transactionTypes"];
  string app_id = 2 [json_name = "appId"];
  string sponsor_id = 3 [json_name = "sponsorId"];
  TransactionDetails transaction_details = 4 [json_name = "transactionDetails"];
}
```

### message TransactionDetails

```protobuf
message TransactionDetails {
  // Masked primary account number, with only the last 4 digits provided
  string payment_token = 1;
  // The total amount to be billed to the cardholder inclusive of any fees assessed. This amount will be in the card
  // issuers currency.
  float cardholder_bill_amount = 2 [json_name = "cardholderBillAmount"];
  // Defines the specific card and cardholder on which to enforce the payment control rule. This is useful when there
  // are multiple people on the same PAN such as a family of husband, wife, and daughter. Cards that are not specified
  // within these details will not have the payment control rule enforced. NOTE: This value is only passed when the
  // nameOnCard value is present in the configured rule.
  string name_on_card = 3 [json_name = "nameOnCard"];
  // The UserIdentifier is a mandatory data element when supporting VTC notifications. It is used by the issuer -or
  // their notification service provider- to uniquely identify the cardholder who has requested the alert message.
  string user_identifier = 4 [json_name = "userIdentifier"];
  // ISO 8583 three-digit currency classification code that identifies the national currency used at the biller
  // currency location.
  string biller_currency_code = 5 [json_name = "billerCurrencyCode"];
  // This value will be set automatically when the decision request is received. Value is in UTC time
  string request_received_time_stamp = 6 [json_name = "requestReceivedTimeStamp"];
  // Masked primary account number, with only the last 4 digits provided
  string primary_account_number = 7;
  // The retrieval reference number assigned by the card acceptor, must be at least 1 and max of 72 characters
  // alphanumeric
  string retrieval_reference_number = 8;
  // The unique transaction ID assigned by processing gateway. In case of NON visa processed transaction, this will be
  // a Unique tranID from Issuer system.
  string transaction_id = 9 [json_name = "transactionID"];
  ExchangeRateDetails exchange_rate_details = 10 [json_name = "exchangeRateDetails"];
}
```

### message ExchangeRateDetails

```protobuf
message ExchangeRateDetails {
  // Contains the Rate Table ID used to perform any required multi-currency processing
  string fx_rate_table_identifier = 1 [json_name = "fxRateTableIdentifier"];
  // Flag that indicates the completeness of the information provided for Visa and ECB rates. In case of INCOMPLETE
  // status, some of the exchange rate details might be missing. INCOMPLETE Status indicates that system was not able
  // to calculate some of the Enhanced FX related fields
  string fx_status = 2 [json_name = "fxStatus"];
  // Contains the rate used by VisaNet to convert the transaction amount (field 4) to the cardholder billing amount
  // (field 6) including the optional issuer fee (OIF)
  float visa_exchange_rate = 3 [json_name = "visaExchangeRate"];
  // Designates whether a transaction is eligible for the exchange rate that was established at processing to persist
  // and be applied at clearing.Allowed values are ELIGIBLE,NON_ELIGIBLE APPLIED.
  string pfx_indicator = 4 [json_name = "pfxIndicator"];
  // Markup Calculated as ((Visa Rate-ECB Rate)/ECB Rate)*100. Returned with 2 decimals rounded (5.0589 = 5.06),
  // positive or negative value is valid. May not be available when status is INCOMPLETE
  float exchange_rate_mark_up = 5 [json_name = "exchangeRateMarkUp"];
  message ECBExchangeRateInfo {
    // Last GMT date and time when the exchangeRate was pulled from ECB to Visa systems
    string last_updated_date = 1 [json_name = "lastUpdatedDate"];
    // Latest available ECB rate for the currencyCode and billerCurrencyCode pair in the Visa systems
    float exchange_rate = 2 [json_name = "exchangeRate"];
  }
  ECBExchangeRateInfo ecb_exchange_rate_info = 6 [json_name = "ecbExchangeRateInfo"];
  message UserInformation {
    // ApplicationDefinedAttributes by the issuer. These can be used to enrich the VTC notification alerts.
    repeated string application_defined_attributes = 1 [json_name = "applicationDefinedAttributes"];
    // Identifier for the issuer to map the user name
    string banking_identifier = 2 [json_name = "bankingIdentifier"];
    // Name of the user who configured the control.Name is mandatory if UserInformation exists.
    string name = 3 [json_name = "name"];
  }
  UserInformation user_information = 7 [json_name = "userInformation"];
  MerchantInfo merchant_info = 8 [json_name = "merchantInfo"];
  TransactionOutcome transaction_outcome = 9 [json_name = "transactionOutcome"];
}
```

### message MerchantInfo

```protobuf
message MerchantInfo {
  // The city for which the merchant is located.
  string city = 1 [json_name = "city"];
  // Three letter country code at Merchant location.
  string country_code = 2 [json_name = "countryCode"];
  // The five to nine digit postal (or zip) code for the merchant.
  string postal_code = 3 [json_name = "postalCode"];
  // The total transaction amount in local merchant currency.
  float transaction_amount = 4 [json_name = "transactionAmount"];
  // ISO 8583 four-digit merchant classification code that identifies the merchant by their business line.
  string merchant_category_code = 5 [json_name = "merchantCategoryCode"];
  // The terminal ID of the card acceptor
  string card_acceptor_terminal_id = 6 [json_name = "cardAcceptorTerminalID"];
  // The name of the merchants business.
  string name = 7 [json_name = "name"];
  // Address of the merchant
  repeated string address_lines = 8 [json_name = "addressLines"];
  // The two or three letter state or region code.
  string region = 9 [json_name = "region"];
  // ISO 8583 three-digit currency classification code that identifies the national currency used at the merchant
  // location.
  string currency_code = 10 [json_name = "currencyCode"];
}
```

### message TransactionOutcome

```protobuf
message TransactionOutcome {
  // the vtc document ID which caused this notification.
  string ctc_document_id = 1 [json_name = "ctcDocumentID"];
  // This value will be set automatically when the decision response was returned. Value is in UTC time
  string decision_response_time_stamp = 2 [json_name = "decisionResponseTimeStamp"];
  // Used to determine whether the transaction was approved or declined Enum Values: * APPROVED , * DECLINED ,
  string transaction_approved = 3 [json_name = "transactionApproved"];
  // The responseCode due to which the transaction was declined
  string decline_response_code = 4 [json_name = "declineResponseCode"];
  // The alertNotification document ID created to the alert.
  string notification_id = 5 [json_name = "notificationID"];
  // The decisionID corresponding to the alerts.
  string decision_id = 6 [json_name = "decisionID"];
  // The list of alerts triggered for a transaction. Provides the list of alerts that should be delivered to the
  // cardholder as a result of the transaction. If the transaction was declined, then only a single alert detail
  // will be present, otherwise multiple alert conditions may have been triggered by the transaction.
  repeated AlertDetails alert_details = 7 [json_name = "alertDetails"];
}
```

### message AlertDetails

```protobuf
message AlertDetails {
  // The triggering app ID for the control that triggered the alert/decline
  string triggering_app_id = 1 [json_name = "triggeringAppID"];
  // The rule category for the control that triggered the alert/decline
  // Enum Values: * PCT_GLOBAL , * PCT_TRANSACTION , * PCT_MERCHANT
  string rule_category = 2 [json_name = "ruleCategory"];
  // The threshold amount for the control that triggered the alert/decline, only present if thresholdAmount breech
  // caused the trigger
  float threshold_amount = 3 [json_name = "thresholdAmount"];
  // The user identifier that defined the control, only present if defined on the control. The user identifier can be
  // correlated within the cardholder facing application to uniquely identify the user that defined the rule.
  string user_identifier = 4 [json_name = "userIdentifier"];
  // The reason for the alert, only present for alert rule details
  // Enum Values: * DECLINE_ALL , * DECLINE_BREACHED_AMT , * ALERT_BREACHED_AMT , * DECLINE_BY_ISSUER ,
  // * DECLINE_BY_SPEND_LIMIT , * ALERT_BREACHED_SPEND , * FX_ALERTS
  string alert_reason = 5 [json_name = "alertReason"];
  // The rule type for the control that triggered the alert/decline
  // Enum Values: * TCT_ATM_WITHDRAW , * TCT_AUTO_PAY , * TCT_BRICK_AND_MORTAR , * TCT_CROSS_BORDER , * TCT_E_COMMERCE ,
  // * TCT_CONTACTLESS , * TCT_PURCHASE_RETURN , * TCT_OCT , * MCT_ADULT_ENTERTAINMENT , * MCT_AIRFARE , * MCT_ALCOHOL ,
  // * MCT_APPAREL_AND_ACCESSORIES , * MCT_AUTOMOTIVE , * MCT_CAR_RENTAL , * MCT_ELECTRONICS , * MCT_SPORT_AND_RECREATION ,
  // * MCT_GAMBLING , * MCT_GAS_AND_PETROLEUM , * MCT_GROCERY , * MCT_HOTEL_AND_LODGING , * MCT_HOUSEHOLD ,
  // * MCT_PERSONAL_CARE , * MCT_SMOKE_AND_TOBACCO , * MCT_DINING ,
  string rule_type = 6 [json_name = "ruleType"];
  // The controlTargetType that defined the control Level that triggered the alert/decline.
  string control_target_type = 7 [json_name = "controlTargetType"];
  message UserInformation {
    repeated string application_defined_attributes = 1 [json_name = "applicationDefinedAttributes"];
    string banking_identifier = 2 [json_name = "bankingIdentifier"];
    string name = 3 [json_name = "name"];
  }
  UserInformation user_information = 8 [json_name = "userInformation"];
}
```

### message Response

```protobuf
message Response {}
```
