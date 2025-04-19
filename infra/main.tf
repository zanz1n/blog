terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.91"
    }

    random = {
      source  = "hashicorp/random"
      version = "~> 3.7"
    }

    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0.1"
    }
  }

  required_version = ">= 1.2.0"
}

provider "aws" {
  region = var.aws_region

  access_key = var.aws_access_key_id
  secret_key = var.aws_secret_access_key
}

provider "docker" {
  registry_auth {
    address  = local.aws_ecr_repository
    username = data.aws_ecr_authorization_token.token.user_name
    password = data.aws_ecr_authorization_token.token.password
  }
}
