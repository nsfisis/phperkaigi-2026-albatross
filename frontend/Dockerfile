FROM node:22.14 AS builder

ARG ALBATROSS_HOST
ARG ALBATROSS_BASE_PATH

WORKDIR /app
ENV NODE_ENV=production

COPY package.json package-lock.json ./
COPY frontend/package.json frontend/
RUN npm install -w frontend --include=dev

COPY frontend/ frontend/
ENV ALBATROSS_BASE_PATH="$ALBATROSS_BASE_PATH"
RUN npm -w frontend run build

################################################################################
FROM nginx:1.27-alpine

COPY --from=builder /app/frontend/dist /usr/share/nginx/html
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
