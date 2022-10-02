# Developer Manual

> The purpose of an onboarding and development section is to make it easy for a new developer to onboard to the team, begin contributing code, add features to the microservice, and introduce new changes into the deployment pipeline.

## Prerequisites and setup

Onboarding if you are new to fabric

- [MacBook setup](https://backstage.fabric.gcpnp.anz/docs/default/system/fabric/general/onboarding/macbook/)
- [Softwares](https://backstage.fabric.gcpnp.anz/docs/default/system/fabric/general/onboarding/softwares/)
- [GCP setup](https://backstage.fabric.gcpnp.anz/docs/default/system/fabric/general/onboarding/gcp/)
- [Github](https://backstage.fabric.gcpnp.anz/docs/default/system/fabric/general/onboarding/github/)
- [Non-Tech](https://backstage.fabric.gcpnp.anz/docs/default/system/fabric/general/onboarding/non-tech/)

Tools

- [Goland IDE](https://www.jetbrains.com/go/download/)
- grpcurl (brew install grpcurl) helpful gRPC curl alternative
- make (brew install make)
- [prototool](https://github.com/uber/prototool/blob/dev/docs/install.md)

## Development

**Step-by-step guide to setting up the service**

- checkout the codebase `git@github.com:anzx/fabric-cards.git`
- run `make docker-build` to build everything.
- run `make docker-stubs` to only build stubs.

**Development cycle and deployment pipeline of the service**

### CI/CD

Build pipelines & tools used for quality control:

- [Spinnaker Cards](https://spinnaker.gcp.anz/#/projects/fabric/applications/cards/clusters)
- [SonarQube](https://sonarqube.platform-services.services-platdev.x.gcpnp.anz/dashboard?id=ghb%7Cfabric-cards) - which
  has our code quality analysis reports
- [BlackDuck](https://blackduck.platform-services.services-platdev.x.gcpnp.anz/api/projects/76ce306d-ba8e-4e32-869d-7030ee082c58)
- [CloudBuild](https://console.cloud.google.com/cloud-build/builds?project=anz-x-fabric-np-641432&amp;rapt=AEjHL4PiPSVte4-GhyElgmrVX3rrvValfMvR3Lfx36zu10xzCjtQBvuTfD7Xdz8w09YL4rBPTWF-FrQ7-nJPYKjYCfcTq33WnA&amp;pageState=(%22builds%22:(%22f%22:%22%255B%257B_22k_22_3A_22Trigger%2520Name_22_2C_22t_22_3A10_2C_22v_22_3A_22_5C_22CARDSMASTER_5C_22_22_2C_22s_22_3Atrue_2C_22i_22_3A_22triggerName_22%257D%255D%22)))
- [Github Actions](https://github.com/anzx/fabric-cards/actions)

## Quality Assurance and Testing

**Unit test**

```sh
make test
```

**[Behavior test](https://github.com/anzx/fabric-cards/tree/master/test/integration/cards)**

```sh
make behavior-cards
```

**[Integration test](https://github.com/anzx/fabric-cards/tree/master/test/integration/cards)**

```sh
# run with local env
go test -tags=integration ./test/integration/cards -env local -test.v

# run with sit env
go test -tags=integration ./test/integration/cards -env sit -test.v
```

**[PNV test](https://github.com/anzx/fabric-cards/tree/master/test/pnv/cards)**

```sh
# assume you are in the root folder of fabric-cards
cd test/pnv

# run pnv test with local env
go run cmd/main.go --test-dir=./cards/local --output-dir=./ --format=pretty

# run pnv env with pnv en. Make sure you are connected to vpn
go run cmd/main.go --test-dir=./cards/pnv --output-dir=./ --format=pretty
```




