package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"terraform-provider-juicefscloud/internal/juicefs"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSourceWithConfigure = &VolumeDataSource{}

func NewVolumeDataSource() datasource.DataSource {
	return &VolumeDataSource{}
}

// VolumeDataSource defines the data source implementation.
type VolumeDataSource struct {
	client *juicefs.Client
}

type VolumeAccessRulesDataSourceModel struct {
	IpRange    types.String `tfsdk:"ip_range"`
	Token      types.String `tfsdk:"token"`
	ReadOnly   types.Bool   `tfsdk:"read_only"`
	AppendOnly types.Bool   `tfsdk:"append_only"`
}

func (VolumeAccessRulesDataSourceModel) schema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"ip_range": schema.StringAttribute{
				MarkdownDescription: "IP range for access rules",
				Required:            false,
				Optional:            false,
				Computed:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Token for access rules",
				Required:            false,
				Optional:            false,
				Computed:            true,
			},
			"read_only": schema.BoolAttribute{
				MarkdownDescription: "Read-only access",
				Required:            false,
				Optional:            false,
				Computed:            true,
			},
			"append_only": schema.BoolAttribute{
				MarkdownDescription: "Append-only access",
				Required:            false,
				Optional:            false,
				Computed:            true,
			},
		},
	}
}

func (VolumeAccessRulesDataSourceModel) attrType() map[string]attr.Type {
	return map[string]attr.Type{
		"ip_range":    types.StringType,
		"token":       types.StringType,
		"read_only":   types.BoolType,
		"append_only": types.BoolType,
	}
}

// VolumeDataSourceModel describes the data source data model.
type VolumeDataSourceModel struct {
	Id          types.Int64  `tfsdk:"id"`
	AccessRules types.List   `tfsdk:"access_rules"`
	Owner       types.Int64  `tfsdk:"owner"`
	Size        types.Int64  `tfsdk:"size"`
	Inodes      types.Int64  `tfsdk:"inodes"`
	Created     types.String `tfsdk:"created"`
	Uuid        types.String `tfsdk:"uuid"`
	Name        types.String `tfsdk:"name"`
	Region      types.Int64  `tfsdk:"region"`
	Bucket      types.String `tfsdk:"bucket"`
	TrashTime   types.Int64  `tfsdk:"trash_time"`
	BlockSize   types.Int64  `tfsdk:"block_size"`
	Compress    types.String `tfsdk:"compress"`
	Compatible  types.Bool   `tfsdk:"compatible"`
	Extend      types.String `tfsdk:"extend"`
	Storage     types.String `tfsdk:"storage"`
}

func (d *VolumeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (d *VolumeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Volume data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required:            false,
				Optional:            false,
				Computed:            true,
				MarkdownDescription: "Volume identifier",
			},
			"access_rules": schema.ListNestedAttribute{
				MarkdownDescription: "Specify access rules for the volume.",
				Required:            false,
				Optional:            false,
				Computed:            true,
				NestedObject:        VolumeAccessRulesDataSourceModel{}.schema(),
			},
			"owner": schema.Int64Attribute{
				MarkdownDescription: "Owner of the volume",
				Required:            false,
				Optional:            false,
				Computed:            true,
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "Size of the volume",
				Required:            false,
				Optional:            false,
				Computed:            true,
			},
			"inodes": schema.Int64Attribute{
				MarkdownDescription: "Number of inodes",
				Required:            false,
				Optional:            false,
				Computed:            true,
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "Creation time of the volume",
				Required:            false,
				Optional:            false,
				Computed:            true,
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "UUID of the volume",
				Required:            false,
				Optional:            false,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the volume",
				Required:            true,
				Optional:            false,
				Computed:            false,
			},
			"region": schema.Int64Attribute{
				MarkdownDescription: "Region of the volume",
				Required:            false,
				Optional:            false,
				Computed:            true,
			},
			"bucket": schema.StringAttribute{
				MarkdownDescription: "Bucket for the volume",
				Required:            false,
				Optional:            false,
				Computed:            true,
			},
			"trash_time": schema.Int64Attribute{
				MarkdownDescription: "Trash time of the volume",
				Required:            false,
				Optional:            false,
				Computed:            true,
			},
			"block_size": schema.Int64Attribute{
				MarkdownDescription: "Block size of the volume",
				Required:            false,
				Optional:            false,
				Computed:            true,
			},
			"compress": schema.StringAttribute{
				MarkdownDescription: "Compression type for the volume",
				Required:            false,
				Optional:            false,
				Computed:            true,
			},
			"compatible": schema.BoolAttribute{
				MarkdownDescription: "Compatibility mode for the volume",
				Required:            false,
				Optional:            false,
				Computed:            true,
			},
			"extend": schema.StringAttribute{
				MarkdownDescription: "Extended attributes for the volume",
				Required:            false,
				Optional:            false,
				Computed:            true,
			},
			"storage": schema.StringAttribute{
				MarkdownDescription: "Storage type for the volume",
				Required:            false,
				Optional:            false,
				Computed:            true,
			},
		},
	}
}

func (d *VolumeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

func (d *VolumeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VolumeDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	volumes, err := d.client.GetVolumes()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read volume, got error: %s", err))
		return
	}

	found := false
	for _, volume := range volumes {
		if volume.Name == data.Name.ValueString() {
			found = true

			tflog.Trace(ctx, fmt.Sprintf("found matching volume name %s", volume.Name))
			data.Id = types.Int64Value(volume.Id)
			accessRules := make([]VolumeAccessRulesDataSourceModel, 0)
			for _, accessRule := range volume.AccessRules {
				accessRules = append(accessRules, VolumeAccessRulesDataSourceModel{
					IpRange:    types.StringValue(accessRule.IpRange),
					Token:      types.StringValue(accessRule.Token),
					ReadOnly:   types.BoolValue(accessRule.ReadOnly),
					AppendOnly: types.BoolValue(accessRule.AppendOnly),
				})
			}
			rules, diag := types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: VolumeAccessRulesDataSourceModel{}.attrType(),
			}, accessRules)
			resp.Diagnostics.Append(diag...)
			if resp.Diagnostics.HasError() {
				return
			}
			data.AccessRules = rules
			data.Owner = types.Int64Value(volume.Owner)
			if volume.Size != nil {
				data.Size = types.Int64Value(*volume.Size)
			}
			if volume.Inodes != nil {
				data.Inodes = types.Int64Value(*volume.Inodes)
			}
			data.Created = types.StringValue(volume.Created.Format(time.RFC3339))
			data.Uuid = types.StringValue(volume.Uuid)
			data.Name = types.StringValue(volume.Name)
			data.Region = types.Int64Value(volume.Region)
			data.Bucket = types.StringValue(volume.Bucket)
			data.TrashTime = types.Int64Value(volume.TrashTime)
			data.BlockSize = types.Int64Value(volume.BlockSize)
			data.Compress = types.StringValue(volume.Compress)
			data.Compatible = types.BoolValue(volume.Compatible)
			if volume.Extend != nil {
				data.Extend = types.StringValue(*volume.Extend)
			}
			if volume.Storage != nil {
				data.Storage = types.StringValue(*volume.Storage)
			}
		}
	}

	if !found {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find volume %s", data.Name.ValueString()))
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
