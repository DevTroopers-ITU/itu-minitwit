Vagrant.configure("2") do |config|
  config.vm.box = "dummy"
  config.vm.synced_folder ".", "/vagrant", disabled: true
  config.ssh.private_key_path = ENV["SSH_KEY_PATH"] || "~/.ssh/id_rsa"
  config.ssh.username = "root"

  config.vm.provider :hetznercloud do |hcloud|
    hcloud.token = ENV["HCLOUD_TOKEN"]
    hcloud.image = "ubuntu-22.04"
    hcloud.location = "nbg1"
    hcloud.server_type = "cx23"
    hcloud.ssh_keys = [ENV["HCLOUD_SSH_KEY_NAME"] || "leo"]
  end

  config.vm.provision "shell", inline: <<-SHELL
    # Install Docker
    apt-get update -qq
    apt-get install -y -qq docker.io docker-compose-v2 git > /dev/null 2>&1

    # Clone and deploy
    cd /root
    git clone -b master https://github.com/DevTroopers-ITU/itu-minitwit.git
    cd itu-minitwit
    docker compose up --build -d
  SHELL
end
