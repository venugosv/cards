# VerifyPIN

!!! Warning "Deprecated"
    This endpoint is deprecated and closed behind feature toggle. Do not use it.


This API allows client to verify a pin.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| VerifyPIN | [VerifyPINRequest](#fabric.service.card.v1beta1.VerifyPINRequest) | [VerifyPINResponse](#fabric.service.card.v1beta1.VerifyPINResponse) | VerifyPIN verifies if the PIN is correct.

<a name="fabric.service.card.v1beta1.VerifyPINRequest"></a>

### VerifyPINRequest

VerifyPINRequest is the request payload for the CardAPI VerifyPIN endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tokenized_card_number | string |  | Tokenized string as the card number is required in the body |
| encrypted_pin_block | string |  | Encrypted PIN block is created using the client side Echidna package and is required in the request payload |

<a name="fabric.service.card.v1beta1.VerifyPINResponse"></a>

### VerifyPINResponse

VerifyPINResponse is the response payload for the CardAPI VerifyPIN endpoint, indicates if PIN is correct or not

<a name="fabric.service.card.v1beta1.ReplaceRequest.Reason"></a>
