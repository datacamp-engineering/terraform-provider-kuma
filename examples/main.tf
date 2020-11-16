provider "kuma" {
  host = "http://localhost:5681"
  api_token = "test123"
}

resource "kuma_mesh" "yolo2" {
  name = "yolo2"
}

resource "kuma_dataplane" "yolo_data" {
  mesh = kuma_mesh.yolo2.name
  name = "yolo_data"
}
