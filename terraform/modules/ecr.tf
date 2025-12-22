resource "aws_ecr_repository" "kanke_sample_ecr" {
  name                 = "code-stash-server"
  image_tag_mutability = "MUTABLE"

  tags = {
    Name = "code-stash-ecr"
  }
}

# docker push 用に出力しておく
output "ecr_repository_url" {
  value = aws_ecr_repository.kanke_sample_ecr.repository_url
}
