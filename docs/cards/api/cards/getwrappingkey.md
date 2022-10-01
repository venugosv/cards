# GetWrappingKey

This API allows client to get a public key used by the salt SDK for set/change PIN on a card.

!!! info "SALT SDK Required"
    Consumers of this rpc will require the SDK by SALT to create the encrypted PIN Block. Find
    more information [here](https://github.com/anzx/fabric-cards/tree/master/docs/integration/salt.md)

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetWrappingKey | [GetWrappingKeyRequest](#fabric.service.card.v1beta1.GetWrappingKeyRequest) | [GetWrappingKeyResponse](#fabric.service.card.v1beta1.GetWrappingKeyResponse) | GetWrappingKey returns the wrapping key i.e. RSA public key&#39;s modulus and public exponent (hex encoded), SKI, X509 Certificate, etc Client can use this public key to encrypt pin block


<a name="fabric.service.card.v1beta1.GetWrappingKeyRequest"></a>

### GetWrappingKeyRequest

GetWrappingKeyRequest is the request payload for the CardAPI GetWrappingKey endpoint

<a name="fabric.service.card.v1beta1.GetWrappingKeyResponse"></a>

### GetWrappingKeyResponse

GetWrappingKeyResponse is the response payload for the CardAPI GetWrappingKey endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| encoded_key | string |  | Base64 encoded public key information in a format that can be interpreted by the Salt client side SDK. |

```json
{
  "encodedKey": "string"
}
```
### Example

```shell
grpcurl \
-H "env: $ENV" \
-H "service: cards" \
-H "Authorization: Bearer $TOKEN" \
fabric.gcpnp.anz:443 fabric.service.card.v1beta1.CardAPI/GetWrappingKey
```
