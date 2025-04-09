/*
  This block defines the file structure and service setup for deploying multiple components (e.g., Lotus, Curio, and Yugabyte).
  - global_files: Files always installed, regardless of service (e.g., environment scripts, wallet files).
  - services: Per-service definitions that specify:
      1. script_files: Shell scripts with optional template variables.
      2. service_files: Systemd .service units for each component.
      3. config_files: Configuration TOML or similar files for the service.
  - file_categories: Categorizes file types to determine how they're handled and where they're found on disk.
  - write_files: Dynamically builds a list of files to write based on the data in `services` plus any global files.
  - all_service_filenames: A flattened list of all .service files used for enabling and starting the services at once.
  - systemd_enable_cmd / systemd_start_cmd: Single commands to enable/start all services, letting systemd manage dependency order.
  - runcmd_steps: Additional commands to run on instance startup (kernel tweaks, installing each component, etc.), then systemd reload and service enable/start.
  - cloud_init: Final rendered user data (writes files, runs commands) for the cloud-init process.

  Integration with cloud-init:
    - The `cloud_init` variable is a rendered template that follows the standard cloud-config format (see `cloud-init.yaml.tpl`).
    - The lists `write_files` and `runcmd_steps` are inserted into that template, forming the final user-data passed to the instance at boot.
    - Cloud-init processes `write_files` to place scripts/configs onto the system with specified permissions, and then executes `runcmd_steps` in order.

  To add a new service:
    1. Create a key under `services` (e.g., "myservice") with `script_files`, `service_files`, and optionally `config_files`.
    2. Place actual scripts in the specified directory (scripts/) and service units in services/ if needed.
    3. Make sure to fill out `filename`, `target_path`, `permissions`, and (if needed) `is_template`/`vars`.
    4. Terraform automatically picks them up in `write_files` and will install/enable them on the instance.
*/

locals {
  scripts_dir = "${path.module}/scripts"
  services_dir = "${path.module}/services"
  configs_dir = "${path.module}/configs"

  global_files = [
    {
      path = "/usr/local/bin/install_rust.sh"
      permissions = "0755"
      content     = file("${local.scripts_dir}/install-rust.sh")
    },
    {
      path = "/usr/local/bin/install_go.sh"
      permissions = "0755"
      content     = templatefile("${local.scripts_dir}/install-go.sh", {
        GO_VERSION = var.go_version
      })
    },
    {
      path = "/usr/local/bin/service-ready.sh"
      permissions = "0755"
      content     = file("${local.scripts_dir}/service-ready.sh")
    },
    {
      path  = "/opt/lotus_wallet_bls.json"
      permissions = "0400"
      content = file(var.lotus_wallet_bls_file)
    },
    {
      path  = "/opt/lotus_wallet_delegated.json"
      permissions = "0400"
      content = file(var.lotus_wallet_delegated_file)
    },
    {
      path  = "/opt/service.pem"
      permissions = "0400"
      content = file(var.curio_service_pem_key_file)
    }
  ]

  services = {
    "lotus" = {
      script_files = [
        {
          filename    = "install-lotus.sh"
          target_path = "/usr/local/bin/install-lotus.sh"
          permissions = "0755"
          is_template = true
          vars        = {
            LOTUS_VERSION      = var.lotus_version
            LOTUS_BUILD_TARGET = var.filecoin_network
          }
        },
        {
          filename    = "lotus-import-snapshot.sh"
          target_path = "/usr/local/bin/lotus-import-snapshot.sh"
          permissions = "0755"
          is_template = true
          vars        = {
            LOTUS_SNAPSHOT_URL = var.lotus_snapshot_url
          }
        },
        {
          filename    = "lotus-import-wallets.sh"
          target_path = "/usr/local/bin/lotus-import-wallets.sh"
          permissions = "0755"
          is_template = false
          vars        = {}
        }
      ]
      service_files = [
        {
          filename    = "lotus.service"
          target_path = "/etc/systemd/system/lotus.service"
          permissions = "0644"
        },
        {
          filename    = "lotus-prestart.service"
          target_path = "/etc/systemd/system/lotus-prestart.service"
          permissions = "0644"
        },
        {
          filename    = "lotus-ready.service"
          target_path = "/etc/systemd/system/lotus-ready.service"
          permissions = "0644"
        },
        {
          filename    = "lotus-poststart.service"
          target_path = "/etc/systemd/system/lotus-poststart.service"
          permissions = "0644"
        }
      ]
    }
    "curio" = {
      script_files = [
        {
          filename    = "install-curio.sh"
          target_path = "/usr/local/bin/install-curio.sh"
          permissions = "0755"
          is_template = true
          vars        = {
            CURIO_VERSION      = var.curio_version
            CURIO_BUILD_TARGET = var.filecoin_network
          }
        },
        {
          filename    = "curio-poststart.sh"
          target_path = "/usr/local/bin/curio-poststart.sh"
          permissions = "0755"
          is_template = false
          vars        = {}
        }
      ]
      service_files = [
        {
          filename    = "curio.service"
          target_path = "/etc/systemd/system/curio.service"
          permissions = "0644"
        },
        {
          filename    = "curio-prestart.service"
          target_path = "/etc/systemd/system/curio-prestart.service"
          permissions = "0644"
        },
        {
          filename    = "curio-ready.service"
          target_path = "/etc/systemd/system/curio-ready.service"
          permissions = "0644"
        },
        {
          filename    = "curio-poststart.service"
          target_path = "/etc/systemd/system/curio-poststart.service"
          permissions = "0644"
        }
      ]
      config_files = [
         {
           filename    = "curio-pdp.toml"
           target_path = "/opt/curio-pdp.toml"
           permissions = "0400"
           is_template = true
           vars        = {
             CURIO_DOMAIN_NAME = "${var.app}.${var.domain}"
           }
         },
        {
          filename    = "curio-storage.toml"
          target_path = "/opt/curio-storage.toml"
          permissions = "0400"
          is_template = false
          vars        = {}
        },
      ]
    }
    "yugabyte" = {
      script_files = [
        {
          filename    = "install-yugabyte.sh"
          target_path = "/usr/local/bin/install-yugabyte.sh"
          permissions = "0755"
          is_template = true
          vars        = {
            YUGABYTE_VERSION = var.yugabyte_version
          }
        }
      ]
      service_files = [
        {
          filename    = "yugabyte.service"
          target_path = "/etc/systemd/system/yugabyte.service"
          permissions = "0644"
        },
        {
          filename    = "yugabyte-ready.service"
          target_path = "/etc/systemd/system/yugabyte-ready.service"
          permissions = "0644"
        }
      ]
    }
  }

  file_categories = [
    { key = "script_files",  base_dir = local.scripts_dir, default_template = false },
    { key = "service_files", base_dir = local.services_dir, default_template = false },
    { key = "config_files",  base_dir = local.configs_dir,  default_template = false },
  ]

  write_files = flatten([
    for svc_name, svc_def in local.services : flatten([
      for cat in local.file_categories : (
        try(svc_def[cat.key], []) != [] ? [for f in svc_def[cat.key] : {
          path        = f.target_path
          permissions = f.permissions
          content     = lookup(f, "is_template", cat.default_template) ? templatefile("${cat.base_dir}/${f.filename}", f.vars) : file("${cat.base_dir}/${f.filename}")
        }] : []
      )
    ])
  ])

  # Flatten all .service filenames into a single list
  all_service_filenames = flatten([
    for svc_name, svc_def in local.services : [
      for uf in svc_def.service_files : basename(uf.filename)
    ]
  ])

  # build one string for systemctl enable and one for systemctl start by joining all the service filenames in a single line.
  # enabling and starting everything with a single command lets systemd manage the order in which services start based
  # on their dependencies: `After` and `Want` units
  systemd_enable_cmd = "systemctl enable ${join(" ", local.all_service_filenames)}"
  systemd_start_cmd  = "systemctl start ${join(" ", local.all_service_filenames)}"


  runcmd_steps = flatten([
    [
      # NB: https://docs.curiostorage.org/installation#system-configuration
      ["sysctl -w net.core.rmem_max=2097152"],
      ["sysctl -w net.core.rmem_default=2097152"],
      ["/usr/local/bin/install-lotus.sh"],
      ["/usr/local/bin/install-curio.sh"],
      ["/usr/local/bin/install-yugabyte.sh" ],
      # NB: required by curio-prestart.service
      ["echo ADDRESS=$(cat /opt/lotus_wallet_bls.json | jq -r .Address) > /run/curio-address.env"],
      ["systemctl daemon-reload"]
    ],
    local.systemd_enable_cmd,
    local.systemd_start_cmd,
  ])

  # Merge global files with all per-service files
  all_write_files = concat(local.global_files, local.write_files)

  # Build our final user data by passing write_files into the template
  cloud_init = templatefile("${path.module}/cloud-init.yaml.tpl", {
    write_files = local.all_write_files
    runcmd_steps = local.runcmd_steps
  })
}