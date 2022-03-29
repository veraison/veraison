# KV Store

The key-values (KV) store is Veraison storage layer.  It is used for both endorsements and trust anchors.

It is intentionally "dumb": we assume that the filtering smarts are provided by the plugins.

The key is a string synthesised deterministically from a structured endorsement / trustanchor "identifier".  It is formatted according to a custom URI format -- [see below](#uri-format)).

The value is an array of JSON strings each containing an endorsement or trust anchor data associated with that key.  The data is opaque to the KV store and varies depending on the attestation format.  The only invariant enforced by the KV store is that it is valid JSON.

The `IKVStore` interface defines the required methods for storing, fetching and deleting KV objects.  Note that (for the moment) there is no method for patching data in place.  Interface methods for initialising and orderly terminating the underlying DB are also exposed.

This package contains two implementations of the `IKVStore`:

1. `SQL`, supporting different SQL engines (e.g., SQLite, PostgreSQL, etc. -- [see below](#sql-drivers)),
1. `Memory`, a thread-safe in-memory associative array intended for testing.

A `New` method can be used to create either of these from a `Config` object.

## SQL drivers

To use a SQL backend the calling code needs to (anonymously) import the supporting driver.

For example, to use PostgreSQL:
```go
import _ "github.com/lib/pq"
```
Instead, to use SQLite:
```go
import _ "github.com/mattn/go-sqlite3"
```

## SQL schemas

```sql
CREATE TABLE endorsement (
  key text NOT NULL,
  vals text NOT NULL
);

CREATE TABLE trustanchor (
  key text NOT NULL,
  vals text NOT NULL
);
```

## URI format

```abnf
scheme ":" authority path-absolute
```

where:

* `scheme` encodes the attestation format (e.g., "psa", "tcg-dice",
"tpm-enacttrust", "open-dice", "tcg-tpm", etc.)
* `authority` encodes the tenant
* `path-absolute` encodes the parts of the key, identified positionally.  Missing optional parts are encoded as empty path segments.

Attestation technology specific code (i.e., plugins) must provide their own synthesis functions.

### Examples

PSA

* Trust Anchor ID
  * `psa-iot://`TenantID.Fmt()`/`ImplID.Fmt()`/`InstID.Fmt()`
* Software ID (Model is optional)
  * `psa-iot://`TenantID.Fmt()`/`ImplID.Fmt()`/`Model.Fmt()
  * `psa-iot://`TenantID.Fmt()`/`ImplID.Fmt()`/`


EnactTrust TPM

* Trust Anchor ID
  * `tpm-enacttrust://`TenantID.Fmt()`/`NodeID.Fmt()
* Software ID
  * `tpm-enacttrust://`TenantID.Fmt()`/`NodeID.Fmt()

