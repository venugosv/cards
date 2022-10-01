### Block Card

BlockCard Provides an endpoint to completely block all functionality of a card temporarily

| Method Name | Request Type | Response Type |
| ----------- | ------------ | ------------- |
| BlockCard | [BlockCardRequest](#BlockCardRequest) | [BlockCardResponse](#BlockCardResponse)  |

### BlockCardRequest

BlockCardRequest is the request payload for the CardControlsAPI BlockCard endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tokenized_card_number | string |  | Tokenized string as the card number is required in the body |
| action | [BlockCardRequest.Action](#BlockCardRequest.Action) |  | Action for block is required in the body with the value from the Action enum set. |

```json
{
  "tokenizedCardNumber": "string",
  "action": "ACTION_BLOCK"
}
```

### BlockCardResponse

BlockCardResponse is the response payload for the CardControlsAPI BlockCard endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| eligibilities | [fabric.service.eligibility.v1beta1.Eligibility](https://backstage.fabric.gcpnp.anz/docs/default/API/fabric.service.eligibility.v1beta1.CardEligibilityAPI/api/can/) | repeated | Possible operations that can be performed on this card |

```json
{
  "eligibilities": [
    "ELIGIBILITY_UNBLOCK"
  ]
}
```

### BlockCardRequest.Action

| Name | Number | Description |
| ---- | ------ | ----------- |
| ACTION_UNKNOWN_UNSPECIFIED | 0 |  |
| ACTION_BLOCK | 1 | Block will block an unblocked card in the ANZ Core |
| ACTION_UNBLOCK | 2 | Unblock will unblock a blocked card in the ANZ Core |

