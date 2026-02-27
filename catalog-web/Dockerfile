FROM docker.io/node:20-alpine AS builder

WORKDIR /app

# Copy submodule package files first
COPY WebSocket-Client-TS/package*.json /app/WebSocket-Client-TS/
COPY Media-Types-TS/package*.json /app/Media-Types-TS/
COPY UI-Components-React/package*.json /app/UI-Components-React/
COPY Media-Browser-React/package*.json /app/Media-Browser-React/
COPY Media-Player-React/package*.json /app/Media-Player-React/
COPY Collection-Manager-React/package*.json /app/Collection-Manager-React/
COPY Dashboard-Analytics-React/package*.json /app/Dashboard-Analytics-React/
COPY Auth-Context-React/package*.json /app/Auth-Context-React/
COPY Catalogizer-API-Client-TS/package*.json /app/Catalogizer-API-Client-TS/

# Copy main app package files
COPY catalog-web/package*.json /app/catalog-web/

# Install dependencies for submodules
WORKDIR /app/Media-Types-TS
RUN npm ci --legacy-peer-deps && npm run build || true

WORKDIR /app/WebSocket-Client-TS
RUN npm ci --legacy-peer-deps && npm run build || true

WORKDIR /app/Catalogizer-API-Client-TS
RUN npm ci --legacy-peer-deps && npm run build || true

WORKDIR /app/Auth-Context-React
RUN npm ci --legacy-peer-deps && npm run build || true

WORKDIR /app/UI-Components-React
RUN npm ci --legacy-peer-deps && npm run build || true

WORKDIR /app/Media-Browser-React
RUN npm ci --legacy-peer-deps && npm run build || true

WORKDIR /app/Media-Player-React
RUN npm ci --legacy-peer-deps && npm run build || true

WORKDIR /app/Collection-Manager-React
RUN npm ci --legacy-peer-deps && npm run build || true

WORKDIR /app/Dashboard-Analytics-React
RUN npm ci --legacy-peer-deps && npm run build || true

# Copy submodule source
COPY WebSocket-Client-TS/ /app/WebSocket-Client-TS/
COPY Media-Types-TS/ /app/Media-Types-TS/
COPY UI-Components-React/ /app/UI-Components-React/
COPY Media-Browser-React/ /app/Media-Browser-React/
COPY Media-Player-React/ /app/Media-Player-React/
COPY Collection-Manager-React/ /app/Collection-Manager-React/
COPY Dashboard-Analytics-React/ /app/Dashboard-Analytics-React/
COPY Auth-Context-React/ /app/Auth-Context-React/
COPY Catalogizer-API-Client-TS/ /app/Catalogizer-API-Client-TS/

# Build main app
COPY catalog-web/ /app/catalog-web/
WORKDIR /app/catalog-web
RUN npm ci --legacy-peer-deps
RUN npm run build || npm run build -- --skipLibCheck

FROM docker.io/nginx:alpine

COPY --from=builder /app/catalog-web/dist /usr/share/nginx/html
COPY catalog-web/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
