package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-juicefscloud/internal/juicefs"
	"time"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VolumeResource{}
var _ resource.ResourceWithImportState = &VolumeResource{}

func NewVolumeResource() resource.Resource {
	return &VolumeResource{}
}

// VolumeResource defines the resource implementation.
type VolumeResource struct {
	client *juicefs.Client
}

type VolumeAccessRulesResourceModel struct {
	IpRange    types.String `tfsdk:"ip_range"`
	Token      types.String `tfsdk:"token"`
	ReadOnly   types.Bool   `tfsdk:"read_only"`
	AppendOnly types.Bool   `tfsdk:"append_only"`
}

func (VolumeAccessRulesResourceModel) schema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"ip_range": schema.StringAttribute{
				MarkdownDescription: "IP range for access rules",
				Computed:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Token for access rules",
				Computed:            true,
			},
			"read_only": schema.BoolAttribute{
				MarkdownDescription: "Read-only access",
				Computed:            true,
			},
			"append_only": schema.BoolAttribute{
				MarkdownDescription: "Append-only access",
				Computed:            true,
			},
		},
	}
}

func (VolumeAccessRulesResourceModel) attrType() map[string]attr.Type {
	return map[string]attr.Type{
		"ip_range":    types.StringType,
		"token":       types.StringType,
		"read_only":   types.BoolType,
		"append_only": types.BoolType,
	}
}

// VolumeResourceModel describes the resource data model.
type VolumeResourceModel struct {
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

func (r *VolumeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (r *VolumeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Volume resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required:            false,
				Optional:            false,
				Computed:            true,
				MarkdownDescription: "volume identifier",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"access_rules": schema.ListNestedAttribute{
				MarkdownDescription: "Specify access rules for the volume.",
				Required:            false,
				Optional:            false,
				Computed:            true,
				NestedObject:        VolumeAccessRulesResourceModel{}.schema(),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"owner": schema.Int64Attribute{
				MarkdownDescription: "Owner of the volume",
				Required:            false,
				Optional:            false,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "Size of the volume",
				Required:            false,
				Optional:            false,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"inodes": schema.Int64Attribute{
				MarkdownDescription: "Number of inodes",
				Required:            false,
				Optional:            false,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "Creation time of the volume",
				Required:            false,
				Optional:            false,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "UUID of the volume",
				Required:            false,
				Optional:            false,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the volume",
				Required:            true,
				Optional:            false,
				Computed:            false,
			},
			"region": schema.Int64Attribute{
				MarkdownDescription: "Region of the volume",
				Required:            true,
				Optional:            false,
				Computed:            false,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"bucket": schema.StringAttribute{
				MarkdownDescription: "Bucket for the volume",
				Required:            false,
				Optional:            false,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"trash_time": schema.Int64Attribute{
				MarkdownDescription: "Trash time of the volume",
				Required:            false,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"block_size": schema.Int64Attribute{
				MarkdownDescription: "Block size of the volume",
				Required:            false,
				Optional:            false,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"compress": schema.StringAttribute{
				MarkdownDescription: "Compression type for the volume",
				Required:            false,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"compatible": schema.BoolAttribute{
				MarkdownDescription: "Compatibility mode for the volume",
				Required:            false,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"extend": schema.StringAttribute{
				MarkdownDescription: "Extended attributes for the volume",
				Required:            false,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"storage": schema.StringAttribute{
				MarkdownDescription: "Storage type for the volume",
				Required:            false,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
		},
	}
}

func (r *VolumeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*juicefs.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *juicefs.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *VolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VolumeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := juicefs.CreateVolumeRequest{
		Name:   data.Name.ValueString(),
		Region: data.Region.ValueInt64(),
	}
	if !data.Bucket.IsNull() {
		bucket := data.Bucket.ValueString()
		apiReq.Bucket = &bucket
	}
	if !data.BlockSize.IsNull() {
		blockSize := data.BlockSize.ValueInt64()
		apiReq.BlockSize = &blockSize
	}
	if !data.TrashTime.IsNull() {
		trashTime := data.TrashTime.ValueInt64()
		apiReq.TrashTime = &trashTime
	}
	if !data.Compress.IsNull() {
		compress := data.Compress.ValueString()
		apiReq.Compress = &compress
	}
	if !data.Compatible.IsNull() {
		compatible := data.Compatible.ValueBool()
		apiReq.Compatible = &compatible
	}
	if !data.Extend.IsNull() {
		extend := data.Extend.ValueString()
		apiReq.Extend = &extend
	}
	if !data.Storage.IsNull() {
		storage := data.Storage.ValueString()
		apiReq.Storage = &storage
	}

	volume, err := r.client.CreateVolume(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("volume created: name=%s ID=%d", volume.Name, volume.Id))

	for {
		ready, err := r.client.IsVolumeReady(volume.Id)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to check volume status, got error: %s", err))
			return
		}
		if ready {
			break
		}
		time.Sleep(1 * time.Second)
	}

	volume, err = r.client.GetVolume(volume.Id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get volume, got error: %s", err))
		return
	}
	data.Id = types.Int64Value(volume.Id)
	accessRules := make([]VolumeAccessRulesResourceModel, 0)
	for _, accessRule := range volume.AccessRules {
		accessRules = append(accessRules, VolumeAccessRulesResourceModel{
			IpRange:    types.StringValue(accessRule.IpRange),
			Token:      types.StringValue(accessRule.Token),
			ReadOnly:   types.BoolValue(accessRule.ReadOnly),
			AppendOnly: types.BoolValue(accessRule.AppendOnly),
		})
	}
	rules, diag := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: VolumeAccessRulesResourceModel{}.attrType(),
	}, accessRules)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.AccessRules = rules
	data.Owner = types.Int64Value(volume.Owner)
	if volume.Size != nil {
		data.Size = types.Int64Value(*volume.Size)
	} else {
		data.Size = types.Int64Null()
	}
	if volume.Inodes != nil {
		data.Inodes = types.Int64Value(*volume.Inodes)
	} else {
		data.Inodes = types.Int64Null()
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

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VolumeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	volumes, err := r.client.GetVolumes()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read volumes, got error: %s", err))
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
				AttrTypes: VolumeAccessRulesResourceModel{}.attrType(),
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
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VolumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VolumeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VolumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VolumeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteVolume(data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
		return
	}
}

func (r *VolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
