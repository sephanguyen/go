opendistro_security.ssl.transport.pemcert_filepath: decrypted/node.pem
opendistro_security.ssl.transport.pemkey_filepath: decrypted/node-key.pem
opendistro_security.ssl.transport.pemtrustedcas_filepath: decrypted/root-ca.pem
opendistro_security.ssl.transport.enforce_hostname_verification: false
opendistro_security.ssl.http.enabled: true
opendistro_security.allow_default_init_securityindex: true
opendistro_security.authcz.admin_dn:
  - 'CN={{ include "elastic.adminDn" . }}'
opendistro_security.nodes_dn:
  - 'CN=elasticsearch-{{ include "elastic.fullname" . }}-*'
opendistro_security.audit.type: internal_elasticsearch
opendistro_security.enable_snapshot_restore_privilege: true
opendistro_security.check_snapshot_restore_write_privileges: true
opendistro_security.restapi.roles_enabled: ["all_access", "security_rest_api_access"]
cluster.routing.allocation.disk.threshold_enabled: false
opendistro_security.audit.config.disabled_rest_categories: NONE
opendistro_security.audit.config.disabled_transport_categories: NONE

opendistro_security.ssl.http.clientauth_mode: OPTIONAL
opendistro_security.ssl.http.keystore_type: PKCS12
opendistro_security.ssl.http.keystore_filepath: /usr/share/elasticsearch/config/decrypted/keystore.jks
opendistro_security.ssl.http.keystore_password: elasticsecret
{{- if .Values.elasticsearch.snapshot.enabled }}
path.repo: ["/mnt/snapshots"]
{{- end }}
