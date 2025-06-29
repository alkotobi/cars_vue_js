# Production HTTPS Setup for cars.merhab.com

To enable microphone access in the chat feature, your production site needs to run over HTTPS.

## Option 1: Let's Encrypt (Recommended - Free)

### 1. Install Certbot

```bash
# For Ubuntu/Debian
sudo apt update
sudo apt install certbot

# For CentOS/RHEL
sudo yum install certbot
```

### 2. Get SSL Certificate

```bash
# For Apache
sudo certbot --apache -d cars.merhab.com

# For Nginx
sudo certbot --nginx -d cars.merhab.com

# For standalone (if using custom server)
sudo certbot certonly --standalone -d cars.merhab.com
```

### 3. Auto-renewal

```bash
# Test auto-renewal
sudo certbot renew --dry-run

# Add to crontab for automatic renewal
sudo crontab -e
# Add this line:
0 12 * * * /usr/bin/certbot renew --quiet
```

## Option 2: Manual SSL Certificate

### 1. Generate CSR

```bash
openssl req -new -newkey rsa:2048 -nodes -keyout cars.merhab.com.key -out cars.merhab.com.csr
```

### 2. Get Certificate from CA

Submit the CSR to your certificate authority and download the certificate.

### 3. Install Certificate

Place the certificate files in `/etc/ssl/certs/` and configure your web server.

## Web Server Configuration

### Apache Configuration

```apache
<VirtualHost *:80>
    ServerName cars.merhab.com
    Redirect permanent / https://cars.merhab.com/
</VirtualHost>

<VirtualHost *:443>
    ServerName cars.merhab.com
    DocumentRoot /var/www/cars.merhab.com

    SSLEngine on
    SSLCertificateFile /etc/letsencrypt/live/cars.merhab.com/fullchain.pem
    SSLCertificateKeyFile /etc/letsencrypt/live/cars.merhab.com/privkey.pem

    # Security headers
    Header always set Strict-Transport-Security "max-age=31536000; includeSubDomains"
    Header always set X-Frame-Options DENY
    Header always set X-Content-Type-Options nosniff

    # Your existing configuration here...
</VirtualHost>
```

### Nginx Configuration

```nginx
server {
    listen 80;
    server_name cars.merhab.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name cars.merhab.com;

    ssl_certificate /etc/letsencrypt/live/cars.merhab.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/cars.merhab.com/privkey.pem;

    # Security settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;

    root /var/www/cars.merhab.com;
    index index.html;

    # Your existing configuration here...
}
```

## Frontend Deployment

### 1. Build for Production

```bash
npm run build
```

### 2. Deploy to Server

Upload the `dist` folder contents to your web server's document root.

### 3. Update .htaccess (for Apache)

```apache
# Force HTTPS
RewriteEngine On
RewriteCond %{HTTPS} off
RewriteRule ^(.*)$ https://%{HTTP_HOST}%{REQUEST_URI} [L,R=301]

# Handle Vue Router
RewriteBase /
RewriteRule ^index\.html$ - [L]
RewriteCond %{REQUEST_FILENAME} !-f
RewriteCond %{REQUEST_FILENAME} !-d
RewriteRule . /index.html [L]
```

## Testing

### 1. Check HTTPS

Visit `https://cars.merhab.com` and verify:

- ✅ SSL certificate is valid
- ✅ Site loads without mixed content warnings
- ✅ HTTP redirects to HTTPS

### 2. Test Microphone

- Open the chat feature
- Try to record a voice message
- Should now work without permission errors

## Troubleshooting

### Common Issues:

1. **Mixed Content Warnings**

   - Ensure all resources (images, scripts, API calls) use HTTPS
   - Check browser console for HTTP requests

2. **Certificate Not Trusted**

   - Verify certificate installation
   - Check certificate chain
   - Ensure intermediate certificates are included

3. **Redirect Loop**

   - Check web server configuration
   - Verify SSL certificate paths
   - Clear browser cache

4. **API Calls Failing**
   - Ensure API endpoint uses HTTPS
   - Check CORS configuration
   - Verify proxy settings

## Security Best Practices

1. **Enable HSTS** (HTTP Strict Transport Security)
2. **Use Strong Ciphers**
3. **Disable Weak Protocols** (TLS 1.0/1.1)
4. **Regular Certificate Renewal**
5. **Monitor Certificate Expiry**

## Quick Check Commands

```bash
# Check SSL certificate
openssl s_client -connect cars.merhab.com:443 -servername cars.merhab.com

# Test HTTPS redirect
curl -I http://cars.merhab.com

# Check certificate expiry
echo | openssl s_client -servername cars.merhab.com -connect cars.merhab.com:443 2>/dev/null | openssl x509 -noout -dates
```
