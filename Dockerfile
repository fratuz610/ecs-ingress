# build stage
FROM golang:1-alpine AS build
WORKDIR /src
COPY . .
RUN go build -o /out/rev-proxy .

# serve stage
FROM nginx:stable-alpine as serve
COPY --from=build /out/rev-proxy /app/rev-proxy
COPY --from=build /src/data/upstreams.conf.tmpl /app/upstreams.conf.tmpl
COPY --from=build /src/data/nginx.conf /app/nginx.conf

EXPOSE 80
CMD /app/rev-proxy