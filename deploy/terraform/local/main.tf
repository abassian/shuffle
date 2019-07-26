# Configure Docker provider and connect to the local Docker socket
provider "docker" {
  host = "unix:///var/run/docker.sock"
}

# Create a new docker network
resource "docker_network" "private_network" {
  name            = "monet"
  check_duplicate = true
  driver          = "bridge"

  ipam_config {
    subnet  = "172.77.5.0/24"
    gateway = "172.77.5.254"
  }
}

# Create shuffle containers
resource "docker_container" "shuffle" {
  count = "${var.nodes}"

  name     = "node${count.index}"
  hostname = "node${count.index}"

  image = "abassian/shuffle:${var.version}"

  networks = ["${docker_network.private_network.name}"]

  # The conf files are mounted in a volume. shuffle, executed by the user
  # specified below, will read and write to this volume. So the user, needs
  # permissions on the host machine (host_path at least). Here, you want to
  # provide the same user that created the /conf folder.
  user = "${var.user}"
  env = ["HOME=/home/${var.user}"]
  volumes {
    host_path      = "${var.conf}/node${count.index}"
    container_path = "/home/${var.user}/.shuffle"
    read_only      = false
  }

  #entrypoint = ["tail", "-f", "/dev/null"]
  entrypoint = ["shl", "run", "${var.consensus}"]

  provisioner "local-exec" {
    command = "echo node${count.index} ${self.ip_address} >> ips.dat"
  }
}

output "public_addresses" {
  value = ["${docker_container.shuffle.*.ip_address}"]
}
