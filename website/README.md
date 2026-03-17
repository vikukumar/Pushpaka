# Pushpaka Website

Modern, enterprise-grade website for Pushpaka cloud deployment platform.

Built with **Astro** and **Tailwind CSS** with a beautiful metallic design system.

## 🚀 Quick Start

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## 📁 Project Structure

```
website/
├── src/
│   ├── pages/          # Astro pages (routes)
│   │   ├── index.astro      (Home page)
│   │   ├── features.astro   (Features page)
│   │   ├── install.astro    (Installation guide)
│   │   ├── docs.astro       (Documentation)
│   │   └── releases.astro   (Release tracker)
│   ├── layouts/        # Reusable layouts
│   │   └── Layout.astro
│   ├── components/     # Reusable components
│   │   ├── Header.astro
│   │   └── Footer.astro
│   └── styles/         # Global styles
│       └── global.css
├── public/             # Static assets
├── astro.config.mjs    # Astro configuration
├── tailwind.config.mjs # Tailwind configuration
└── package.json
```

## 🎨 Design System

- **Primary Color**: Indigo (#6366f1)
- **Accent Color**: Cyan (#06b6d4)
- **Background**: Dark blue (#0f172a)
- **Metallic Effect**: Gradient shadows with glassmorphism

## 📄 Pages

- **Home**: Hero section with key features
- **Features**: Comprehensive feature list with categories
- **Installation**: Platform-specific installation guides
- **Documentation**: API docs, core concepts, and guides
- **Releases**: Version history and download links

## 🌐 Deployment

### GitHub Pages (Automatic)

Push to `main` branch → Automatic deployment via GitHub Actions

```bash
git add .
git commit -m "Update website"
git push origin main
```

Website will be available at: `https://vikukumar.github.io/Pushpaka/`

### Manual Deployment

```bash
# Build the site
npm run build

# The `dist/` folder contains the static site
# Deploy to any static hosting provider
```

## 🔧 Tech Stack

- **Framework**: Astro 5.1
- **Styling**: Tailwind CSS 4.0
- **Runtime**: Node.js 20+
- **Deployment**: GitHub Pages

## 📦 Build Artifacts

- `dist/` - Production-ready static site

## 🎯 Features

✨ Modern, responsive design
🎨 Dark/light theme support (extensible)
📱 Mobile-optimized
⚡ Lightning-fast performance
🔍 SEO-friendly
♿ Accessibility-focused

## 📝 Configuration

Edit `astro.config.mjs` to customize:
- Site URL
- Base path
- Integrations

Edit `tailwind.config.mjs` for theme customization.

## 🚀 Performance

- Zero JavaScript by default
- Optimized static HTML
- ~100 Lighthouse performance score
- < 1s page load time

## 📚 Customization

### Colors
Update brand colors in `tailwind.config.mjs`:
```javascript
colors: {
  'brand': { ... },
  'accent': { ... }
}
```

### Content
Edit Astro pages in `src/pages/` to modify content.

### Components
Add new components in `src/components/` and import them.

## 🔗 Links

- [Pushpaka Repository](https://github.com/vikukumar/Pushpaka)
- [Astro Docs](https://docs.astro.build)
- [Tailwind Docs](https://tailwindcss.com)

## 📄 License

MIT - Same as Pushpaka project
