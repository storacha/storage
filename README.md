<div align="center">
  <img src="https://bafybeihuuqarjd3wkswa6mbirykvtsdlwu7yaccqpujnbycvurb6jyq2fm.ipfs.w3s.link/the-racha-centipede.png" alt="Storacha storage node logo" width="180" />
  <h1>Storage Node</h1>
  <p>A storage node that runs on the Storacha network.</p>
</div>

## Usage

### Getting Started

Install [Go](https://go.dev) v1.23.3 or higher.

Next, generate a private key for your storage node. This can be done on any computer, not necessarily the deployment target.

Clone the repo and `cd` into the repo directory:

```sh
git clone https://github.com/storacha/storage.git
cd storage
```

Build and install the CLI tool:

```sh
make install
```

Generate a new identity:

```sh
storage identity gen
```

Make a note of your node identity. The string beginning `Mg` is your private key. Do not share this with anyone.

Next, obtain a delegation allowing your node to publish claims to the Storacha Indexer node(s). Contact the engineers in `#node-providers` on the Storacha Discord - give them your _public_ key (the string beginning with `did:key:`).

### System Requriements

TODO

### Deployment

#### Environment Variables

The environment variables required to start a node are:

```sh
STORAGE_PRIVATE_KEY=             # string beginning Mg...
STORAGE_PUBLIC_URL=              # URL the node will be publically accessible at
STORAGE_PORT=                    # local port to bind the server to
STORAGE_INDEXING_SERVICE_PROOF=  # delegation(s) from the Storacha Indexing node(s)
```

#### Deployment to a VM/Bare Metal

Clone the repo and build the binary as per the [getting started](#getting-started) section. Set environment variables as above. The following command will start the Storage Node daemon:

```sh
storage start
```

#### Deployment to DigitalOcean

The [Dockerfile](./Dockerfile) allows a Storage Node to be deployed to DigitalOcean Apps platform. You'll need to setup a "Spaces Object Storage" bucket to persist data. You must configure the following additional environment variables:

```sh
STORAGE_S3_ENDPOINT=
STORAGE_S3_REGION=
STORAGE_S3_BUCKET=
STORAGE_S3_ACCESS_KEY=
STORAGE_S3_SECRET_KEY=
```

#### Deployment to AWS

The Terraform scripts in `/deploy` allow a Storage Node to be deployed to AWS. See [deploy/README.md](./deploy/README.md) for instructions.

## Contributing

All welcome! Storacha is open-source. Please feel empowered to open a PR or an issue.

## License

Dual-licensed under [Apache 2.0 OR MIT](LICENSE.md)

