// PWA utilities for service worker registration and management

/**
 * Register the service worker
 */
export function registerServiceWorker() {
  if ('serviceWorker' in navigator) {
    window.addEventListener('load', () => {
      navigator.serviceWorker
        .register('/sw.js')
        .then((registration) => {
          console.log('Service Worker registered:', registration);

          // Check for updates every hour
          setInterval(() => {
            registration.update();
          }, 60 * 60 * 1000);

          // Handle updates
          registration.addEventListener('updatefound', () => {
            const newWorker = registration.installing;

            newWorker.addEventListener('statechange', () => {
              if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
                // New service worker available
                showUpdateNotification(registration);
              }
            });
          });
        })
        .catch((error) => {
          console.error('Service Worker registration failed:', error);
        });
    });
  }
}

/**
 * Unregister the service worker
 */
export async function unregisterServiceWorker() {
  if ('serviceWorker' in navigator) {
    const registrations = await navigator.serviceWorker.getRegistrations();
    for (const registration of registrations) {
      await registration.unregister();
    }
  }
}

/**
 * Show update notification to user
 */
function showUpdateNotification(registration) {
  // Create a custom event that can be handled by the app
  const event = new CustomEvent('sw-update', {
    detail: { registration }
  });
  window.dispatchEvent(event);

  // You can also show a native notification here
  if ('Notification' in window && Notification.permission === 'granted') {
    new Notification('Update Available', {
      body: 'A new version is available. Refresh to update.',
      icon: '/icons/icon-192x192.png',
      tag: 'app-update'
    });
  }
}

/**
 * Request notification permission
 */
export async function requestNotificationPermission() {
  if ('Notification' in window && Notification.permission === 'default') {
    const permission = await Notification.requestPermission();
    return permission === 'granted';
  }
  return Notification.permission === 'granted';
}

/**
 * Subscribe to push notifications
 */
export async function subscribeToPushNotifications() {
  if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
    console.warn('Push notifications not supported');
    return null;
  }

  try {
    const registration = await navigator.serviceWorker.ready;

    // Check if already subscribed
    let subscription = await registration.pushManager.getSubscription();

    if (!subscription) {
      // Request permission if not already granted
      const permission = await requestNotificationPermission();
      if (!permission) {
        console.warn('Notification permission denied');
        return null;
      }

      // Create new subscription
      // Note: You need to generate VAPID keys for your application
      // Use: npx web-push generate-vapid-keys
      const vapidPublicKey = import.meta.env.VITE_VAPID_PUBLIC_KEY;

      if (!vapidPublicKey) {
        console.warn('VAPID public key not configured');
        return null;
      }

      subscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(vapidPublicKey)
      });
    }

    return subscription;
  } catch (error) {
    console.error('Failed to subscribe to push notifications:', error);
    return null;
  }
}

/**
 * Unsubscribe from push notifications
 */
export async function unsubscribeFromPushNotifications() {
  if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
    return false;
  }

  try {
    const registration = await navigator.serviceWorker.ready;
    const subscription = await registration.pushManager.getSubscription();

    if (subscription) {
      await subscription.unsubscribe();
      return true;
    }

    return false;
  } catch (error) {
    console.error('Failed to unsubscribe from push notifications:', error);
    return false;
  }
}

/**
 * Check if the app is installed as PWA
 */
export function isInstalled() {
  return (
    window.matchMedia('(display-mode: standalone)').matches ||
    window.navigator.standalone === true
  );
}

/**
 * Check if the app can be installed
 */
export function canInstall() {
  return 'BeforeInstallPromptEvent' in window;
}

/**
 * Handle PWA install prompt
 */
export function setupInstallPrompt() {
  let deferredPrompt = null;

  window.addEventListener('beforeinstallprompt', (e) => {
    // Prevent the mini-infobar from appearing on mobile
    e.preventDefault();

    // Store the event for later use
    deferredPrompt = e;

    // Dispatch custom event that can be handled by the app
    const event = new CustomEvent('pwa-installable', {
      detail: { prompt: deferredPrompt }
    });
    window.dispatchEvent(event);
  });

  window.addEventListener('appinstalled', () => {
    console.log('PWA installed');
    deferredPrompt = null;

    // Dispatch custom event
    const event = new CustomEvent('pwa-installed');
    window.dispatchEvent(event);
  });

  return {
    async showInstallPrompt() {
      if (!deferredPrompt) {
        return false;
      }

      // Show the install prompt
      deferredPrompt.prompt();

      // Wait for the user's response
      const { outcome } = await deferredPrompt.userChoice;

      console.log(`User response to install prompt: ${outcome}`);

      // Clear the saved prompt
      deferredPrompt = null;

      return outcome === 'accepted';
    },

    isPromptAvailable() {
      return deferredPrompt !== null;
    }
  };
}

/**
 * Cache URLs for offline use
 */
export async function cacheUrls(urls) {
  if ('serviceWorker' in navigator && navigator.serviceWorker.controller) {
    navigator.serviceWorker.controller.postMessage({
      type: 'CACHE_URLS',
      urls
    });
  }
}

/**
 * Clear all caches
 */
export async function clearCaches() {
  if ('caches' in window) {
    const cacheNames = await caches.keys();
    await Promise.all(
      cacheNames.map((name) => caches.delete(name))
    );
  }
}

/**
 * Get cache size
 */
export async function getCacheSize() {
  if ('caches' in window && 'storage' in navigator && 'estimate' in navigator.storage) {
    const estimate = await navigator.storage.estimate();
    return {
      usage: estimate.usage,
      quota: estimate.quota,
      percentage: (estimate.usage / estimate.quota) * 100
    };
  }
  return null;
}

/**
 * Check if the device is online
 */
export function isOnline() {
  return navigator.onLine;
}

/**
 * Setup online/offline event listeners
 */
export function setupOnlineListeners(onOnline, onOffline) {
  window.addEventListener('online', onOnline);
  window.addEventListener('offline', onOffline);

  return () => {
    window.removeEventListener('online', onOnline);
    window.removeEventListener('offline', onOffline);
  };
}

/**
 * Helper function to convert VAPID key
 */
function urlBase64ToUint8Array(base64String) {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding)
    .replace(/\-/g, '+')
    .replace(/_/g, '/');

  const rawData = window.atob(base64);
  const outputArray = new Uint8Array(rawData.length);

  for (let i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i);
  }

  return outputArray;
}
