# SetPIN

!!! Success

    An new card will have ELIGIBILITY_SET_PIN eligibility when it has not set its PIN in the past. This is required for successful completion of this RPC.

This API allows client to set a pin on a card.

!!! info "SALT SDK Required"
    Consumers of this rpc will require the SDK by SALT to create the encrypted PIN Block. Find more information [here](https://github.com/anzx/fabric-cards/tree/master/docs/integration/salt.md)

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| SetPIN | [SetPINRequest](#fabric.service.card.v1beta1.SetPINRequest) | [SetPINResponse](#fabric.service.card.v1beta1.SetPINResponse) | SetPIN allows a user to select a new pin on a card, if the card is not active, it will be activated.


<a name="fabric.service.card.v1beta1.SetPINRequest"></a>

### SetPINRequest

SetPINRequest is the request payload for the CardAPI SetPIN endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tokenized_card_number | string |  | Tokenized string as the card number is required in the body |
| encrypted_pin_block | string |  | Encrypted PIN block is created using the client side Echidna package and is required in the request payload |


```json
{
  "tokenizedCardNumber": "9149004651839526",
  "encryptedPINBlock": "A4Tcl/3AJG3UIYCUWaVZ5U5yPC24Jf1Zxl+ShTjzDroP1EVcTSJJbe/pnuCvkxkWAX06KHsyX/tl9cc8C8eBe0+udApiehUe3DPLm2vL9JaLtc9UR7CDRN+Gk636M7MONKcRuiLVzOd8/rqPgxA9pbxdXlOPGg1eX01L5TJ0YbR/S7Pnhb8X8+V2zjmr86VqNajG7PuFg1ZJ2pSXCM82TAeB1YC2JQFJza3vtV09zEdT9zQLN81wYF7qk0wPFgaOYFGRheV9RBnK5ZjF32ak2XZXY0mrwmLDbxdSp3RNj8xJWSpTISWDRe/BOfazAkgRtdxrsqmk9etI81FCbQo9NA=="
}
```

<a name="fabric.service.card.v1beta1.SetPINResponse"></a>

### SetPINResponse

SetPINResponse is the response payload for the CardAPI SetPIN endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| eligibilities | [fabric.service.eligibility.v1beta1.Eligibility](#fabric.service.eligibility.v1beta1.Eligibility) | repeated | Possible operations that can be performed on this card |

```json
{
  "eligibilities": [
    "ELIGIBILITY_APPLE_PAY",
    "ELIGIBILITY_GOOGLE_PAY",
    "ELIGIBILITY_SAMSUNG_PAY",
    "ELIGIBILITY_CARD_ACTIVATION",
    "ELIGIBILITY_CHANGE_PIN",
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
-d "{\"tokenizedCardNumber\": \"$TOKENIZED_CARD_NUMBER\", \"encryptedPinBlock\": \"$PIN_BLOCK\"}" \
fabric.gcpnp.anz:443 fabric.service.card.v1beta1.CardAPI/SetPIN
```
