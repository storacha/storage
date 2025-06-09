# How to Run a Piri UCAN Server

## Prerequisites

This Guide assumes the reader has:

- An existing PDP Node, either via Curio or Piri that is reachable at a FQDM supporting TLS. (TODO link to guides for setting up PDP node with respective implementations)
    - **In the steps to follow this domain will be referenced by the variable `PDP_DOMAIN`**
- The following packages installed: `make`, `go (1.23 or later)`, `git`, `jq`, `nginx`, `certbot python3-certbot-nginx`.
- A Domain Name, e.g. `SpicyStorage.tech` with a DNS record pointing at the machine they intend to run `piri` on.
    - **In the steps to follow this domain name will be referenced by the variable `YOUR_DOMAN`.**
- A machine without existing services occupying ports or special security considerations.
- A basic understanding of Unix-like systems.

## Configuration and Installation

This section walks you through setting up nginx, installing the Piri node, generating a private key, registering with the Delegator, and Starting your Piri Node.

### Configure and Start Nginx

We will be using nginx to handle TLS termination and proxy requests to the Piri node.

1. Create an nginx configuration file: `vim /etc/nginx/sites-available/YOUR_DOMAIN`

    1. Add the following:
        ```nginx
         server {
             server_name YOUR_DOMAIN;
             location / {
                 proxy_pass http://localhost:3000; # piri ucan server port
                 proxy_http_version 1.1;
                 proxy_set_header Upgrade $http_upgrade;
                 proxy_set_header Connection 'upgrade';
                 proxy_set_header Host $host;
                 proxy_set_header X-Real-IP $remote_addr;
                 proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                 proxy_set_header X-Forwarded-Proto $scheme;
                 proxy_cache_bypass $http_upgrade;
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

### Install Piri and Generate a Private Key

1. Clone the Piri Repo and Build
    1. `git clone https://github.com/storacha/piri`
    2. `cd piri`
    3. `make calibnet`
    4. `cp piri /usr/local/bin/piri`
2. Create a Private Key
    1. Generate a Private Key in [PEM format](https://en.wikipedia.org/wiki/Privacy-Enhanced_Mail): `piri id gen -t=pem`
    2. Write the Private Key block to a file, e.g. `service.pem`
    3. Note: A DID may be derived from this file via: `piri id parse service.pem | jq .did`

Note: If running Curio, the Public Key block will need to be added as a service named `storacha` to the Curio PDP configuration system (TODO: add link to Curio doc for guidance on this step.)

### Create a Proof Set

1. Create a Proof Set:

    - `piri proofset create --key-file=service.pem --curio-url=https://PDP_DOMAIN --record-keeper=0x6170dE2b09b404776197485F3dc6c968Ef948505`

2. Monitor Creation Status:

    - `piri proofset status --key-file=service.pem --curio-url=https://PDP_DOMAIN --ref-url=/pdp/proof-set/created/<HASH_FROM_CREATE_OUTPUT>`

    - While the proofset is being created, the previous command will return:

      ```json
       {
         "proofsetCreated": false,
         "txStatus": "pending",
         "ok": false
       }
      ```

    - Once the proofset has been create, the previous command will return:

      ```json
       {
         "proofsetCreated": true,
         "txStatus": "confirmed",
         "ok": true,
         "proofSetId": <YOUR_PROOFSET_ID>
       }
      ```

### Register You Piri Node with the Delegator

1. Share the DID of your Piri node with the Storacha Team.
2. Temporarily start your piri node: `piri start --key-file=service.pem`
3. Visit https://staging.delegator.storacha.network
4. Follow the instruction to receive your Delegation and Configuration Values
5. Stop the Piri node start in step 2 after submitting your details to the Delegator.

### Configure and Start Piri Delegation from Delegator

1. Create a `.env` file containing the environment variables received from the Delegator.
2. Apply the environment variables to your shell, i.e. `source .env`
3. Start your Piri node: `piri start --key-file=service.pem --curio-url=https://PDP_DOMAIN`

**🎉 Congratulations! Your Piri node is now running, and ready to accept data from the storacha warm storage network.**