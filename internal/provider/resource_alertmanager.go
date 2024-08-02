package provider

import (
	"context"
	"fmt"
	"log"

	mimirtool "github.com/grafana/mimir/pkg/mimirtool/client"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource = &alertmanagerResource{}
)

func NewAlertmanagerResource() resource.Resource {
	return &alertmanagerResource{}
}

// alertmanagerResource is the resource implementation.
type alertmanagerResource struct {
	client *mimirtool.MimirClient
}

func (r *alertmanagerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
		[Official documentation](https://grafana.com/docs/mimir/latest/references/http-api/#alertmanager)
		`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The hexadecimal encoding of the SHA1 checksum of the file content.",
				Computed:    true,
			},
			"config_yaml": schema.StringAttribute{
				Description: "The Alertmanager configuration to load in Grafana Mimir as YAML.",
				Required:    true,
				// PlanModifiers: []planmodifier.String{
				// 	stringplanmodifier.RequiresReplace(),
				// },
			},
			"templates_config_yaml": schema.MapAttribute{
				Description: "The templates to load along with the configuration.",
				Optional:    true,
				ElementType: types.StringType,
				// PlanModifiers: []planmodifier.String{
				// 	stringplanmodifier.RequiresReplace(),
				// },
				// Validators: []validator.String{
				// 	stringvalidator.ExactlyOneOf(
				// 		path.MatchRoot("sensitive_content"),
				// 		path.MatchRoot("content_base64"),
				// 		path.MatchRoot("source")),
				// },
			},
		},
	}
}

func (r *alertmanagerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alertmanager"
}

func (r *alertmanagerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertmanagerResourceModelV0

	log.Printf("[NEW] config supplied by user is: %+v\n", req.Config)

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	alertmanagerConfig := plan.Config_yaml.ValueString()
	templates := make(map[string]string)
	//templates := make(map[string]types.String)
	diags = plan.Templates_config_yaml.ElementsAs(ctx, &templates, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// client request
	log.Printf("[NEW] templates is: %+v\n", templates)
	tflog.Info(ctx, "before creating AM config")
	log.Printf("[NEW] r.client is: %+v\n", r.client)
	log.Printf("[NEW] plan is: %+v\n", plan)
	err := r.client.CreateAlertmanagerConfig(ctx, alertmanagerConfig, templates)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Alertmanager config",
			"An unexpected error occurred while creating Alertmanager config\n\n+"+
				fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	// Mimir supports only one alertmanager configuration per tenant as such there is no associated ID
	plan.ID = types.StringValue("alertmanager")
	log.Printf("[NEW] plan is: %+v\n", plan)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *alertmanagerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// var state alertmanagerResourceModelV0
	// diags := req.State.Get(ctx, &state)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// //templates := make(map[string]string)
	// // Get refreshed order value from HashiCups
	// config, templates, err := r.client.GetAlertmanagerConfig(ctx)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Error Reading Alertmanager config",
	// 		"Could not read Alertmanager configuration "+state.ID.ValueString()+": "+err.Error(),
	// 	)
	// 	return
	// }

	// // Overwrite items with refreshed state
	// state.Config_yaml = types.StringValue(config)
	// state.Templates_config_yaml = types.MapValue{types.String, templates}
	// // state.Items = []orderItemModel{}
	// // for _, item := range order.Items {
	// // 	state.Items = append(state.Items, orderItemModel{
	// // 		Coffee: orderItemCoffeeModel{
	// // 			ID:          types.Int64Value(int64(item.Coffee.ID)),
	// // 			Name:        types.StringValue(item.Coffee.Name),
	// // 			Teaser:      types.StringValue(item.Coffee.Teaser),
	// // 			Description: types.StringValue(item.Coffee.Description),
	// // 			Price:       types.Float64Value(item.Coffee.Price),
	// // 			Image:       types.StringValue(item.Coffee.Image),
	// // 		},
	// // 		Quantity: types.Int64Value(int64(item.Quantity)),
	// // 	})
	// // }

	// // Set refreshed state
	// diags = resp.State.Set(ctx, &state)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// var state alertmanagerResourceModelV0

	// diags := req.State.Get(ctx, &state)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// // If the output file doesn't exist, mark the resource for creation.
	// outputPath := state.Filename.ValueString()
	// if _, err := os.Stat(outputPath); os.IsNotExist(err) {
	// 	resp.State.RemoveResource(ctx)
	// 	return
	// }

	// // Verify that the content of the destination file matches the content we
	// // expect. Otherwise, the file might have been modified externally, and we
	// // must reconcile.
	// outputContent, err := os.ReadFile(outputPath)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Read local file error",
	// 		"An unexpected error occurred while reading the file\n\n+"+
	// 			fmt.Sprintf("Original Error: %s", err),
	// 	)
	// 	return
	// }

	// outputChecksum := sha1.Sum(outputContent)
	// if hex.EncodeToString(outputChecksum[:]) != state.ID.ValueString() {
	// 	resp.State.RemoveResource(ctx)
	// 	return
	// }

	//// OLD
	// client := meta.(*client).cli
	// alertmanagerConfig, templates, err := client.GetAlertmanagerConfig(ctx)
	// if errors.Is(err, mimirtool.ErrResourceNotFound) {
	// 	// need to tell terraform the resource does not exist
	// 	tflog.Info(ctx, "No alertmanager mimir side")
	// 	d.SetId("")
	// } else if err != nil {
	// 	return diag.FromErr(err)
	// }
	// d.Set("config_yaml", alertmanagerConfig)
	// d.Set("templates_config_yaml", templates)
	// return nil
}

func (r *alertmanagerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan alertmanagerResourceModelV0

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alertmanagerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var plan alertmanagerResourceModelV0

	// client request
	err := r.client.DeleteAlermanagerConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while deleting Alertmanager configuration",
			fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	plan.ID = types.StringValue("")
	diags := resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// func parseLocalFileContent(plan alertmanagerResourceModelV0) ([]byte, error) {
// 	if !plan.SensitiveContent.IsNull() && !plan.SensitiveContent.IsUnknown() {
// 		return []byte(plan.SensitiveContent.ValueString()), nil
// 	}
// 	if !plan.ContentBase64.IsNull() && !plan.ContentBase64.IsUnknown() {
// 		return base64.StdEncoding.DecodeString(plan.ContentBase64.ValueString())
// 	}

// 	if !plan.Source.IsNull() && !plan.Source.IsUnknown() {
// 		sourceFileContent := plan.Source.ValueString()
// 		return os.ReadFile(sourceFileContent)
// 	}

// 	content := plan.Content.ValueString()
// 	return []byte(content), nil
// }

// Configure adds the provider configured client to the resource.
func (r *alertmanagerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mimirtool.MimirClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

type alertmanagerResourceModelV0 struct {
	ID                    types.String `tfsdk:"id"`
	Config_yaml           types.String `tfsdk:"config_yaml"`
	Templates_config_yaml types.Map    `tfsdk:"templates_config_yaml"`
}

// func resourceAlertManager() *schema.Resource {
// 	return &schema.Resource{
// 		Description: `
// [Official documentation](https://grafana.com/docs/mimir/latest/references/http-api/#alertmanager)
// `,

// 		CreateContext: alertmanagerCreate,
// 		ReadContext:   alertmanagerRead,
// 		UpdateContext: alertmanagerCreate, // There is no PUT, the POST is responsible to overwrite the configuration
// 		DeleteContext: alertmanagerDelete,
// 		Importer: &schema.ResourceImporter{
// 			StateContext: schema.ImportStatePassthroughContext,
// 		},

// 		Schema: map[string]*schema.Schema{
// 			"config_yaml": {
// 				Description: "The Alertmanager configuration to load in Grafana Mimir as YAML.",
// 				Type:        schema.TypeString,
// 				Required:    true,
// 			},
// 			"templates_config_yaml": {
// 				Description: "The templates to load along with the configuration.",
// 				Type:        schema.TypeMap,
// 				Elem:        &schema.Schema{Type: schema.TypeString},
// 				Optional:    true,
// 			},
// 		},
// 	}
// }

// func alertmanagerDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
// 	var diags diag.Diagnostics
// 	client := meta.(*client).cli
// 	err := client.DeleteAlermanagerConfig(ctx)
// 	if err != nil {
// 		return diag.FromErr(err)
// 	}

// 	d.SetId("")
// 	return diags
// }
