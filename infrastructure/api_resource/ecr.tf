resource "aws_ecr_repository" "repository" {
  name = var.name
}

data "aws_ecr_image" "image" {
  repository_name = var.name
  image_tag       = "latest"
}

resource "aws_ecr_lifecycle_policy" "untagged-retention" {
  repository = aws_ecr_repository.repository.name
  policy = <<EOF
    {
        "rules": [
            {
                "rulePriority": 1,
                "description": "Expire untagged images older than 5 days",
                "selection": {
                    "tagStatus": "untagged",
                    "countType": "sinceImagePushed",
                    "countUnit": "days",
                    "countNumber": 5
                },
                "action": {
                    "type": "expire"
                }
            }
        ]
    }
EOF
}