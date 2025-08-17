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
FROM node:22.14

WORKDIR /app
ENV NODE_ENV=production

COPY package.json package-lock.json .
RUN npm install --omit=dev

COPY --from=builder /app/build ./build

ENV HOST="$ALBATROSS_HOST"
ENV PORT=80
EXPOSE 80
CMD ["npm", "run", "start"]
