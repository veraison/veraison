#!/bin/bash

set -eux
set -o pipefail

for t in endorsement trustanchor
do
    echo "CREATE TABLE $t ( key text NOT NULL, vals text NOT NULL );" | \
        sqlite3 veraison-$t.sql
done
