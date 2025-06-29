# HTTPS Test Guide

After deploying your Vue.js application, test these steps to ensure HTTPS is working properly:

## 1. Test HTTPS Access

- Visit: `https://cars.merhab.com`
- Should load without security warnings
- Check browser address bar shows 🔒 (secure)

## 2. Test HTTP Redirect

- Visit: `http://cars.merhab.com`
- Should automatically redirect to `https://cars.merhab.com`

## 3. Test API Calls

- Open browser developer tools (F12)
- Go to Network tab
- Use any feature that makes API calls
- Check that all requests go to `https://cars.merhab.com/api/...`

## 4. Test Microphone Access

- Go to Chat feature
- Click the microphone button
- Should either:
  - ✅ Start recording (if permissions granted)
  - ⚠️ Ask for microphone permission
  - ❌ Show specific error message (not HTTPS error)

## 5. Check Console for Errors

- Open browser developer tools
- Check Console tab for:
  - Mixed content warnings
  - CORS errors
  - API connection errors

## Common Issues & Solutions:

### Mixed Content Warning

- **Problem**: Some resources still loading over HTTP
- **Solution**: Check all API calls use HTTPS

### CORS Error

- **Problem**: API server not configured for HTTPS
- **Solution**: Ensure your API server supports HTTPS

### Microphone Still Not Working

- **Problem**: Browser still blocking microphone
- **Solution**:
  1. Check site is on HTTPS
  2. Clear browser cache
  3. Check browser permissions
  4. Try in incognito/private mode

## Browser Permissions Check:

1. Click the lock icon in address bar
2. Check microphone permission is "Allow"
3. If blocked, click "Allow" and refresh page

## If Still Having Issues:

1. Check browser console for specific error messages
2. Verify your hosting provider has SSL properly configured
3. Ensure all API endpoints are accessible via HTTPS
4. Test with different browsers (Chrome, Firefox, Safari)
