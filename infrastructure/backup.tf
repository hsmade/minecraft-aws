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

