# Rubix - HCP Middleware

Middleware program to access Rubix Wallet Data from HCP Vault

## Build

Run `go build -o bin/`

## Other Location target: Connected Mobile device

**For Linux PCs**
To store the encrypted wallet data to the connected mobile device, pass the absolute mount location of the mobile phone as `storageLocation` request param. The path can be found at: `/run/user/<uid>/gvfs/mtp:host=<DEVICE_VENDOR_PRODUCT_ID>/Internal storage/`. where,

- `uid` is the User ID of the logged in Linux User, get it `id -u USER_NAME`
