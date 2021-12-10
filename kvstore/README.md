# KV Store

The key-value (KV) store is Veraison storage layer.  It is used for both endorsements and trust anchors.

It is intentionally "dumb": we assume that the filtering smarts are provided by the plugin.

The key is a string synthesised deterministically from a structured endorsement / trustanchor identifier.  It is formatted according to a custom URI format -- [see below](#uri-format)).

The value is a string containing the endorsement or trust anchor data in JSON.  This data varies depending on the attestation format.

The `KVStore` interface defines the required methods for storing, fetching and deleting KV objects.  Note that there is no method for patching data in place.  Methods for initialising and orderly terminating the underlying DB are also provided.

This package contains two implementations of the `KVStore`:

1. `SQL`, supporting different SQL engines (e.g., SQLite, PostgreSQL, etc. -- [see below](#sql-drivers)),
1. `Memory`, a thread-safe in-memory associative array intended for testing.

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
  key text PRIMARY KEY,
  val text NOT NULL
);

CREATE TABLE trustanchor (
  key text PRIMARY KEY,
  val text NOT NULL
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
  * `psa://`TenantID.Fmt()`/`ImplID.Fmt()`/`InstID.Fmt()`
* Software ID (Model is optional)
  * `psa://`TenantID.Fmt()`/`ImplID.Fmt()`/`Model.Fmt()
  * `psa://`TenantID.Fmt()`/`ImplID.Fmt()`/`


EnactTrust TPM

* Trust Anchor ID
  * `tpm-enacttrust://`TenantID.Fmt()`/`NodeID.Fmt()
* Software ID
  * `tpm-enacttrust://`TenantID.Fmt()`/`NodeID.Fmt()

