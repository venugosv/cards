---
id: block
title: Block
---
This rpc method allows a client to temporarily Block ALL features and functionality on a given card. This includes all transactions on digital tokens and physical cards. No further operations can be completed on a card in Temporary block status.

View the [Sequence Diagram](https://docs.fabric.gcpnp.anz/docs/services/Card-Controls/Card-Controls#cardcontrolsapi-block) to understand logical flow

## Downstream APIs
| API                          | Purpose                                  | Link
|------------------------------|------------------------------------------|--------------------------
| Entitlements/May             | Entitlements Check                       | [Service Documentation](https://docs.fabric.gcpnp.anz/docs/services/Entitlements/Entitlements)
| CardEligibilityAPI/Can       | Eligibility Check                        | [fabric.service.eligibility.v1beta1.CardEligibilityAPI/Can](https://docs.fabric.gcpnp.anz/docs/services/Card-Eligibility/Card-Eligibility#cardeligibilityapi-can)
| Vault        | de/tokenise card number                  |
| Debit Card Status - Update   | Activate and modify status of a debit card in CTM based on debit card number | [Debit Card Status 1.0.0](https://sandpit.developer.dev.anz/eapicorp01/sandpit/node/987)

## Example:

:::tip Heads Up!
`ACTION_BLOCK` & `ACTION_UNBLOCK` are required actions on this rpc. They are tied to `ELIGIBILITY_BLOCK` & `ELIGIBILITY_UNBLOCK` eligibility items. So, if a card does **not** have `ELIGIBILITY_BLOCK` then it **cannot** use `ACTION_BLOCK`.
:::

### gRPC
Some example `grpcurl` commands that can be run when connected to the ANZ private network. For a list of environments see [Environments](#environments)

```shell script
# Block
# local dev env with stubbed external service (make run-cards)
export TESTAUTH=`NAME="SEAN FRY" make genjwt`
grpcurl -plaintext -H "Authorization: ${TESTAUTH}" -d '{"tokenizedCardNumber":"9978525667965953", "action":"ACTION_BLOCK"}' localhost:8090  fabric.service.cardcontrols.v1beta1.CardControlsAPI/Block


# SIT Block
grpcurl -H "Authorization: Basic OWFmZWIwNzYtMDVhZS00ZjUwLWI4ZTEtZTRjMzZhZTU1OWVlOjkyMjBjNjUzLWY2NjUtNDVhZS1iYTkxLTcyYmRlYTFiOWIyOA==" -d '{"tokenizedCardNumber":"3483942089355737", "action":"ACTION_BLOCK"}' cardcontrols-sit.fabric.gcpnp.anz:443 fabric.service.cardcontrols.v1beta1.CardControlsAPI/Block

# SIT Unblock
grpcurl -H "Authorization: Basic OWFmZWIwNzYtMDVhZS00ZjUwLWI4ZTEtZTRjMzZhZTU1OWVlOjkyMjBjNjUzLWY2NjUtNDVhZS1iYTkxLTcyYmRlYTFiOWIyOA==" -d '{"tokenizedCardNumber":"3483942089355737", "action":"ACTION_UNBLOCK"}' cardcontrols-sit.fabric.gcpnp.anz:443 fabric.service.cardcontrols.v1beta1.CardControlsAPI/Block
```

### [anzctl](https://github.com/anzx/fabric-anzctl)

block:
```shell script
anzctl cardcontrols block <card token>
```

unblock:
```shell script
anzctl cardcontrols unblock <card token>
```

### REST
```http request
POST https://cardcontrols-{{name}}.fabric.gcpnp.anz:443/api/v1alpha2/cardcontrols/{{tokenizedCardNumber}}/block
Authorization: Basic {{authToken}}
Content-Type: application/json

{
  "action": "ACTION_BLOCK"
}
```

### Example Response Payload
```json
{
  "status": "true",
  "eligibilities": [
    "ELIGIBILITY_UNBLOCK"
  ]
}
```
