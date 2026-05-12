data "aws_availability_zones" "aws-az" {
  state = "available"
}

resource "aws_vpc" "vpc" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true

  tags = {
    Name = var.app
  }
}

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = var.app
  }
}

resource "aws_subnet" "private_az" {
  count                   = length(data.aws_availability_zones.aws-az.names)
  vpc_id                  = aws_vpc.vpc.id
  cidr_block              = cidrsubnet(aws_vpc.vpc.cidr_block, 8, count.index + 1)
  availability_zone       = data.aws_availability_zones.aws-az.names[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name = var.app
  }
}

resource "aws_route_table" "private_rt" {
  vpc_id = aws_vpc.vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }

  tags = {
    Name = var.app
  }
}

resource "aws_route_table_association" "private_asc" {
  count          = length(data.aws_availability_zones.aws-az.names)
  subnet_id      = aws_subnet.private_az[count.index].id
  route_table_id = aws_route_table.private_rt.id
}