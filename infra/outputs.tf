output "routes" {
  value = local.routes
}

output "database_url" {
  value     = local.database_url
  sensitive = true
}

output "version" {
  value = local.version
}

output "api_gateway_url" {
  value = aws_apigatewayv2_api.api.api_endpoint
}
