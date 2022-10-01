# AuditTrail

!!! Warning "Internal use only"
    Data contained in the response is sensitive and could result in supporting fraudulent activity and compromise a
    customers safety. Please do not share these details directly with the customer.

This rpc method allows client to view audit history for a card that a user owns.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| AuditTrail | [AuditTrailRequest](#fabric.service.card.v1beta1.AuditTrailRequest) | [AuditTrailResponse](#fabric.service.card.v1beta1.AuditTrailResponse) | Audit Trail

<a name="fabric.service.card.v1beta1.AuditTrailRequest"></a>

### AuditTrailRequest

AuditTrailRequest is the request payload for the CardAPI AuditTrail endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tokenized_card_number | string |  | Tokenized string as the card number is required in the body |

```json
{
  "tokenizedCardNumber": "string"
}
```

<a name="fabric.service.card.v1beta1.AuditTrailResponse"></a>

### AuditTrailResponse

AuditTrailResponse is the response payload for the CardAPI AuditTrail endpoint

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| accounts_linked | int64 |  | Number of Accounts linked |
| total_cards | int64 |  | Total number of cards |
| activated | bool |  | Can be one of the following Activation Status of the Card: | Code | Description | | ---- | ---- | | true | Card is activated | | false | Card is inactive,needs to be activated | |
| card_control_enabled | bool |  | When set to true, it indicates customer has enabled visa card control to better manage their finance. This field will also be included in CTM Non-Monetary events published to CIM, so that it can be passed on to CAM and Falcon. Updates to the indicator will also be logged by CTM to be passed to Base24 every 15 minutes in the ‘trickle feed’ process. |
| merchant_update_enabled | bool |  | When this field is set to true (i.e. Opted-In), it means the customer is allowing Visa to pass on their updated card number and expiry date information to participating merchants, so any payments set up from this card at registered Merchants are not interrupted. Recurring bills will continue, without the need for the customer to advise them. The customer however, has the option to ‘Opt-out’, by contacting ANZ, where-by this new indicator will need to be updated to false. |
| replaced_date | [Date](#Date) |  | Card replacement date |
| replacement_count | int64 |  | Number of times the card has been replaced |
| issue_date | [Date](#Date) |  | Card issue date |
| reissue_date | [Date](#Date) |  | Card reissue date |
| expiry_date | [Date](#Date) |  | Card Expiry Date |
| previous_expiry_date | [Date](#Date) |  | Previous cards expiry date |
| details_changed_date | [Date](#Date) |  | Card Details Changed date |
| closed_date | [Date](#Date) |  | The closed date of card |
| limits | [Limit](#Limit) | repeated |  |
| new_card | [MaskedCard](#MaskedCard) |  | New card if this card has been replaced |
| old_card | [MaskedCard](#MaskedCard) |  | Old card if this card is a replacement of a previously issued card |
| pin_change_date | [Date](#Date) |  | Last PIN change date |
| pin_change_count | int64 |  | Number of times the PIN has been changed |
| last_pin_failed | [Date](#Date) |  | Last PIN fail date |
| pin_failed_count | int64 |  | Number of times the PIN has failed |
| status | string |  | Can be one of the following Status Codes: | Code | | ---- | | Closed | | Delinquent (Return Card) | | Delinquent (Retain Card) | | Issued | | Lost | | Stolen | | Unissued (N&amp;D ICI Cards) | | Temporary Block | | Block ATM | | Block ATM &amp; POS (Exclude CNP)| | Block ATM, POS, CNP &amp; BCH | | Block ATM, POS &amp; CNP | | Block CNP | | Block POS (exclude CNP) | |
| status_changed_date | [Date](#Date) |  | Status Changed date |

```json
{
  "accountsLinked": "1",
  "totalCards": "39",
  "activated": true,
  "merchantUpdateEnabled": true,
  "issueDate": "2016-07-18T00:00:00Z",
  "expiryDate": "2020-07-01T00:00:00Z",
  "detailsChangedDate": "2016-08-01T00:00:00Z",
  "limits": [
    {
      "dailyLimit": "1000",
      "dailyLimitAvailable": "1000",
      "type": "POS"
    },
    {
      "dailyLimit": "1000",
      "dailyLimitAvailable": "1000",
      "type": "ATM"
    }
  ],
  "newCard": {
  },
  "oldCard": {
  },
  "status": "Issued",
  "statusChangedDate": "2016-08-01T00:00:00Z"
}
```

<a name="fabric/type/date.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## Other Types

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
| year | [OptionalInt32](#fabric.type.OptionalInt32) |  | Year of the date. Must be from 1 to 9999, or optional to specify a date without a year. |
| month | [OptionalInt32](#fabric.type.OptionalInt32) |  | Month of a year. Must be from 1 to 12, or optional to specify a year without a month and day. |
| day | [OptionalInt32](#fabric.type.OptionalInt32) |  | Day of a month. Must be from 1 to 31 and valid for the year and month, or optional to specify a year by itself or a year and month where the day isn&#39;t significant. |

<a name="MaskedCard"></a>

### MaskedCard

MaskedCard provides basic identifiable information about a card

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tokenized_card_number | string |  | Tokenized card number is an encrypted reference to the original card number that is approved to be stored |
| last_4_digits | string |  | The last 4 digits of the detokenized card number |

<a name="fabric.type.OptionalInt32"></a>

### OptionalInt32

Wrapper message for `int32`.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | int32 |  |  |

### Limit

Limit provides information about a cards daily limit

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| daily_limit | string |  | Daily limit |
| daily_limit_available | string |  | Available Limit (from the Daily Limit) |
| last_transaction | [Date](#Date) |  | Date of last transaction processed |
| type | string |  | Information the Type of Limit set on the Card, such as POS, or ATM |


## Example

```shell
grpcurl \
-H "env: $ENV" \
-H "service: cards" \
-H "Authorization: Bearer $TOKEN" \
-d "{\"tokenizedCardNumber\": \"$TOKENIZED_CARD_NUMBER\"}" \
fabric.gcpnp.anz:443 fabric.service.card.v1beta1.CardAPI/AuditTrail
```
