# Types

| Type | Description |
| ------------- | ------------|
| [CardControl](#CardControl) | CardControl is an object containing all related attributes of a single control
| [ControlRequest](#ControlRequest) | ControlRequest defines the configurable attributes of a request to set a new or existing control
| [ControlType](#ControlType) | ControlType are the different categories of controls available in visa consumer transaction controls API. Three ControlType categories exist; Transaction, Merchant & Global
| [Duration](#Duration) | The JSON representation for Duration is a String that ends in s to indicate seconds and is preceded by the number of seconds, with nanoseconds expressed as fractional seconds.
| [Timestamp](#Timestamp) | A Timestamp represents a point in time independent of any time zone or calendar, represented as seconds and fractions of seconds at nanosecond resolution in UTC Epoch time.

### CardControl

CardControl is an object containing all related attributes of a single control

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| control_type | [ControlType](#ControlType) |  | ControlType exists as an element of the ControlType enum set |
| impulse_delay_start | [google.protobuf.Timestamp](#Timestamp) |  | Used only for gambling block, defined the impulse delay start time |
| impulse_delay_period | [google.protobuf.Duration](#Duration) |  | Used only for gambling block, defined the impulse delay period |

### ControlRequest

ControlRequest defines the configurable attributes of a request to set a new or existing control

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| control_type | [ControlType](#ControlType) |  | ControlType exists as an element of the ControlType enum set |

### ControlType

ControlType are the different categories of controls available in visa consumer transaction controls api Three
ControlType categories exist; Transaction, Merchant & Global

| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN_UNSPECIFIED | 0 |  |
| TCT_ATM_WITHDRAW | 1 | Used for ATM cash withdrawals. |
| TCT_AUTO_PAY | 2 | Used for recurring or installment payment transactions. |
| TCT_BRICK_AND_MORTAR | 3 | Used for card present transactions. |
| TCT_CROSS_BORDER | 4 | Used when the merchant and issuer’s country code do not match for a card-present transaction.|
| TCT_E_COMMERCE | 5 | Used for card-not-present transactions performed at e-commerce and mail order/telephone order (MOTO) merchants |
| TCT_CONTACTLESS | 6 | Used for contactless purchases in a card present environment. |
| MCT_ADULT_ENTERTAINMENT | 7 | VTC supports a number of merchant control categories (see Customer Rules API MerchantControls child attributes.  Enabling a merchant card control will trigger a VTC response whenever a purchase is made at a merchant with a corresponding MCC. |
| MCT_AIRFARE | 8 |  |
| MCT_ALCOHOL | 9 |  |
| MCT_APPAREL_AND_ACCESSORIES | 10 |  |
| MCT_AUTOMOTIVE | 11 |  |
| MCT_CAR_RENTAL | 12 |  |
| MCT_ELECTRONICS | 13 |  |
| MCT_SPORT_AND_RECREATION | 14 |  |
| MCT_GAMBLING | 15 |  |
| MCT_GAS_AND_PETROLEUM | 16 |  |
| MCT_GROCERY | 17 |  |
| MCT_HOTEL_AND_LODGING | 18 |  |
| MCT_HOUSEHOLD | 19 |  |
| MCT_PERSONAL_CARE | 20 |  |
| MCT_SMOKE_AND_TOBACCO | 21 |  |
| GCT_GLOBAL | 22 | Card-level rules, imposed on all transactions  (e.g. Card On/Off).|


### CardControlResponse

CardControlResponse is the response payload for the CardControlsAPI QueryControls, SetControls &amp; RemoveControls
endpoints

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tokenized_card_number | string |  | A tokenized card number for which the response is related to |
| card_controls | [CardControl](#CardControl) | repeated | A list of controls currently active on the cards transaction control document |

### Duration

A Duration represents a signed, fixed-length span of time represented as a count of seconds and fractions of seconds at
nanosecond resolution. It is independent of any calendar and concepts like "day" or "month". It is related to Timestamp
in that the difference between two Timestamp values is a Duration and it can be added or subtracted from a Timestamp.
Range is approximately +-10,000
years. [link](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#google.protobuf.Duration)

|Field|Type|Description|
|---|---|---|
| `seconds` | `Int64` | Signed seconds of the span of time. Must be from -315,576,000,000 to +315,576,000,000 inclusive.|
| `nanos` | `Int32` | Signed fractions of a second at nanosecond resolution of the span of time. Durations less than one second are represented with a 0 seconds field and a positive or negative nanos field. For durations of one second or more, a non-zero value for the nanos field must be of the same sign as the seconds field. Must be from -999,999,999 to +999,999,999 inclusive.|

### Timestamp

A Timestamp represents a point in time independent of any time zone or calendar, represented as seconds and fractions of
seconds at nanosecond resolution in UTC Epoch time. It is encoded using the Proleptic Gregorian Calendar which extends
the Gregorian calendar backwards to year one. It is encoded assuming all minutes are 60 seconds long, i.e. leap seconds
are "smeared" so that no leap second table is needed for interpretation. Range is from 0001-01-01T00:00:00Z to
9999-12-31T23:59:59.999999999Z. By restricting to that range, we ensure that we can convert to and from RFC 3339 date
strings. See https://www.ietf.org/rfc/rfc3339.txt.

|Field|Type|Description|
|---|---|---|
| `seconds` | `Int64` | Represents seconds of UTC time since Unix epoch 1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to 9999-12-31T23:59:59Z inclusive.|
| `nanos` | `Int32` | Non-negative fractions of a second at nanosecond resolution. Negative second values with fractions must still have non-negative nanos values that count forward in time. Must be from 0 to 999,999,999 inclusive.|
