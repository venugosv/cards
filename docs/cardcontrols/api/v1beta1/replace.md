---
id: replace
title: Replace
---

Client can use this API to replace the current card numbers in Visa card control with a new card number. The new card will have the same card controls.

:::info
This rpc will not be used by ANZx, at this point in time, a card lifecycle event will be managed by the [CSM](https://github.service.anz/dcx/java-api-csm) until a mature migration has taken place.
:::

View the [Sequence Diagram](https://docs.fabric.gcpnp.anz/docs/services/Card-Controls/Card-Controls#cardcontrolsapi-replace) to understand logical flow

## Downstream APIs
| API                          | Purpose                                  | Link
|------------------------------|------------------------------------------|--------------------------
| Entitlements/May             | Entitlements Check                       | [Service Documentation](https://docs.fabric.gcpnp.anz/docs/services/Entitlements/Entitlements)
| CardEligibilityAPI/Can       | Eligibility Check                        | [fabric.service.eligibility.v1beta1.CardEligibilityAPI/Can](https://docs.fabric.gcpnp.anz/docs/services/Card-Eligibility/Card-Eligibility#cardeligibilityapi-can)
| Vault                 | de/tokenise card number                  |
| Visa Card Controls           | card control functionalities from VISA system via B2B DP proxy service | [Visa Card Controls 2.0.0](https://apiau182devprt01.dev.anz/eapicorp01/sb/node/32513)

## Example:
```shell script
# local dev env with stubbed external service (make run-cardcontrol)
export TESTAUTH=`NAME="SEAN FRY" make genjwt`
grpcurl -plaintext -H "Authorization: ${TESTAUTH}" -d '{"currentTokenizedCardNumber":"9149004651839526", "newTokenizedCardNumber":"8912319147207381"}' localhost:8090  fabric.service.cardcontrols.v1beta1.CardControlsAPI/Replace

# sit
grpcurl -H "Authorization: Basic YjU5YzM1YjItYzgyMC00ODI4LWEwOWEtM2U3ZTZmNmQ1NGY5OmY4OTkxYjIwLTE2NzgtNGZhNi05ODdjLTRhMjkwN2JjYzQ5OQ=="  -d '{"currentTokenizedCardNumber":"9978525667965953","newTokenizedCardNumber":"9978525667965953"}' cardcontrols-sit.fabric.gcpnp.anz:443 fabric.service.cardcontrols.v1beta1.CardControlsAPI/Replace
```

### [anzctl](https://github.com/anzx/fabric-anzctl)

```shell script
 anzctl cardcontrols replace <old card token> <new card token>
```

### Example Response Payload
```json
{
  "status": true
}
```
