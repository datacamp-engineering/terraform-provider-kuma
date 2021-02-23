provider "kuma" {
  host      = "http://localhost:5681"
  api_token = "test123"
}

resource "kuma_traffic_permission" "yolo_permission" {
  mesh = "default"
  name = "yolo_permission"

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

output "yolo_permission_name" {
  value = kuma_traffic_permission.yolo_permission.name
}