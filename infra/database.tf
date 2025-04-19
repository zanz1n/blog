resource "random_pet" "database_identifier" {
  length = 2
  prefix = "blog-${var.environment}"
}

resource "random_pet" "database_master_username" {
  length    = 2
  separator = "_"
}

resource "random_password" "database_master_password" {
  length  = 64
  special = false
}

resource "aws_rds_cluster" "database" {
  cluster_identifier = random_pet.database_identifier.id
  database_name      = "blog"

  engine         = "aurora-postgresql"
  engine_mode    = "provisioned"
  engine_version = "16"

  master_username = random_pet.database_master_username.id
  master_password = random_password.database_master_password.result
  # enable_http_endpoint = false
  # backup_retention_period = 1

  skip_final_snapshot = true

  # allow_major_version_upgrade = true
  # storage_type = ""

  serverlessv2_scaling_configuration {
    max_capacity             = 2
    min_capacity             = 0
    seconds_until_auto_pause = 300
  }
}

resource "aws_rds_cluster_instance" "database" {
  cluster_identifier = aws_rds_cluster.database.id
  instance_class     = "db.serverless"
  engine             = aws_rds_cluster.database.engine
  engine_version     = aws_rds_cluster.database.engine_version
}
