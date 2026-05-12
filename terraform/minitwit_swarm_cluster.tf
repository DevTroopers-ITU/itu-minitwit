#  _                _
# | | ___  __ _  __| | ___ _ __
# | |/ _ \/ _` |/ _` |/ _ \ '__|
# | |  __/ (_| | (_| |  __/ |
# |_|\___|\__,_|\__,_|\___|_|

# create cloud vm
resource "digitalocean_droplet" "minitwit-swarm-leader" {
  image    = "docker-20-04"
  name     = "minitwit-swarm-leader"
  region   = var.region
  size     = "s-1vcpu-1gb"
  tags     = ["minitwit-swarm", "minitwit-leader"]
  # add public ssh key so we can access the machine
  ssh_keys = [digitalocean_ssh_key.minitwit.fingerprint]

  # specify a ssh connection
  connection {
    user        = "root"
    host        = self.ipv4_address
    type        = "ssh"
    private_key = file(var.pvt_key)
    timeout     = "2m"
  }

  provisioner "remote-exec" {
    inline = [
      # ufw rules - swarm communication ports
      "ufw allow 2377/tcp",
      "ufw allow 7946/tcp",
      "ufw allow 7946/udp",
      "ufw allow 4789/udp",
      # ufw rules - public ports (traefik handles routing)
      "ufw allow 80/tcp",   # HTTP
      "ufw allow 443/tcp",  # HTTPS
      "ufw allow 8080/tcp", # simulator API
      # ufw rules - SSH
      "ufw allow 22/tcp",
      "ufw enable --force",

      # initialize docker swarm cluster
      "docker swarm init --advertise-addr ${self.ipv4_address}"
    ]
  }
}

resource "null_resource" "swarm-worker-token" {
  depends_on = [digitalocean_droplet.minitwit-swarm-leader]

  # save the worker join token
  provisioner "local-exec" {
    command = "ssh -o 'ConnectionAttempts 3600' -o 'StrictHostKeyChecking no' root@${digitalocean_droplet.minitwit-swarm-leader.ipv4_address} -i ssh_key/terraform 'docker swarm join-token worker -q' > temp/worker_token"
  }
}


#                     _
# __      _____  _ __| | _____ _ __
# \ \ /\ / / _ \| '__| |/ / _ \ '__|
#  \ V  V / (_) | |  |   <  __/ |
#   \_/\_/ \___/|_|  |_|\_\___|_|

# create cloud vm
resource "digitalocean_droplet" "minitwit-swarm-worker" {
  # create workers after the leader
  depends_on = [null_resource.swarm-worker-token]

  # number of vms to create
  count = 2

  image    = "docker-20-04"
  name     = "minitwit-swarm-worker-${count.index}"
  region   = var.region
  size     = "s-1vcpu-1gb"
  tags     = ["minitwit-swarm"]
  # add public ssh key so we can access the machine
  ssh_keys = [digitalocean_ssh_key.minitwit.fingerprint]

  # specify a ssh connection
  connection {
    user        = "root"
    host        = self.ipv4_address
    type        = "ssh"
    private_key = file(var.pvt_key)
    timeout     = "2m"
  }

  provisioner "file" {
    source      = "temp/worker_token"
    destination = "/root/worker_token"
  }

  provisioner "remote-exec" {
    inline = [
      # ufw rules - swarm communication ports only (no public ports on workers)
      "ufw allow 2377/tcp",
      "ufw allow 7946/tcp",
      "ufw allow 7946/udp",
      "ufw allow 4789/udp",
      # ufw rules - SSH
      "ufw allow 22/tcp",
      "ufw enable --force",

      # join swarm cluster as worker
      "docker swarm join --token $(cat worker_token) ${digitalocean_droplet.minitwit-swarm-leader.ipv4_address}"
    ]
  }
}


#      _       _        _
#   __| | __ _| |_ __ _| |__   __ _ ___  ___
#  / _` |/ _` | __/ _` | '_ \ / _` / __|/ _ \
# | (_| | (_| | || (_| | |_) | (_| \__ \  __/
#  \__,_|\__,_|\__\__,_|_.__/ \__,_|___/\___|

resource "digitalocean_droplet" "minitwit-db" {
  image    = "ubuntu-24-04-x64"
  name     = "minitwit-db"
  region   = var.region
  size     = "s-1vcpu-1gb"
  tags     = ["minitwit-db"]
  # add public ssh key so we can access the machine
  ssh_keys = [digitalocean_ssh_key.minitwit.fingerprint]

  # specify a ssh connection
  connection {
    user        = "root"
    host        = self.ipv4_address
    type        = "ssh"
    private_key = file(var.pvt_key)
    timeout     = "2m"
  }
  provisioner "remote-exec" {
    inline = [
      # ufw rules - only SSH and PostgreSQL
      "ufw allow 22/tcp",
      "ufw allow 5432/tcp",
      "ufw enable --force",

      # install PostgreSQL
      "apt-get update",
      "apt-get install -y postgresql",

      # create database and user
      "sudo -u postgres psql -c \"CREATE USER minitwit WITH PASSWORD '${var.db_password}';\"",
      "sudo -u postgres psql -c \"CREATE DATABASE minitwit OWNER minitwit;\"",

      # allow external connections to PostgreSQL
      "echo \"listen_addresses='*'\" >> /etc/postgresql/16/main/postgresql.conf",
      "echo \"host all all 0.0.0.0/0 md5\" >> /etc/postgresql/16/main/pg_hba.conf",
      "systemctl restart postgresql"
    ]
  }
}


#  _____ _                      _ _
# |  ___(_)_ __ _____      __ _| | |___
# | |_  | | '__/ _ \ \ /\ / / _` | | / __|
# |  _| | | | |  __/\ V  V / (_| | | \__ \
# |_|   |_|_|  \___| \_/\_/ \__,_|_|_|___/

# firewall for all swarm nodes (leader + workers)
resource "digitalocean_firewall" "minitwit-swarm" {
  name = "minitwit-swarm"
  tags = ["minitwit-swarm"]

  # SSH from anywhere
  inbound_rule {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  # swarm management - only from other swarm nodes
  inbound_rule {
    protocol    = "tcp"
    port_range  = "2377"
    source_tags = ["minitwit-swarm"]
  }

  # node communication TCP - only from other swarm nodes
  inbound_rule {
    protocol    = "tcp"
    port_range  = "7946"
    source_tags = ["minitwit-swarm"]
  }

  # node communication UDP - only from other swarm nodes
  inbound_rule {
    protocol    = "udp"
    port_range  = "7946"
    source_tags = ["minitwit-swarm"]
  }

  # overlay network - only from other swarm nodes
  inbound_rule {
    protocol    = "udp"
    port_range  = "4789"
    source_tags = ["minitwit-swarm"]
  }

  # allow all outbound traffic
  outbound_rule {
    protocol              = "tcp"
    port_range            = "all"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "udp"
    port_range            = "all"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "icmp"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}

# separate firewall for leader only - public traffic
resource "digitalocean_firewall" "minitwit-leader" {
  name = "minitwit-leader"
  tags = ["minitwit-leader"]

  # HTTP from anywhere (Traefik)
  inbound_rule {
    protocol         = "tcp"
    port_range       = "80"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  # HTTPS from anywhere (Traefik)
  inbound_rule {
    protocol         = "tcp"
    port_range       = "443"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  # simulator API from anywhere
  inbound_rule {
    protocol         = "tcp"
    port_range       = "8080"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  # allow all outbound traffic
  outbound_rule {
    protocol              = "tcp"
    port_range            = "all"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "udp"
    port_range            = "all"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}

# firewall for database - only allow swarm nodes in
resource "digitalocean_firewall" "minitwit-db" {
  name = "minitwit-db"
  tags = ["minitwit-db"]

  # SSH from anywhere
  inbound_rule {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  # PostgreSQL only from swarm nodes
  inbound_rule {
    protocol    = "tcp"
    port_range  = "5432"
    source_tags = ["minitwit-swarm"]
  }

  # allow all outbound traffic
  outbound_rule {
    protocol              = "tcp"
    port_range            = "all"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "udp"
    port_range            = "all"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}


#   ___        _               _
#  / _ \ _   _| |_ _ __  _   _| |_ ___
# | | | | | | | __| '_ \| | | | __/ __|
# | |_| | |_| | |_| |_) | |_| | |_\__ \
#  \___/ \__,_|\__| .__/ \__,_|\__|___/
#                 |_|

output "minitwit-swarm-leader-ip-address" {
  value = digitalocean_droplet.minitwit-swarm-leader.ipv4_address
}

output "minitwit-swarm-worker-ip-address" {
  value = digitalocean_droplet.minitwit-swarm-worker.*.ipv4_address
}

output "minitwit-db-ip-address" {
  value = digitalocean_droplet.minitwit-db.ipv4_address
}