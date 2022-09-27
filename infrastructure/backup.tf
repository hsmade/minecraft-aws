resource "aws_s3_bucket" "backup_bucket" {
  bucket_prefix = "minecraft-servers-backup"
}

resource "aws_s3_bucket_server_side_encryption_configuration" "backup_bucket" {
  bucket = aws_s3_bucket.backup_bucket.bucket
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
    bucket_key_enabled = false
  }
}

resource "aws_s3_bucket_acl" "backup_bucket" {
  bucket = aws_s3_bucket.backup_bucket.bucket
  acl    = "private"
}

# see main.tf
data "aws_iam_policy_document" "ecs_backup_restore" {
  statement {
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:GetObjectVersion",
      "s3:PutObject",
      "s3:DeleteObject",
      "s3:DeleteObjectVersion",
      "s3:PutObjectACL",
    ]
    resources = [
      "arn:aws:s3:::${aws_s3_bucket.backup_bucket.bucket}:/*",
      "arn:aws:s3:::${aws_s3_bucket.backup_bucket.bucket}:",
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "s3:ListBucket",
    ]
    resources = [
      "arn:aws:s3:::${aws_s3_bucket.backup_bucket.bucket}:",
    ]
  }
}

resource "aws_iam_policy" "ecs_backup_restore" {
  policy = data.aws_iam_policy_document.ecs_backup_restore.json
  name   = "ecs_backup_restore"
}
