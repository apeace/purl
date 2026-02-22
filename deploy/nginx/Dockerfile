FROM node:22-alpine AS builder

WORKDIR /app
COPY package.json package-lock.json ./
COPY lib/package.json ./lib/
COPY purl/package.json ./purl/
RUN npm ci
COPY lib/ ./lib/
COPY purl/ ./purl/
ENV VITE_API_URL=https://api.justpurl.com
RUN npm run build -w @purl/purl

FROM nginx:alpine

COPY deploy/nginx/nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /app/purl/dist /usr/share/nginx/html

EXPOSE 80 443
