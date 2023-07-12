package pkg

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var aksMarkdown string
var azureContainerAppMarkdown string
var awsEksMarkdown string
var googleMarkdown string

func init() {
	aksMarkdown = readFileAsString("azurerm_kubernetes_cluster.markdown")
	azureContainerAppMarkdown = readFileAsString("azurerm_container_app.markdown")
	awsEksMarkdown = readFileAsString("aws_eks_cluster.markdown")
	googleMarkdown = readFileAsString("google_container_cluster.markdown")
}

func readFileAsString(path string) string {
	c, _ := os.ReadFile(path)
	return string(c)
}

func TestDocumentParse_RootArgumentDescription(t *testing.T) {
	cases := []struct {
		resourceType string
		expected     string
		document     string
		path         string
	}{
		{
			resourceType: "azurerm_container_app",
			document:     azureContainerAppMarkdown,
			path:         "container_app_environment_id",
			expected:     `(Required) The ID of the Container App Environment within which this Container App should exist. Changing this forces a new resource to be created.`,
		},
		{
			resourceType: "azurerm_kubernetes_cluster",
			document:     aksMarkdown,
			path:         "name",
			expected:     `(Required) The name of the Managed Kubernetes Cluster to create. Changing this forces a new resource to be created.`,
		},
		{
			resourceType: "aws_eks_cluster",
			document:     awsEksMarkdown,
			path:         "name",
			expected:     "(Required) Name of the cluster. Must be between 1-100 characters in length. Must begin with an alphanumeric character, and must only contain alphanumeric characters, dashes and underscores (`^[0-9A-Za-z][A-Za-z0-9\\-_]+$`).",
		},
		{
			resourceType: "google_container_cluster",
			document:     googleMarkdown,
			path:         "name",
			expected:     "(Required) The name of the cluster, unique within the project and location.",
		},
		{
			resourceType: "google_container_cluster",
			document:     googleMarkdown,
			path:         "vertical_pod_autoscaling",
			expected:     "(Optional) Vertical Pod Autoscaling automatically adjusts the resources of pods controlled by it. Structure is [documented below](#nested_vertical_pod_autoscaling).",
		},
	}
	for i := 0; i < len(cases); i++ {
		c := cases[i]
		t.Run(c.resourceType, func(t *testing.T) {
			d := newDocument(c.resourceType)
			d.getContent = doc(c.document)
			args, err := d.parseDocument()
			actual := args[c.path]
			require.NoError(t, err)
			assert.Equal(t, c.expected, actual.desc)
		})
	}
}

func TestDocumentParse_NestedBlockArgumentDescription(t *testing.T) {
	cases := []struct {
		resourceType string
		expected     string
		document     string
		path         string
	}{
		{
			resourceType: "azurerm_container_app",
			document:     azureContainerAppMarkdown,
			path:         "secret.name",
			expected:     `(Required) The Secret name.`,
		},
		{
			resourceType: "azurerm_kubernetes_cluster",
			document:     aksMarkdown,
			path:         "aci_connector_linux.subnet_name",
			expected:     `(Required) The subnet name for the virtual nodes to run.`,
		},
		{
			resourceType: "azurerm_kubernetes_cluster",
			document:     aksMarkdown,
			path:         "api_server_access_profile.authorized_ip_ranges",
			expected:     `(Optional) Set of authorized IP ranges to allow access to API server, e.g. ["198.51.100.0/24"].`,
		},
		{
			resourceType: "azurerm_kubernetes_cluster",
			document:     aksMarkdown,
			path:         "linux_os_config.swap_file_size_mb",
			expected:     `(Optional) Specifies the size of the swap file on each node in MB. Changing this forces a new resource to be created.`,
		},
		{
			resourceType: "aws_eks_cluster",
			document:     awsEksMarkdown,
			path:         "encryption_config.resources",
			expected:     "(Required) List of strings with resources to be encrypted. Valid values: `secrets`.",
		},
		{
			resourceType: "aws_eks_cluster",
			document:     awsEksMarkdown,
			path:         "provider.key_arn",
			expected:     "(Required) ARN of the Key Management Service (KMS) customer master key (CMK). The CMK must be symmetric, created in the same region as the cluster, and if the CMK was created in a different account, the user must have access to the CMK. For more information, see [Allowing Users in Other Accounts to Use a CMK in the AWS Key Management Service Developer Guide](https://docs.aws.amazon.com/kms/latest/developerguide/key-policy-modifying-external-accounts.html).",
		},
		{
			resourceType: "google_container_cluster",
			document:     googleMarkdown,
			path:         "addons_config.horizontal_pod_autoscaling",
			expected:     "(Optional) The status of the Horizontal Pod Autoscaling addon, which increases or decreases the number of replica pods a replication controller has based on the resource usage of the existing pods. It is enabled by default; set `disabled = true` to disable.",
		},
	}
	for i := 0; i < len(cases); i++ {
		c := cases[i]
		t.Run(fmt.Sprintf("%s.%s", c.resourceType, c.path), func(t *testing.T) {
			d := newDocument(c.resourceType)
			d.getContent = doc(c.document)
			args, err := d.parseDocument()
			actual := args[c.path]
			require.NoError(t, err)
			assert.Equal(t, c.expected, actual.desc)
		})
	}
}

func TestDocumentParse_Timeouts(t *testing.T) {
	cases := []struct {
		resourceType string
		document     string
	}{
		{
			resourceType: "azurerm_container_app",
			document:     azureContainerAppMarkdown,
		},
		{
			resourceType: "azurerm_kubernetes_cluster",
			document:     aksMarkdown,
		},
		{
			resourceType: "aws_eks_cluster",
			document:     awsEksMarkdown,
		},
		{
			resourceType: "google_container_cluster",
			document:     googleMarkdown,
		},
	}
	for i := 0; i < len(cases); i++ {
		c := cases[i]
		t.Run(c.resourceType, func(t *testing.T) {
			d := newDocument(c.resourceType)
			d.getContent = doc(c.document)
			args, err := d.parseDocument()
			actual := args["timeouts.create"]
			require.NoError(t, err)
			assert.NotEqual(t, "", actual.desc)
		})
	}
}

func doc(d string) func(string) (string, error) {
	return func(string) (string, error) {
		return d, nil
	}
}
