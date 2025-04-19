resource "random_pet" "ecr_repository" {
  prefix = "blog-${var.environment}"
  length = 2
}

resource "aws_ecr_repository" "images" {
  name                 = random_pet.ecr_repository.id
  image_tag_mutability = "MUTABLE"

  force_delete = true

  # image_scanning_configuration {
  #   scan_on_push = true
  # }
}

resource "null_resource" "blog_docker_image" {
  provisioner "local-exec" {
    working_dir = ".."

    command = <<EOF
    docker login \
    --username ${local.aws_ecr_username} \
    --password ${local.aws_ecr_password} \
    ${local.aws_ecr_repository}

    docker buildx build \
    --build-arg VERSION=${local.version} \
    -t "${aws_ecr_repository.images.repository_url}:latest" \
    -f lambda.dockerfile .

    docker push "${aws_ecr_repository.images.repository_url}:latest"
    EOF
  }

  triggers = {
    run_at = timestamp()
  }

  depends_on = [aws_ecr_repository.images]
}
