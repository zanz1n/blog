resource "random_pet" "lambda" {
  prefix = "blog-${var.environment}"
  length = 2
}

resource "aws_iam_role" "lambda_exec" {
  name               = random_pet.lambda.id
  assume_role_policy = data.aws_iam_policy_document.allow_public_access.json
}

resource "aws_iam_role_policy_attachment" "name" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_lambda_function" "lambda" {
  function_name = random_pet.lambda.id

  role = aws_iam_role.lambda_exec.arn

  image_uri    = "${aws_ecr_repository.images.repository_url}:latest"
  package_type = "Image"

  logging_config {
    log_format = "JSON"
  }

  memory_size = 512
  timeout     = 10

  environment {
    variables = {
      ENVIRONMENT     = var.environment,
      NO_COLOR        = "1"
      DATABASE_URL    = local.database_url
      LOG_LEVEL       = "INFO",
      BCRYPT_COST     = 12,
      REQUEST_TIMEOUT = 8,
      STATIC_ASSETS   = "/static"
    }
  }

  image_config {
    command = ["-json-logs"]
  }

  depends_on = [null_resource.blog_docker_image]
}

resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${aws_lambda_function.lambda.function_name}"
  retention_in_days = 7
}

resource "random_pet" "lambda_logging" {
  prefix = "blog-${var.environment}"
  length = 2
}

resource "aws_iam_policy" "lambda_logging" {
  name        = random_pet.lambda_logging.id
  path        = "/"
  description = "IAM policy for logging from a lambda"
  policy      = data.aws_iam_policy_document.lambda_logging.json
}

resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = aws_iam_policy.lambda_logging.arn
}
