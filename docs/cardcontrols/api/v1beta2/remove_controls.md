### Remove Controls

RemoveControls to remove control(s) for a given tokenized card number

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| RemoveControls | [RemoveControlsRequest](#RemoveControlsRequest) | [CardControlResponse](#CardControlResponse) |

```json
{
  "tokenizedCardNumber": "string",
  "cardControls": [
    {
      "controlType": "UNKNOWN_UNSPECIFIED"
    }
  ]
}
```

### RemoveControlsRequest

RemoveControls is the request payload for the CardControlsAPI RemoveControls endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tokenized_card_number | string |  | Tokenized string as the card number is required in the body |
| control_types | [ControlType](./types.md#ControlType) | repeated | List of ControlType that intend to be removed, also required |

```json
{
  "tokenizedCardNumber": "3902445551647359",
  "controlTypes": [
    "TCT_E_COMMERCE"
  ]
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
