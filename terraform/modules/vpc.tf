locals {
  vpc_cidr_block = var.vpc_cidr["sample"]

  public_subnets = {
    a = { cidr = var.subnets_cidr["sample_a_public"], az = var.az["a"] }
  }

  private_subnets = {
    a = { cidr = var.subnets_cidr["sample_a_private"], az = var.az["a"] }
  }
}

# VPC
resource "aws_vpc" "code_stash_vpc" {
  cidr_block           = local.vpc_cidr_block
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = "code-stash-vpc"
  }
}

# インターネットゲートウェイ
resource "aws_internet_gateway" "code_stash_gw" {
  vpc_id = aws_vpc.code_stash_vpc.id

  tags = {
    Name = "code-stash-igw"
  }
}

# EIP for NAT
resource "aws_eip" "code_stash_nat_eip_a" {
  tags = {
    Name = "code-stash-nat-eip-a"
  }
}

# Public Subnet
resource "aws_subnet" "code_stash_public_subnet" {
  for_each                = local.public_subnets
  vpc_id                  = aws_vpc.code_stash_vpc.id
  cidr_block              = each.value.cidr
  availability_zone       = each.value.az
  map_public_ip_on_launch = true

  tags = {
    Name = "code-stash-subnet-public-${each.key}"
  }
}

# Private Subnet
resource "aws_subnet" "code_stash_private_subnet" {
  for_each          = local.private_subnets
  vpc_id            = aws_vpc.code_stash_vpc.id
  cidr_block        = each.value.cidr
  availability_zone = each.value.az

  tags = {
    Name = "code-stash-subnet-private-${each.key}"
  }
}

# NAT Gateway
resource "aws_nat_gateway" "code_stash_ngw_a" {
  allocation_id = aws_eip.code_stash_nat_eip_a.id
  subnet_id     = aws_subnet.code_stash_public_subnet["a"].id

  tags = {
    Name = "code-stash-nat-gw-a"
  }
}

# Public Route Table
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.code_stash_vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.code_stash_gw.id
  }

  tags = {
    Name = "code-stash-public-rt"
  }
}

# Attach to Public Subnet
resource "aws_route_table_association" "public_a" {
  subnet_id      = aws_subnet.code_stash_public_subnet["a"].id
  route_table_id = aws_route_table.public.id
}

# Private Route Table
resource "aws_route_table" "private_a" {
  vpc_id = aws_vpc.code_stash_vpc.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.code_stash_ngw_a.id
  }

  tags = {
    Name = "code-stash-private-rt-a"
  }
}

# プライベートサブネットへの関連付け
resource "aws_route_table_association" "private_a" {
  subnet_id      = aws_subnet.code_stash_private_subnet["a"].id
  route_table_id = aws_route_table.private_a.id
}
