#!/bin/bash
CERT_OUTDIR=${CERT_OUTDIR:-$PWD}
CERT_OUTFILE=${CERT_OUTFILE:-local.crt}
cert_path="$CERT_OUTDIR"/"$CERT_OUTFILE"

echo "===================================="
echo exporting cert

if kubectl exec -n emulator deploy/letsencrypt -- sh -c "apk add curl > /dev/null; curl -ksS https://localhost:15000/intermediates/0" > "$cert_path";
then
  echo "------------------------------------"
  echo exported cert in "$cert_path"
  echo "===================================="
else
  echo "------------------------------------"
  echo failed to export cert
  echo "===================================="
fi

# If you also want to install the cert system-wide (so that hasura CLI works with HTTPS),
# set "--install" or "-i" argument. Requires root privileges.
# Note that this step:
#   - is not recommended by Pebble.
#   - does not affect Google Chrome certificate store (you still need to install cert manually into Google Chrome)
arg=$1
if [ "$arg" == "--install" -o "$arg" == "-i" ]; then
  sudo cp "$cert_path" /usr/local/share/ca-certificates/pebble.crt
  sudo update-ca-certificates
fi
