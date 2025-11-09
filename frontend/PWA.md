# Progressive Web App (PWA) Implementation

This document describes the PWA implementation in the SaaS Blueprint application.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Components](#components)
- [Service Worker](#service-worker)
- [Offline Support](#offline-support)
- [Push Notifications](#push-notifications)
- [Installation](#installation)
- [Testing](#testing)
- [Configuration](#configuration)

## Overview

The SaaS Blueprint is implemented as a Progressive Web App (PWA), providing:

- **Installability**: Users can install the app on their devices
- **Offline Support**: Core functionality works without internet connection
- **Push Notifications**: Real-time notifications even when app is closed
- **App-like Experience**: Standalone mode without browser UI
- **Automatic Updates**: Service worker handles updates seamlessly

## Features

### ✅ Core PWA Features

- [x] Web App Manifest (`manifest.json`)
- [x] Service Worker with caching strategies
- [x] Offline fallback page
- [x] Install prompt with custom UI
- [x] Update prompt for new versions
- [x] Online/offline status indicator
- [x] Push notification support
- [x] Background sync capabilities
- [x] App shortcuts for quick actions
- [x] Theme color and splash screen

### 📱 Platform Support

- **Android**: Full PWA support with installation and notifications
- **iOS**: Install via "Add to Home Screen", limited notification support
- **Desktop**: Install via browser's install button (Chrome, Edge, etc.)

## Architecture

```
frontend/
├── public/
│   ├── manifest.json          # PWA manifest configuration
│   ├── sw.js                  # Service worker
│   ├── offline.html           # Offline fallback page
│   └── icons/                 # PWA icons (various sizes)
├── src/
│   ├── utils/
│   │   └── pwa.js            # PWA utility functions
│   └── components/
│       ├── PWAInstallPrompt.jsx   # Install prompt UI
│       ├── PWAUpdatePrompt.jsx    # Update prompt UI
│       └── OnlineStatus.jsx       # Online/offline indicator
```

## Components

### PWAInstallPrompt

Custom UI component for the app installation prompt.

**Features:**
- Shows when app is installable
- Auto-dismisses after 7 days if declined
- Remembers user preference
- Gradient design matching app theme

**Usage:**
```jsx
import PWAInstallPrompt from './components/PWAInstallPrompt'

function App() {
  return (
    <>
      {/* Your app content */}
      <PWAInstallPrompt />
    </>
  )
}
```

### PWAUpdatePrompt

Notifies users when a new version is available.

**Features:**
- Detects service worker updates
- One-click update and reload
- Shows loading state during update
- Can be dismissed to update later

**Usage:**
```jsx
import PWAUpdatePrompt from './components/PWAUpdatePrompt'

function App() {
  return (
    <>
      {/* Your app content */}
      <PWAUpdatePrompt />
    </>
  )
}
```

### OnlineStatus

Visual indicator for network connectivity.

**Features:**
- Shows when user goes offline
- Shows brief message when back online
- Auto-hides when online after 3 seconds
- Positioned at top center for visibility

**Usage:**
```jsx
import OnlineStatus from './components/OnlineStatus'

function App() {
  return (
    <>
      {/* Your app content */}
      <OnlineStatus />
    </>
  )
}
```

## Service Worker

### Caching Strategies

The service worker implements multiple caching strategies:

#### 1. **Precache** (Install-time)
Static assets cached during service worker installation:
- `/` (root)
- `/index.html`
- `/manifest.json`
- `/offline.html`

#### 2. **Network First** (Runtime)
API requests and HTML pages:
- Try network first
- Fall back to cache if offline
- Show offline page as last resort

#### 3. **Cache First** (Runtime)
Images and static assets:
- Check cache first
- Fetch from network if not cached
- Store in cache for future use

#### 4. **Stale While Revalidate** (Runtime)
JavaScript, CSS, and fonts:
- Serve from cache immediately
- Update cache in background
- Fresh content on next visit

### Cache Management

Three separate caches are maintained:

1. **`saas-blueprint-v1`**: Static assets and precached files
2. **`runtime-cache-v1`**: API responses and dynamic content
3. **`image-cache-v1`**: Images for better separation

Cache versioning ensures old caches are cleaned up on updates.

## Offline Support

### Offline Capabilities

The app works offline with these features:

1. **Cached Pages**: Previously visited pages load from cache
2. **Cached API Responses**: Recent API data available offline
3. **Offline Fallback**: Custom offline page for unavailable content
4. **Auto-sync**: Queued requests sent when back online

### Offline Page

The custom offline page (`/offline.html`):
- Beautiful, branded design
- Shows connection status
- Auto-reloads when online
- Retry button for manual check

### Testing Offline Mode

```bash
# In Chrome DevTools
1. Open DevTools (F12)
2. Go to Application tab
3. Select Service Workers
4. Check "Offline" checkbox
5. Reload the page
```

## Push Notifications

### Setup Push Notifications

1. **Generate VAPID Keys**:
```bash
npx web-push generate-vapid-keys
```

2. **Add to Environment**:
```env
VITE_VAPID_PUBLIC_KEY=your_public_key
```

3. **Request Permission**:
```javascript
import { requestNotificationPermission, subscribeToPushNotifications } from './utils/pwa'

// Request permission
const granted = await requestNotificationPermission()

// Subscribe to notifications
if (granted) {
  const subscription = await subscribeToPushNotifications()
  // Send subscription to your backend
}
```

### Handling Notifications

The service worker automatically handles:
- Push notifications
- Notification clicks
- Opening the app when clicked
- Focusing existing windows

## Installation

### User Installation

Users can install the app via:

1. **Chrome (Desktop)**:
   - Click install icon in address bar
   - Or: Menu → Install SaaS Blueprint

2. **Chrome (Android)**:
   - Banner prompt appears automatically
   - Or: Menu → Add to Home Screen

3. **Safari (iOS)**:
   - Share button → Add to Home Screen

### Programmatic Installation

```javascript
import { setupInstallPrompt } from './utils/pwa'

const installHandler = setupInstallPrompt()

// Check if prompt is available
if (installHandler.isPromptAvailable()) {
  // Show custom install button
  const accepted = await installHandler.showInstallPrompt()
}
```

## Testing

### PWA Audit

1. **Lighthouse** (Chrome DevTools):
```bash
# Open Chrome DevTools
# Lighthouse tab → PWA category → Generate report
```

Should score 100 in PWA category with all checks passing.

2. **PWA Builder**:
Visit https://www.pwabuilder.com/ and enter your URL.

### Device Testing

Test on actual devices:

**Android**:
- Chrome (full support)
- Samsung Internet (full support)
- Firefox (limited support)

**iOS**:
- Safari (install only, limited notifications)
- Chrome (uses Safari engine, same limitations)

**Desktop**:
- Chrome (full support)
- Edge (full support)
- Firefox (limited support)

### Manual Testing

```bash
# 1. Build the app
npm run build

# 2. Serve production build
npx serve -s dist

# 3. Test on localhost:3000
# - Check manifest loads
# - Check service worker registers
# - Test offline mode
# - Test install prompt
```

## Configuration

### Manifest Configuration

Edit `public/manifest.json`:

```json
{
  "name": "Your App Name",
  "short_name": "Short Name",
  "theme_color": "#3b82f6",
  "background_color": "#ffffff",
  "start_url": "/",
  "display": "standalone"
}
```

### Service Worker Configuration

Edit `public/sw.js`:

```javascript
// Update cache version when deploying
const CACHE_NAME = 'saas-blueprint-v2'

// Add more URLs to precache
const PRECACHE_ASSETS = [
  '/',
  '/index.html',
  '/your-page.html'
]
```

### Environment Variables

```env
# Push notifications (optional)
VITE_VAPID_PUBLIC_KEY=your_vapid_public_key
```

## Best Practices

### 1. **Update Strategy**

- Increment cache version on each deployment
- Test service worker updates before deploying
- Provide clear update prompts to users

### 2. **Cache Management**

- Don't cache authentication tokens
- Set appropriate cache TTLs
- Clean up old caches regularly

### 3. **Offline Experience**

- Show clear offline indicators
- Cache critical API responses
- Provide meaningful offline content

### 4. **Performance**

- Optimize image assets
- Use lazy loading for routes
- Minimize service worker script size

### 5. **Icons**

- Provide all required icon sizes
- Use maskable icons for Android
- Test icons on actual devices

## Troubleshooting

### Service Worker Not Registering

1. Check HTTPS (required except localhost)
2. Check browser console for errors
3. Verify `sw.js` is accessible at `/sw.js`

### Install Prompt Not Showing

1. Check manifest is valid (DevTools → Application → Manifest)
2. Ensure all PWA criteria are met (Lighthouse audit)
3. Clear site data and reload
4. Check if already installed

### Offline Mode Not Working

1. Check service worker is active
2. Verify caching strategies in `sw.js`
3. Check network tab for cache hits
4. Clear caches and re-test

### Updates Not Applying

1. Check cache version is incremented
2. Verify update prompt is shown
3. Clear service worker and caches
4. Hard reload (Ctrl+Shift+R)

## Resources

- [MDN PWA Guide](https://developer.mozilla.org/en-US/docs/Web/Progressive_web_apps)
- [web.dev PWA](https://web.dev/progressive-web-apps/)
- [Workbox Documentation](https://developers.google.com/web/tools/workbox)
- [PWA Builder](https://www.pwabuilder.com/)
- [Can I Use: Service Workers](https://caniuse.com/serviceworkers)

## Browser Support

| Feature | Chrome | Firefox | Safari | Edge |
|---------|--------|---------|--------|------|
| Service Worker | ✅ | ✅ | ✅ | ✅ |
| Install Prompt | ✅ | ❌ | ✅* | ✅ |
| Push Notifications | ✅ | ✅ | ❌ | ✅ |
| Background Sync | ✅ | ❌ | ❌ | ✅ |

*Safari requires manual "Add to Home Screen"

## License

This PWA implementation is part of the SaaS Blueprint project (MIT License).
