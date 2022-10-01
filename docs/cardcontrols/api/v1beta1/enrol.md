---
id: enrol
title: Enrol
---

:::note
The Transaction Controls Client Host Callback API will be hosted external to Visa. Card Enrollment callback will be hosted by the authorizing agent.
:::

## rpc: Enrol

Enrollment callback service hosted by the issuer or issuers authorizing agent.

:::info
This rpc will not be used by ANZx, at this point in time, a VISA callback event will be managed by the [CSM](https://github.service.anz/dcx/java-api-csm) until a mature migration has taken place.
:::

[Diagram](https://docs.fabric.gcpnp.anz/docs/services/fabricapis/fabric_service_cardcontrols_v1alpha3/fabric_service_cardcontrols_v1alpha3#cardcontrolsapi-enrol)

## rpc: Disenrol

Enrollment callback service hosted by the issuer or issuers authorizing agent.

:::info
This rpc will not be used by ANZx, at this point in time, a VISA callback event will be managed by the [CSM](https://github.service.anz/dcx/java-api-csm) until a mature migration has taken place.
:::

[Diagram](https://docs.fabric.gcpnp.anz/docs/services/fabricapis/fabric_service_cardcontrols_v1alpha3/fabric_service_cardcontrols_v1alpha3#cardcontrolsapi-disenrol)
