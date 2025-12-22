# ALB
resource "aws_alb" "code_stash_alb" {
  name               = "code-stash-alb"
  internal           = false
  load_balancer_type = "application"

  subnets = [
    aws_subnet.code_stash_public_subnet["a"].id
  ]

  security_groups = [aws_security_group.code_stash_alb_sg.id]

  tags = {
    Name = "code-stash-alb"
  }
}

# ターゲットグループ
resource "aws_lb_target_group" "code_stash_tg" {
  name        = "code-stash-tg"
  port        = 8085
  protocol    = "HTTP"
  vpc_id      = aws_vpc.code_stash_vpc.id
  target_type = "ip"

  health_check {
    path                = "/health"
    protocol            = "HTTP"
    matcher             = "200-399"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 2
  }

  tags = {
    Name = "code-stash-tg"
  }
}


# リスナー
resource "aws_lb_listener" "kanke_sample_http_listener" {
  load_balancer_arn = aws_alb.code_stash_alb.arn
  port              = 80
  protocol          = "HTTP"
  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.code_stash_tg.arn
  }

  tags = {
    Name = "code-stash-http-listener"
  }
}

# アクセス用DNS
output "alb_dns_name" {
  description = "The DNS name of the Application Load Balancer"
  value       = aws_alb.code_stash_alb.dns_name
}
