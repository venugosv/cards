# List

You get a card... you get a card, everyone gets a card! Start right here, at the beginning of cards... list all cards
associated to a given `personaID`. No need for a request payload, simply login and hit us up
at `fabric.service.card.v1beta1.CardAPI/List`. Its worth nothing that a single persona can have many cards, so you need
to be prepared for an array to be returned.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| List | [ListRequest](#fabric.service.card.v1beta1.ListRequest) | [ListResponse](#fabric.service.card.v1beta1.ListResponse) | List returns all card on which the given persona_id is able to perform READ operations.

<a name="fabric.service.card.v1beta1.ListRequest"></a>

### ListRequest

ListRequest is the request payload for the CardAPI List endpoint

<a name="fabric.service.card.v1beta1.ListResponse"></a>

### ListResponse

ListResponse is the response payload for the CardAPI List endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cards | [Card](#fabric.service.card.v1beta1.Card) | repeated | List of cards, and their details is returned for all entitled and eligible cards associated with a personaID |

```json
{
  "cards": [
    {
      "name": "ARYANNA KIHN",
      "tokenizedCardNumber": "3006571927277477",
      "last4Digits": "9137",
      "status": "Issued",
      "expiryDate": {
        "year": {
          "value": 2020
        },
        "month": {
          "value": 10
        }
      },
      "accountNumbers": [
        "210418704"
      ],
      "eligibilities": [
        "ELIGIBILITY_APPLE_PAY",
        "ELIGIBILITY_GOOGLE_PAY",
        "ELIGIBILITY_SAMSUNG_PAY",
        "ELIGIBILITY_CARD_ACTIVATION",
        "ELIGIBILITY_SET_PIN",
        "ELIGIBILITY_CARD_REPLACEMENT_LOST",
        "ELIGIBILITY_CARD_REPLACEMENT_STOLEN",
        "ELIGIBILITY_CARD_REPLACEMENT_DAMAGED",
        "ELIGIBILITY_CARD_CONTROLS"
      ],
      "wallets": {
      }
    },
    {
      "name": "ARYANNA KIHN",
      "tokenizedCardNumber": "4508439374353901",
      "last4Digits": "3952",
      "status": "Lost",
      "expiryDate": {
        "year": {
          "value": 2020
        },
        "month": {
          "value": 9
        }
      },
      "accountNumbers": [
        "210418704"
      ],
      "eligibilities": [
        "ELIGIBILITY_CARD_REPLACEMENT_LOST"
      ],
      "newTokenizedCardNumber": "3006571927277477",
      "wallets": {
      }
    }
  ]
}
```

<a name="fabric.service.card.v1beta1.Card"></a>

### Card

Card is a debit card.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Name is the first and last name assigned to the card |
| tokenized_card_number | string |  | Tokenized card number is an encrypted reference to the original card number that is approved to be stored |
| last_4_digits | string |  | The last 4 digits of the detokenized card number |
| status | string |  | As defined in our core system, what state the card is in |
| expiry_date | [fabric.type.Date](#Date) |  | Card expiry date |
| account_numbers | string | repeated | Linked account numbers the card is configured to spend from |
| eligibilities | [fabric.service.eligibility.v1beta1.Eligibility](#fabric.service.eligibility.v1beta1.Eligibility) | repeated | Possible operations that can be performed on this card |
| new_tokenized_card_number | string |  | New card token if present if this card has been replaced(Lost/stolen) |
| wallets | [Wallets](#Wallets) |  | Shows the number of times this card has been tokenized into each type of digital Wallet |
| card_controls_enabled | [bool](#bool) |  | Indicates if a visa card control exists against the card |

<a name="fabric.service.card.v1beta1.Limit"></a>

### Limit

Limit provides information about a cards daily limit

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| daily_limit | string |  | Daily limit |
| daily_limit_available | string |  | Available Limit (from the Daily Limit) |
| last_transaction | [fabric.type.Date](#Date) |  | Date of last transaction processed |
| type | string |  | Information the Type of Limit set on the Card, such as POS, or ATM |

<a name="fabric.service.card.v1beta1.MaskedCard"></a>

### MaskedCard

MaskedCard provides basic identifiable information about a card

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tokenized_card_number | string |  | Tokenized card number is an encrypted reference to the original card number that is approved to be stored |
| last_4_digits | string |  | The last 4 digits of the detokenized card number |

<a name="fabric.service.card.v1beta1.Wallets"></a>

### Wallets

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| other | uint32 |  | Number of devices registered to the &#39;other&#39; wallet. |
| fitness | uint32 |  | Number of devices registered to the &#39;Fitness&#39; wallet (eg. FitBit, Garmin, etc) |
| apple_pay | uint32 |  | Number of devices registered to the &#39;ApplePay&#39; wallet |
| e_commerce | uint32 |  | Number of devices registered to the &#39;eCommerce wallet (eg. NetFlix, PayPal, etc – essentially merchant that use a token instead of keeping Card on File)&#39; |
| samsung_pay | uint32 |  | Number of devices registered to the &#39;SamsungPay&#39; wallet |
| google_pay | uint32 |  | Number of devices registered to the &#39;GooglePay&#39; wallet |

<a name="fabric/service/card/v1beta1/card_api.proto"></a>

<a name="Date"></a>

### Date

Represents a whole or partial calendar date, such as a birthday. The time of day and time zone are either specified
elsewhere or are insignificant. The date is relative to the Gregorian Calendar. This can represent one of the following:

* A full date, with non-zero year, month, and day values
* A month and day value, with a optional year, such as an anniversary
* A year on its own, with optional month and day values
* A year and month value, with a optional day, such as a credit card expiration date

Related types are `google.type.TimeOfDay` and `google.protobuf.Timestamp`.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| year | [OptionalInt32](#OptionalInt32) |  | Year of the date. Must be from 1 to 9999, or optional to specify a date without a year. |
| month | [OptionalInt32](#OptionalInt32) |  | Month of a year. Must be from 1 to 12, or optional to specify a year without a month and day. |
| day | [OptionalInt32](#OptionalInt32) |  | Day of a month. Must be from 1 to 31 and valid for the year and month, or optional to specify a year by itself or a year and month where the day isn&#39;t significant. |

<a name="MaskedCard"></a>
<a name="fabric.type.OptionalInt32"></a>

### OptionalInt32

Wrapper message for `int32`.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | int32 |  |  |

<p align="right"><a href="#top">Top</a></p>

## Example

```shell
grpcurl \
-H "env: $ENV" \
-H "service: cards" \
-H "Authorization: Bearer $TOKEN" \
fabric.gcpnp.anz:443 fabric.service.card.v1beta1.CardAPI/List
```
