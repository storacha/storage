# How to Run a Piri PDP Server

## Prerequisites

This Guide assumes the reader has:

- The following packages installed: `make`, `go (1.23 or later)`, `git`, `jq`, `nginx`, `certbot python3-certbot-nginx`.
- A Domain Name, e.g. `SpicyStorage.tech` with a DNS record pointing at the machine they intend to run `piri` on.
    - **In the steps to follow this domain name will be referenced by the variable `YOUR_DOMAN`.**
- A machine without existing services occupying ports or special security considerations.
- A basic understanding of Unix-like systems.
- A Lotus Node synced to the Filecoin Calibration Network
    - Basic understanding of Filecoin Primitives and Addresses 
- TODO: Something about hardware requirements and disk space.

## Configuration and Installation

This section walks you through setting up nginx, installing the Piri node, created a Delegated Filecoin Address, and Starting your Piri PDP Server.

### Configure and Start Nginx

We will be using nginx to handle TLS termination and proxy requests to the Piri node.

1. Create an nginx configuration file: `vim /etc/nginx/sites-available/YOUR_DOMAIN`

    1. Add the following:
        ```nginx
         server {
            server_name YOUR_DOMAIN;
            # Allow unlimited file upload size (0 = no limit) 
            # since piri accepts data uploads from clients.
            client_max_body_size 0;
            # Increase timeout for receiving client request body (default is 60s)
            # Prevents timeout errors during slow uploads of large files.
            client_body_timeout 300s;
            # Increase timeout for receiving client request headers (default is 60s)
            # Useful for clients with slow connections.
            client_header_timeout 300s;
            # Increase timeout for transmitting response to client (default is 60s)
            # Prevents timeout when sending large responses back.
            send_timeout 300s;
            location / {
                proxy_pass http://localhost:3001; # piri pdp server port
                proxy_http_version 1.1;
                proxy_set_header Upgrade $http_upgrade;
                proxy_set_header Connection 'upgrade';
                proxy_set_header Host $host;
                proxy_set_header X-Real-IP $remote_addr;
                proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                proxy_set_header X-Forwarded-Proto $scheme;
                proxy_cache_bypass $http_upgrade;
                
                # Disable request buffering - stream uploads directly to backend
                # This saves memory and allows the backend to process uploads as 
                # they arrive.
                proxy_request_buffering off;
            }
            listen 443 ssl;
            ssl_certificate /etc/letsencrypt/live/YOUR_DOMAIN/fullchain.pem;
            ssl_certificate_key /etc/letsencrypt/live/YOUR_DOMAIN/privkey.pem;
            include /etc/letsencrypt/options-ssl-nginx.conf;
            ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;
        }
        server {
            if ($host = YOUR_DOMAIN) {
                return 301 https://$host$request_uri;
            }
            listen 80;
            server_name YOUR_DOMAIN;
            return 404;
        }
        ```

2. Enable the site

    1. `ln -s /etc/nginx/sites-available/YOUR_DOMAIN /etc/nginx/sites-enabled/`

3. Test and reload nginx

    1. `nginx -t`
    2. `systemctl reload nginx`

4. Obtain an SSL certificate

    1. `certbot --nginx -d YOUR_DOMAIN`
    2. Follow the prompts to complete the certificate installation.

### Create and Export a Delegated Filecoin Address

1. Create a [delegated address](https://docs.filecoin.io/smart-contracts/filecoin-evm-runtime/address-types#delegated-addresses) via Lotus:
   - `lotus wallet new delegated`
     - Example Output: `t410fzmmaqcn3j6jidbyrqqsvayejt6sskofwash6zoi`
2. Fund the wallet
   1. Request funds for the calibration network via the [faucet](https://faucet.calibnet.chainsafe-fil.io/funds.html).
3. Export the wallet to a file:
   1. `lotus wallet export YOUR_ADDRESS > YOUR_ADDRESS.hex`

### Install Piri

1. Clone the Piri Repo and Build
    1. `git clone https://github.com/storacha/piri`
    2. `cd piri`
    3. `make calibnet`
    4. `cp piri /usr/local/bin/piri`

### Import Wallet to Piri

1. `piri wallet import YOUR_ADDRESS.hex`
   1. Example Output: `imported wallet 0x7469B47e006D0660aB92AE560b27A1075EEcF97F successfully!`
2. Note: You may list wallets held by the piri node via: `piri wallet list`
   1. Example Output: `Address:  0x7469B47e006D0660aB92AE560b27A1075EEcF97F`

### Start the PDP Server

1. `piri serve pdp --lotus-client-host=wss://LOTUS_ENDPOINT/rpc/v1 --eth-client-host=wss://LOTUS_ENDPOINT/rpc/v1 --pdp-address=<ETH_ADDRESS_FROM_WALLET_LIST>`
   - Example, using address from previous command: 
   - `piri serve pdp --lotus-client-host=wss://LOTUS_ENDPOINT/rpc/v1 --eth-client-host=wss://LOTUS_ENDPOINT/rpc/v1 --pdp-address=0x7469B47e006D0660aB92AE560b27A1075EEcF97F`

**🎉 Congratulations! Your Piri PDP Server is now running.**