#!/bin/bash
#
# The MIT License
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.
#

set -e

PKCS12_PASSWORD="changeit"
INFLUXDB_SERVER_NAME="influxdb"
OTHER_SERVER_NAME="other-server"

# Gen Server .crt and .key with SAN (Subject Alternative Names) then convert to .p12 file
for SERVER_NAME in "$INFLUXDB_SERVER_NAME" "$OTHER_SERVER_NAME"; do
  echo "### Generate server key and certificate for $SERVER_NAME..."
  openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 | openssl pkcs8 -topk8 -nocrypt -out ../internal/test/certificates/"${SERVER_NAME}.key"

  openssl req -new -x509 -key ../internal/test/certificates/"${SERVER_NAME}.key" -out ../internal/test/certificates/"${SERVER_NAME}.crt" -days 3650 \
    -subj "/CN=localhost/O=Development/C=US" \
    -addext "subjectAltName = DNS:localhost,IP:127.0.0.1"

  openssl pkcs12 -export -in ../internal/test/certificates/"${SERVER_NAME}.crt" -inkey ../internal/test/certificates/"${SERVER_NAME}.key" \
    -out ../internal/test/certificates/"${SERVER_NAME}.p12" -name "myalias" -password "pass:$PKCS12_PASSWORD"
done

# Gen Client .crt and .key then convert to .p12 file
echo "### Generate client key..."
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 | openssl pkcs8 -topk8 -nocrypt -out ../internal/test/certificates/client.key
openssl req -new -x509 -key ../internal/test/certificates/client.key -out ../internal/test/certificates/client.crt -days 3650 \
  -subj "/CN=test-client/O=Development/C=US"
openssl pkcs12 -export -in ../internal/test/certificates/client.crt -inkey ../internal/test/certificates/client.key \
  -out ../internal/test/certificates/client.p12 -name "myalias" -password "pass:$PKCS12_PASSWORD"

#Encrypt pkcs8 for client.key for testing
openssl pkcs8 -topk8 -v2 aes-256-cbc -iter 1000000 -in ../internal/test/certificates/client.key -out ../internal/test/certificates/client_pkcs8.key