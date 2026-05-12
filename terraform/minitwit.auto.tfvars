# Frankfurt region — all droplets deployed here
region  = "fra1"

# SSH keys are required before running bootstrap.sh
# Run these commands from the terraform/ folder:
# mkdir ssh_key && ssh-keygen -t rsa -b 4096 -q -N '' -f ./ssh_key/terraform
# Note: ssh_key/ is gitignored and must be generated locally each time

pub_key = "ssh_key/terraform.pub"
pvt_key = "ssh_key/terraform"