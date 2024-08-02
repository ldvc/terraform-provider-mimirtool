package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"mimirtool": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
}

/// OLD


// // testAccProviderFactories is a static map containing only the main provider instance
// var testAccProviderFactories map[string]func() (*schema.Provider, error)

// // testAccProvider is the "main" provider instance
// //
// // This Provider can be used in testing code for API calls without requiring
// // the use of saving and referencing specific ProviderFactories instances.
// //
// // testAccPreCheck(t) must be called before using this provider instance.
// var testAccProvider *schema.Provider

// var testAccProviders map[string]*schema.Provider

// // testAccProviderConfigure ensures testAccProvider is only configured once
// //
// // The testAccPreCheck(t) function is invoked for every test and this prevents
// // extraneous reconfiguration to the same values each time. However, this does
// // not prevent reconfiguration that may happen should the address of
// // testAccProvider be errantly reused in ProviderFactories.
// var testAccProviderConfigure sync.Once

// func init() {
// 	testAccProvider = New("dev")()
// 	testAccProviders = map[string]*schema.Provider{
// 		"mimirtool": New("dev")(),
// 	}

// 	// Always allocate a new provider instance each invocation, otherwise gRPC
// 	// ProviderConfigure() can overwrite configuration during concurrent testing.
// 	testAccProviderFactories = map[string]func() (*schema.Provider, error){
// 		"mimirtool": func() (*schema.Provider, error) {
// 			return New("dev")(), nil
// 		},
// 	}
// }

// func TestProvider(t *testing.T) {
// 	if err := New("dev")().InternalValidate(); err != nil {
// 		t.Fatalf("err: %s", err)
// 	}
// }

// // testAccPreCheck verifies required provider testing configuration. It should
// // be present in every acceptance test.
// //
// // These verifications and configuration are preferred at this level to prevent
// // provider developers from experiencing less clear errors for every test.
// func testAccPreCheck(t *testing.T) {
// 	testAccProviderConfigure.Do(func() {
// 		// Since we are outside the scope of the Terraform configuration we must
// 		// call Configure() to properly initialize the provider configuration.
// 		err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 	})
// }
