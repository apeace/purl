FROM golang:1.25-alpine AS api-docs

WORKDIR /app
COPY api/go.mod api/go.sum ./
RUN go mod download
RUN go install github.com/swaggo/swag/cmd/swag@latest
COPY api/ .
RUN swag init --dir . --output docs

FROM node:22-alpine AS builder

WORKDIR /app
COPY package.json package-lock.json ./
COPY lib/package.json ./lib/
COPY purl/package.json ./purl/
RUN npm ci
COPY lib/ ./lib/
COPY purl/ ./purl/
COPY --from=api-docs /app/docs/ ./api/docs/
RUN npm run generate -w @purl/lib
ENV VITE_API_URL=https://api.justpurl.com
RUN npm run build -w @purl/purl

FROM nginx:alpine

COPY deploy/nginx/nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /app/purl/dist /usr/share/nginx/html

EXPOSE 80 443
