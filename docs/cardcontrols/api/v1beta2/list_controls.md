### List Controls

ListControls retrieves the list of Transaction Controls for all customers cards.

| Method Name | Request Type | Response Type |
| ----------- | ------------ | ------------- |
| ListControls | [ListControlsRequest](#ListControlsRequest) | [ListControlsResponse](#ListControlsResponse) |

### ListControlsRequest

ListControlsRequest is the request payload for CardControlsAPI ListControls endpoint

### ListControlsResponse

ListResponse is the response payload for CardControlsAPI ListControls endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| card_controls | [CardControlResponse](./types.md#CardControlResponse) | repeated | returns a CardControlResponse for each cards transaction control document |

```json
{
  "card_controls": [
    {
      "tokenized_card_number": "string",
      "card_controls": [
        {
          "control_type": "UNKNOWN_UNSPECIFIED",
          "impulse_delay_start": "2022-03-14T11:59:45.028Z",
          "impulse_delay_period": "string"
        }
      ]
    }
  ]
}
```
