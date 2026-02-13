FROM node:22.14 AS builder

ARG ALBATROSS_HOST
ARG ALBATROSS_BASE_PATH

WORKDIR /app
ENV NODE_ENV=production

COPY package.json package-lock.json .
RUN npm install --include=dev

COPY . .
ENV ALBATROSS_BASE_PATH="$ALBATROSS_BASE_PATH"
RUN npm run build

################################################################################
FROM nginx:1.27-alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
