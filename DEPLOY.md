# Deploying to Amazon ECS Express Mode

The Docker image contains the Go API and the built React interface. Amazon ECS Express Mode runs it as a Fargate service and provisions the supporting web infrastructure. PostgreSQL remains on Supabase.

## Prerequisites

- An AWS account and AWS CLI configured with access to Amazon ECR and Amazon ECS.
- Docker with Buildx.
- An ECR repository in your AWS region.
- A VPC with public subnets, or your own subnet configuration for the Express Mode service.

Set the values for your AWS account and region:

~~~sh
export AWS_REGION=us-east-1
export ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
export ECR_REPO=cellphone-repair
export IMAGE_URI="$ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPO:latest"
~~~

Create the repository if needed:

~~~sh
aws ecr create-repository --repository-name "$ECR_REPO" --region "$AWS_REGION"
~~~

## Build and push the image

Build for Linux AMD64 and push the image to ECR:

~~~sh
aws ecr get-login-password --region "$AWS_REGION" \
  | docker login --username AWS --password-stdin "$ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com"

docker buildx build --platform linux/amd64 --push -t "$IMAGE_URI" .
~~~

## Create the Express Mode service

In the [Amazon ECS console](https://console.aws.amazon.com/ecs/v2/), choose **Express mode** and create a service using the image URI above.

Configure the service with:

- **Container port:** 8080
- **Health check path:** /health
- **Environment variables:** DATABASE_URL, JWT_SECRET, and the Supabase settings listed in .env.example
- **Task execution role and infrastructure role:** create or select the roles requested by the console

Set secrets in the service configuration or reference AWS Secrets Manager values. Do not add .env to the image or store production credentials in the repository. Leave the TWILIO_* variables unset to use the development SMS logger; configure them to send real texts.

When the service is active, ECS provides an HTTPS application URL. Express Mode provisions a load balancer and TLS for the service. See the [ECS Express Mode console guide](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/express-service-first-run.html) for the full creation flow.

## Database setup

Run migrations against the production database before sending traffic:

~~~sh
go run ./cmd/migrate
go run ./cmd/seedadmin -email admin@example.com -password 'use-a-strong-password'
~~~

Run these commands from an environment with the production database settings available. Create the admin account only once.

## Deploying an update

Build and push a new image, then update the ECS Express Mode service to deploy the new image. See the [ECS Express Mode update guide](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/express-service-update-full.html) for the available update workflow.
