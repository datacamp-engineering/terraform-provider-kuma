provider "kuma" {
  host      = "http://localhost:5681"
  api_token = "test123"
}

# resource "kuma_traffic_permission" "yolo_permission" {
#   mesh = "default"
#   name = "yolo_permission"

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

resource "kuma_retry" "yolo_retry" {
  mesh = "default"
  name = "yolo_retry"

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
        numretries = 5
        pertrytimeout = "200ms"
        backoff {
          baseinterval = "20ms"
          maxinterval = "1s"
        }
        retriablestatuscodes = [500,504]
      }
    }
}


# output "yolo_permission_name" {
#   # value = kuma_traffic_permission.yolo_permission.name
#   # value2 = kuma_retry.yolo_retry.name
# }
