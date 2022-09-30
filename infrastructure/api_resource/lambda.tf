resource "aws_ecr_repository" "repository" {
  name = var.name
}

data "aws_ecr_image" "image" {
  repository_name = var.name
  image_tag       = "latest"
}

data "aws_iam_policy_document" "lambda" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["lambda.amazonaws.com"]
      type        = "Service"
    }
    effect = "Allow"
  }
}

data "aws_iam_policy_document" "iam_role" {
  statement {
    effect  = "Allow"
    actions = var.iam_actions
    resources = [
      "*"
    ]
  }
}

resource "aws_iam_policy" "iam_role" {
  name   = "allows_for_${var.name}"
  policy = data.aws_iam_policy_document.iam_role.json
}

resource "aws_iam_role" "iam_role" {
  name               = var.name
  assume_role_policy = data.aws_iam_policy_document.lambda.json
}

resource "aws_iam_role_policy_attachment" "iam_role" {
  role       = aws_iam_role.iam_role.name
  policy_arn = aws_iam_policy.iam_role.arn
}

resource "aws_iam_role_policy_attachment" "lambda_policy" {
  role       = aws_iam_role.iam_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_lambda_function" "function" {
  function_name = var.name
  role          = aws_iam_role.iam_role.arn
  image_uri     = "${aws_ecr_repository.repository.repository_url}@${data.aws_ecr_image.image.id}"
  timeout       = "30"
  package_type  = "Image"
  environment {
    variables = {
      CLUSTER     = var.cluster_name
      DNS_ZONE_ID = var.dns_zone_id
      DNS_ZONE    = var.dns_zone
      SUBNETS     = join(",", var.subnets)
      CORS_DOMAIN = var.cors_domain
    }
  }

  depends_on = [
    aws_ecr_repository.repository
  ]
}
