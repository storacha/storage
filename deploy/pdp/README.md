# PDP Deployment (Lotus + Curio + Yugabyte)

This Terraform configuration deploys a PDP node using three major components:
1. **Lotus** – Interacts with the Filecoin network.
2. **Curio** – Coordinates storage, PDP (Provable Data Possession) services, and network communication.
3. **Yugabyte** – Provides a distributed database for Curio’s metadata.

## Quick Start

1. **Create/Configure Wallets & Keys**
    - Ensure you have:
        - **BLS Lotus wallet** (for configuring Curio).
        - **Delegated Lotus wallet** (for interacting with the PDP smart contract).
        - A **PEM file** (`service.pem`) to authenticate a Storacha storage node with the Curio API.

2. **Create `terraform.tfvars`**  
   Provide the file paths to your wallet JSON files and PEM key file:
   ```hcl
   lotus_wallet_bls_file       = "./bls.json"
   lotus_wallet_delegated_file = "./delegated.json"
   curio_service_pem_key_file  = "./service.pem"
   
3. **Plan & Apply**
    ```bash
    tofu plan -var-file=terraform.tfvars
    tofu apply -var-file=terraform.tfvars
    ```
   Once apply completes, note the public IP address in the output and confirm the node is also available at https://pdp.storacha.network.



## Detailed Setup Steps

There are three requirements to deploy this terraform:
1. A BLS Lotus wallet with funds for configuring Curio (TODO: ideally this isn't required, but for now it is.)
2. A Delegated Lotus wallet with funds for interation with the PDP Smart-Contract.
3. A PEM file containing a public key for authenticating a storacha storage node with the curio API.

### 1. Create Lotus wallets
**a. Create a BLS wallet:**
```bash
export BLS_WALLET=$(lotus wallet new bls) \
&& lotus wallet export "$BLS_WALLET" \
| xxd -r -p \
| jq -c --arg address "$BLS_WALLET" '.Address = $address' \
> bls.json
```

**b. Create a Delegated wallet:**
```bash
export DELEGATED_WALLET=$(lotus wallet new delegated) \
&& lotus wallet export "$DELEGATED_WALLET" \
| xxd -r -p \
| jq -c --arg address "$DELEGATED_WALLET" '.Address = $address' \
> delegated.json
```

**c. Fund both wallets using the [calibration faucet](https://faucet.calibnet.chainsafe-fil.io/funds.html):**
- BLS wallet address: `cat bls.json | jq '.Address'` 
- Delegated wallet address: `cat delegated.json | jq '.Address'`

### 2. Create a PEM file the storacha storage node
_draw the rest of the !%^&$#@ owl..._

**TODO: we may need to add support to the `storage identity` command for this.**


You need a PEM file (`service.pem`) so Curio can authorize API requests from the Storacha storage node.


Generally this looks something like:
`./storage identity generate > service.pem`

## 3. Populate `terraform.tfvars`

Create a file named `terraform.tfvars` with the necessary variables:
```terraform
lotus_wallet_bls_file       = "./bls.json"
lotus_wallet_delegated_file = "./delegate.json"
curio_service_pem_key_file  = "./service.pem"
```
You can also customize things like instance size, region, or volume size if needed.


## 4. Deploy
a. **Plan:**
```
tofu plan -var-vars=terraform.tfvars
```

b. **Apply:**
```
tofu apply -var-vars=terraform.tfvars
```

When finished, the Terraform output shows your instance’s public IP. You can also reach it at https://pdp.storacha.network.


## Monitoring & Next Steps
### 1. Access the Instance

Use the provided private key from [1Password](https://start.1password.com/open/i?a=SJ2Q5WC77NHLDFRUAUF6HYKKHU&v=rof22gakdxoldc6exdtsqmo5tu&i=walq46vmpwcvxd6mqvhzkxps6e&h=storachainc.1password.com) (different from service.pem) to SSH: 
```bash
$ ssh -i /path/to/key.pem ubuntu@pdp.storacha.network
```
_(Replace `/path/to/key.pem` with the actual file path.)_


### 2. Check Service States
All key services must reach active before the node is fully ready. You can watch them in real time:
```bash
watch -n 2 '
  for s in yugabyte yugabyte-ready lotus-prestart lotus lotus-ready lotus-poststart curio-prestart curio curio-ready curio-poststart
  do
    echo "$s: $(systemctl is-active $s)" | ccze -A
  done
'
```
Once all show `active`, the node is operational.

### 3. Review Logs
To view live logs:
```bash
journalctl -u yugabyte -u yugabyte-ready \
  -u lotus-prestart -u lotus -u lotus-ready -u lotus-poststart \
  -u curio-prestart -u curio -u curio-ready -u curio-poststart \
  -f -o short-iso | ccze -A
```

### 4. Monitor Resource Usage
Use top to see CPU and memory usage for Lotus and Curio:
```bash
top -p $(pgrep -d',' -f lotus),$(pgrep -d',' -f curio)
```

### 5. Next Steps
Optionally Deploy and Connect a Storacha Storage Node to the Curio instance for PDP interactions
