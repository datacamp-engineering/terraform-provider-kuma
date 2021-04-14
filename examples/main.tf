provider "kuma" {
  host      = "http://localhost:5681"
  api_token = "test123"
}

resource "kuma_traffic_permission" "test_permission" {
  mesh = "default"
  name = "test_permission"

  sources {
    match = {
      "kuma.io/service" = "*"
    }
  }

  destinations {
    match = {
      "kuma.io/service" = "*"
    }
  }
}

resource "kuma_retry" "test_retry" {
  mesh = "default"
  name = "test_retry"

  sources {
    match = {
      "kuma.io/service" = "*"
    }
  }

  destinations {
    match = {
      "kuma.io/service" = "*"
    }
  }
  conf {
      http {
        num_retries = 5
        per_try_timeout = "200ms"
        backoff {
          base_interval = "20ms"
          max_interval = "1s"
        }
        retriable_status_codes = [500,504]

      }
      grpc {
        num_retries = 5
        per_try_timeout = "300ms"
        backoff {
          base_interval = "20ms"
          max_interval = "1s"
        }

      }
    }
}


output "test_permission_name" {
  value = kuma_traffic_permission.test_permission.name
}

output "test_retry_name" {
  value = kuma_retry.test_retry.name
}
