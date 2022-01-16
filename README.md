# test-convexfinance

My experiment to try call incentive calculation extracted from Booster contract of Convex Finance.
This program gets required variable from the contract to use formula defined in `_earmarkRewards()` function
in the contract.

## Binding
Go binding was generated using `https://github.com/cryptoriums/contraget`.

## How to Use
* git clone git@github.com:arknable/test-convexfinance.git
* cd `test-convexfinanc`
* run `go mod tidy`
* set environment variable `NODE_URLS` to main net node.
* run `go run main.go`