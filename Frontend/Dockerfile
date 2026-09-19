# Builds the SPA with Vite, then serves the static output with nginx.
FROM node:24-alpine AS build

WORKDIR /app

# Dependencies first, so this layer is only rebuilt when the manifests change.
# npm ci installs exactly the lockfile and fails if it and package.json disagree.
COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# tsc runs as part of this, so a type error fails the image build rather than
# shipping a broken bundle.
RUN npm run build

FROM nginx:1.27-alpine

# The entrypoint substitutes ${API_UPSTREAM} from the environment into this
# template, so the upstream can change without rebuilding the image.
COPY nginx.conf.template /etc/nginx/templates/default.conf.template
COPY --from=build /app/dist /usr/share/nginx/html

ENV API_UPSTREAM="http://api:8080"

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1/ >/dev/null || exit 1
