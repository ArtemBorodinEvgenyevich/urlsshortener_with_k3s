# ===========================================
# Base stage - общие зависимости
# ===========================================
FROM node:20-alpine AS base

WORKDIR /app

# Копируем package files
COPY package.json package-lock.json ./

# ===========================================
# Development stage - для разработки
# ===========================================
FROM base AS development

ARG VITE_API_URL=http://localhost:9091

# Устанавливаем все зависимости (включая dev)
RUN npm install

# Копируем исходный код
COPY . .

# Vite dev server по умолчанию на порту 5173
EXPOSE 5173

# Запускаем dev server с доступом извне (--host)
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

# ===========================================
# Production stage - сборка
# ===========================================
FROM base AS production

ARG VITE_API_URL=""

RUN npm ci

COPY . .

ENV VITE_API_URL=${VITE_API_URL}
RUN npm run build

# ===========================================
# Runtime stage - nginx для раздачи статики
# ===========================================
FROM nginx:alpine AS runtime

COPY --from=production /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]

