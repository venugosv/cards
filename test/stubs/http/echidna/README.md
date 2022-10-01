![Fabric](../docs/images/fabric.png)

# Stub Echidna
Echidna-stub - Returns a predefined response to all PIN requests

# getWrappingKey
Path: `/daw/card-and-pin-services/getWrappingKey`

|   |   |
|---|---|
|Required Header   | `X-Request-ID`   |
|HTTP Method   | POST   |

Response:
```json
{
  "getWrappingKeyResponse": {
    "method": "getWrappingKey",
    "result": {
      "code": "0",
      "message": "Get wrapping key operation successful.",
      "encodedKey": "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArI6WKTJMLVfpaG+Mkaj4IVX3/2dbtHvacI9sKfutMsg5It6pEvFf9oYoIWMQkxFARf14ds0+1t83sm6foPHm4HZ0oP2GX0iiFdALEZr3C6C2FXAoQQXYMGeczoeta0IwF75B3Pr6VETQjf7niL00MF0n/McsE9tu9VTOFjq6LkvZgOnBe9wG+f0nvdx29FAPzIjdpBoZ27Ingmtnmtk2T9oadY5vXE2ruIhjU2rL/8aPPN8LtvlWrcV0y+YW2l4EMGenAFYMu4jh6R5deNfartmNotJgbzHFcD7EpXJivzYgdMvea2Dy7AjlC5cic4ijcna750HhfMoFFNqf6T7psQIDAQAB"
    },
    "logmessages": {
      "wantLevel": "INFO",
      "item": []
    }
  }
}
```

# selectPIN
Path: `/daw/card-and-pin-services/selectPIN`

|   |   |
|---|---|
|Required Header   | `X-Request-ID`   |
|HTTP Method   | POST   |

Response:
```json
{
  "selectPINResponse":{
    "method":"selectPIN",
    "result":{
      "code":"0",
      "message":"Select PIN operation successful."},
    "logmessages":{
      "wantLevel":"INFO",
      "item":[]
    }
  }
}
```

# verifyPIN
Path: `/daw/card-and-pin-services/verifyPIN`

|   |   |
|---|---|
|Required Header   | `X-Request-ID`   |
|HTTP Method   | POST   |

Response:
```json
{
  "verifyPINResponse":{
    "method":"verifyPIN",
    "result":{
      "code":"0",
      "message":"Verify PIN operation successful."
      },
    "logmessages":{
      "wantLevel":"INFO",
      "item":[]
    }
  }
}
```

# changePIN
Path: `/daw/card-and-pin-services/changePIN`

|   |   |
|---|---|
|Required Header   | `X-Request-ID`   |
|HTTP Method   | POST   |

Response:
```json
{
  "changePINResponse":{
    "method":"changePIN",
    "result":{
      "code":"0",
      "message":"Change PIN operation successful."},
    "logmessages":{
      "wantLevel":"INFO",
      "item":[]
    }
  }
}
```
