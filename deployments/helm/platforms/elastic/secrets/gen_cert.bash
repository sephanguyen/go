#!/bin/bash
# ORG=manabie ENV=staging NAME=admin ./gen_cert.bash
# subject=CN = admin.manabie.staging.search

openssl genrsa -out ./$ORG/$ENV/$NAME-key-temp.pem 2048
openssl pkcs8 -inform PEM -outform PEM -in ./$ORG/$ENV/$NAME-key-temp.pem -topk8 -nocrypt -v1 PBE-SHA1-3DES -out ./$ORG/$ENV/$NAME-key.pem
openssl req -new -key ./$ORG/$ENV/$NAME-key.pem -out ./$ORG/$ENV/$NAME.csr -subj "/CN=$NAME.$ORG.$ENV.search"
openssl x509 -req -in ./$ORG/$ENV/$NAME.csr -CA ./$ORG/$ENV/root-ca.pem -CAkey ./$ORG/$ENV/root-ca-key.pem -CAcreateserial -sha256 -out ./$ORG/$ENV/$NAME.pem -days 3650

rm ./$ORG/$ENV/$NAME-key-temp.pem
rm ./$ORG/$ENV/$NAME.csr
