package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-juicefscloud/internal/juicefs"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSourceWithConfigure = &RegionDataSource{}

func NewRegionDataSource() datasource.DataSource {
	return &RegionDataSource{}
}

type RegionDataSource struct {
	client *juicefs.Client
}

type RegionDataSourceModel struct {
	Id    types.Int64  `tfsdk:"id"`
	Cloud types.Int64  `tfsdk:"cloud"`
	Name  types.String `tfsdk:"name"`
	Desp  types.String `tfsdk:"desp"`
}

func (r *RegionDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_region"
}

func (r *RegionDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Region data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required:            false,
				Optional:            false,
				Computed:            true,
				MarkdownDescription: "Region ID",
			},
			"cloud": schema.Int64Attribute{
				Required:            true,
				Optional:            false,
				Computed:            false,
				MarkdownDescription: "Cloud ID",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Optional:            false,
				Computed:            false,
				MarkdownDescription: "Region name",
			},
			"desp": schema.StringAttribute{
				Required:            false,
				Optional:            false,
				Computed:            true,
				MarkdownDescription: "Region description",
			},
		},
	}
}

func (r *RegionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	r.client = client
}

func (r *RegionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RegionDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	regions, err := r.client.GetRegions()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read regions, got error: %s", err))
	}
	found := false
	for _, region := range regions {
		if region.Cloud == data.Cloud.ValueInt64() && region.Name == data.Name.ValueString() {
			found = true
			data.Id = types.Int64Value(region.Id)
			data.Cloud = types.Int64Value(region.Cloud)
			data.Name = types.StringValue(region.Name)
			data.Desp = types.StringValue(region.Desp)
		}
	}

	if !found {
		resp.Diagnostics.AddError("Region Not Found", fmt.Sprintf("Unable to find region %s", data.Name.ValueString()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
