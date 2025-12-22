resource "aws_wafv2_web_acl" "code_stash_waf" {
  name  = "code-stash-waf"
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "limit-10-requests-per-second"
    priority = 1
    action {
      block {}
    }

    statement {
      rate_based_statement {
        limit              = 3000
        aggregate_key_type = "IP"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "limit-10-requests-per-second"
      sampled_requests_enabled   = true
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "code-stash-alb-waf-acl"
    sampled_requests_enabled   = true
  }
}

# ALBにアタッチ
resource "aws_wafv2_web_acl_association" "code_stash_alb_waf_attach" {
  resource_arn = aws_lb_listener.kanke_sample_http_listener.load_balancer_arn
  web_acl_arn  = aws_wafv2_web_acl.code_stash_waf.arn
}
