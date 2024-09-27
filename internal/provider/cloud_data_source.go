package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
	"terraform-provider-juicefscloud/internal/juicefs"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSourceWithConfigure = &CloudDataSource{}

func NewCloudDataSource() datasource.DataSource {
	return &CloudDataSource{}
}

// CloudDataSource defines the data source implementation.
type CloudDataSource struct {
	client *juicefs.Client
}

type CloudDataSourceModel struct {
	ID      types.Int64  `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Storage types.String `tfsdk:"storage"`
}

func (c *CloudDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_cloud"
}

func (c *CloudDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Cloud provider data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required:            false,
				Optional:            false,
				Computed:            true,
				MarkdownDescription: "Cloud identifier",
			},
			"name": schema.StringAttribute{
				Required:            false,
				Optional:            true,
				Computed:            false,
				MarkdownDescription: "Cloud name",
			},
			"storage": schema.StringAttribute{
				Required:            false,
				Optional:            true,
				Computed:            false,
				MarkdownDescription: "Cloud storage",
			},
		},
	}
}

func (c *CloudDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*juicefs.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *juicefs.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	c.client = client
}

func (c *CloudDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data CloudDataSourceModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	clouds, err := c.client.GetClouds()
	if err != nil {
		response.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read clouds, got error: %s", err))
		return
	}

	found := false
	for _, cloud := range clouds {
		if cloud.Name == data.Name.ValueString() || cloud.Storage == data.Storage.ValueString() {
			found = true
			data.ID = types.Int64Value(cloud.ID)
			if !data.Name.IsNull() && data.Name.ValueString() != cloud.Name {
				response.Diagnostics.AddError("Cloud Name Mismatch", "Cloud name mismatch, you should use name or storage to identify the cloud, not both.")
			}
			data.Name = types.StringValue(cloud.Name)
			if !data.Storage.IsNull() && data.Storage.ValueString() != cloud.Storage {
				response.Diagnostics.AddError("Cloud Storage Mismatch", "Cloud storage mismatch, you should use name or storage to identify the cloud, not both.")
			}
			data.Storage = types.StringValue(cloud.Storage)
		}
	}

	if !found {
		names := make([]string, len(clouds))
		storages := make([]string, len(clouds))
		for i, cloud := range clouds {
			names[i] = cloud.Name
			storages[i] = cloud.Storage
		}
		response.Diagnostics.AddError("Cloud Not Found", fmt.Sprintf("avaliable cloud names are: %s\navaliable storages are: %s", strings.Join(names, ", "), strings.Join(storages, ", ")))
		return
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}
