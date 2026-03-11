FROM nginx:alpine

# 移除默认配置
RUN rm /etc/nginx/conf.d/default.conf

# 拷贝自定义 Nginx 配置
COPY nginx.conf /etc/nginx/conf.d/

# 拷贝静态资源到 Nginx HTML 目录
COPY public/ /usr/share/nginx/html/

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
