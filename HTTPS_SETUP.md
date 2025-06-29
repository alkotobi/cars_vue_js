# HTTPS Setup for Development

This project requires HTTPS for microphone access in the chat feature. Modern browsers only allow microphone access over secure connections.

## Quick Setup

1. **Generate SSL Certificates** (already done):

   ```bash
   ./generate-ssl.sh
   ```

2. **Start Development Server**:

   ```bash
   ./dev.sh
   ```

3. **Access the Application**:
   - Frontend: https://localhost:5173
   - API: https://localhost:8000

## Manual Setup (if needed)

If you need to regenerate the certificates:

1. **Generate SSL Certificates**:

   ```bash
   chmod +x generate-ssl.sh
   ./generate-ssl.sh
   ```

2. **Start PHP Server with HTTPS**:

   ```bash
   php -S localhost:8000 server.php --ssl-cert localhost.pem --ssl-key localhost-key.pem
   ```

3. **Start Vite Dev Server**:
   ```bash
   npm run dev
   ```

## Browser Security Warning

When accessing the site for the first time, your browser will show a security warning because we're using a self-signed certificate. This is normal for development.

**To proceed:**

1. Click "Advanced" or "Show Details"
2. Click "Proceed to localhost (unsafe)" or similar option
3. The site will work normally with microphone access enabled

## Production

For production deployment, ensure your domain (cars.merhab.com) has a valid SSL certificate from a trusted certificate authority.

## Troubleshooting

- **"Could not access microphone"**: Make sure you're using HTTPS and have allowed microphone permissions
- **SSL Certificate Error**: Accept the self-signed certificate warning in your browser
- **Port Already in Use**: Stop any existing servers and try again
