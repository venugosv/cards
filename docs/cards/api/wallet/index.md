# WalletAPI

WalletAPI provides an interface to generate digital payment tokens. Apple Pay In-App Provisioning provides a credit or
debit card issuer the ability to initiate the card provisioning process for Apple Pay directly from the issuer’s iOS
app. Cardholders will find the In-App Provisioning feature an extremely convenient method to provision their payment
details into their iOS devices by avoiding the need to input card details manually. Issuers will also find In-App
Provisioning an effective component of a seamless mobile banking experience. By driving the provisioning of cards via
their iOS mobile apps, issuers can create a unified interface for card provisioning and their other banking services.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateApplePaymentToken | [CreateApplePaymentTokenRequest](./apple.md#CreateApplePaymentTokenRequest) | [CreateApplePaymentTokenResponse](./apple.md#CreateApplePaymentTokenResponse) | CreateApplePaymentToken generates the payload, OTP and key that Apple require to put a payment token into the Apple Wallet for in-app provisioning. This service will prepare the payment data payload for the user, generate an ephemeral key pair &amp; encrypt the payload with a shared key derived from the Apple public certificates and generated private ephemeral key. Then it will deliver the encrypted payload and ephemeral public key back to the app. The issuer host will also generate a cryptographic OTP per the Payment Network Operator (PNO) or service provider specifications and pass that to the iOS app as well |
| CreateGooglePaymentToken | [CreateGooglePaymentTokenRequest](./google.md#CreateGooglePaymentTokenRequest) | [CreateGooglePaymentTokenResponse](./google.md#CreateGooglePaymentTokenResponse) |
