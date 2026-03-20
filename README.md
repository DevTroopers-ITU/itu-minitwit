# ITU-MiniTwit

## Deployment

Vagrant + Hetzner Cloud. (can be changed at any point)

### You need

- Vagrant ([download](https://www.vagrantup.com/downloads))
- `vagrant plugin install vagrant-hetznercloud`
- One-liner to create and add a hetznercloud dummy box
  `mkdir /tmp/hetzner-dummy && cd /tmp/hetzner-dummy && echo '{"provider":"hetznercloud"}' > metadata.json && tar czf hetzner-dummy.box metadata.json && vagrant box add hetzner-dummy.box --name dummy --provider hetznercloud && cd ~`

- Your SSH key on Hetzner Console (ask Leo if unsure)`

### Environment variables

Get the token from Leo. SSH key name is one of these Peter is missing (`peter-juul`, `haakon`, `apoorva`, `leo`).

Create a `.env` file (same folder as this README):

```bash
nano .env
```

Paste the following and save (`Ctrl+O`, `Ctrl+X`):

```bash
export HCLOUD_TOKEN="your-token-here"
export HCLOUD_SSH_KEY_NAME="your-name-here"
export SSH_KEY_PATH="~/.ssh/id_ed25519"   # or ~/.ssh/id_rsa
export SECRET_KEY=""                      # insert random string (longer the better)
export DATABASE_URL=""                    # insert the Postgres URL (ask for it in Discord)
export POSTGRES_PASSWORD=""               # insert the Postgres password (ask for it in Discord)
```

This file is in `.gitignore` so it won't be committed. You only create it once, but you need to load it every time you open a new terminal:

```bash
source .env
ssh-add ~/.ssh/id_ed25519                  # same key as SSH_KEY_PATH
```

### Plugin bugfix (required)

The vagrant-hetznercloud plugin has a typo bug. Run this once after installing the plugin:

```bash
sed -i '' 's/option\[:location\]/options[:location]/;s/option\[:datacenter\]/options[:datacenter]/;s/option\[:user_data\]/options[:user_data]/' \
  ~/.vagrant.d/gems/3.3.8/gems/vagrant-hetznercloud-0.0.1/lib/vagrant-hetznercloud/action/create_server.rb
```

### Run

```bash
source .env                          # load environment variables first
vagrant up --provider=hetznercloud
```

App will be at `http://<server-ip>:8080`. Takes a few minutes to build.

`vagrant ssh` to get into the server, `vagrant destroy` to tear it down.

## Testing

We have two layers of tests:

Tests the code directly without starting a server. Runs fast and requires no setup:

```bash
go test -v
```

### Go tests (unit/integration)

Tests the code directly without starting a server. Runs fast and requires no setup:

```bash
go test -v
```

This runs both `main_test.go` (web UI tests) and `sim_api_test.go` (simulator API tests).

### Python E2E tests (integration against running server)

Tests the API over HTTP against a running server. Requires Python with `pytest` and `requests`.

Install dependencies (one time):

```bash
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

Run the server and tests:

```bash
# Terminal 1 — start the server
go run .

# Terminal 2 — run the tests
pytest python-references/minitwit_sim_api_test.py -v
```

Note: On macOS, port 5000 is taken by AirPlay Receiver — that's why we use port 8080.

### Test development

If you just want to make virtual machines on hetzner with the development branch, you can run:

```bash
source .env                          # load environment variables first
DEPLOY_BRANCH=dev vagrant up --provider=hetznercloud
```
