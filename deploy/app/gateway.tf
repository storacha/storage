locals {
  web_functions = { for k, v in local.functions : k => v if (v.route != "") }
}

resource "aws_apigatewayv2_api" "api" {
  name        = "${terraform.workspace}-${var.app}-api"
  description = "${terraform.workspace} ${var.app} API Gateway"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_route" "routes" {
  for_each = local.web_functions
  api_id =  aws_apigatewayv2_api.api.id
  route_key = each.value.route
  authorization_type = "NONE"
  target = "integrations/${aws_apigatewayv2_integration.integrations[each.key].id}"
}

resource "aws_apigatewayv2_integration" "integrations" {
  for_each = local.web_functions
  api_id             = aws_apigatewayv2_api.api.id
  integration_uri =  aws_lambda_function.lambda[each.key].invoke_arn
  payload_format_version = "2.0"
  integration_type    = "AWS_PROXY"
  connection_type = "INTERNET"
}

resource "aws_apigatewayv2_deployment" "deployment" {
  depends_on = [aws_apigatewayv2_integration.integrations]
  triggers = {
    redeployment = sha1(join(",", concat(
      [ for k, v in local.web_functions : jsonencode(aws_apigatewayv2_route.routes[k]) ],
      [ for k, v in local.web_functions : jsonencode(aws_apigatewayv2_integration.integrations[k]) ])
    ))
  }

  api_id = aws_apigatewayv2_api.api.id
  description = "${terraform.workspace} ${var.app} API Deployment"
  lifecycle {
    create_before_destroy = true
  }
}

data "terraform_remote_state" "shared" {
  backend = "s3"
  config = {
    bucket = "storacha-terraform-state"
    key    = "storacha/storage/shared.tfstate"
    region = "us-west-2"
  }
}

resource "aws_acm_certificate" "cert" {
  domain_name       = terraform.workspace == "prod" ? "${var.app}.storacha.network" : "${terraform.workspace}.${var.app}.storacha.network"
  validation_method = "DNS"
  
  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_route53_record" "cert_validation" {
  allow_overwrite = true
  name    = tolist(aws_acm_certificate.cert.domain_validation_options)[0].resource_record_name
  type    = tolist(aws_acm_certificate.cert.domain_validation_options)[0].resource_record_type
  zone_id = data.terraform_remote_state.shared.outputs.primary_zone.zone_id
  records = [tolist(aws_acm_certificate.cert.domain_validation_options)[0].resource_record_value]
  ttl     = 60
}

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn = aws_acm_certificate.cert.arn
  validation_record_fqdns = [ aws_route53_record.cert_validation.fqdn ]
}
resource "aws_apigatewayv2_domain_name" "custom_domain" {
  domain_name = "${terraform.workspace}.${var.app}.storacha.network"

  domain_name_configuration {
    certificate_arn = aws_acm_certificate_validation.cert.certificate_arn
    endpoint_type = "REGIONAL"
    security_policy = "TLS_1_2"
  }
}
resource "aws_apigatewayv2_stage" "stage" {
  api_id = aws_apigatewayv2_api.api.id
  name   = "$default"
  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_apigatewayv2_api_mapping" "api_mapping" {
  api_id      = aws_apigatewayv2_api.api.id
  stage  = aws_apigatewayv2_stage.stage.id
  domain_name = aws_apigatewayv2_domain_name.custom_domain.id
}

resource "aws_route53_record" "api_gateway" {
  zone_id = data.terraform_remote_state.shared.outputs.primary_zone.zone_id
  name    = "${terraform.workspace}.${var.app}.storacha.network"
  type    = "A"

  alias {
    name                   = aws_apigatewayv2_domain_name.custom_domain.domain_name_configuration[0].target_domain_name
    zone_id                = aws_apigatewayv2_domain_name.custom_domain.domain_name_configuration[0].hosted_zone_id
    evaluate_target_health = false
  }
}