# Deployment

Deployment of a Storage Node to AWS is managed by Terraform which you can invoke from the project root with `make`.

First, install OpenTofu e.g.

```sh
brew install opentofu
```

or for Linux distributions that support Snap:

```sh
snap install --classic opentofu
```

...for other Operating Systems see: https://opentofu.org/docs/intro/install

### AWS settings

The Terraform configuration will fetch AWS settings (such as credentials and the region to deploy resources to) from your local AWS configuration. Although an installation of the AWS CLI is not strictly required, it can be a convenient way to manage these settings.

OpenTofu will go to the same places as the AWS CLI to find settings, which means it will read environment variables such as `AWS_REGION` and `AWS_PROFILE` and the `~/.aws/config` and `~/.aws/credentials` files.

Make sure you are using the correct AWS profile and region before invoking `make` targets.

### `.env`

You need to first create a `.env` with relevant vars. Copy `.env.tpl` to `.env`. Set required variables, and any optional variables you want to set. Explanations for all variables can be found in the template.

### Deployment commands

Note that these commands will call needed prerequisites -- `make apply` will essentially do all of these start to finish.

#### `make lambdas`

This will simply compile the lambdas locally and put then in the `build` directory.

#### `make init`

You should only need to run this once -- initializes your terraform deployment and workspace. Make sure you've set `TF_WORKSPACE` first!

If the `make init` fails you will need to execute `tofu init` directly from the `deploy/app` folder to install the required dependencies, and it will update the `.terraform.lock.hcl` file if needed.

#### `make validate`

This will validate your terraform configuration -- good to run to check errors in any changes you make to terraform configs.

#### `make plan`

This will plan a deployment, but not execute it -- useful to see ahead what changes will happen when you run the next deployment.

#### `make apply`

The big kahuna! This will deploy all of your changes, including redeploying lambdas if any of code changes.
