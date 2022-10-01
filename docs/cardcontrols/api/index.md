# CardControlsAPI

Interfaces with Visa Consumer Transaction Controls and allows cardholders to place restrictions on their enrolled cards
that define when, where and how those cards can be used. Cardholders can turn their cards on and off or can restrict
their use in certain situations using a number of different triggers and thresholds.

## v1beta2

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListControls | [ListControlsRequest](./v1beta2/list_controls.md#ListControlsRequest) | [ListControlsResponse](./v1beta2/list_controls.md#ListControlsResponse) | ListControls retrieves the list of Transaction Controls for all customers cards. |
| QueryControls | [QueryControlsRequest](./v1beta2/query_controls.md#QueryControlsRequest) | [CardControlResponse](./v1beta2/./types.md#CardControlResponse) | QueryControls retrieves the list of Controls for a given tokenized card number |
| SetControls | [SetControlsRequest](./v1beta2/set_controls.md#SetControlsRequest) | [CardControlResponse](./v1beta2/./types.md#CardControlResponse) | SetControls to add or update control(s) for a given tokenized card number |
| RemoveControls | [RemoveControlsRequest](./v1beta2/remove_controls.md#RemoveControlsRequest) | [CardControlResponse](./v1beta2/./types.md#CardControlResponse) | RemoveControls to remove control(s) for a given tokenized card number |
| TransferControls | [TransferControlsRequest](./v1beta2/transfer_controls.md#TransferControlsRequest) | [TransferControlsResponse](./v1beta2/transfer_controls.md#TransferControlsResponse) | TransferControls If a card becomes lost/stolen then its existing control document can be updated with the new account number. Replace the tokenizedCardNumber of a control document with a new tokenizedCardNumber (Card Replacement) |
| BlockCard | [BlockCardRequest](./v1beta2/block_card.md#BlockCardRequest) | [BlockCardResponse](./v1beta2/block_card.md#BlockCardResponse) | BlockCard Provides an endpoint to completely block all functionality of a card temporarily |

## v1beta1

| Method Name | Description |
| ----------- | ------------|
| [Block](./v1beta1/block.md) | DEPRECATED |
| [Disenrol](./v1beta1/enrol.md) | DEPRECATED |
| [Enrol](./v1beta1/enrol.md) | DEPRECATED |
| [List](./v1beta1/list.md) | DEPRECATED |
| [Query](./v1beta1/query.md) | DEPRECATED |
| [Remove](./v1beta1/remove.md) | DEPRECATED |
