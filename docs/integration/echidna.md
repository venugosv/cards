# Echidna

Salt Echidna-RemotePIN application provides the following REST Web Service APIs to perform PIN related operations. Upon
receiving the requests through the Web Services, the RemotePIN service communicates with the Tandem server to perform
the PIN operation then passes through the appropriate responses back to the caller application. Typically, the caller
application is an internal bank authentication web application, which proxies the user end requests (requests from a
bank mobile application for an example). This specification is based on version 18.4.0.34493 spec.

Connection Via APIc - https://sandpit.developer.dev.anz/eapicorp01/sandpit/node/1218

## SALT SDK

Salt Mobile SDK is an embedded library with a collection of APIs that enable multi-factor authentication of user
identity and transaction signing capabilities within existing mobile apps such as mobile banking. All the ‘Connected
Token’ features and capabilities of Salt mSign and Salt mCode are included in the Salt Mobile SDK embedded token. We use
the SDK for encrypting the card pins before communicating with backend servers.

### Context

ANZRemotePin is the CocoaPod that wraps the SaltSDK.framework provided by the Salt Group. It is used to encrypt the card
PIN entered by the customer in the set or change card PIN flow. This document outlines the integration of ANZRemotePIN
in ANZ App and the maintenance approach.

### Why we are using SaltSDK

The Australian Payments Network Limited, which ANZ is a member of has standards/requirements that strongly recommend the
card PIN to be encrypted as soon as possible using specific encryption standards.
See [Issuers Code Set Volume 2]( https://www.auspaynet.com.au/resources/cards-and-devices). Sections 3.1b i and 3.2.1d
describe this.

Section 3.7.2 specifies that ISO 9564.1 format 3 is the preferred encrypted PIN format, and SaltSDK is used because it
supports that format.

### Current integration

The SaltSDK is a closed source framework maintained by the Salt Group. DCX is not responsible for, nor has visibility
of, the source code for the SaltSDK.

The SDK is stored as a zip file in Artifactory, and integrated into ANZ App as a CocoaPod via `ANZRemotePin.podspec`
which is located in `Frameworks/ANZRemotePin` folder. The pod spec references the Artifactory zip file as its source
location.

The SDK is manually uploaded to Artifactory by an engineer with Artifactory write access.

### How to use it in Objective C?

Given the static library is wrote in Objective-C, the integration is very straightforward. Put #import "RemotePin.h" in
the code then we could use the function provided in the SDK.

### How to use it in Swift?

Things get a bit tricky here. It is not a framework. It can't be used directly in Swift. The solution is to manually map
and export it as a module for Swift to consume by adding a module.modulemap file to the SDK folder.

```
module RemotePin {
    header "include/RemotePin/RemotePin.h"
    export *
}
```

After that, we can import RemotePin in the code then use the function provided in the SDK.

### Artifactory

The SaltSDK framework can be found by searching
in [Artifactory](https://artifactory.service.anz/artifactory/anzapp-ios-lib-local/salt) for SaltSDK.framework within the
anzapp-ios-lib-local repository. Find the Android
version [on Artifactory](https://artifactory.service.anz/artifactory/anzapp-android-lib-local/salt/) within the
anzapp-android-lib-local repository/

SDK PDF documentation is available in Artifactory along-side the SDK zip archive.

### Key contacts and current distribution mechanism

The SaltSDK is distributed as SaltSDK.framework (universal framework) by the CAT Team. Salt Mobile SDK is a 3rd party
SDK, but the relationship with Salt is held by the Echidna team the CAT tribe, so contact with Salt should via the
contacts below to avoid additional/unnecessary engagement costs with Salt.

* Manoharan, Karthik <Karthik.Manoharan@anz.com> (Product Owner, CAT)
* Rasika Arachchige <rarachchige@saltgroup.com.au>, <Rasika.MenmendaArachchige@anz.com> (Salt Team)
* Doxey <john.doxey@anz.com> (ANZ CAT Tribe Development team, manages the framework)

### Potential problems and fixes

For the best practise of managing the third party library, we should create a new artefact and manage it through
CocoaPods.

### Useful Documentation

- [RemotePIN System Sequence Diagram v1.0](../assets/ANZ-RemotePIN System Sequence Diagram v1.0.docx)
- [RemotePIN Web Service API Specification v1.4.0](../assets/Salt RemotePIN Web Service API Specification v1.4.0.doc)
- [RemotePIN Mobile SDK overview v2](../assets/SaltGroup-RemotePIN Mobile SDK overview v2.docx)

### Reference

- https://confluence.service.anz/display/DIS/Card+and+Pin+Services+Interface+Specs


