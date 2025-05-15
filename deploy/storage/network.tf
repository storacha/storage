resource "aws_vpc" "storage_vpc" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name  = "${var.app}-storage-vpc"
    Owner = var.owner
    Team  = var.team
    Org   = var.org
  }
}

# Internet Gateway
resource "aws_internet_gateway" "storage_igw" {
  vpc_id = aws_vpc.storage_vpc.id
  tags = {
    Name  = "${var.app}-storage-igw"
    Owner = var.owner
    Team  = var.team
    Org   = var.org
  }
}

# Public Subnet
resource "aws_subnet" "storage_public_subnet" {
  vpc_id                  = aws_vpc.storage_vpc.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = "${var.region}a"
  map_public_ip_on_launch = true
  tags = {
    Name  = "${var.app}-storage-public-subnet"
    Owner = var.owner
    Team  = var.team
    Org   = var.org
  }
}

# Route Table
resource "aws_route_table" "storage_public_rt" {
  vpc_id = aws_vpc.storage_vpc.id
  tags = {
    Name  = "${var.app}-storage-public-rt"
    Owner = var.owner
    Team  = var.team
    Org   = var.org
  }
}

# Default route to the Internet
resource "aws_route" "storage_public_route_igw" {
  route_table_id         = aws_route_table.storage_public_rt.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.storage_igw.id
}

# Associate public subnet with the route table
resource "aws_route_table_association" "storage_public_rta" {
  subnet_id      = aws_subnet.storage_public_subnet.id
  route_table_id = aws_route_table.storage_public_rt.id
}