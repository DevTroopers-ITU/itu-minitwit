# ITU-MiniTwit

## Deployment

Vagrant + Hetzner Cloud. (can be changed at any point)

### You need

- Vagrant ([download](https://www.vagrantup.com/downloads))
- `vagrant plugin install vagrant-hetznercloud`
-  One-liner to create and add a hetznercloud dummy box
`mkdir /tmp/hetzner-dummy && cd /tmp/hetzner-dummy && echo '{"provider":"hetznercloud"}' > metadata.json && tar czf hetzner-dummy.box metadata.json && vagrant box add hetzner-dummy.box --name dummy --provider hetznercloud && cd ~`

- Your SSH key on Hetzner Console (ask Leo if unsure)`

### Environment variables

Get the token from Leo. SSH key name is one of these Peter is missing (`peter-juul`, `haakon`, `apoorva`, `leo`).

```bash
export HCLOUD_TOKEN="..."
export HCLOUD_SSH_KEY_NAME="your-name-here"
ssh-add ~/.ssh/id_rsa OR ssh-add ~/.ssh/id_ed25519
```

### Plugin bugfix (required)

The vagrant-hetznercloud plugin has a typo bug. Run this once after installing the plugin:

```bash
sed -i '' 's/option\[:location\]/options[:location]/;s/option\[:datacenter\]/options[:datacenter]/;s/option\[:user_data\]/options[:user_data]/' \
  ~/.vagrant.d/gems/3.3.8/gems/vagrant-hetznercloud-0.0.1/lib/vagrant-hetznercloud/action/create_server.rb
```

### Run

```bash
vagrant up --provider=hetznercloud
```

App will be at `http://<server-ip>:5001`. Takes a few minutes to build.

`vagrant ssh` to get into the server, `vagrant destroy` to tear it down.
