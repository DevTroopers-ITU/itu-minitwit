#!/bin/bash
set -e
echo -e "\n--> Bootstrapping Minitwit\n"

echo -e "\n--> Loading environment variables from secrets file\n"
# shellcheck source=/dev/null
source secrets

echo -e "\n--> Checking that environment variables are set\n"
# check that all variables are set
[ -z "$TF_VAR_do_token" ] && echo "TF_VAR_do_token is not set" && exit
[ -z "$SPACE_NAME" ] && echo "SPACE_NAME is not set" && exit
[ -z "$STATE_FILE" ] && echo "STATE_FILE is not set" && exit
[ -z "$AWS_ACCESS_KEY_ID" ] && echo "AWS_ACCESS_KEY_ID is not set" && exit
[ -z "$AWS_SECRET_ACCESS_KEY" ] && echo "AWS_SECRET_ACCESS_KEY is not set" && exit
[ -z "$TF_VAR_db_password" ] && echo "TF_VAR_db_password is not set" && exit
[ -z "$TF_VAR_secret_key" ] && echo "TF_VAR_secret_key is not set" && exit
[ -z "$TF_VAR_discord_webhook_url" ] && echo "TF_VAR_discord_webhook_url is not set" && exit

echo -e "\n--> Initializing terraform\n"
# initialize terraform
terraform init \
    -backend-config "bucket=$SPACE_NAME" \
    -backend-config "key=$STATE_FILE" \
    -backend-config "access_key=$AWS_ACCESS_KEY_ID" \
    -backend-config "secret_key=$AWS_SECRET_ACCESS_KEY"

# check that everything looks good
echo -e "\n--> Validating terraform configuration\n"
terraform validate

# create infrastructure
echo -e "\n--> Creating Infrastructure\n"
terraform apply -auto-approve

# grab outputs from terraform
echo -e "\n--> Fetching infrastructure details\n"
LEADER_IP=$(terraform output -raw minitwit-swarm-leader-ip-address)
DB_IP=$(terraform output -raw minitwit-db-ip-address)
PUBLIC_IP=$(terraform output -raw public_ip)

# build database connection string and create docker secrets on the swarm
echo -e "\n--> Creating Docker secrets on swarm\n"
DB_URL="postgresql://minitwit:${TF_VAR_db_password}@${DB_IP}:5432/minitwit"
ssh -o 'StrictHostKeyChecking no' "root@${LEADER_IP}" -i ssh_key/terraform \
    "echo '$DB_URL' | docker secret create database_url -"
ssh -o 'StrictHostKeyChecking no' "root@${LEADER_IP}" -i ssh_key/terraform \
    "echo '$TF_VAR_secret_key' | docker secret create secret_key -"
ssh -o 'StrictHostKeyChecking no' "root@${LEADER_IP}" -i ssh_key/terraform \
    "echo '$TF_VAR_discord_webhook_url' | docker secret create discord_webhook_url -"

# copy stack file to leader and deploy
# ../docker-stack.yml points to the stack file in the root of the repo
echo -e "\n--> Copying stack file to leader\n"
scp -o 'StrictHostKeyChecking no' \
    -i ssh_key/terraform \
    ../docker-stack.yml \
    "root@${LEADER_IP}:~/docker-stack.yml"

# deploy the stack to the cluster
echo -e "\n--> Deploying the Minitwit stack to the cluster\n"
ssh \
    -o 'StrictHostKeyChecking no' \
    "root@${LEADER_IP}" \
    -i ssh_key/terraform \
    'docker stack deploy minitwit -c docker-stack.yml'

echo -e "\n--> Done bootstrapping Minitwit"
echo -e "--> The db will need a moment to initialize, this can take up to a couple of minutes..."
echo -e "--> Site will be available @ https://$PUBLIC_IP"
echo -e "--> ssh to swarm leader with 'ssh root@$LEADER_IP -i ssh_key/terraform'"
echo -e "--> To remove the infrastructure run: terraform destroy -auto-approve"