# Cars Management System

This template should help get you started developing with Vue 3 in Vite.

## Environment Setup

### Google Maps API Key

To use the map functionality, you need to set up a Google Maps API key:

1. **Get an API Key**: Go to [Google Cloud Console](https://console.cloud.google.com/)
2. **Create a new project** or select an existing one
3. **Enable the Maps JavaScript API** in the APIs & Services section
4. **Create credentials** (API Key) in the Credentials section
5. **Add your domain** to the authorized domains list (e.g., `https://cars.merhab.com`)

### Environment Variables

Create a `.env` file in the root directory with the following variables:

```env
# Google Maps API Key
VITE_GOOGLE_MAPS_API_KEY=YOUR_GOOGLE_MAPS_API_KEY_HERE

# Other environment variables can be added here
# VITE_API_BASE_URL=https://your-api-domain.com
# VITE_APP_TITLE=Cars Management System
```

## Recommended IDE Setup

[VSCode](https://code.visualstudio.com/) + [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar) (and disable Vetur).

## Customize configuration

See [Vite Configuration Reference](https://vite.dev/config/).

## Project Setup

```sh
npm install
```

### Compile and Hot-Reload for Development

```sh
npm run dev
```

### Compile and Minify for Production

```sh
npm run build
```
