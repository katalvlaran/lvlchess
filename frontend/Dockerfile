# 1. build stage
FROM node:16-alpine as build
WORKDIR /app

# Копируем package.json и (при наличии) package-lock.json
COPY package*.json ./
# Устанавливаем зависимости
RUN npm install
# RUN npm ci

# Копируем исходники
COPY . ./

# Собираем production build
RUN npm run build

# 2. final stage
FROM nginx:alpine
# Копируем build-результат (папку build/) из первого контейнера
COPY --from=build /app/build /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
