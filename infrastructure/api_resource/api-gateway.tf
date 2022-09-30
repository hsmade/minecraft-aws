resource "aws_lambda_permission" "allow_api_gateway" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.function.function_name
  principal     = "apigateway.amazonaws.com"

  # More: http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-control-access-using-iam-policies-to-invoke-api.html
  source_arn = "arn:aws:execute-api:${var.region}:${var.account_id}:${var.rest_api_id}/*/${aws_api_gateway_method.method.http_method}/${var.path}"
}

resource "aws_api_gateway_method" "method" {
  authorization = "NONE"
  http_method   = var.method
  resource_id   = var.resource_id
  rest_api_id   = var.rest_api_id
}


resource "aws_api_gateway_integration" "integration" {
  http_method             = aws_api_gateway_method.method.http_method
  resource_id             = var.resource_id
  rest_api_id             = var.rest_api_id
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.function.invoke_arn
}

