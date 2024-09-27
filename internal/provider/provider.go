package provider

import (
	"context"
	"log"
	"terraform-provider-juicefscloud/internal/juicefs"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure juicefsCloudProvider satisfies various provider interfaces.
var _ provider.Provider = &juicefsCloudProvider{}
var _ provider.ProviderWithFunctions = &juicefsCloudProvider{}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &juicefsCloudProvider{
			version: version,
		}
	}
}

// juicefsCloudProvider defines the provider implementation.
type juicefsCloudProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// juicefsCloudProviderModel describes the provider data model.
type juicefsCloudProviderModel struct {
	Endpoint  types.String `tfsdk:"endpoint"`
	AccessKey types.String `tfsdk:"access_key"`
	SecretKey types.String `tfsdk:"secret_key"`
}

func (p *juicefsCloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "juicefscloud"
	resp.Version = p.version
}

func (p *juicefsCloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "JuiceFS API endpoint, default to https://juicefs.com/api/v1",
				Required:            false,
				Optional:            true,
			},
			"access_key": schema.StringAttribute{
				MarkdownDescription: "JuiceFS API access key",
				Required:            true,
				Optional:            false,
			},
			"secret_key": schema.StringAttribute{
				MarkdownDescription: "JuiceFS API secret key",
				Required:            true,
				Optional:            false,
			},
		},
	}
}

func (p *juicefsCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data juicefsCloudProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := "https://juicefs.com/api/v1"
	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
		log.Printf("Using configured API endpoint %s", endpoint)
	}

	accessKey := data.AccessKey.ValueString()
	secretKey := data.SecretKey.ValueString()

	client := &juicefs.Client{
		Endpoint:  endpoint,
		AccessKey: accessKey,
		SecretKey: secretKey,
	}

	_, err := client.GetRegions()
	if err != nil {
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *juicefsCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVolumeResource,
	}
}

func (p *juicefsCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCloudDataSource,
		NewRegionDataSource,
		NewVolumeDataSource,
	}
}

func (p *juicefsCloudProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewEmptyTrashFunction,
	}
}
