![Fabric](../docs/images/fabric.png)

# Stub Visa Consumer Transaction Controls
Visa-stub - provides a simple http server for the visa consumer transaction controls service

**Overall supported Credit Card numbers: 462239051234XXXX**

view the [scratch file](scratch.http) for example calls

### Supporting the following visa functionality

- Query By PAN
    - POST /vctc/customerrules/consumertransactioncontrols/inquiries/cardinquiry
        - 4622390512341XXX - All Controls in returned document: Global, merchant and transaction
        - 4622390512342XXX - Global Controls in returned document
        - 4622390512343XXX - No Controls in returned document
        - Default - Not Enrolled
- Enrol By PAN
    - POST /vctc/customerrules/consumertransactioncontrols
        - 462239051234XXXX - Success
        - Default - Not Enrolled
- Card Replacement
    - POST /vctc/customerrules/consumertransactioncontrols/accounts/accountupdate
        - 462239051234XXXX - Success
        - Default - Failure
- Create Control
    - POST /vctc/customerrules/consumertransactioncontrols/{{DocumentID}}/rules
         - any value for doc id will return successful creation of documentID
- Update Control
    - PUT /vctc/customerrules/consumertransactioncontrols/{{DocumentID}}/rules
        - any value for doc id will return successful creation of documentID

NOTE: this stub has **no state** it will not maintain modifications to the control document

Still to come:
- Callbacks

