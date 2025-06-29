#!/bin/bash

# Generate self-signed SSL certificate for localhost
echo "Generating SSL certificates for localhost..."

# Generate private key
openssl genrsa -out localhost-key.pem 2048

# Generate certificate signing request
openssl req -new -key localhost-key.pem -out localhost.csr -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"

# Generate self-signed certificate
openssl x509 -req -in localhost.csr -signkey localhost-key.pem -out localhost.pem -days 365

# Clean up CSR file
rm localhost.csr

echo "SSL certificates generated successfully!"
echo "Files created:"
echo "  - localhost-key.pem (private key)"
echo "  - localhost.pem (certificate)"
echo ""
echo "You can now start the development server with HTTPS support." 