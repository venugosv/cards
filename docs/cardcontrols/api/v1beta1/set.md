---
id: set
title: Set
---

This API allows client to set card controls on a card. If a card is not enrolled, it will enrol the card and set the card controls on it.

View the [Sequence Diagram](https://docs.fabric.gcpnp.anz/docs/services/Card-Controls/Card-Controls#cardcontrolsapi-set) to understand logical flow

## Synopsis
The controls that exist on a card are a result of the **union** of controls that exist in the transaction control document. Add controls to a card by providing a `tokenizedCardNumber` and the controls you want to add:

```json
{
  "tokenizedCardNumber":"JFYR4hmKLWqmIr8IGoyQSmxYPqjXzoQQDvIlSM8ML3w",
  "cardControls":{
    "controlType":"MCT_GAMBLING"
  }
}
```

## Downstream APIs
| API                          | Purpose                                  | Link
|------------------------------|------------------------------------------|--------------------------
| Entitlements/May             | Entitlements Check                       | [Service Documentation](https://docs.fabric.gcpnp.anz/docs/services/Entitlements/Entitlements)
| CardEligibilityAPI/Can       | Eligibility Check                        | [fabric.service.eligibility.v1beta1.CardEligibilityAPI/Can](https://docs.fabric.gcpnp.anz/docs/services/Card-Eligibility/Card-Eligibility#cardeligibilityapi-can)
| Vault                 | de/tokenise card number                  |
| Visa Card Controls           | card control functionalities from VISA system via B2B DP proxy service | [Visa Card Controls 2.0.0](https://apiau182devprt01.dev.anz/eapicorp01/sb/node/32513)

## Example:
### gRPC

local:
```shell script
# local dev env with stubbed external service (make run-cardcontrol)
export TESTAUTH=`NAME="SEAN FRY" make genjwt`
grpcurl -plaintext -H "Authorization: ${TESTAUTH}" -d '{"tokenizedCardNumber":"9978525667965953", "cardControls":{"controlType":"MCT_GAMBLING"}}' localhost:8090  fabric.service.cardcontrols.v1beta1.CardControlsAPI/Set
```

sit:
```shell script
grpcurl -H "Authorization: Basic YjU5YzM1YjItYzgyMC00ODI4LWEwOWEtM2U3ZTZmNmQ1NGY5OmY4OTkxYjIwLTE2NzgtNGZhNi05ODdjLTRhMjkwN2JjYzQ5OQ=="  -d '{"tokenizedCardNumber":"9978525667965953","cardControls":{"controlType":"MCT_GAMBLING"}}' cardcontrols-sit.fabric.gcpnp.anz:443 fabric.service.cardcontrols.v1beta1.CardControlsAPI/Set
```

### [anzctl](https://github.com/anzx/fabric-anzctl)

```shell script
anzctl cardcontrols set <card token> --control-type TCT_E_COMMERCE  --control-type TCT_CROSS_BORDER
```

### REST
```http request
POST https://cardcontrols-{{name}}.fabric.gcpnp.anz:443/api/v1alpha2/cardcontrols/{{tokenizedCardNumber}}/set
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
    },
    {
      "controlType": "MCT_HOUSEHOLD",
      "controlEnabled": true
    }
  ]
}
```
