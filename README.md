# ECS Ingress

A simple NGINX based solution that allows to reverse proxy to services deployed on ECS.
It's designed to run as a deamon on the ECS cluster and provide reverse proxying to the ECS services in the cluster.

## How does it work

ECS Ingress is a small golang executable loosly modelled after [nginx-ingress from k8s](https://kubernetes.github.io/ingress-nginx/) but designed to be significantly simpler.

* It works by launching and managing a vanilla NGINX instance with a custom specified NGINX configuration bundle stored on S3.
* The NGINX configuration must reference a dynamically modified upstreams file `/app/nginx/upstreams.conf` where one upstream is created per [ECS Service](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs_services.html).
* It automatically sends config reloads to NGINX once the S3 config bundle OR the ECS configuration changes - because of new deployments, failovers or nginx config changes.
* It plays nicely with [AWS CodeDeploy](https://docs.aws.amazon.com/codedeploy/latest/userguide/welcome.html) so that the nginx configuration to be version controlled and automatically zipped to S3 upon change.

## Notes

* A valid NGINX configuration is required for the container to start properly. Subsequent configuration changes are accepted only if the new configuration passes the nginx config test.
* Only `RUNNING` tasks are dynamically injected inside the upstreams file.
* If a ECS service has no tasks running - because of failover or errors - a placeholder backend endpoint marked as DOWN is set to prevent missing reference errors in the main configuration file.
* ECS Ingress combines the NGINX logs and its internal ones in 1 stdout/stderr stream.
* ECS and Nginx config changes are polled every 10 seconds. Currently API requests against AWS resources are unmetered and free. S3 file updates are billed at the [current S3 GET request pricing](https://aws.amazon.com/s3/pricing/).

## Deployment
* ECS Ingress is designed to be deployed as a DAEMON in a ECS cluster with [HOST](https://docs.docker.com/network/host/) networking configuration binding on the ports opened by NGINX. The NGINX listening port numbers need to be referenced in the [ECS Task Definition](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definitions.html) for the DEAMON service. 
* it can be placed behind a [Network Load Balancer](https://aws.amazon.com/elasticloadbalancing/network-load-balancer/) for HTTPs translation on the ELB or directly referenced from Route53/your DNS provider through multiple A records - one per ECS Cluster instance.

## Environment Variables

| ENV Variable  | Default value | Meaning |
| ------------- | ------------- | ------- |
| `AWS_CLUSTER_NAME`  | `default` | the name of the ECS Cluster to reference |
| `AWS_REGION`  | `ap-southeast-2` | the AWS Region id |
| `NGINX_CONFIG_FILE_NAME` | `nginx.conf` | the nginx config file to reference in the S3 bundle |
| `NGINX_CONFIG_BUNDLE_S3_BUCKET` |  | the S3 bucket for the config bundle |
| `NGINX_CONFIG_BUNDLE_S3_KEY` |  | the S3 key for the config bundle |


## Example Nginx config bundle

```
user nobody;
worker_processes auto;
pid /run/nginx.pid;

events {
  worker_connections 768;
}

http {
  sendfile on;
  tcp_nopush on;
  tcp_nodelay on;
  keepalive_timeout 65;
  types_hash_max_size 2048;
  server_tokens off;

  # all upstreams
  # this is the dynamic reference that always needs to be there
  include /app/nginx/upstreams.conf;

  server {

    server_name www.example.com;
    
    location / {
      return 403;
    }

    location /apis {
      # api-prod should be defined inside /app/nginx/upstreams.conf and is dynamically updated
      proxy_pass http://api-prod;
    }

  } 
  
  server {

    server_name app.example.com;
    
    location / {
      proxy_pass http://app-ui-prod;
    }

    location /v2/api {
      proxy_pass http://app-api-prod;
    }

  } 
  
}

# Example with TCP stream connector
stream {

  # all upstreams
  # this needs to be repeated here as it's context sensitive - http and stream
  include /app/nginx/upstreams.conf;

  log_format basic '$remote_addr [$time_local] '
                 '$protocol $status $bytes_sent $bytes_received '
                 '$session_time';

  server {
    listen                  1883 so_keepalive=on;
    proxy_pass              mqtt-server:1883;
    proxy_connect_timeout 1s;
  }

}
```




## Roadmap
* Automatic support for Route53 updates to reflect changes in the instances attached to a ECS cluster
* [Letsencrypt](https://letsencrypt.org/) support to automatically generate new HTTPS certificates
* [Gossip protocol](https://github.com/hashicorp/memberlist) coordination across running containers in a cluster to coordinate Letsencrypt requests
* Move to [openresty](https://openresty.org/en/) to avoid potentially costly config reloads from NGINX

## Caveat emptor
Use at your own risk.

## License
MIT / Apache2

Author [Stefano Fratini](https://www.linkedin.com/in/stefanofratini610/)