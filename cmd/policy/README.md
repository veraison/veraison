## Example usage

NOTE: All commands allow specifying a tenant with -t. If this is not specified,
then tenant 1 is usually assumed. The only exceptions are list command (which
provides output for all tenants), and verify command (which will read tenant ID
from the provided evidence context).

    ./policy list

List all policies in Veraison store. This will output a list of tenant ID and
format name pairs.

    ./policy -t 2 list

Same as above but only list policies for tenant ID 2.

    ./policy set -f my_policies.zip

Add policies from my_policies.zip for tenant 1, overwriting any existing policies.

   ./policy -c /opt/veraison/config/ get -t 3 -o my_policies.zip psa dice

Get policies for psa and dice token formats for tenant 3, using Veraison
configuration from /opt/veraison/config/, and write them to my_policies.zip.

 ./policy verify -e test/endorsements.json test/evidence-context.json

Run stored policy against input inside test/evidence-context.json, using
endorsements read from test/endorsements.json.

## Policy zip format

Policies can be read out from the store into a zip archive. A similar archive
can be used to add policies to the store. Each policy is contained within a
top-level directory within the archive, named after the token format. The
directory contains two files: query_map.json is a JSON serialization of the
query map -- a mapping indicating how endorsements query parameters can be
obtained from the token claims; the second file is "rules" and contains the
rules to be used for evaluating token evidence against endorsements. The format
of the rules must match what is expected by the Policy Engine configured in a
particular Veraison deployment (e.g. in REGO format for the OPA engine).

There is an example layout of a zip archive:

    my_policies.zip
     |
     |--psa
     |   |
     |   |-- query_map.json
     |   |-- rules
     |
     |--dice
     |   |
     |   |-- query_map.json
     |   |-- rules
