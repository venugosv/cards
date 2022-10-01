### Query Controls

QueryControls retrieves the list of Controls for a given tokenized card number

| Method Name | Request Type | Response Type |
| ----------- | ------------ | ------------- |
| QueryControls | [QueryControlsRequest](#QueryControlsRequest) | [CardControlResponse](#CardControlResponse) |

### QueryControlsRequest

QueryControlsRequest is the request payload for the CardControlsAPI QueryControls endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tokenized_card_number | string |  | Tokenized string as the card number is required in the body |


```json
{
  "tokenizedCardNumber": "string"
}
```

### CardControlResponse

CardControlResponse is the response payload for the CardControlsAPI QueryControls, SetControls &amp; RemoveControls
endpoints

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tokenized_card_number | string |  | A tokenized card number for which the response is related to |
| card_controls | [CardControl](./types.md#CardControl) | repeated | A list of controls currently active on the cards transaction control document |

```json
{
  "tokenized_card_number": "string",
  "card_controls": [
    {
      "control_type": "UNKNOWN_UNSPECIFIED",
      "impulse_delay_start": "2022-03-14T12:00:29.061Z",
      "impulse_delay_period": "string"
    }
  ]
}
```
