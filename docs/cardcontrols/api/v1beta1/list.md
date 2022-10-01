---
id: list
title: List
---

This API allows client to list all the visa card controls placed on all the cards belong to current user.

View the [Sequence Diagram](https://docs.fabric.gcpnp.anz/docs/services/Card-Controls/Card-Controls#cardcontrolsapi-list) to understand logical flow

## Synopsis
List to receive the controls are applied to all the cards.

## Example:
### gRPC

local:
```shell script
# local dev env with stubbed external service (make run-cardcontrol)
export TESTAUTH=`NAME="SEAN FRY" make genjwt`
grpcurl -plaintext -H "Authorization: ${TESTAUTH}" localhost:8090  fabric.service.cardcontrols.v1beta1.CardControlsAPI/List
```

sit:
```shell script
# The card number needs to be replaced with a available test card in sit when you run the command
grpcurl -H "Authorization: Basic YjU5YzM1YjItYzgyMC00ODI4LWEwOWEtM2U3ZTZmNmQ1NGY5OmY4OTkxYjIwLTE2NzgtNGZhNi05ODdjLTRhMjkwN2JjYzQ5OQ==" cardcontrols-sit.fabric.gcpnp.anz:443 fabric.service.cardcontrols.v1beta1.CardControlsAPI/List
```

### [anzctl](https://github.com/anzx/fabric-anzctl)

```shell script
anzctl cardcontrols query <card token>
```


### REST
```http request
POST https://cardcontrols-{{name}}.fabric.gcpnp.anz:443/api/v1alpha2/cardcontrols/{{tokenizedCardNumber}}/query
Authorization: Basic {{authToken}}
Content-Type: application/json

{
  "tokenizedCardNumber": "JFYR4hmKLWqmIr8IGoyQSmxYPqjXzoQQDvIlSM8ML3w"
}
```

### Example Response Payload
```json
{
  "cardControls": [
    {
      "controlType": "GCT_GLOBAL",
      "controlEnabled": true
    }
  ]
}
```

## Downstream APIs
| API                          | Purpose                                  | Link
|------------------------------|------------------------------------------|--------------------------
| Entitlements/May             | Entitlements Check                       | [Service Documentation](https://docs.fabric.gcpnp.anz/docs/services/Entitlements/Entitlements)
| CardEligibilityAPI/Can       | Eligibility Check                        | [fabric.service.eligibility.v1beta1.CardEligibilityAPI/Can](https://docs.fabric.gcpnp.anz/docs/services/Card-Eligibility/Card-Eligibility#cardeligibilityapi-can)
| Vault                 | de/tokenise card number                  |
| Visa Card Controls           | card control functionalities from VISA system via B2B DP proxy service | [Visa Card Controls 2.0.0](https://apiau182devprt01.dev.anz/eapicorp01/sb/node/32513)
