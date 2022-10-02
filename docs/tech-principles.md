# Tech Principles

The principles and practises enable each application team for easier sharing, remembering and assessing against current
design work.

List out the Tech Principles that are intgrated with your team.

**Build for Run**: Our services are robust and easy to support.

Our service is self-contained and stateless. The service is dockerized and can be deployed to any platform that supports
docker. Also the service is versionised which means we can easier rollback to a working version when things went wrong.

**Build for Change**:  Rapid change expected and easy to manage without downtime.

- Microservice architect & proper layered components designs makes it easy for us to make change to our code.
- We ensure change scope is in small, bite-sized chunks to enable a simple peer-reviewing experience.
- Development branches do not live for extended periods of time, our service IS the trunk not the tree.
- Automated testing and deployment pipelines make iterating to production quick and safe.

**Build Quality In**: Quality is built-in and not a phase to be completed later.

In general, we are following the best industry practices and build quality into our service. Things we are doing:

- Unit test: 90+% coverage
- Structural test: testing the different scenarios for the connected logical flow
- Behavioural test: attempts to test the system from the consumers point of view
- Integration test: test systems integration in real integrated environment
- Performance test: make sure our code not only runs but also runs in fast speed
- Linting: ensure the code is clean and consistent
- Checkmarx, BlackDuck & CodeQL: find potential security problem

**Always Automated**: End-to-end driven automation mandated.

We automate everything that we can automated. Things we automated:

- Automated deployment of our service means the latest build of our service is ready to be consumed, across many
  environments at all times.
- Document generation is automated to reduce the risk that docs are out sync with the code
- Build and quality check are automated for every PR
- Integration tests & PNV tests are automated in the integrated environment

**Always Observable**: Observability tools/dashboards/alerting built in parallel with services.

- Utilising the open-telemetry standards, Cards is prepared from day one to be exporting runtime metrics and tracing.
- All logs in Cards exported to Stackdriver where they take part in
  the [logging solution](https://confluence.service.anz/display/ABT/OVR-019+Logging+Solution+for+ANZx) designed for
  ANZx.
- Masked HTTP requests and responses logged out to provide context around what is going out and coming in from our
  downstream dependencies.

**Always Stable**: So that we can ensure there is no data duplication or reconciliation issues between any of the data
source and ourselves, Cards does not manage a database.

- PnV and Integration tests run as part of the deployment pipeline.
- Support for multiple versions of our API when new changes are introduced.
- Environments built to scale when load increases.

**Always Secure**: Security is everyone's responsibility. Security resources embedded in the platform.

- Plaintext PANs are a serious concern of ours, therefore, Cards does not require a plaintext PAN in its contract. Our
  consumers will provide a tokenized version of a PAN which cannot be used in fraudulent activity, making Fabric-cards
  more secure by far than the current ANZ approved pattern for handling cards.

- It's great to log, but its very easy to end up with someones banking details or card numbers in the log history -
  forever. To combat this, we mask all PII data from our logs.

- Using external tools to make sure our code is secure:
  - Checkmarx scan for each release
  - BlackDuck scan for each PR
  - Github CodeQL scan for each PR

**Cloud Native**: We leverage the best of open source and cloud native services.

Fabric Cards runs on a GCP GKE cluster and has plans to move to a serverless Google Cloud Run environment in the near
future.

**Easy to Contribute**: We have well documented services that are easy to contribute to.

- Generated off our proto definitions are sequence diagrams and model object uml.
- documentation housed in fabricapis & Cards repo is synced with fabric-docs automatically.
- source code is organized in a proper folder structure and the code is easy to understand
- Sharing common practices with other fabric services, so it is easy for other fabric devs to contribute to the codebase

> For more details on Tech Principles and the practices, go
to [Tech Principles](https://confluence.service.anz/display/ABT/ANZx+Tech+Principles)





