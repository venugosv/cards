# GetDetails

!!! Success

    An card will need ELIGIBILITY_GET_DETAILS eligibility for successful completion of this RPC.

!!! Danger "PCI Data"
    The data provided in the response payload is protected by the Payment Card Industry Data Security Standard. As such,
    this endpoint must not be consumed by bankers.

This API allows client to get card details such as card number in its plain form, cvc, expireTime etc

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetDetails | [GetDetailsRequest](#fabric.service.card.v1beta1.GetDetailsRequest) | [GetDetailsResponse](#fabric.service.card.v1beta1.GetDetailsResponse) | GetDetails returns unmasked attributes of a given tokenised card number In order to run the command, you need to replace the tokenizedCardNumber with one available in sit.

<a name="fabric.service.card.v1beta1.GetDetailsRequest"></a>

### GetDetailsRequest

GetDetailsRequest is the request payload for the CardAPI GetDetails endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tokenized_card_number | string |  | Tokenized string as the card number is required in the body |

```json
{
  "tokenizedCardNumber": "JFYR4hmKLWqmIr8IGoyQSmxYPqjXzoQQDvIlSM8ML3w"
}
```

<a name="fabric.service.card.v1beta1.GetDetailsResponse"></a>

### GetDetailsResponse

GetDetailsResponse is the response payload for the CardAPI GetDetails endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Name is the first and last name assigned to the card |
| card_number | string |  | Plain card number |
| cvc | string |  | Card verification code |
| expiry_date | [fabric.type.Date](#fabric.type.Date) |  | Card expiry date |
| eligibilities | [fabric.service.eligibility.v1beta1.Eligibility](#fabric.service.eligibility.v1beta1.Eligibility) | repeated | Possible operations that can be performed on this card |

```json
{
  "name": "Bernie Sanders",
  "cardNumber": "4645790062754063",
  "expiryTime": "2020-07-01T00:00:00Z",
  "cvc": "123",
  "eligibilities": [
    "ELIGIBILITY_APPLE_PAY",
    "ELIGIBILITY_GOOGLE_PAY",
    "ELIGIBILITY_SAMSUNG_PAY",
    "ELIGIBILITY_SET_PIN",
    "ELIGIBILITY_CHANGE_PIN",
    "ELIGIBILITY_CARD_REPLACEMENT_LOST",
    "ELIGIBILITY_CARD_REPLACEMENT_STOLEN",
    "ELIGIBILITY_CARD_REPLACEMENT_DAMAGED",
    "ELIGIBILITY_CARD_CONTROLS",
    "ELIGIBILITY_BLOCK"
  ]
}
```

### Example

```shell
grpcurl \
-H "env: $ENV" \
-H "service: cards" \
-H "Authorization: Bearer $TOKEN" \
-d "{\"tokenizedCardNumber\": \"$TOKENIZED_CARD_NUMBER\"}" \
fabric.gcpnp.anz:443 fabric.service.card.v1beta1.CardAPI/GetDetails
```

