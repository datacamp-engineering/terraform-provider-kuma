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

resource "kuma_circuit_breaker" "test_circuit_breaker" {
  mesh = "default"
  name = "test_circuit_breaker"

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
    interval = "5s"
    base_ejection_time = "30s"
    max_ejection_percent = 20
    split_external_and_local_errors = true
    detectors {    
      total_errors {
        consecutive = 20
      }
      gateway_errors {
        consecutive = 10
      }
      local_errors {
        consecutive = 7
      }
      standard_deviation {
        request_volume = 10
        minimum_hosts = 5
        factor = 1.9
      }
      failure {
        request_volume = 10
        minimum_hosts = 5
        threshold = 85
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

output "test_circuit_breaker_name" {
  value = kuma_circuit_breaker.test_circuit_breaker.name
}
