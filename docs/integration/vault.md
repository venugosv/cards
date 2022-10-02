# Vault

The Transform secrets engine handles secure data transformation and tokenization against provided input value.
Transformation methods may encompass NIST vetted cryptographic standards such as format-preserving encryption (FPE) via
FF3-1, but can also be pseudonymous transformations of the data through other means, such as masking.

- **Contact**: Eaas-ADP support [eaasadpsupport@anz.com](mailto:eaasadpsupport@anz.com)
- **Approle**: gcpiamrole-fabric-encode.common

## Terraform

[Non-production](https://github.service.anz/IAM/au_adp_tenants/blob/production/namespaces/eaas/fabric/common.tfvars)

```terraform
gcpiamroles = [
  {
    transformrole_suffix = "common"
    decide_role          = true
    encode_constraints   = {
      bound_projects         = ["anz-x-fabric-cde-np-ba0f52"]
      bound_service_accounts = [
        "cards-pnv@anz-x-fabric-cde-np-ba0f52.iam.gserviceaccount.com",
        "cards-preprod@anz-x-fabric-cde-np-ba0f52.iam.gserviceaccount.com",
        "cards-st@anz-x-fabric-cde-np-ba0f52.iam.gserviceaccount.com",
        "cards-sit@anz-x-fabric-cde-np-ba0f52.iam.gserviceaccount.com",
        "cards-sit-n@anz-x-fabric-cde-np-ba0f52.iam.gserviceaccount.com",
        "cards-intpnv@anz-x-fabric-cde-np-ba0f52.iam.gserviceaccount.com",
        "cards-preprod-k@anz-x-fabric-cde-np-ba0f52.iam.gserviceaccount.com",
      ]
      token_bound_cidrs      = ["0.0.0.0/0"]
    },
    ...
```

[Production](https://github.service.anz/IAM/au_adp_tenants/blob/development/namespaces/eaas/fabric/prod.tfvars)

```terraform
gcpiamroles = [
  {
    transformrole_suffix = "common"
    decide_role          = true
    encode_constraints   = {
      bound_projects         = ["anz-x-fabric-cde-prod-d3ac9b"]
      bound_service_accounts = [
        "cards-prod@anz-x-fabric-cde-prod-d3ac9b.iam.gserviceaccount.com",
      ]
      token_bound_cidrs      = [
        "10.180.64.0/19",
        "10.148.10.0/24"
      ] # Whitelist PROD LB range because XFF header injection isn't supported by the current LB
    },
    ...
```
