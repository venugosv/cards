# CreateGooglePaymentToken

With Push Provisioning, issuers can easily develop their own fully functional payment apps. This lets issuers provision
cards and perform common actions like setting the default payment token and deleting a token from the wallet. Once
integrated, customers can use their credit or debit cards to pay in apps and in stores with an NFC-enabled device.

When users tap the push provisioning button, they'll see a screen that asks them for confirmation before provisioning
their card. Once they confirm, users will need to accept their issuer’s terms of service before their card is added to
the device. They will see a confirmation screen before they return to the issuer app.

| Method Name | Request Type | Response Type |
| ----------- | ------------ | ------------- |
| CreateGooglePaymentToken | [CreateGooglePaymentTokenRequest](#CreateGooglePaymentTokenRequest) | [CreateGooglePaymentTokenResponse](#CreateGooglePaymentTokenResponse) |

### CreateGooglePaymentTokenRequest

CreateGooglePaymentTokenRequest is the request payload for CreateGooglePaymentToken, containing the tokenized card
number, desired network & a stable hardware ID of the requesting device.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tokenized_card_number` | string |  | This is a tokenized string representing an encrypted card fpan |
| `card_network` | [CardNetwork](#CardNetwork) |  | The card payment network |

```json
{
  "tokenizedCardNumber": "string",
  "cardNetwork": "CARD_NETWORK_UNSPECIFIED"
}
```

### CreateGooglePaymentTokenResponse

CreateGooglePaymentTokenResponse is the response payload for CreateGooglePaymentToken, containing an Opaque payment
card, the encrypted object exchanged with TSPs for a token and the address of the user to be stored on file. We also
return several enumerated constant values as for use in the TapAndPay namespace, these include the token provider & card
payment network.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `opaque_payment_card` | `[]byte`  |  | The Opaque Payment card (OPC) is base64 encoded, encrypted object that's created by the issuer and passed to Google Pay during push provisioning. |
| `token_provider` | [TokenProvider](#TokenProvider)  |  | The token provider specifies the tokenization service used to create a given token |
| `card_network` | [CardNetwork](#CardNetwork)  |  | The card payment network |
| `user_address` | [Address](#Address)  |  | The user’s address. |

```json
{
  "opaquePaymentCard": "",
  "tokenProvider": "TOKEN_PROVIDER_UNSPECIFIED",
  "cardNetwork": "CARD_NETWORK_UNSPECIFIED",
  "userAddress": {
    "lineOne": "string",
    "lineTwo": "string",
    "countryCode": "string",
    "locality": "string",
    "administrativeArea": "string",
    "name": "string",
    "phoneNumber": "string",
    "postalCode": "string"
  }
}
```

## Example

```shell
grpcurl \
-H "env: $ENV" \
-H "service: cards" \
-H "Authorization: Bearer $TOKEN" \
-d "" \
fabric.gcpnp.anz:443 fabric.service.card.v1beta1.WalletAPI/CreateGooglePaymentToken
```

## Types

These are the messages/enums referenced in the Requests/Responses in this API.

### Messages

#### Address

https://github.com/anzx/fabricapis/blob/master/proto/fabric/service/card/v1beta1/wallet_api.proto

|Field|Type|Required|
|---|---|---|
| `line_one` | `String` |  |
| `line_two` | `String` |  |
| `country_code` | `String` |  |
| `locality` | `String` |  |
| `administrative_area` | `String` |  |
| `name` | `String` |  |
| `phone_number` | `String` |  |
| `postal_code` | `String` |  |

### Enums

#### CardNetwork

https://github.com/anzx/fabricapis/blob/master/proto/fabric/service/card/v1beta1/wallet_api.proto

|Value|
|---|
| `CARD_NETWORK_UNSPECIFIED` |
| `CARD_NETWORK_AMEX` |
| `CARD_NETWORK_DISCOVER` |
| `CARD_NETWORK_MASTERCARD` |
| `CARD_NETWORK_QUICPAY` |
| `CARD_NETWORK_PRIVATE_LABEL` |
| `CARD_NETWORK_VISA` |
| `CARD_NETWORK_MIR` |

#### TokenProvider

https://github.com/anzx/fabricapis/blob/master/proto/fabric/service/card/v1beta1/wallet_api.proto

|Value|
|---|
| `TOKEN_PROVIDER_UNSPECIFIED` |
| `TOKEN_PROVIDER_AMEX` |
| `TOKEN_PROVIDER_DISCOVER` |
| `TOKEN_PROVIDER_JCB` |
| `TOKEN_PROVIDER_MASTERCARD` |
| `TOKEN_PROVIDER_VISA` |
| `TOKEN_PROVIDER_MIR` |
