variable "environment" {
  type        = string
  description = "The environment of the project"
  default     = "prod"

  validation {
    condition     = var.environment == "prod" || var.environment == "dev"
    error_message = "The environment must be dev or prod"
  }
}

variable "aws_region" {
  type        = string
  description = "The in which the project will be deployed"
  default     = "sa-east-1"
}

variable "aws_access_key_id" {
  type        = string
  description = "AWS access key id to deploy the project"
  sensitive   = true
}

variable "aws_secret_access_key" {
  type        = string
  description = "AWS secret key to deploy the project"
  sensitive   = true
}
