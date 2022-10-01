# ChangePIN

!!! Warning "Deprecated"
    This endpoint is deprecated and closed behind feature toggle. Do not use it. Use [ResetPIN](./resetPIN.md) instead

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ChangePIN | [ChangePINRequest](#fabric.service.card.v1beta1.ChangePINRequest) | [ChangePINResponse](#fabric.service.card.v1beta1.ChangePINResponse) | ChangePIN allows a user to change a pin on a card.

<a name="fabric.service.card.v1beta1.ChangePINRequest"></a>

### ChangePINRequest

ChangePINRequest is the request payload for the CardAPI ChangePIN endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tokenized_card_number | string |  | Tokenized string as the card number is required in the body |
| encrypted_pin_block_old | string |  | Encrypted PIN block is created using the client side Echidna package for the existing pin and is required in the request payload |
| encrypted_pin_block_new | string |  | Encrypted PIN block is created using the client side Echidna package for the new pin and is required in the request payload |

<a name="fabric.service.card.v1beta1.ChangePINResponse"></a>

### ChangePINResponse

ChangePINResponse is the response payload for the CardAPI ChangePIN endpoint
