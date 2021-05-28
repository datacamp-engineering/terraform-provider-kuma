provider "kuma" {
  host      = "http://localhost:5681"
  api_token = "test123"
}

# resource "kuma_traffic_permission" "test_permission" {
#   mesh = "default"
#   name = "test_permission"

#   sources {
#     match = {
#       "kuma.io/service" = "*"
#     }
#   }

#   destinations {
#     match = {
#       "kuma.io/service" = "*"
#     }
#   }
# }

# resource "kuma_retry" "test_retry" {
#   mesh = "default"
#   name = "test_retry"

#   sources {
#     match = {
#       "kuma.io/service" = "*"
#     }
#   }

#   destinations {
#     match = {
#       "kuma.io/service" = "*"
#     }
#   }
#   conf {
#       http {
#         num_retries = 5
#         per_try_timeout = "200ms"
#         backoff {
#           base_interval = "20ms"
#           max_interval = "1s"
#         }
#         retriable_status_codes = [500,504]

#       }
#       grpc {
#         num_retries = 5
#         per_try_timeout = "300ms"
#         backoff {
#           base_interval = "20ms"
#           max_interval = "1s"
#         }

#       }
#     }
# }

# resource "kuma_circuit_breaker" "test_circuit_breaker" {
#   mesh = "default"
#   name = "test_circuit_breaker"

#   sources {
#     match = {
#       "kuma.io/service" = "*"
#     }
#   }

#   destinations {
#     match = {
#       "kuma.io/service" = "*"
#     }
#   }
#   conf {
#     interval = "5s"
#     base_ejection_time = "30s"
#     max_ejection_percent = 20
#     split_external_and_local_errors = true
#     detectors {    
#       total_errors {
#         consecutive = 20
#       }
#       gateway_errors {
#         consecutive = 10
#       }
#       local_errors {
#         consecutive = 7
#       }
#       standard_deviation {
#         request_volume = 10
#         minimum_hosts = 5
#         factor = 1.9
#       }
#       failure {
#         request_volume = 10
#         minimum_hosts = 5
#         threshold = 85
#       }
#     }
#   }
# }

resource "kuma_proxy_template" "test_proxy_template" {
  mesh = "default"
  name = "test_proxy_template"

  selectors {
    match = {
      "kuma.io/service" = "*"
    }
  }

  conf {
    imports =["default-proxy"]
    modifications{
      cluster{
        operation = "add"
        value = <<EOT
          name: test-cluster
          connectTimeout: 5s
          type: STATIC
          EOT
        match{
          name = "test_cluster"
          origin = "inbound"
        }
       }
      listener{
        operation = "add"
        value = <<EOT
          name: test-listener
          address:
            socketAddress:
              address: 192.168.0.1
              portValue: 8080
          EOT
        match{
          name = "test_cluster"
          origin = "inbound"
        }
      }
      network_filter{
        operation = "addFirst"
        value = <<EOT
          name: envoy.filters.network.tcp_proxy
          typedConfig:
            '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
            idleTimeout: 10s
          EOT
        match{
          name = "envoy.filters.network.tcp_proxy"
          listener_name = "test-listener"
          origin = "inbound"
        }
      }
      http_filters{
        operation = "addFirst"
        value = <<EOT
          name: envoy.filters.http.gzip
          typedConfig:
            '@type': type.googleapis.com/envoy.config.filter.http.gzip.v2.Gzip
            memoryLevel: 9
          EOT
        match{
          name = "envoy.filters.network.tcp_proxy"
          listener_name = "test-listener"
          origin = "inbound"
        }
      }
      # virtual_host{
      #   operation = "add"
      #   value = <<EOT
      #     name: backend
      #     domains:
      #     - "*"
      #     routes:
      #     - match:
      #         prefix: /
      #       route:
      #         cluster: backend
      #     EOT
      #   match{
          
      #   }
      # }
    }
  }
}


# output "test_permission_name" {
#   value = kuma_traffic_permission.test_permission.name
# }

# output "test_retry_name" {
#   value = kuma_retry.test_retry.name
# }

# output "test_circuit_breaker_name" {
#   value = kuma_circuit_breaker.test_circuit_breaker.name
# }
