provider "dnsimple" {
  token    = var.dnstoken
  account  = var.dnsaccount
  sandbox  = false
  prefetch = false
}

data "dnsimple_zone" "apex" {
  name = var.dnsapex
}

data "dnsimple_certificate" "apexcert" {
  domain         = var.dnsapex
  certificate_id = var.dnscertid
}

resource "dnsimple_zone_record" "www" {
  zone_name = var.dnsapex
  name      = "www"
  value     = aws_lb.applb.dns_name
  type      = "CNAME"
  ttl       = 300
}

resource "aws_acm_certificate" "apexcert" {
  private_key       = data.dnsimple_certificate.apexcert.private_key
  certificate_body  = data.dnsimple_certificate.apexcert.server_certificate
  certificate_chain = join("\n", data.dnsimple_certificate.apexcert.certificate_chain)
}