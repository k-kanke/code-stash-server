locals {
  container_image_tag = coalesce(var.container_image_tag, "latest")
  container_name      = "code-stash-app"
  container_port      = 8085
}

# ECS Cluster
resource "aws_ecs_cluster" "code_stash_cluster" {
  name = "code-stash-ecs-cluster"

  tags = {
    Name = "code-stash-ecs-cluster"
  }
}

# Task Definition
resource "aws_ecs_task_definition" "code_stash_task" {
  family                   = "code-stash-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = tostring(var.ecs_cpu)
  memory                   = tostring(var.ecs_memory)
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn

  container_definitions = jsonencode([
    {
      name      = local.container_name
      image     = "${aws_ecr_repository.code_stash_ecr.repository_url}:${local.container_image_tag}"
      essential = true
      portMappings = [
        {
          containerPort = local.container_port
          hostPort      = local.container_port
          protocol      = "tcp"
        }
      ]
    }
  ])

  tags = {
    Name = "code-stash-task"
  }
}

# ECS Service
resource "aws_ecs_service" "code_stash_service" {
  name             = "code-stash-service"
  cluster          = aws_ecs_cluster.code_stash_cluster.id
  task_definition  = aws_ecs_task_definition.code_stash_task.arn
  desired_count    = var.ecs_desired_count
  launch_type      = "FARGATE"
  platform_version = "LATEST"

  network_configuration {
    subnets         = [aws_subnet.kanke_sample_private_subnet["a"].id]
    security_groups = [aws_security_group.kanke_sample_ecs_sg.id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.code_stash_tg.arn
    container_name   = local.container_name
    container_port   = local.container_port
  }

  depends_on = [aws_lb_listener.kanke_sample_http_listener]

  tags = {
    Name = "code-stash-service"
  }
}

# IAMロール
data "aws_iam_policy_document" "ecs_task_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

# ECRリポジトリからimageをpullしてくる際に必要
resource "aws_iam_role" "ecs_task_execution_role" {
  name               = "code-stash-ecs-task-execution-role"
  assume_role_policy = data.aws_iam_policy_document.ecs_task_assume_role.json

  tags = {
    Name = "code-stash-ecs-task-execution-role"
  }
}

# IAM Policy をロールにアタッチ
resource "aws_iam_role_policy_attachment" "ecs_task_execution_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}
