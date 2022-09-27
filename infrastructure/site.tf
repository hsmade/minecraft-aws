resource "aws_s3_bucket" "site_bucket" {
  bucket_prefix = "minecraft-servers-site"
}

resource "aws_s3_bucket_acl" "site_bucket" {
  bucket = aws_s3_bucket.site_bucket.bucket
  acl    = "public-read"
}

resource "aws_s3_bucket_cors_configuration" "site_bucket" {
  bucket = aws_s3_bucket.site_bucket.bucket
  cors_rule {
    allowed_methods = ["GET"]
    allowed_origins = [
      "https://${var.domain_name}"
    ]
  }
}

resource "aws_s3_bucket_website_configuration" "site" {
  bucket = aws_s3_bucket.site_bucket.bucket

  index_document {
    suffix = "index.html"
  }
}

module "template_files" {
  source = "hashicorp/dir/template"

  base_dir = "${path.module}/site"
  template_vars = {
    # Pass in any values that you wish to use in your templates.
    server_list   = module.servers_list.url
    server_start  = module.server_start.url
    server_status = module.server_status.url
    server_stop   = module.server_stop
  }
}

resource "aws_s3_object" "site" {
  for_each = module.template_files.files

  bucket       = aws_s3_bucket.site_bucket.bucket
  key          = each.key
  content_type = each.value.content_type
  source       = each.value.source_path
  content      = each.value.content
  etag         = each.value.digests.md5
  acl          = "public_read"
}
