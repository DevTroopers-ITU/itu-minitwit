# ITU-MiniTwit

## Deployment

Vagrant + Hetzner Cloud. (can be changed at any point)

### You need

- Vagrant ([download](https://www.vagrantup.com/downloads))
- `vagrant plugin install vagrant-hetznercloud`
- Your SSH key on Hetzner Console (ask Leo if unsure)

### Environment variables

Get the token from Leo. SSH key name is one of these Peter is missing (`peter-juul`, `haakon`, `apoorva`, `leo`).

```bash
export HCLOUD_TOKEN="..."
export HCLOUD_SSH_KEY_NAME="your-name-here"
ssh-add ~/.ssh/id_rsa
```

### Run

```bash
vagrant up --provider=hetznercloud
```

App will be at `http://<server-ip>:5001`. Takes a few minutes to build.

`vagrant ssh` to get into the server, `vagrant destroy` to tear it down.
