Author - Claude
# Minitwit Infrastructure

This folder contains the Terraform configuration for provisioning and deploying
the Minitwit application on DigitalOcean. Running `bootstrap.sh` will set up
the entire infrastructure from scratch and deploy the app automatically.

## Overview

The infrastructure consists of 4 Droplets:

- **minitwit-swarm-leader** — Docker Swarm manager/leader, runs Traefik for routing
- **minitwit-swarm-worker-0** — Docker Swarm worker, runs app containers
- **minitwit-swarm-worker-1** — Docker Swarm worker, runs app containers
- **minitwit-db** — Standalone PostgreSQL database, not part of the swarm

## File Structure

| File | Purpose |
|------|---------|
| `bootstrap.sh` | One-time setup script — runs everything in order |
| `provider.tf` | Declares variables and sets up the DigitalOcean provider |
| `backend.tf` | Configures DO Spaces as remote storage for Terraform state |
| `minitwit_swarm_cluster.tf` | Creates all Droplets and firewall rules |
| `ip.tf` | Creates and assigns the floating IP to the swarm leader |
| `ssh_key.tf` | Uploads your SSH public key to DigitalOcean |
| `minitwit.auto.tfvars` | Non-sensitive variable values (region, key paths) |
| `secrets_template` | Template showing which secrets need to be filled in |
| `secrets` | Your actual secret values — never committed to GitHub |

## Prerequisites

### 1. Install Terraform
Follow the instructions at https://learn.hashicorp.com/tutorials/terraform/install-cli

### 2. Generate SSH keys
Run these commands from the `terraform/` folder:
```bash
mkdir ssh_key && ssh-keygen -t rsa -b 4096 -q -N '' -f ./ssh_key/terraform
```
Note: `ssh_key/` is gitignored and must be generated locally every time you
spin up the infrastructure from scratch.

### 3. Create a DO Spaces bucket
The Terraform state file is stored remotely in a DigitalOcean Spaces bucket
so it's accessible from any machine. Create one manually in the DO dashboard:
- Go to Spaces Object Storage → Create Bucket
- Name it `devtroopers-minitwit-terraform`
- Choose the Frankfurt (fra1) region

### 4. Generate DO Spaces access keys
In the DO dashboard go to Spaces Object Storage → Access Keys and generate
a new key pair. You'll need both the Access Key ID and the Secret Key.

### 5. Generate a DO API token
In the DO dashboard go to API → Tokens and generate a new token with Full Access.

### 6. Set up your secrets file
Copy the template and fill in the blanks:
```bash
cp secrets_template secrets
```

Then open `secrets` and fill in all the empty values:
```bash
# DigitalOcean API token
export TF_VAR_do_token=your_do_token_here

# DO Spaces credentials
export AWS_ACCESS_KEY_ID=your_spaces_key_id
export AWS_SECRET_ACCESS_KEY=your_spaces_secret_key

# Database password — choose your own
export TF_VAR_db_password=your_database_password

# Secret key for signing session cookies — any long random string
export TF_VAR_secret_key=your_random_secret_string

# Discord webhook URL for Grafana alerts
export TF_VAR_discord_webhook_url=your_discord_webhook_url
```

## Spinning up the infrastructure

Once all prerequisites are done, run:
```bash
bash bootstrap.sh
```

This will:
1. Load and validate your secrets
2. Initialize Terraform with the DO Spaces backend
3. Validate the Terraform configuration
4. Create all Droplets and firewall rules
5. Install PostgreSQL on the DB Droplet and create the database
6. Create Docker secrets on the swarm (database_url, secret_key, discord_webhook_url)
7. Copy the stack file to the swarm leader
8. Deploy the Minitwit stack to the cluster

When done, the site will be available at the printed public IP. Point your
DNS records at this IP to restore the domain.

## Tearing down the infrastructure

```bash
source secrets
terraform destroy -auto-approve
```

This will delete all Droplets, firewalls and the floating IP. The DO Spaces
bucket and its state file are NOT deleted automatically — do that manually
in the DO dashboard if needed.

## Scaling

To scale the number of workers edit `minitwit_swarm_cluster.tf` and change
the `count` value on the worker Droplet resource, then run:
```bash
source secrets
terraform apply -auto-approve
```

Terraform will only create or remove the difference — existing Droplets are
left untouched.