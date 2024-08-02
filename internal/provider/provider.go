package provider

import (
	"context"
	"os"
	"strconv"

	"github.com/grafana/dskit/crypto/tls"
	mimirtool "github.com/grafana/mimir/pkg/mimirtool/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure mimirtoolProvider satisfies various provider interfaces.
var _ provider.Provider = &mimirtoolProvider{}

// TODO: check if mandatory?
var _ provider.ProviderWithFunctions = &mimirtoolProvider{}

// mimirtoolProvider defines the provider implementation.
type mimirtoolProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *mimirtoolProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mimirtool"
	resp.Version = p.version
}

func (p *mimirtoolProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A super cool description",
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				MarkdownDescription: "Address to use when contacting Grafana Mimir. May alternatively be set via the `MIMIRTOOL_ADDRESS` or `MIMIR_ADDRESS` environment variable.",
				Optional:            true,
				// Validators: []validator.String{
				// 	attribute_validator.UrlWithScheme(supportedProxySchemesStr()...),
				// 	stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("from_env")),
				// },
			},
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "Tenant ID to use when contacting Grafana Mimir. May alternatively be set via the `MIMIRTOOL_TENANT_ID` or `MIMIR_TENANT_ID` environment variable.",
				Optional:            true,
			},
			"api_user": schema.StringAttribute{
				MarkdownDescription: "API user to use when contacting Grafana Mimir. May alternatively be set via the `MIMIRTOOL_API_USER` or `MIMIR_API_USER` environment variable.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key to use when contacting Grafana Mimir. May alternatively be set via the `MIMIRTOOL_API_KEY` or `MIMIR_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"auth_token": schema.StringAttribute{
				MarkdownDescription: "Authentication token for bearer token or JWT auth when contacting Grafana Mimir. May alternatively be set via the `MIMIRTOOL_AUTH_TOKEN` or `MIMIR_AUTH_TOKEN` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"tls_key_path": schema.StringAttribute{
				MarkdownDescription: "Client TLS key file to use to authenticate to the MIMIR server. May alternatively be set via the `MIMIRTOOL_TLS_KEY_PATH` or `MIMIR_TLS_KEY_PATH` environment variable.",
				Optional:            true,
			},
			"tls_cert_path": schema.StringAttribute{
				MarkdownDescription: "Client TLS certificate file to use to authenticate to the MIMIR server. May alternatively be set via the `MIMIRTOOL_TLS_CERT_PATH` or `MIMIR_TLS_CERT_PATH` environment variable.",
				Optional:            true,
			},
			"tls_ca_path": schema.StringAttribute{
				MarkdownDescription: "Certificate CA bundle to use to verify the MIMIR server's certificate. May alternatively be set via the `MIMIRTOOL_TLS_CA_PATH` or `MIMIR_TLS_CA_PATH` environment variable.",
				Optional:            true,
			},
			"insecure_skip_verify": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS certificate verification. May alternatively be set via the `MIMIRTOOL_INSECURE_SKIP_VERIFY` or `MIMIR_INSECURE_SKIP_VERIFY` environment variable.",
				Optional:            true,
			},
			"prometheus_http_prefix": schema.StringAttribute{
				MarkdownDescription: "Path prefix to use for rules. May alternatively be set via the `MIMIRTOOL_PROMETHEUS_HTTP_PREFIX` or `MIMIR_PROMETHEUS_HTTP_PREFIX` environment variable.",
				Optional:            true,
			},
			"alertmanager_http_prefix": schema.StringAttribute{
				MarkdownDescription: "Path prefix to use for alertmanager. May alternatively be set via the `MIMIRTOOL_ALERTMANAGER_HTTP_PREFIX` or `MIMIR_ALERTMANAGER_HTTP_PREFIX` environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *mimirtoolProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Debug(ctx, "Configuring provider")
	var providerConfig alertmanagerProviderModelV0

	diags := req.Config.Get(ctx, &providerConfig)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if providerConfig.Address.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("address"),
			"Unknown HashiCups API Host",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_HOST environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	address := os.Getenv("MIMIRTOOL_ADDRESS")
	if address == "" {
		address = os.Getenv("MIMIR_ADDRESS")
	}

	tenantID := os.Getenv("MIMIRTOOL_TENANT_ID")
	if address == "" {
		tenantID = os.Getenv("MIMIR_TENANT_ID")
	}

	apiUser := os.Getenv("MIMIRTOOL_API_USER")
	if apiUser == "" {
		apiUser = os.Getenv("MIMIR_API_USER")
	}

	apiKey := os.Getenv("MIMIRTOOL_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("MIMIR_API_KEY")
	}

	authToken := os.Getenv("MIMIRTOOL_AUTH_TOKEN")
	if authToken == "" {
		authToken = os.Getenv("MIMIR_AUTH_TOKEN")
	}

	tlsKeyPath := os.Getenv("MIMIRTOOL_TLS_KEY_PATH")
	if tlsKeyPath == "" {
		tlsKeyPath = os.Getenv("MIMIR_TLS_KEY_PATH")
	}

	tlsCertPath := os.Getenv("MIMIRTOOL_TLS_CERT_PATH")
	if tlsCertPath == "" {
		tlsCertPath = os.Getenv("MIMIR_TLS_CERT_PATH")
	}

	tlsCAPath := os.Getenv("MIMIRTOOL_TLS_CA_PATH")
	if tlsCAPath == "" {
		tlsCAPath = os.Getenv("MIMIR_TLS_CA_PATH")
	}

	insecureSkipVerify := os.Getenv("MIMIRTOOL_INSECURE_SKIP_VERIFY")
	if insecureSkipVerify == "" {
		insecureSkipVerify = os.Getenv("MIMIR_INSECURE_SKIP_VERIFY")
	}

	prometheusHTTPPrefix := os.Getenv("MIMIRTOOL_PROMETHEUS_HTTP_PREFIX")
	if prometheusHTTPPrefix == "" {
		prometheusHTTPPrefix = os.Getenv("MIMIR_PROMETHEUS_HTTP_PREFIX")
	}

	alertmanagerHTTPPrefix := os.Getenv("MIMIRTOOL_ALERTMANAGER_HTTP_PREFIX")
	if alertmanagerHTTPPrefix == "" {
		alertmanagerHTTPPrefix = os.Getenv("MIMIR_ALERTMANAGER_HTTP_PREFIX")
	}

	// blop
	if !providerConfig.Address.IsNull() {
		address = providerConfig.Address.ValueString()
	}

	if !providerConfig.Tenant_id.IsNull() {
		tenantID = providerConfig.Tenant_id.ValueString()
	}

	if !providerConfig.Api_user.IsNull() {
		apiUser = providerConfig.Api_user.ValueString()
	}

	if !providerConfig.Api_key.IsNull() {
		apiKey = providerConfig.Api_key.ValueString()
	}

	if !providerConfig.Auth_token.IsNull() {
		authToken = providerConfig.Auth_token.ValueString()
	}

	if !providerConfig.Tls_key_path.IsNull() {
		tlsKeyPath = providerConfig.Tls_key_path.ValueString()
	}

	if !providerConfig.Tls_cert_path.IsNull() {
		tlsCertPath = providerConfig.Tls_cert_path.ValueString()
	}

	if !providerConfig.Tls_ca_path.IsNull() {
		tlsCAPath = providerConfig.Tls_ca_path.ValueString()
	}
	insecureSkipVerifyBool, err := strconv.ParseBool(insecureSkipVerify)
	if !providerConfig.Insecure_skip_verify.IsNull() {
		insecureSkipVerifyBool = providerConfig.Insecure_skip_verify.ValueBool()
	}

	if !providerConfig.Prometheus_http_prefix.IsNull() {
		prometheusHTTPPrefix = providerConfig.Prometheus_http_prefix.ValueString()
	}

	if !providerConfig.Alertmanager_http_prefix.IsNull() {
		alertmanagerHTTPPrefix = providerConfig.Alertmanager_http_prefix.ValueString()
	}

	if address == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("address"),
			"Missing Mimir URL",
			"The provider cannot create the Mimirtool API client as there is a missing or empty value for the Mimirtool API address. "+
				"Set the address value in the configuration or use the MIMIRTOOL_ADDRESS (or MIMIR_ADDRESS) environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "address", address)
	ctx = tflog.SetField(ctx, "tenant_id", tenantID)
	ctx = tflog.SetField(ctx, "auth_token", authToken)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "auth_token")

	tflog.Debug(ctx, "Creating Mimirtool client")

	// Create a new Mimirtool client using the configuration values
	// TODO: verify how to set user-agent
	// https://pkg.go.dev/github.com/grafana/mimir/pkg/mimirtool/client#UserAgent
	// mimirtool.UserAgent()

	client, err := mimirtool.New(mimirtool.Config{
		AuthToken: authToken,
		User:      apiUser,
		Key:       apiKey,
		Address:   address,
		ID:        tenantID,
		TLS: tls.ClientConfig{
			CAPath:             tlsCAPath,
			CertPath:           tlsCertPath,
			KeyPath:            tlsKeyPath,
			InsecureSkipVerify: insecureSkipVerifyBool,
		},
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Mimirtool API Client",
			"An unexpected error occurred when creating the Mimirtool API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Mimirtool Client Error: "+err.Error(),
		)
		return
	}

	// Make the HashiCups client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Mimirtool client", map[string]any{"success": true})
}

func (p *mimirtoolProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAlertmanagerResource,
	}
}

func (p *mimirtoolProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *mimirtoolProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &mimirtoolProvider{
			version: version,
		}
	}
}

type alertmanagerProviderModelV0 struct {
	Address                  types.String `tfsdk:"address"`
	Tenant_id                types.String `tfsdk:"tenant_id"`
	Api_user                 types.String `tfsdk:"api_user"`
	Api_key                  types.String `tfsdk:"api_key"`
	Auth_token               types.String `tfsdk:"auth_token"`
	Tls_key_path             types.String `tfsdk:"tls_key_path"`
	Tls_cert_path            types.String `tfsdk:"tls_cert_path"`
	Tls_ca_path              types.String `tfsdk:"tls_ca_path"`
	Insecure_skip_verify     types.Bool   `tfsdk:"insecure_skip_verify"`
	Prometheus_http_prefix   types.String `tfsdk:"prometheus_http_prefix"`
	Alertmanager_http_prefix types.String `tfsdk:"alertmanager_http_prefix"`
}

/// OLD CODE
// func init() {
// 	// Set descriptions to support markdown syntax, this will be used in document generation
// 	// and the language server.
// 	schema.DescriptionKind = schema.StringMarkdown

// 	// Customize the content of descriptions when output. For example you can add defaults on
// 	// to the exported descriptions if present.
// 	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
// 	// 	desc := s.Description
// 	// 	if s.Default != nil {
// 	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
// 	// 	}
// 	// 	return strings.TrimSpace(desc)
// 	// }
// }

// // New returns a newly created provider
// func New(version string) func() *schema.Provider {
// 	return func() *schema.Provider {
// 		p := &schema.Provider{
// 			Schema: map[string]*schema.Schema{
// 				// In order to allow users to use both terraform and mimirtool cli let's use the same envvar names
// 				// We shall accept two envvar name: one to respect terraform convention <provider>_<resource_name> and the other one from mimirtool.
// 				// terraform convention will be taken into account first.
// 				"address": {
// 					Type:         schema.TypeString,
// 					Required:     true,
// 					DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"MIMIRTOOL_ADDRESS", "MIMIR_ADDRESS"}, nil),
// 					Description:  "Address to use when contacting Grafana Mimir. May alternatively be set via the `MIMIRTOOL_ADDRESS` or `MIMIR_ADDRESS` environment variable.",
// 					ValidateFunc: validation.IsURLWithHTTPorHTTPS,
// 				},
// 				"tenant_id": {
// 					Type:        schema.TypeString,
// 					Optional:    true,
// 					DefaultFunc: schema.MultiEnvDefaultFunc([]string{"MIMIRTOOL_TENANT_ID", "MIMIR_TENANT_ID"}, nil),
// 					Description: "Tenant ID to use when contacting Grafana Mimir. May alternatively be set via the `MIMIRTOOL_TENANT_ID` or `MIMIR_TENANT_ID` environment variable.",
// 				},
// 				"api_user": {
// 					Type:        schema.TypeString,
// 					Optional:    true,
// 					DefaultFunc: schema.MultiEnvDefaultFunc([]string{"MIMIRTOOL_API_USER", "MIMIR_API_USER"}, nil),
// 					Description: "API user to use when contacting Grafana Mimir. May alternatively be set via the `MIMIRTOOL_API_USER` or `MIMIR_API_USER` environment variable.",
// 				},
// 				"api_key": {
// 					Type:        schema.TypeString,
// 					Optional:    true,
// 					Sensitive:   true,
// 					DefaultFunc: schema.MultiEnvDefaultFunc([]string{"MIMIRTOOL_API_KEY", "MIMIR_API_KEY"}, nil),
// 					Description: "API key to use when contacting Grafana Mimir. May alternatively be set via the `MIMIRTOOL_API_KEY` or `MIMIR_API_KEY` environment variable.",
// 				},
// 				"auth_token": {
// 					Type:        schema.TypeString,
// 					Optional:    true,
// 					Sensitive:   true,
// 					DefaultFunc: schema.MultiEnvDefaultFunc([]string{"MIMIRTOOL_AUTH_TOKEN", "MIMIR_AUTH_TOKEN"}, nil),
// 					Description: "Authentication token for bearer token or JWT auth when contacting Grafana Mimir. May alternatively be set via the `MIMIRTOOL_AUTH_TOKEN` or `MIMIR_AUTH_TOKEN` environment variable.",
// 				},
// 				"tls_key_path": {
// 					Type:        schema.TypeString,
// 					Optional:    true,
// 					DefaultFunc: schema.MultiEnvDefaultFunc([]string{"MIMIRTOOL_TLS_KEY_PATH", "MIMIR_TLS_KEY_PATH"}, nil),
// 					Description: "Client TLS key file to use to authenticate to the MIMIR server. May alternatively be set via the `MIMIRTOOL_TLS_KEY_PATH` or `MIMIR_TLS_KEY_PATH` environment variable.",
// 				},
// 				"tls_cert_path": {
// 					Type:        schema.TypeString,
// 					Optional:    true,
// 					DefaultFunc: schema.MultiEnvDefaultFunc([]string{"MIMIRTOOL_TLS_CERT_PATH", "MIMIR_TLS_CERT_PATH"}, nil),
// 					Description: "Client TLS certificate file to use to authenticate to the MIMIR server. May alternatively be set via the `MIMIRTOOL_TLS_CERT_PATH` or `MIMIR_TLS_CERT_PATH` environment variable.",
// 				},
// 				"tls_ca_path": {
// 					Type:        schema.TypeString,
// 					Optional:    true,
// 					DefaultFunc: schema.MultiEnvDefaultFunc([]string{"MIMIRTOOL_TLS_CA_PATH", "MIMIR_TLS_CA_PATH"}, nil),
// 					Description: "Certificate CA bundle to use to verify the MIMIR server's certificate. May alternatively be set via the `MIMIRTOOL_TLS_CA_PATH` or `MIMIR_TLS_CA_PATH` environment variable.",
// 				},
// 				"insecure_skip_verify": {
// 					Type:        schema.TypeBool,
// 					Optional:    true,
// 					DefaultFunc: schema.MultiEnvDefaultFunc([]string{"MIMIRTOOL_INSECURE_SKIP_VERIFY", "MIMIR_INSECURE_SKIP_VERIFY"}, nil),
// 					Description: "Skip TLS certificate verification. May alternatively be set via the `MIMIRTOOL_INSECURE_SKIP_VERIFY` or `MIMIR_INSECURE_SKIP_VERIFY` environment variable.",
// 				},
// 				"prometheus_http_prefix": {
// 					Type:        schema.TypeString,
// 					Optional:    true,
// 					DefaultFunc: schema.MultiEnvDefaultFunc([]string{"MIMIRTOOL_PROMETHEUS_HTTP_PREFIX", "MIMIR_PROMETHEUS_HTTP_PREFIX"}, "/prometheus"),
// 					Description: "Path prefix to use for rules. May alternatively be set via the `MIMIRTOOL_PROMETHEUS_HTTP_PREFIX` or `MIMIR_PROMETHEUS_HTTP_PREFIX` environment variable.",
// 				},
// 				"alertmanager_http_prefix": {
// 					Type:        schema.TypeString,
// 					Optional:    true,
// 					DefaultFunc: schema.MultiEnvDefaultFunc([]string{"MIMIRTOOL_ALERTMANAGER_HTTP_PREFIX", "MIMIR_ALERTMANAGER_HTTP_PREFIX"}, "/alertmanager"),
// 					Description: "Path prefix to use for alertmanager. May alternatively be set via the `MIMIRTOOL_ALERTMANAGER_HTTP_PREFIX` or `MIMIR_ALERTMANAGER_HTTP_PREFIX` environment variable.",
// 				},
// 			},
// 			DataSourcesMap: map[string]*schema.Resource{},
// 			ResourcesMap: map[string]*schema.Resource{
// 				"mimirtool_ruler_namespace": resourceRulerNamespace(),
// 				"mimirtool_alertmanager":    resourceAlertManager(),
// 			},
// 		}

// 		p.ConfigureContextFunc = configure(version, p)

// 		return p
// 	}
// }

// func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
// 	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
// 		var (
// 			diags diag.Diagnostics
// 			err   error
// 		)
// 		p.UserAgent("terraform-provider-mimirtool", version)

// 		c := &client{}

// 		c.cli, err = getDefaultMimirClient(d)
// 		if err != nil {
// 			return nil, diag.FromErr(err)
// 		}
// 		return c, diags
// 	}
// }

// func getDefaultMimirClient(d *schema.ResourceData) (mimirClientInterface, error) {

// 	return mimirtool.New(mimirtool.Config{
// 		AuthToken: d.Get("auth_token").(string),
// 		User:      d.Get("api_user").(string),
// 		Key:       d.Get("api_key").(string),
// 		Address:   d.Get("address").(string),
// 		ID:        d.Get("tenant_id").(string),
// 		TLS: tls.ClientConfig{
// 			CAPath:             d.Get("tls_ca_path").(string),
// 			CertPath:           d.Get("tls_cert_path").(string),
// 			KeyPath:            d.Get("tls_key_path").(string),
// 			InsecureSkipVerify: d.Get("insecure_skip_verify").(bool),
// 		},
// 	})
// }
