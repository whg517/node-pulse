# Production Dockerfile for the Pulse frontend.
# Multi-stage: builds the Vite production bundle with Node, then serves the
# static assets with nginx. nginx rewrites unknown paths to index.html so the
# client-side BrowserRouter (history mode) handles deep links and refreshes.
#
# Build context: ./frontend   (run from repo root: docker build -f frontend/Dockerfile frontend/)
# The backend API is expected at /api and /ws on the same origin (see
# nginx.conf); point it at the Pulse server via a reverse proxy in production.

# ---- build stage ----
FROM node:25-alpine AS builder

WORKDIR /app

# Cache dependency install. Copy .npmrc alongside the manifests so the
# project's legacy-peer-deps setting (required for react-simple-maps under
# React 19) is honoured by npm ci.
COPY package*.json .npmrc ./
RUN npm ci

COPY . .

# Build with same-origin relative API URLs. The nginx reverse proxy in this
# image forwards /api and /ws to the Pulse backend, so the SPA must NOT embed
# an absolute host (the .env.production default is localhost:6532, which only
# works for local dev). Empty value = relative paths resolved by the browser.
RUN VITE_API_BASE_URL= npm run build

# ---- runtime stage ----
FROM nginx:1.27-alpine

# Replace the default site config with our SPA-aware one.
COPY --from=builder /app/dist /usr/share/nginx/html
# Build context is ./frontend, so nginx.conf is at the context root.
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

HEALTHCHECK --interval=10s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -q -O /dev/null http://localhost/ || exit 1
