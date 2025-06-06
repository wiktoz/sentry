# Dockerfile
FROM arm32v5/nginx:latest

# Remove default site and add custom static content
RUN rm -rf /usr/share/nginx/html/*

COPY ./dist /usr/share/nginx/html

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
