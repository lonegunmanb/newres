# `newres` - Terraform Resource Generation Tool

![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/lonegunmanb/newres/build.yml)

`newres` is a command-line tool that generates Terraform configuration files for a specified resource type. It automates the process of creating variables.tf and main.tf files, making it easier to get started with Terraform and reducing the time spent on manual configuration.

## Features

Supports multiple Terraform providers, including AWS, Azure, Google Cloud Platform, and more.

Generates `variables.tf` and `main.tf` files based on the specified resource type.

Supports two different generation modes: `UniVariable` and `MultipleVariables`:
* `MultipleVariables` (default): Generates separate variable blocks for each attribute and nested block of the resource.
* `UniVariable`: Generates a single variable block for the entire resource with nested blocks as attributes.

To use `newres`, you'll need to have Go installed and build the tool using the provided source code:

```shell
go install github.com/lonegunmanb/newres/v3@latest
```

Once you've built the tool, you can use it with the following command:

```shell
newres -dir [DIRECTORY] [-u] [-r RESOURCE_TYPE]
```

* `-dir [DIRECTORY]`: Required. The directory path where the generated files will be stored.
* `-r RESOURCE_TYPE`: Required. The resource type to generate configuration for (e.g., `aws_instance`, `azurerm_virtual_machine`, `google_compute_instance`).
* `-u`: Optional. If set, the tool will generate the resource configuration in UniVariable mode. If not set, MultipleVariables mode will be used by default.

For example, to generate configuration files for an Azure resource group in the current working directory, you would run:

```shell
newres -dir ./ -r azurerm_resource_group
```

The result looks like:

```hcl
variable "resource_group_location" {
  type        = string
  description = "(Required) The Azure Region where the Resource Group should exist. Changing this forces a new Resource Group to be created."
  nullable    = false
}

variable "resource_group_name" {
  type        = string
  description = "(Required) The Name which should be used for this Resource Group. Changing this forces a new Resource Group to be created."
  nullable    = false
}

variable "resource_group_tags" {
  type        = map(string)
  default     = null
  description = "(Optional) A mapping of tags which should be assigned to the Resource Group."
}

variable "resource_group_timeouts" {
  type = object({
    create = optional(string)
    delete = optional(string)
    read   = optional(string)
    update = optional(string)
  })
  default     = null
  description = <<-EOT
 - `create` - (Defaults to 1 hour and 30 minutes) Used when creating the Resource Group.
 - `delete` - (Defaults to 1 hour and 30 minutes) Used when deleting the Resource Group.
 - `read` - (Defaults to 5 minutes) Used when retrieving the Resource Group.
 - `update` - (Defaults to 1 hour and 30 minutes) Used when updating the Resource Group.
EOT
}

resource "azurerm_resource_group" "this" {
  location = var.resource_group_location
  name     = var.resource_group_name
  tags     = var.resource_group_tags

  dynamic "timeouts" {
    for_each = var.resource_group_timeouts == null ? [] : var.resource_group_timeouts
    content {
      create = timeouts.value.create
      delete = timeouts.value.delete
      read   = timeouts.value.read
      update = timeouts.value.update
    }
  }
}
```

After running the command, you should find `variables.tf` and `main.tf` files in the specified directory, containing the generated Terraform configuration for the specified resource type.

**Note**: You can run the command multiple times with different resource types, and the newly added resource blocks and variable blocks will be appended to the existing `main.tf` and `variables.tf` files, allowing you to easily expand your Terraform configuration without manual editing.

## Limitations

### Sometimes optional attributes might be required

`newres` has a known limitation when dealing with certain nested blocks in the Terraform plugin SDK. In some cases, a nested block may be marked as an attribute instead of a nested block, as shown in this example: https://github.com/hashicorp/terraform-provider-azurerm/blob/v3.62.1/internal/services/recoveryservices/site_recovery_replicated_vm_resource.go#L182-L187.

When this occurs, the JSON schema returned by the Terraform CLI will treat these nested blocks as attributes, and `newres` will try to restore these "attributes" back to nested blocks. Consequently, all attributes inside these affected nested blocks will be marked as required in the generated configuration, as the corresponding schema information is lost in the process.

Please be aware of this limitation when using `newres` and ensure to double-check the generated configuration files for accuracy, especially when dealing with resources that exhibit this behavior.

### Required nested block would also be generated as `dynamic` block

To simplify the implementation, now required nested block would be generated like:

```hcl
dynamic "default_node_pool" {
  for_each = [var.kubernetes_cluster_default_node_pool]
  content {
    name                          = default_node_pool.value.name
    vm_size                       = default_node_pool.value.vm_size
  }
}
```

Even `var.kubernetes_cluster_default_node_pool` is a required object, that's because considering there could be required nested block inside a required nested block. This `dynamic` block implementation could simplify the iterator to `<block_name>.value.<attribute_name>`.

# Supported Providers and Documentation Limitations

`newres` currently supports variable block description generation for the following providers:
* AWS (`aws`)
* Azure Resource Manager (`azurerm`)
* Azure Active Directory (`azuread`)
* Google Cloud Platform (`google`)

Please note that there is no unified and strict rule for provider documentation. As a result, `newres` may not always parse the documentation correctly for all providers and resources. This tool is designed to help automate the generation of Terraform configuration files, but it is still essential to review the generated files for accuracy.

We are not planning to spend significant time on improving documentation parsing for every provider and resource. However, if you encounter issues or have suggestions, please feel free to open a pull request or submit an issue on the project's GitHub repository, and we will consider addressing them on a case-by-case basis.

## Contributing

If you'd like to contribute to the project or report any issues, please feel free to open a pull request or submit an issue on the project's GitHub repository.

## License

`newres` is released under the MIT License.