# Stubs

## How to simulate latencies

For pnv test, stubs are going to be used for non fabric downstream dependencies.
To make it more realistic, our stubs exposed a restful API(`POST: http://stubhost:9070/simulateDelay`) for setting latency based on url.
The request body for this endpoint is a `map[string]int64` where the key is an url pattern, and value is the latency you want to apply to any urls matches the key.

For example, to set 300ms latency for ctm calls where url contains `debit-card`, simply do:
```bash
curl --header "Content-Type: application/json" \
   --request POST \
   --data '{"debit-card":300}' \
   http://localhost:9070/simulateDelay
```
 **Note:** In the future, we can add some extra randomness to the latency if needed.
