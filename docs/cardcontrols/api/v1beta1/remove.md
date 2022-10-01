---
id: remove
title: Remove
---

This API allows client to remove a card control placed on a card. Most of the controls are removed immediately, but for the gambling control, the control will stay effective for 24 hours before it is disabled.

View the [Sequence Diagram](https://docs.fabric.gcpnp.anz/docs/services/Card-Controls/Card-Controls#cardcontrolsapi-remove) to understand logical flow

## Synopsis
The controls that exist on a card are a result of the **union** of controls that exist in the transaction control document. Remove a control from a card by providing a `tokenizedCardNumber`

```json
{
  "tokenizedCardNumber":"JFYR4hmKLWqmIr8IGoyQSmxYPqjXzoQQDvIlSM8ML3w",
  "controlTypes":[
    "MCT_ADULT_ENTERTAINMENT"
  ]
}
```

## Example:
### gRPC
local:
```shell script
# local dev env with stubbed external service (make run-cardcontrol)
export TESTAUTH=`NAME="SEAN FRY" make genjwt`
grpcurl -plaintext -H "Authorization: ${TESTAUTH}" -d '{"tokenizedCardNumber":"9978525667965953", "controlTypes":["MCT_ADULT_ENTERTAINMENT"]}' localhost:8090  fabric.service.cardcontrols.v1beta1.CardControlsAPI/Remove
```

sit:
```shell script
grpcurl -H "Authorization: Basic YjU5YzM1YjItYzgyMC00ODI4LWEwOWEtM2U3ZTZmNmQ1NGY5OmY4OTkxYjIwLTE2NzgtNGZhNi05ODdjLTRhMjkwN2JjYzQ5OQ=="  -d '{"tokenizedCardNumber":"9978525667965953", "controlTypes":["MCT_GAMBLING"]}' cardcontrols-sit.fabric.gcpnp.anz:443 fabric.service.cardcontrols.v1beta1.CardControlsAPI/Remove
```

### [anzctl](https://github.com/anzx/fabric-anzctl)

```shell script
anzctl cardcontrols remove <card token> --control-type TCT_E_COMMERCE
```

### REST
```http request
POST https://cardcontrols-{{name}}.fabric.gcpnp.anz:443/api/v1alpha2/cardcontrols/{{tokenizedCardNumber}}/remove
Authorization: Basic {{authToken}}
Content-Type: application/json

{
  "controlTypes":[
    "TCT_CONTACTLESS"
  ]
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
