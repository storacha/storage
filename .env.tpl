# required

TF_WORKSPACE= # your name. Note: "staging" and "prod" are reserved for deployments from CI
TF_VAR_private_key= # generate using CLI: `./storage identity gen`
TF_VAR_indexing_service_proof= # obtain from Storacha support
TF_VAR_allowed_account_ids='["0"]' # set to your AWS account ID
TF_VAR_domain=storacha.network # domain name to use for the deployment (will be prefixed with app name unless "prod" deployment)

# optional - AWS config

TF_VAR_region=us-west-2 # AWS region to deploy all resources

# tags applied to AWS resources (useful for cost accounting)
TF_VAR_app=storage
TF_VAR_owner=storacha
TF_VAR_team="Storacha Engineer"
TF_VAR_org=Storacha

# optional - curio integration

TF_VAR_use_pdp=false
TF_VAR_pdp_proofset=
TF_VAR_curio_url=

# optional - external (S3 compatible) blob bucket

TF_VAR_use_external_blob_bucket=false
TF_VAR_external_blob_bucket_endpoint= # API endpoint for the external bucket
TF_VAR_external_blob_bucket_region=
TF_VAR_external_blob_bucket_name=
TF_VAR_external_blob_bucket_domain= # public domain for accessing bucket
TF_VAR_external_blob_bucket_access_key_id=
TF_VAR_external_blob_bucket_secret_access_key=

# optional - indexing service configuration

TF_VAR_indexing_service_did=did:web:indexer.storacha.network
TF_VAR_indexing_service_url=https://indexer.storacha.network

# optional - debugging

# Setting this variable enables tracing for lambdas based on HTTP handlers.
# Currently, only Honeycomb is supported as the tracing backend. You can create
# a Honeycomb account and get an API key from honeycomb.io.
TF_VAR_honeycomb_api_key=
