variable "az" {
  type        = map(string)
  description = "availability_zone"

  default = {
    a = "ap-northeast-1a"
  }
}

variable "vpc_cidr" {
  type        = map(string)
  description = "Map of CIDR blocks for VPC"

  default = {
    "sample" = "10.0.0.0/16"
  }
}

variable "subnets_cidr" {
  type        = map(string)
  description = "Map of subnet CIDR blocks in vpc"

  default = {
    sample_a_public  = "10.0.0.0/20"
    sample_a_private = "10.0.64.0/20"
  }
}

variable "container_image_tag" {
  type        = string
  description = "Docker image tag to deploy from ECR"
  default     = ""
}

variable "ecs_desired_count" {
  type        = number
  description = "Number of ECS tasks to run"
  default     = 1
}

variable "ecs_cpu" {
  type        = number
  description = "CPU units for the Fargate task"
  default     = 256
}

variable "ecs_memory" {
  type        = number
  description = "Memory (MB) for the Fargate task"
  default     = 512
}
