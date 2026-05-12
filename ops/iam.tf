data "aws_iam_policy_document" "ecspolicy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      identifiers = ["ecs-tasks.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_iam_role" "ecstaskexec" {
  name               = "${var.app}-ecs-task-exec"
  assume_role_policy = data.aws_iam_policy_document.ecspolicy.json

  tags = {
    Name = "${var.app}-ecs-task-exec"
  }
}

resource "aws_iam_role" "ecstask" {
  name               = "${var.app}-ecs-task"
  assume_role_policy = data.aws_iam_policy_document.ecspolicy.json

  tags = {
    Name = "${var.app}-ecs-task"
  }
}

resource "aws_iam_policy_attachment" "ecstaskexec" {
  name       = "${var.app}-ecs-task-exec"
  roles      = [aws_iam_role.ecstaskexec.name]
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_policy_attachment" "ecstask" {
  name       = "${var.app}-ecs-task"
  roles      = [aws_iam_role.ecstask.name]
  policy_arn = "arn:aws:iam::aws:policy/AmazonS3FullAccess"
}

resource "aws_iam_role_policy" "ecstaskexec_secrets" {
  name = "secrets-access"
  role = aws_iam_role.ecstaskexec.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = "arn:aws:secretsmanager:us-west-2:821445872109:secret:grafana/otel/header-z2qbpC"
      }
    ]
  })
}
