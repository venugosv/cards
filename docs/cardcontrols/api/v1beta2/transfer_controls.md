### Transfer Controls

TransferControls If a card becomes lost/stolen then its existing control document can be updated with the new account number. Replace the tokenizedCardNumber of a control document with a new tokenizedCardNumber (Card Replacement)

| Method Name | Request Type | Response Type |
| ----------- | ------------ | ------------- |
| TransferControls | [TransferControlsRequest](#TransferControlsRequest) | [TransferControlsResponse](#TransferControlsResponse) |

### TransferControlsRequest

TransferControlsRequest is the request payload for the CardControlsAPI TransferControls endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| current_tokenized_card_number | string |  | Tokenized card number string for the existing card is required |
| new_tokenized_card_number | string |  | Tokenized card number string for the new card is required |

```json
{
  "currentTokenizedCardNumber": "string",
  "newTokenizedCardNumber": "string"
}
```
### TransferControlsResponse

TransferControlsResponse is the response payload for the CardControlsAPI TransferControls endpoint, will confirm if the
request is successful
