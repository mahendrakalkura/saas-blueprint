# PWA Icons

This directory should contain PWA icons in various sizes for optimal display across different devices and platforms.

## Required Icon Sizes

The following icon sizes are required for PWA support:

- `icon-16x16.png` - Favicon (16x16)
- `icon-32x32.png` - Favicon (32x32)
- `icon-72x72.png` - iOS Home Screen Icon
- `icon-96x96.png` - Android Home Screen Icon
- `icon-128x128.png` - Chrome Web Store
- `icon-144x144.png` - Windows Tile Icon
- `icon-152x152.png` - iOS Home Screen Icon (iPad)
- `icon-192x192.png` - Android Chrome Icon (Recommended)
- `icon-384x384.png` - Android Chrome Icon
- `icon-512x512.png` - Android Chrome Splash Screen (Required)

## Generating Icons

### Option 1: Using PWA Asset Generator (Recommended)

```bash
# Install globally
npm install -g pwa-asset-generator

# Generate all icons from a single source image (1024x1024 or larger recommended)
pwa-asset-generator logo.png icons/ --icon-only --path-override ""
```

### Option 2: Using ImageMagick

If you have a source logo file (e.g., `logo.png`), you can use ImageMagick to generate all required sizes:

```bash
# Install ImageMagick (if not already installed)
# macOS: brew install imagemagick
# Ubuntu: sudo apt-get install imagemagick

# Generate icons
convert logo.png -resize 16x16 icon-16x16.png
convert logo.png -resize 32x32 icon-32x32.png
convert logo.png -resize 72x72 icon-72x72.png
convert logo.png -resize 96x96 icon-96x96.png
convert logo.png -resize 128x128 icon-128x128.png
convert logo.png -resize 144x144 icon-144x144.png
convert logo.png -resize 152x152 icon-152x152.png
convert logo.png -resize 192x192 icon-192x192.png
convert logo.png -resize 384x384 icon-384x384.png
convert logo.png -resize 512x512 icon-512x512.png
```

### Option 3: Online Tools

- [PWA Asset Generator Online](https://www.pwabuilder.com/)
- [Favicon Generator](https://realfavicongenerator.net/)
- [App Icon Generator](https://appicon.co/)

## Design Guidelines

### Icon Design Best Practices

1. **Size**: Start with a high-resolution source image (1024x1024 or larger)
2. **Format**: Use PNG format with transparency
3. **Safe Zone**: Keep important content within the center 80% of the icon
4. **Background**: Consider both light and dark backgrounds
5. **Simplicity**: Keep the design simple and recognizable at small sizes
6. **Brand**: Ensure consistency with your brand identity

### Maskable Icons

For Android adaptive icons, consider creating maskable icons:

```json
{
  "src": "/icons/icon-192x192.png",
  "sizes": "192x192",
  "type": "image/png",
  "purpose": "any maskable"
}
```

Maskable icons should have:
- Important content in the center 40% (safe zone)
- Background that extends to edges
- No transparency in corners

Test your maskable icons: https://maskable.app/

## Verification

After adding your icons, verify they work correctly:

1. **Local Testing**: Run your app locally and check the browser console for any missing icon warnings
2. **Lighthouse**: Run a Lighthouse PWA audit in Chrome DevTools
3. **Device Testing**: Test on actual devices (iOS and Android)

## Current Status

⚠️ **Placeholder icons needed!**

This directory currently doesn't contain any icons. Please generate and add your app's icons following the instructions above.

## Screenshots

For a better PWA experience, also add screenshots to the `screenshots/` directory:

- `desktop.png` - Desktop screenshot (1280x720 or larger)
- `mobile.png` - Mobile screenshot (750x1334 or similar mobile aspect ratio)

These screenshots will be used in the app installation prompt and app stores.
