# build stage
FROM golang:1-alpine AS build
WORKDIR /src
COPY . .
RUN go build -o /out/ecs-ingress .

# serve stage
FROM nginx:stable-alpine as serve
COPY --from=build /out/ecs-ingress /app/ecs-ingress
COPY --from=build /src/data/upstreams.conf.tmpl /app/nginx/upstreams.conf.tmpl

EXPOSE 80
CMD /app/ecs-ingress