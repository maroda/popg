output "bucket_name" {
  description = "Ephemeral application data"
  value       = aws_s3_bucket.appdata.id
}

output "certificate" {
  description = "Domain holding the TLS cert"
  value       = data.dnsimple_certificate.apexcert.domain
}

output "cname-www" {
  description = "Endpoint CNAME for popg service"
  value       = dnsimple_zone_record.www.value
}

output "www" {
  description = "FQDN for the popg API"
  value       = dnsimple_zone_record.www.qualified_name
}

output "lb_dns_popg" {
  description = "popg load balancer endpoint"
  value       = aws_lb.applb.dns_name
}
