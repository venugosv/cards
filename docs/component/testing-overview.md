[Home](../../README.md) / Testing Overview
# Testing Overview
Fabric Card uses the Go Test standard library with the aid of [testify](https://github.com/stretchr/testify)
for assertions.

At a minimum all APIs provided by a service should have automated functional API tests against mock data that cover:

- Every path defined in the API specification
- Every HTTP verb defined in the API specification
- Every response code defined in the API specification
- Every example specified in the API specification

These tests should validate the expected data returned by the service and that the response conforms to the contract.
Any important business logic implementation should also be covered with functional API tests.
It is assumed data validation of inputs covered via unit tests, if this is not the case these should also fall within the functional API testing scope.


 #### We currently do the following forms of testing:

- *Unit* - per-function based testing

 #### We intend do the following forms of testing as we mature:
- *E2E(Integration)* should cover an end to end execution of all functions provided by the service.
A single variation of each scenario should be covered, with every variation covered by unit tests.

    - *Structural* - the internal structure/design/implementation of the item being tested is known to the tester.
Includes Lower level based integration and validation testing of microservice code and dependency interactions
(such as database).

    - *Behaviour* - a software testing method in which the internal structure/ design/ implementation of the item being
tested is not known to the tester. Includes Contract testing of API interface provided by API.

- *Performance Testing* - validating expectations from a performance perspective, ensuring the service meets a target
request rate and response time criteria. A tool we have investigated is [ghz](https://ghz.sh/)

## Unit Tests
There is an expectation that all new code is tested, achieving **100% unit test coverage** before being
merged into master. The overall project aims to achieve **95%** coverage in unit tests.

## Behavior Tests
The behavior tests run locally against a docker-compose environment providing stubbed services where necessary and should test mainly integrations and happy paths.

## e2e Tests

## Regression Tests

## Security Testing

Read more on ANZx API testing standards [here](https://confluence.service.anz/display/ABT/API+Testing)
