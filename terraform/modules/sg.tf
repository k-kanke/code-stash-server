# Security Group for ALB
resource "aws_security_group" "code_stash_alb_sg" {
  name   = "code-stash-alb-sg"
  vpc_id = aws_vpc.code_stash_vpc.id

  # インバウンドルール
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # アウトバウンドルール
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "code-stash-alb-sg"
  }
}

# ECSサービス用セキュリティグループ
resource "aws_security_group" "code_stash_ecs_sg" {
  name   = "code-stash-ecs-service-sg"
  vpc_id = aws_vpc.code_stash_vpc.id

  # インバウンドルール
  ingress {
    from_port       = 8085
    to_port         = 8085
    protocol        = "tcp"
    security_groups = [aws_security_group.code_stash_alb_sg.id]
  }

  # アウトバウンドルール
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "code-stash-ecs-service-sg"
  }
}
