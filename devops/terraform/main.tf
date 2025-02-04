terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# 🌐 Configure Cloudflare provider
provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

# 🚢 Configure Kubernetes provider
provider "kubernetes" {
  config_path = "~/.kube/config"
}

# 🌐 VPC Configuration
module "vpc" {
  source = "terraform-aws-modules/vpc/aws"

  name = "lambda-kms-vpc"
  cidr = "10.0.0.0/16"

  azs             = ["${var.aws_region}a", "${var.aws_region}b", "${var.aws_region}c"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]

  enable_nat_gateway = true
  single_nat_gateway = true

  tags = {
    Environment = var.environment
    Project     = "lambda-kms"
  }
}

# 🚀 EKS Cluster
module "eks" {
  source = "terraform-aws-modules/eks/aws"

  cluster_name    = "lambda-kms-${var.environment}"
  cluster_version = "1.27"

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  eks_managed_node_groups = {
    default = {
      min_size     = 2
      max_size     = 5
      desired_size = 3

      instance_types = ["t3.medium"]
      capacity_type  = "ON_DEMAND"
    }
  }

  tags = {
    Environment = var.environment
    Project     = "lambda-kms"
  }
}

# 💾 EBS CSI Driver
resource "aws_iam_role" "ebs_csi" {
  name = "lambda-kms-ebs-csi"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRoleWithWebIdentity"
        Effect = "Allow"
        Principal = {
          Federated = module.eks.oidc_provider_arn
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ebs_csi" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy"
  role       = aws_iam_role.ebs_csi.name
}

# 🔐 KMS Key for Encryption
resource "aws_kms_key" "lambda_kms" {
  description             = "KMS key for Xyphos service"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  tags = {
    Environment = var.environment
    Project     = "lambda-kms"
  }
}

# 📦 ECR Repository
resource "aws_ecr_repository" "lambda_kms" {
  name = "lambda-kms"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "KMS"
    kms_key        = aws_kms_key.lambda_kms.arn
  }

  tags = {
    Environment = var.environment
    Project     = "lambda-kms"
  }
}

# 🪣 Create R2 bucket for backups
resource "cloudflare_r2_bucket" "backups" {
  account_id = var.cloudflare_account_id
  name       = "${var.environment}-xyphos-backups"
  location   = "WEUR"  # Western Europe
}

# 🔑 Create R2 API token for backups
resource "cloudflare_api_token" "backup_token" {
  name = "xyphos-backup-${var.environment}"

  policy {
    permission_groups = [
      cloudflare_api_token_permission_groups.r2_write,
    ]
    resources = {
      "com.cloudflare.api.account.${var.cloudflare_account_id}" = "*"
    }
  }
}

# 🚀 Deploy Xyphos to Kubernetes
resource "kubernetes_namespace" "xyphos" {
  metadata {
    name = "xyphos-${var.environment}"
  }
}

# 🔐 Create Kubernetes secrets
resource "kubernetes_secret" "xyphos_secrets" {
  metadata {
    name      = "xyphos-secrets"
    namespace = kubernetes_namespace.xyphos.metadata[0].name
  }

  data = {
    "jwt-secret"          = base64encode(var.jwt_secret)
    "r2-account-id"       = base64encode(var.cloudflare_account_id)
    "r2-access-key-id"    = base64encode(cloudflare_api_token.backup_token.id)
    "r2-access-key-secret" = base64encode(cloudflare_api_token.backup_token.value)
  }
}

# ⚙️ Create Kubernetes ConfigMap
resource "kubernetes_config_map" "xyphos_config" {
  metadata {
    name      = "xyphos-config"
    namespace = kubernetes_namespace.xyphos.metadata[0].name
  }

  data = {
    "r2-bucket-name"  = cloudflare_r2_bucket.backups.name
    "backup-interval" = var.backup_interval
  }
}

# 💾 Create PersistentVolume for BadgerDB
resource "kubernetes_persistent_volume_claim" "xyphos_data" {
  metadata {
    name      = "xyphos-data"
    namespace = kubernetes_namespace.xyphos.metadata[0].name
  }

  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        storage = "10Gi"
      }
    }
  }
}

# Variables
variable "environment" {
  description = "Environment (e.g., dev, staging, prod)"
  type        = string
}

variable "cloudflare_api_token" {
  description = "Cloudflare API token"
  type        = string
  sensitive   = true
}

variable "cloudflare_account_id" {
  description = "Cloudflare account ID"
  type        = string
}

variable "jwt_secret" {
  description = "JWT secret for authentication"
  type        = string
  sensitive   = true
}

variable "backup_interval" {
  description = "Backup interval duration (e.g., 24h)"
  type        = string
  default     = "24h"
}

# Outputs
output "r2_bucket_name" {
  description = "Name of the R2 bucket for backups"
  value       = cloudflare_r2_bucket.backups.name
}

output "backup_token_id" {
  description = "ID of the R2 API token for backups"
  value       = cloudflare_api_token.backup_token.id
  sensitive   = true
} 