provider "kuma" {
  host      = "http://localhost:5681"
  api_token = "test123"
}

data "kuma_mesh" "default" {
  name = "default"
}

resource "kuma_dataplane" "yolo_data" {
  mesh = data.kuma_mesh.default.name
  name = "yolo_data"

  networking {
    address = "10.0.0.1"

    # gateway {
    #   tags = {
    #     "kuma.io/service" = "test"
    #   }
    # }

    inbound {
      port = 3000
      tags = {
        "Project" : "unicorn"
        "kuma.io/service" = "test"
      }
    }

    outbound {
      port    = 80
      service = "some-backend"
    }

    outbound {
      port    = 8000
      service = "some-other-backend"
    }
  }
}
