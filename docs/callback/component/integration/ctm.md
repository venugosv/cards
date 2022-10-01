# CTM

Update preferences on a debit card such as card controls.

- The card controls indicator can be updated for Visa Debit Cards and Access Debit Cards. Updates to this flag will be
  logged by CTM to be passed to Base24 every 15 minutes in the ‘trickle feed’ process. Base24 will maintain this on
  their CRDD database, which will be used to decide if a card transactions needs to be authorised at Visa as well as
  CTM.

for further documentation, checkout the cards CTM documentation [here](https://backstage.fabric.gcpnp.anz/docs/default/Component/fabric-cards/integration/ctm)
