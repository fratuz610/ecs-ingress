# ECS Ingress

A simple NGINX based solution that allows to reverse proxy to services deployed on ECS.
It's designed to run as a deamon on the ECS cluster and provide reverse proxying to the ECS services in the cluster.

Available on Docker Hub: [fratuz610/ecs-ingress](https://hub.docker.com/r/fratuz610/ecs-ingress)

## Rationale

[Amazon ECS](https://aws.amazon.com/ecs/) is the cheaper and simpler proprietary alternative to K8s.

By leveraging the battle proven flexibility of NGINX, ECS Ingress lets you deploy a **sophisticated reverse proxy solution that supports HTTP/s, TCP and UDP load balancing with URL rewriting across any ECS Service** without paying the cost and wrestling with the limitedness of the Amazon provided load balancing solutions (ELBs/ALBs/NLBs).

## How does it work

ECS Ingress is a small golang executable loosly modelled after [nginx-ingress from k8s](https://kubernetes.github.io/ingress-nginx/) but significantly simpler.

* It works by launching and managing a vanilla NGINX instance with a custom specified NGINX configuration bundle stored on S3 and downloaded locally.
* The NGINX configuration in turns references a dynamically modified upstreams file `/app/nginx/upstreams.conf` continously kept in sync with the ECS cluster tasks and services.
* It automatically sends config reloads to NGINX every time the S3 config bundle OR the ECS configuration changes - because of new deployments, failovers or nginx config changes.
* It plays nicely with [AWS CodeDeploy](https://docs.aws.amazon.com/codedeploy/latest/userguide/welcome.html) so that the nginx configuration to be version controlled and automatically zipped to S3 upon new commits.

## Notes

* A valid NGINX configuration is required for the container **to start properly**. Subsequent configuration changes are accepted only if the new configuration passes the nginx config test without service disruptions in case of errors.
* AWS API calls are authenticated using ECS Role or AWS IAM credentials. See below.
* Only `RUNNING` tasks are dynamically injected inside the upstreams file. If a ECS service has no tasks running - because of failover or errors - a placeholder backend endpoint marked as DOWN is set to prevent missing reference errors in the main configuration file.
* ECS Ingress combines the NGINX logs and its internal ones in 1 stdout/stderr stream for easy ingestion into Cloudwatch Logs.
* ECS and Nginx config changes are polled **every 10 seconds**. Currently API requests against AWS resources are unmetered and **free**. S3 file requests are billed at the [current S3 GET request pricing](https://aws.amazon.com/s3/pricing/).

## Deployment
* ECS Ingress is designed to be deployed as a DAEMON in a ECS cluster with [HOST](https://docs.docker.com/network/host/) networking configuration binding on the ports opened by NGINX. The NGINX listening port numbers need to be referenced in the [ECS Task Definition](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definitions.html) for the DEAMON service. 
* it can be placed behind a [Network Load Balancer](https://aws.amazon.com/elasticloadbalancing/network-load-balancer/) for HTTPs translation on the ELB or directly referenced from Route53/your DNS provider through multiple A records - one per ECS Cluster instance.

## Environment Variables

| ENV Variable  | Default value | Meaning |
| ------------- | ------------- | ------- |
| `AWS_CLUSTER_NAME`  | `default` | the name of the ECS Cluster to reference |
| `AWS_REGION`  | `ap-southeast-2` | the AWS Region id |
| `NGINX_CONFIG_FILE_NAME` | `nginx.conf` | the nginx config file to reference in the S3 bundle |
| `NGINX_CONFIG_BUNDLE_S3_BUCKET` |  | the S3 bucket for the config bundle. |
| `NGINX_CONFIG_BUNDLE_S3_KEY` |  | the S3 key for the config bundle.<br/>Must be a ZIP file containing at least the `NGINX_CONFIG_FILE_NAME` file.<br/>It's unzipped in the `/app/nginx/` folder |
| `AWS_ACCESS_KEY_ID`| | the AWS Access Key to access the AWS Services.<br/>Leave blank if using ECS Roles. |
| `AWS_SECRET_ACCESS_KEY` | | the AWS Secret Access Key to access the AWS Services.<br/>Leave blank if using ECS Roles. |

## Example Nginx config file with HTTP load balancing

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

    server_name app.example.com;
    
    location / {
      # app-ui-prod should be the name of the ECS service
      proxy_pass http://app-ui-prod;
    }

    location /v2/api {
      # app-api-prod should be the name of the ECS service
      proxy_pass http://app-api-prod;
    }

  } 
  
}
```

## Example Nginx config file with HTTPS support

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

    server_name app.example.io;

    listen 443 ssl;
    listen [::]:443 ssl;

    ssl_certificate /app/nginx/fullchain.pem;
    ssl_certificate_key /app/nginx/privkey.pem;

    # we enable only more recent protocols
    ssl_protocols TLSv1.1 TLSv1.2;

    # as suggested by Nginx we prioritize newer ciphers
    ssl_ciphers  HIGH:!aNULL:!MD5;

    # we cache the ssl session parameters
    # to reduce the CPU load on the web server
    ssl_session_cache   shared:SSL:10m;
    ssl_session_timeout 10m;

    # we increase the keep alive timeout
    # to improve socket reuse and reduce
    # the need for SSL handshakes
    keepalive_timeout 70;

    location / {
      proxy_pass http://example-ui-prod;
    }

    location /v1/api {
      proxy_pass http://example-api-prod;
    }

  }
  
}
```

This example requires a bundle with the following files

| File | Description |
| ---- | ----------- |
| `nginx.conf` | the main nginx config file. The name be customized via `NGINX_CONFIG_FILE_NAME` env variable. |
| `privkey.pem` | the certificate private key |
| `fullchain.pem` | the certiticate public key chain |


## Example Nginx config file with TCP load balancing - MQTT

```
user nobody;
worker_processes auto;
pid /run/nginx.pid;

events {
  worker_connections 768;
}

# Example with TCP stream connector
stream {

  # all upstreams
  # this needs to be repeated here as it's context sensitive - http and stream
  include /app/nginx/upstreams.conf;

  server {
    listen                  1883 so_keepalive=on;
    proxy_pass              mqtt-server:1883;
    proxy_connect_timeout 1s;
  }

}
```

## Example Nginx config file with reverse proxying of Pgsql deployed on the cluster itself :)

```
user nobody;
worker_processes auto;
pid /run/nginx.pid;

events {
  worker_connections 768;
}

# PGSQL Connector to the postgres-prod upstream
stream {

  # all upstreams
  include /app/nginx/upstreams.conf;

  server {
    listen                  5432 so_keepalive=on;
    proxy_pass              postgres-prod;

    # allows access only internally
    allow  172.17.0.0/16;
    deny   all;
  }

}
```

You can connect to Pgsql on `172.17.0.1:5432` from each container in the cluster.

## AWS IAM resources

ECS Ingress works best when a ECS role is associated with the container via ECS Task definition like in the following example:

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeInstances",
                "ecs:ListServices",
                "ecs:ListTasks",
                "ecs:DescribeTasks",
                "ecs:DescribeContainerInstances"
            ],
            "Resource": "*"
        },
        {
            "Sid": "ListObjectsInBucket",
            "Effect": "Allow",
            "Action": [
                "s3:ListBucket"
            ],
            "Resource": [
                "arn:aws:s3:::example-bucket-name"
            ]
        },
        {
            "Sid": "AllObjectActions",
            "Effect": "Allow",
            "Action": "s3:*Object",
            "Resource": [
                "arn:aws:s3:::example-bucket-name/*"
            ]
        }
    ]
}
```

Alternatively a IAM User with equal access can be used and referenced via the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` env variables.

## Roadmap
* Automatic support for Route53 updates to reflect changes in the instances attached to a ECS cluster
* [Slack Hooks](https://api.slack.com/messaging/webhooks) support for automatic update notifications
* [Letsencrypt](https://letsencrypt.org/) support to automatically generate new HTTPS certificates
* [Gossip protocol](https://github.com/hashicorp/memberlist) coordination across running containers in a cluster to coordinate Letsencrypt requests
* Move to [openresty](https://openresty.org/en/) to avoid potentially costly config reloads from NGINX

## Caveat emptor
Use at your own risk.

## License
MIT / Apache2

Author [Stefano Fratini](https://www.linkedin.com/in/stefanofratini610/)