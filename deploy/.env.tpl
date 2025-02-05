################################################################################
# REQUIRED
################################################################################

# Your name. Note: "staging" and "prod" are reserved for deployments from CI.
TF_WORKSPACE=
# Generate using CLI: `./storage identity gen`.
TF_VAR_private_key=
# A delegation granting the DID of this storage node (public key corresponding
# to the value set in `TF_VAR_private_key`) "claim/cache" on an indexer node.
# Obtain from Storacha support.
TF_VAR_indexing_service_proof=
# Set to your AWS account ID.
TF_VAR_allowed_account_ids=["0"]
# Domain name to use for the deployment. Automatically prefixed with app name
# (see `TF_VAR_app`) and workspace name (unless workspace is "prod").
# i.e. workspace.app.domain or app.domain if workspace == "prod".
TF_VAR_domain=storacha.network

################################################################################
# OPTIONAL
################################################################################

# AWS config ###################################################################

# AWS region to deploy all resources.
TF_VAR_region=us-west-2

# Tags applied to AWS resources (useful for cost accounting).
TF_VAR_app=storage
TF_VAR_owner=storacha
TF_VAR_team=Storacha Engineer
TF_VAR_org=Storacha

# Curio integration ############################################################

TF_VAR_use_pdp=false
TF_VAR_pdp_proofset=0
TF_VAR_curio_url=

# External (S3 compatible) blob bucket #########################################

TF_VAR_use_external_blob_bucket=false
# API endpoint for the external bucket.
TF_VAR_external_blob_bucket_endpoint=
TF_VAR_external_blob_bucket_region=
TF_VAR_external_blob_bucket_name=
# Public domain for accessing bucket.
TF_VAR_external_blob_bucket_domain=
TF_VAR_external_blob_bucket_access_key_id=
TF_VAR_external_blob_bucket_secret_access_key=

# Indexing service configuration ###############################################

TF_VAR_indexing_service_did=did:web:indexer.storacha.network
TF_VAR_indexing_service_url=https://indexer.storacha.network

# Debugging ####################################################################

# Setting this variable enables tracing for lambdas based on HTTP handlers.
# Currently, only Honeycomb is supported as the tracing backend. You can create
# a Honeycomb account and get an API key from honeycomb.io.
TF_VAR_honeycomb_api_key=
