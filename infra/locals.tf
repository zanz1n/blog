locals {
  aws_ecr_repository = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com"

  routes = jsondecode(data.external.routes.result.data)

  database_url = "postgres://${aws_rds_cluster.database.master_username}:${aws_rds_cluster.database.master_password}@${aws_rds_cluster.database.endpoint}:${aws_rds_cluster.database.port}/${aws_rds_cluster.database.database_name}"

  version = "${var.environment}-${data.external.git_version.result.version}"

  aws_ecr_username = data.aws_ecr_authorization_token.token.user_name
  aws_ecr_password = data.aws_ecr_authorization_token.token.password
}
