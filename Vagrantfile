branch = ENV["DEPLOY_BRANCH"] || "master"

Vagrant.configure("2") do |config|
  config.vm.box = "dummy"
  config.vm.synced_folder ".", "/vagrant", disabled: true
  config.ssh.private_key_path = ENV["SSH_KEY_PATH"] || "~/.ssh/id_rsa"
  config.ssh.username = "root"
  config.vm.provision "file", source: ".env", destination: "/root/.env"

  config.vm.provider :hetznercloud do |hcloud|
    hcloud.token = ENV["HCLOUD_TOKEN"]
    hcloud.image = "ubuntu-22.04"
    hcloud.location = "nbg1"
    hcloud.server_type = "cx23"
    hcloud.ssh_keys = ["leo", "haakon-2", "apoorva", "peter-juul", "phkt@archlinux"]
  end

  config.vm.provision "shell", inline: <<-SHELL
    apt-get update -qq
    apt-get install -y -qq docker.io docker-compose-v2 git > /dev/null 2>&1

    if [ -d /root/itu-minitwit ]; then
      echo "Repo already exists, pulling latest..."
      cd /root/itu-minitwit
      git pull origin #{branch}
    else
      echo "Cloning repo..."
      git clone -b #{branch} https://github.com/DevTroopers-ITU/itu-minitwit.git /root/itu-minitwit
    fi

    cp /root/.env /root/itu-minitwit/.env

    cd /root/itu-minitwit
    docker compose up --build -d

    echo "================================================"
    echo " MiniTwit is running at http://$(curl -s -4 ifconfig.me):$(grep -m1 'EXPOSE' /root/itu-minitwit/Dockerfile | awk '{print $2}')"
    echo "================================================"
  SHELL
end