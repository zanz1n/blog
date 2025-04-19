data "aws_iam_policy_document" "allow_public_access" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

data "aws_iam_policy_document" "lambda_logging" {
  statement {
    effect = "Allow"

    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:aws:logs:*:*:*"]
  }
}

data "external" "git_version" {
  working_dir = ".."
  program = [
    "bash",
    "-c",
    "echo '{\"version\":\"$(git rev-parse HEAD | head -c8)\"}'"
  ]
}

data "external" "routes" {
  working_dir = ".."
  program = [
    "make",
    "-s",
    "gen-routes"
  ]
}

data "aws_caller_identity" "current" {}

data "aws_ecr_authorization_token" "token" {}
