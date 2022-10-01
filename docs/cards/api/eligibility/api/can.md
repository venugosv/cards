# Can

Can returns the eligible status of particular operation for a given card

View
the [Sequence Diagram](https://docs.fabric.gcpnp.anz/docs/services/Card-Eligibility/Card-Eligibility#cardeligibilityapi-can)
to understand logical flow

#### Downstream APIs

| API                          | Purpose                                  | Link
|------------------------------|------------------------------------------|--------------------------
| Entitlements/May             | Entitlements Check                       | [Service Documentation](https://docs.fabric.gcpnp.anz/docs/services/Entitlements/Entitlements)
| CardEligibilityAPI/Can       | Eligibility Check                        | [fabric.service.eligibility.v1beta1.CardEligibilityAPI/Can](https://docs.fabric.gcpnp.anz/docs/services/Card-Eligibility/Card-Eligibility#cardeligibilityapi-can)
| Vault        | de/tokenise card number                  |
| Debit Card Inquiry           | Retrieve Card details using the CTM Card Inquiry service | [Debit Card Inquiry 1.0.0](https://sandpit.developer.dev.anz/eapicorp01/sandpit/node/985)

#### Example:

Some example `grpcurl` commands that can be run when connected to the ANZ private network. For a list of environments
see [Environments](#environments)

```shell script
grpcurl -H "Authorization: Basic YmJhZTRlZDctNTE0MC00ZTIwLThiNzUtYTJlZTRiNDc0NTJjOmRzVlA1eEJISzdra0FNZWxsMU1BWWZyS29OcmJQMXht" -d "{\"tokenizedCardNumber\":\"gSzqPOO2lmdbvs8UwsIGwYX78qp00jhdhvCx3fama7g\",\"eligibility\":\"ELIGIBILITY_CARD_REPLACEMENT_DAMAGED\"}" cards-sit.fabric.gcpnp.anz:443 fabric.service.eligibility.v1beta1.CardEligibilityAPI/Can
```

##### Example Response Payload

```json
{
  "eligible": "true"
}
```
