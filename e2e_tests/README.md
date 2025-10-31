# End-to-end tests
Tests check collecting metrics for different exporter configs.

Cases:
* launching the exporter without specifying the configuration file;
* launching the exporter with an empty configuration file;
* launch of the exporter specifying the configuration file for TLS and without basic auth;
* launch of the exporter specifying the configuration file for TLS (certificates inline) and without basic auth;
* launch of the exporter specifying the configuration file for TLS and basic auth;
* launch of the exporter specifying the configuration file for basic auth and without TLS;
* launch of the exporter specifying the configuration file for TLS and client certificate (any cert);
* launch of the exporter specifying the configuration file for TLS and client certificate (signed cert).

## Generate certificates and keys
The certificates and keys in `e2e_tests` directory are used only for end-to-end tests and are not used for actual services.

### Generate server.crt and server.key

Generate a self-signed SSL certificate and private key using OpenSSL. The certificate's subject is set to `/O=medusa_exporter/OU=medusa_exporter`. 

```bash
openssl req -x509 -newkey rsa:2048 -keyout server.key -out server.crt -days 36500 -nodes -subj "/O=medusa_exporter/OU=medusa_exporter" -sha256
```

### Generate user.pem and user.key

Generate user certificate and private key and sign it with the server certificate.

```bash
# Generate user.key and user.csr.
openssl req -newkey rsa:2048 -nodes -keyout user.key -out user.csr -subj "/O=medusa_exporter/OU=medusa_exporter"
# Sign user.csr with server.crt.
openssl x509 -req -in user.csr -CA server.crt -CAkey server.key -CAcreateserial -out user.crt -days 36500
# Convert user.crt to user.pem.
openssl x509 -outform pem -in user.crt -out user.pem
```

## Build image for tests
### Build with Release Medusa (Default)
Build using the official Medusa release from PyPI:

```bash
docker build \
  --file e2e_tests/Dockerfile \
  --tag medusa_exporter_e2e \
  .
```

or with specific Cassandra and Medusa versions:

```bash
docker build \
  --file e2e_tests/Dockerfile \
  --tag medusa_exporter_e2e \
  --build-arg CASSANDRA_VERSION=5.0.6 \
  --build-arg MEDUSA_VERSION=0.25.1 \
  --build-arg MEDUSA_BUILD_TYPE=release \
  .
```

### Build with Custom Medusa
Build using a custom Medusa version from a specific Git repository/branch:

```bash
docker build \
  --file e2e_tests/Dockerfile \
  --tag medusa_exporter_e2e \
  --build-arg MEDUSA_BUILD_TYPE=custom \
  --build-arg MEDUSA_REPO_TAG=add_json_output \
  --build-arg MEDUSA_REPO_URL=https://github.com/woblerr/cassandra-medusa \
  .
```