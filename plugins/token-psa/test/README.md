`psa-token.cbor` used by the tests is compiled from the YAML description using
a utility that is included with iat-verifier tool that can be found here:

    https://git.trustedfirmware.org/TF-M/trusted-firmware-m.git/tree/tools/iat-verifier

To install the tool, get the source and then

    pip install <path/to/iat-verfier>

After the tool has been installed, you can re-create the token with the
following command run inside this directory:

    compile_token -k key.pem psa-token.yaml -o psa-token.cbor

(`compile_token` should be added to path when iat-verifier is installed.)

Note: there is an analogous `decompile_token` command that can be used to get
a YAML description from a CBOR-encoded token.
