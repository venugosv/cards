# Activate

!!! Success

    An **inactive** card will have ELIGIBILITY_CARD_ACTIVATION. This is required for successful completion of this RPC.

The activation RPC is expected to be used at the beginning of the card lifecycle. A card is issued in an inactive state
and cannot begin spending out of its linked account until it is activated. We require the `last6Digits` as user input as
proof of physical possession to prevent inadvertent unauthorized spend.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Activate | [ActivateRequest](#fabric.service.card.v1beta1.ActivateRequest) | [ActivateResponse](#fabric.service.card.v1beta1.ActivateResponse) | Activate activates an inactive card

<a name="fabric.service.card.v1beta1.ActivateRequest"></a>

### ActivateRequest

ActivateRequest is the request payload for the CardAPI Activate endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tokenized_card_number | string |  | Tokenized string as the card number is required in the body |
| last_6_digits | string |  | Last 6 digits of the non-tokenized card number is required in the body |

```json
{
  "tokenizedCardNumber": "string",
  "last6Digits": "string"
}
```

<a name="fabric.service.card.v1beta1.ActivateResponse"></a>

### ActivateResponse

ActivateResponse is the response payload for the CardAPI Activate endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| eligibilities | [fabric.service.eligibility.v1beta1.Eligibility](#fabric.service.eligibility.v1beta1.Eligibility) | repeated | Possible operations that can be performed on this card |

```json
{
  "eligibility": [
    "ELIGIBILITY_APPLE_PAY",
    "ELIGIBILITY_GOOGLE_PAY",
    "ELIGIBILITY_SAMSUNG_PAY",
    "ELIGIBILITY_CHANGE_PIN",
    "ELIGIBILITY_CARD_REPLACEMENT_LOST",
    "ELIGIBILITY_CARD_REPLACEMENT_STOLEN",
    "ELIGIBILITY_CARD_REPLACEMENT_DAMAGED",
    "ELIGIBILITY_CARD_CONTROLS",
    "ELIGIBILITY_BLOCK",
    "ELIGIBILITY_GET_DETAILS"
  ]
}
```

### Example

```shell
grpcurl \
-H "env: $ENV" \
-H "service: cards" \
-H "Authorization: Bearer $TOKEN" \
-d "{\"tokenizedCardNumber\": \"$TOKENIZED_CARD_NUMBER\", \"last6Digits\": \"$LAST_6_DIGITS\"}" \
fabric.gcpnp.anz:443 fabric.service.card.v1beta1.CardAPI/Activate
```
