# Queries

## Overview

This document contains definitions of "standard" endorsement queries. Queries
are the interface by which endorsements are exposed to the rest of the system.
A query definition consists of the following parts.

1. Name: the name is how a query is referenced. It must be unique in a deployment.
2. Parameters: the descriptions of the parameters the query expects, i.e. the
   input against which the query is run.
3. Result: the description of the matches returned by the query. (note: this
   documents the format of a single match -- the query result is a slice of
   such matches).

## Standard Queries

### hardware_id

#### Parameters

**platform_id**: a `string` representation of the platform identifier extracted from the
evidence.

#### Result

A `string` containing the hardware identifier for the platform from which the
evidence was extracted.

### software_components

#### Parameters

**platform_id**: a `string` representation of the platform identifier extracted from the
evidence.

#### Result

A slice of `SoftwareEndorsement` instances containing the measurements of endorsed
software component versions associated with a particular platform.
