# ITU-MiniTwit

## Deployment

Production runs on a 3-node Docker Swarm cluster on DigitalOcean fronted by Traefik with Let's Encrypt TLS. Pushes to `master` trigger GitHub Actions, which builds the images, pushes them to GHCR, and runs `docker stack deploy` on the manager. See [docs/operations/docker-swarm.md](docs/operations/docker-swarm.md) for the full setup and [docs/architecture/architecture.md](docs/architecture/architecture.md) for the topology.

## Spinning up the infrastructure (optional)

Terraform + DigitalOcean. All config lives in [terraform/](terraform/).

### You need

- [Terraform CLI](https://learn.hashicorp.com/tutorials/terraform/install-cli)
- A DigitalOcean account with a Full Access API token
- A DigitalOcean Spaces bucket named `devtroopers-minitwit-terraform` in `fra1` with a Spaces access key pair (for remote Terraform state)

Generate a dedicated SSH key pair from the `terraform/` folder (do this once per machine):

```bash
mkdir ssh_key && ssh-keygen -t rsa -b 4096 -q -N '' -f ./ssh_key/terraform
```

`ssh_key/` is gitignored and must be regenerated each time you provision from scratch.

### Secrets file

Copy the template and fill in the blanks:

```bash
cp terraform/secrets_template terraform/secrets
nano terraform/secrets
```

```bash
export TF_VAR_do_token=          # DigitalOcean API token
export SPACE_NAME=devtroopers-minitwit-terraform
export STATE_FILE=minitwit/terraform.tfstate
export AWS_ACCESS_KEY_ID=        # DO Spaces access key
export AWS_SECRET_ACCESS_KEY=    # DO Spaces secret key
export TF_VAR_db_password=       # choose a database password
export TF_VAR_secret_key=        # random string for session signing
export TF_VAR_discord_webhook_url=  # Discord webhook for Grafana alerts
```

This file is in `.gitignore` so it won't be committed.

### Run

```bash
cd terraform
bash bootstrap.sh
```

This provisions all Droplets, sets up PostgreSQL, creates Docker secrets, and deploys the stack. When done the public IP is printed — site will be at `https://<ip>`.

SSH to the swarm leader with `ssh root@<leader-ip> -i ssh_key/terraform`. Tear everything down with:

```bash
source secrets && terraform destroy -auto-approve
```

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

To provision a throw-away cluster from the `dev` branch, point `docker-stack.yml` at the dev image tags before running `bootstrap.sh`, or push dev images to GHCR and redeploy manually from the leader.
