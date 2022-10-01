# CardAPI

CardAPI provides an interface for debit card management functionality. Allowing users to control all aspects of their
physical card lifecycle. From activation, selecting a PIN and replacing the card, this API will provide end to end self
service.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| [List](./list.md) | ListRequest | ListResponse | List returns all card on which the given persona_id is able to perform READ operations.
| [GetDetails](./getdetails.md) | GetDetailsRequest | GetDetailsResponse | GetDetails returns unmasked attributes of a given tokenised card number In order to run the command, you need to replace the tokenizedCardNumber with one available in sit.
| [Activate](./activate.md) | ActivateRequest | ActivateResponse | Activate activates an inactive card
| [GetWrappingKey](./getwrappingkey.md) | GetWrappingKeyRequest | GetWrappingKeyResponse | GetWrappingKey returns the wrapping key i.e. RSA public key&#39;s modulus and public exponent (hex encoded), SKI, X509 Certificate, etc Client can use this public key to encrypt pin block
| [ResetPIN](./resetPIN.md) | ResetPINRequest | ResetPINResponse | ResetPIN allows a user to reset a new pin on a card without providing the old pin.
| [SetPIN](./setPIN.md) | SetPINRequest | SetPINResponse | SetPIN allows a user to select a new pin on a card, if the card is not active, it will be activated.
| [VerifyPIN](./verifyPIN.md) | VerifyPINRequest | VerifyPINResponse | VerifyPIN verifies if the PIN is correct.
| [ChangePIN](./changePIN.md) | ChangePINRequest | ChangePINResponse | ChangePIN allows a user to change a pin on a card.
| [Replace](./replace.md) | ReplaceRequest | ReplaceResponse | Replace allows a user to trigger a card lifecycle event in case of lost/stolen/damaged
| [AuditTrail](./audittrail.md) | AuditTrailRequest | AuditTrailResponse | Audit Trail

