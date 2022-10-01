[Home](../../README.md) / Repository Map

# Repository Map
```
├── build
├── cmd
│   ├── cardcontrols
│   └── cards
├── config
│   ├── cardcontrols
│   └── cards
├── docs
├── internal
│   └── service
│       ├── cards
│       ├── controls
│       └── eligibility
├── pkg
│   ├── identity
│   ├── integration
│   ├── middleware
│   ├── ops
│   └── util
├── sonar-project.properties
├── test
│   ├── behavior
│   ├── fixtures
│   ├── structural
│   └── stubs
└── vendor
```
- [build](../../build) \
Directory for all build related files, includes the likes of cloudbuild & dockerfiles.
- [cmd](../../cmd) \
Directory in which entry-point files are located for compilation targets/required at startup, such as the application config & cli or the initialisation of open cencus
- [config](../../config) \
Directory in which contains the application config files for all environments and platform config files needed for Spinaker
- [docs](../../docs) \
Directory for storage of subjective READMEs. Main README markdown file is stored in the [root directory](../..)
- [internal](../../internal) \
Directory of Cards-specific application code, not intended to be shared with others.
    - [service](../../internal/service) - gRPC api definition and related "commands" to orchestrate multiple integrations and return back to user.
- [pkg](../../pkg) \
Directory of common packages that aren't specific to the Cards business logic, packages that could potentially be shared with others or extracted to a common library
    - [integration](../../pkg/integration) - Directory of downstream integration code, expected to be independent packages for each integration.
    - [middleware](../../pkg/middleware) - middleware layer for grpc server, think interceptors for auth and tracing.
- [test/stubs](../../test/stubs) \
Includes a server builder used to create mocks for unit testing
- [test/fixtures](../../test/fixtures) \
Directory of stubs of downstream services
- [test/structural](../../test/structural) \
Directory contains all the structural tests
- [test/behavior](../../test/structural) \
Directory contains all the behavior tests
- [vendor](../../vendor) \
Typical Golang project vendor folder - stores copies of project dependencies
