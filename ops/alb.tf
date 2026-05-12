resource "aws_lb" "applb" {
  name               = var.app
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.pubaccess.id]
  subnets            = aws_subnet.private_az.*.id
  idle_timeout       = 3600

  tags = {
    Name = var.app
  }
}

resource "aws_lb_listener" "public" {
  load_balancer_arn = aws_lb.applb.arn
  port              = var.port
  protocol          = "HTTP"

  default_action {
    type = "redirect"
    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_lb_target_group" "lbtarget" {
  name        = var.app
  port        = var.port
  protocol    = "HTTP"
  target_type = "ip"
  vpc_id      = aws_vpc.vpc.id

  health_check {
    healthy_threshold   = "3"
    interval            = "30"
    protocol            = "HTTP"
    matcher             = "200"
    timeout             = "3"
    unhealthy_threshold = "2"
    path                = "/"
    port                = var.port
  }
}

resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.applb.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_acm_certificate.apexcert.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.lbtarget.arn
  }
}