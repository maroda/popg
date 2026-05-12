resource "aws_s3_bucket" "appdata" {
  bucket = "${var.app}-${var.dnsapex}"

  tags = {
    Name = var.app
  }
}

