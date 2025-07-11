package pkg

import (
	tfjson "github.com/hashicorp/terraform-json"
	alicloud "github.com/lonegunmanb/terraform-alicloud-schema/generated"
	aws_v2 "github.com/lonegunmanb/terraform-aws-schema/v2/generated"
	aws_v3 "github.com/lonegunmanb/terraform-aws-schema/v3/generated"
	aws_v4 "github.com/lonegunmanb/terraform-aws-schema/v4/generated"
	aws_v5 "github.com/lonegunmanb/terraform-aws-schema/v5/generated"
	aws_v6 "github.com/lonegunmanb/terraform-aws-schema/v6/generated"
	awscc "github.com/lonegunmanb/terraform-awscc-schema/generated"
	azapi "github.com/lonegunmanb/terraform-azapi-schema/generated"
	azapi_v2 "github.com/lonegunmanb/terraform-azapi-schema/v2/generated"
	azuread_v2 "github.com/lonegunmanb/terraform-azuread-schema/v2/generated"
	azuread_v3 "github.com/lonegunmanb/terraform-azuread-schema/v3/generated"
	azure_v2 "github.com/lonegunmanb/terraform-azurerm-schema/v2/generated"
	azure_v3 "github.com/lonegunmanb/terraform-azurerm-schema/v3/generated"
	azure_v4 "github.com/lonegunmanb/terraform-azurerm-schema/v4/generated"
	bytebase "github.com/lonegunmanb/terraform-bytebase-schema/generated"
	google_v2 "github.com/lonegunmanb/terraform-google-schema/v2/generated"
	google_v3 "github.com/lonegunmanb/terraform-google-schema/v3/generated"
	google_v4 "github.com/lonegunmanb/terraform-google-schema/v4/generated"
	google_v5 "github.com/lonegunmanb/terraform-google-schema/v5/generated"
	google_v6 "github.com/lonegunmanb/terraform-google-schema/v6/generated"
	helm_v2 "github.com/lonegunmanb/terraform-helm-schema/v2/generated"
	helm_v3 "github.com/lonegunmanb/terraform-helm-schema/v3/generated"
	kubernetes_v2 "github.com/lonegunmanb/terraform-kubernetes-schema/v2/generated"
	local "github.com/lonegunmanb/terraform-local-schema/v2/generated"
	modtm "github.com/lonegunmanb/terraform-modtm-schema/generated"
	null "github.com/lonegunmanb/terraform-null-schema/v3/generated"
	random "github.com/lonegunmanb/terraform-random-schema/v3/generated"
	template "github.com/lonegunmanb/terraform-template-schema/v2/generated"
	time "github.com/lonegunmanb/terraform-time-schema/generated"
	tls "github.com/lonegunmanb/terraform-tls-schema/v4/generated"
)

var resourceSchemas = make(map[string]*tfjson.Schema, 0)

func init() {
	resources := []map[string]*tfjson.Schema{
		alicloud.Resources,
		azure_v2.Resources,
		azure_v3.Resources,
		azure_v4.Resources,
		azuread_v2.Resources,
		azuread_v3.Resources,
		azapi.Resources,
		azapi_v2.Resources,
		awscc.Resources,
		aws_v2.Resources,
		aws_v3.Resources,
		aws_v4.Resources,
		aws_v5.Resources,
		aws_v6.Resources,
		bytebase.Resources,
		google_v2.Resources,
		google_v3.Resources,
		google_v4.Resources,
		google_v5.Resources,
		google_v6.Resources,
		helm_v2.Resources,
		helm_v3.Resources,
		kubernetes_v2.Resources,
		local.Resources,
		null.Resources,
		random.Resources,
		template.Resources,
		time.Resources,
		tls.Resources,
		modtm.Resources,
	}
	for _, schemas := range resources {
		mergeSchemas(resourceSchemas, schemas)
	}
}

func mergeSchemas(s1, s2 map[string]*tfjson.Schema) {
	for k, v := range s2 {
		s1[k] = v
	}
}
