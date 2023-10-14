# How to create self-signed certificates for registry

1. Create a private key
```
openssl genrsa -des3 -out domain.key 2048

```

2. Creating a Certificate Signing Request

```
openssl req -key domain.key -new -out domain.csr
```

```
Enter pass phrase for domain.key:
You are about to be asked to enter information that will be incorporated
into your certificate request.
What you are about to enter is what is called a Distinguished Name or a DN.
There are quite a few fields but you can leave some blank
For some fields there will be a default value,
If you enter '.', the field will be left blank.
-----
Country Name (2 letter code) [AU]:VN
State or Province Name (full name) [Some-State]:HCM                        
Locality Name (eg, city) []:HCM
Organization Name (eg, company) [Internet Widgits Pty Ltd]:Manabie
Organizational Unit Name (eg, section) []:Platform
Common Name (e.g. server FQDN or YOUR name) []:kind-reg.actions-runner-system.svc
Email Address []:email@email.com

Please enter the following 'extra' attributes
to be sent with your certificate request
A challenge password []:
An optional company name []:
```

3. Create a Self-Signed Root CA

```
openssl req -x509 -sha256 -days 1825 -newkey rsa:2048 -keyout rootCA.key -out rootCA.crt
```

4. Sign Our CSR With Root CA
- Create domain.ext file as below:

```
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
subjectAltName = @alt_names
[alt_names]
DNS.1 = kind-reg.actions-runner-system.svc
```

5. Create certificate

```
openssl x509 -req -CA rootCA.crt -CAkey rootCA.key -in domain.csr -out domain.crt -days 3650 -CAcreateserial -extfile domain.ext
```

6. Create secrets in kubernetes:

```
kubectl -n actions-runner-system create secret generic kind-shared-registry-secret\
  --from-file=tls.cert=domain.crt\
  --from-file=tls.key=domain.key

 kubectl -n actions-runner-system create secret generic kind-shared-registry-ca\    
  --from-file=ca.crt=rootCA.crt 
```

7. Mount secret to registry and runners