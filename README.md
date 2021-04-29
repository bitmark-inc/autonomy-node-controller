# Autonomy Pod Controller

![Run Testcases](https://github.com/bitmark-inc/autonomy-pod-controller/actions/workflows/run_testcases.yaml/badge.svg??branch=main)

This is controller of autonomy pod which enables a user to communicate with his
bitcoind via whisper protocol.

## Pre-requisite

- go 1.16


## Build

```
make pod-controller
```

## Run

First, create a `config.yaml` by copy `config.yaml.sample` and configure it properly. Then run:

```
make run-pod-controller
```

## Generate mock interfaces for testing

```
mockgen -source=store.go -destination store_mock.go -package main
```

## Usage

The request to the pod controller is of the following form:

```
{
    "id":      "test",
    "command": "<command>",
    "args": {}
}
```

### Command examples

---

### bind

#### Args

```
{}
```

#### Returns

```
{
  "identity": "did:key:zQ3shaZgv7c7EetwDBsnXxFQ9B9WcH8Kfr7bTLDYaj3cDE8Ec",
  "nonce": "5e8976ef",
  "timestamp": "1619667646464",
  "signature": "30450221009d22dc9541ce4f9c5e5d91b0206715757342c67e84ac2c3a75523eb7e4b3da80022068711157a13b71c115d2d2ee983f1d125c22b979f4415e89620e2dae74c67f4d"
}
```

---

### bind_ack

#### Args

```
{
  "nonce": "5e8976ef"
}
```

#### Returns

```
{
  "status": "ok"
}
```

---

### create_wallet

#### Args

```
{
  "descriptor": "wsh(sortedmulti(2,[119dbcab/48h/1h/0h/2h]tpubDFYr9xD4WtT3yDBdX2qT2j2v6ZruqccwPKFwLguuJL99bWBrk6D2Lv1aPpRbFnw1sQUU9DM7ScMAkPRJqR1iXKhWMBNMAJ45QCTuvSZbzzv/0/*,[e650dc93/48h/1h/0h/2h]tpubDEijNAeHVNmm6wHwspPv4fV8mRkoMimeVCk47dExpN9e17jFti12BdjzL8MX17GvKEekRzknNuDoLy1Q8fujYfsWfCvjwYmjjENUpzwDy6B/0/*,[<fingerprint>/48h/1h/0h/2h]<xpub>/0/*))"
}
```

#### Returns

```
{
  "descriptor": "wsh(sortedmulti(2,[119dbcab/48h/1h/0h/2h]tpubDFYr9xD4WtT3yDBdX2qT2j2v6ZruqccwPKFwLguuJL99bWBrk6D2Lv1aPpRbFnw1sQUU9DM7ScMAkPRJqR1iXKhWMBNMAJ45QCTuvSZbzzv/0/*,[e650dc93/48h/1h/0h/2h]tpubDEijNAeHVNmm6wHwspPv4fV8mRkoMimeVCk47dExpN9e17jFti12BdjzL8MX17GvKEekRzknNuDoLy1Q8fujYfsWfCvjwYmjjENUpzwDy6B/0/*,[<fingerprint>/48h/1h/0h/2h]<xpub>/0/*))"
}
```

---

### finish_psbt

#### Args

```
{
  "psbt": "cHNidP8BAH0CAAAAAUtbuwAXBenDq8soKdRpZVRcDx3om/g1s+/EUOlp1aw2AQAAAAD+////AooCAAAAAAAAIgAgptJpKonsWEFQBfZlFGIflekEgQhAs5wG3ESE0p3Vz5ftIwAAAAAAABYAFKIV4bxghFvY5te+QvOT5keZpGGwAAAAAAABAHECAAAAASAt8DNNC05TTcogbRwahlBvRaXG52a6sZJ549IkFKBlAQAAAAD+////AuHrRQAJAAAAFgAU7GjPiRbOfeES4pnRFuS3f466eCkQJwAAAAAAABYAFJz79oL5/lWICJ9Yfk7g1z/tKw671RoeAAEBHxAnAAAAAAAAFgAUnPv2gvn+VYgIn1h+TuDXP+0rDrsiBgOBuLfcCkFk/1B6LkCBn4uZuP0EV2cz+PhiYCSmSVNjthDMZttGAAAAgAAAAIACAACAAAAiAgPsIGy6eXigvHcW/9xgIVOJI2Ujj7/vWYtJg6noe8mgDhDMZttGAAAAgAEAAIABAACAAA=="
}
```

#### Returns

```
{
  "txid": "dee5b21ef0e839c39f7ee1b690f1b0e63155af35ca85b6e1a50d7803b008b561"
}
```

---

### set_member

#### Args

```
{
  "member_did": "did:key:zQ3shk5bp53SwcW4685TwY1BuieKLTLEPLSJRG8Qknu9oddRj",
  "access_mode": 2
}
```

#### Returns

```
{
  "status": "ok"
}
```

---

### remove_member

#### Args

```
{
  "member_did": "did:key:zQ3shk5bp53SwcW4685TwY1BuieKLTLEPLSJRG8Qknu9oddRj"
}
```

#### Returns

```
{
  "status": "ok"
}
```