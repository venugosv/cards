# FAQ

**Is there currently a limit on how many times a user can replace a card in quick succession in ctm? If not, are you
aware of falcon identifying this behaviour and preventing it?**

There’s no limit on number of times a card can be replaced. If a card is replaced more than once per business day (
weekend/public holiday and next day are considered one business day) only the latest replacement will be actioned, all
prior replacements will be dropped.

**What does temp block do in CTM?**

The temp block on the card will stop all ATM, EFTPOS (Merchant) and digital wallet transactions.

**If a card temp blocked in CTM – will merchant transactions with card information saved on file be blocked? Example:
Netflix subscription.**

The block will be in ANZ (CTM and Tandem) and Visa. It depends on what the merchant does when a payment is due.
Normally, they would send an Authorisation request to ANZ to authorise the payment, and Tandem would decline. Sometimes,
merchants do not do that, instead they send a settlement value transaction to ANZ which is force post, and the
transaction would go through and posted to the account. In this case, the bank has a right to charge back (i.e. return
the tran to merchant).

**Is there any way to determine that a card is ordered to display message like ‘Your new card has been ordered’ to a
customer?**

CTM debit card inquiry API returns a field called ReplacedDate, we can use it to determine if a card is recently
ordered. List API will then return a `TokenizedNewCardNumber` field in the old card until the new card is
activated/delivered. UI can check if this field exists to determine whether it should show the ‘Your new card has been
ordered’ message.




