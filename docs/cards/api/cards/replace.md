# Replace

!!! Success

    An card will need ELIGIBILITY_CARD_REPLACEMENT_LOST, ELIGIBILITY_CARD_REPLACEMENT_STOLEN or ELIGIBILITY_CARD_REPLACEMENT_DAMAGED eligibility for successful completion of this RPC.


Life happens, we lose things or sometimes they get damaged, even stolen. Fabric has your back. This RPC allows client to
request a replacement card for lost/stolen/damaged card

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Replace | [ReplaceRequest](#fabric.service.card.v1beta1.ReplaceRequest) | [ReplaceResponse](#fabric.service.card.v1beta1.ReplaceResponse) | Replace allows a user to trigger a card lifecycle event in case of lost/stolen/damaged

<a name="fabric.service.card.v1beta1.ReplaceRequest"></a>

### ReplaceRequest

ReplaceRequest is the request payload for the CardAPI Replace endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tokenized_card_number | string |  | Tokenized string as the card number is required in the body |
| reason | [ReplaceRequest.Reason](#fabric.service.card.v1beta1.ReplaceRequest.Reason) |  | Reason for replacement is required in the body with the value from the Reason enum set. |

<a name="fabric.service.card.v1beta1.ReplaceRequest.Reason"></a>

### ReplaceRequest.Reason

| Name | Number | Description |
| ---- | ------ | ----------- |
| REASON_UNKNOWN_UNSPECIFIED | 0 |  |
| REASON_LOST | 1 | Lost will generate a new card number and new physical card |
| REASON_STOLEN | 2 | Stolen will generate a new card number and new physical card |
| REASON_DAMAGED | 3 | Stolen will reuse the existing card number and only create a new physical card |

```json
{
  "tokenizedCardNumber": "4508439374353901",
  "reason": "REASON_LOST"
}
```

<a name="fabric.service.card.v1beta1.ReplaceResponse"></a>

### ReplaceResponse

ReplaceResponse is the response payload for the CardAPI Replace endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| new_tokenized_card_number | string |  | New tokenized card number as a result of the replacement |
| eligibilities | [fabric.service.eligibility.v1beta1.Eligibility](#fabric.service.eligibility.v1beta1.Eligibility) | repeated | Possible operations that can be performed on this card |

```json
{
  "newTokenizedCardNumber": "3006571927277477",
  "eligibilities": [
    "ELIGIBILITY_APPLE_PAY",
    "ELIGIBILITY_GOOGLE_PAY",
    "ELIGIBILITY_SAMSUNG_PAY",
    "ELIGIBILITY_CARD_ACTIVATION",
    "ELIGIBILITY_SET_PIN",
    "ELIGIBILITY_CARD_REPLACEMENT_LOST",
    "ELIGIBILITY_CARD_REPLACEMENT_STOLEN",
    "ELIGIBILITY_CARD_REPLACEMENT_DAMAGED",
    "ELIGIBILITY_CARD_CONTROLS"
  ]
}
```

## Example

```shell
grpcurl \
-H "env: $ENV" \
-H "service: cards" \
-H "Authorization: Bearer $TOKEN" \
-d "{\"tokenizedCardNumber\": \"$TOKENIZED_CARD_NUMBER\", \"reason\": \"$REASON\"}" \
fabric.gcpnp.anz:443 fabric.service.card.v1beta1.CardAPI/Replace
```
