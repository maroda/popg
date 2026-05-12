resource "aws_ecs_cluster" "appcluster" {
  name = var.app
}

resource "aws_cloudwatch_log_group" "cwlogs" {
  name              = var.app
  retention_in_days = 0

  tags = {
    Name = var.app
  }
}

resource "aws_ecs_task_definition" "apptask" {
  family                   = var.app
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecstaskexec.arn

  depends_on = [aws_cloudwatch_log_group.cwlogs]

  container_definitions = <<TASK
[
  {
  "image": "${var.repository}:${var.release}",
  "name": "${var.app}",
  "environment": [
  {
    "name": "GM_PASSWORD",
    "value": "${var.gm_password}"
  }
],
  "essential": true,
        "logConfiguration": {
            "logDriver": "awslogs",
            "options": {
                "awslogs-group": "${var.app}",
                "awslogs-region": "${var.aws_region}",
                "awslogs-stream-prefix": "ecs"
              }
          },
      "portMappings": [
        {
          "protocol": "tcp",
          "containerPort": ${var.port}
        }
      ]
    }
]
TASK
}

resource "aws_ecs_service" "appserv" {
  name             = var.app
  cluster          = aws_ecs_cluster.appcluster.arn
  task_definition  = aws_ecs_task_definition.apptask.arn
  launch_type      = "FARGATE"
  platform_version = "LATEST"

  desired_count = var.tcount

  lifecycle {
    create_before_destroy = true
    ignore_changes        = [desired_count]
  }

  force_new_deployment = true

  network_configuration {
    security_groups  = [aws_security_group.taskaccess.id]
    subnets          = aws_subnet.private_az.*.id
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.lbtarget.arn
    container_name   = var.app
    container_port   = var.port
  }

}

