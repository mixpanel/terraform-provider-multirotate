package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &MultiRotateProvider{}
var _ provider.ProviderWithFunctions = &MultiRotateProvider{}

// MultiRotateProvider defines the provider implementation.
type MultiRotateProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// MultiRotateProviderModel describes the provider data model.
type MultiRotateProviderModel struct {
	Timestamp types.String `tfsdk:"timestamp"`
}

func (p *MultiRotateProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "multirotate"
	resp.Version = p.version
}

func (p *MultiRotateProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Multi Rotate Provider is a provider that allows you to easily rotate multiple instances of an object on a regular basis.
`,
		Attributes: map[string]schema.Attribute{
			"timestamp": schema.StringAttribute{
				MarkdownDescription: "The current timestamp.  Only useful for testing.",
				Optional:            true,
			},
		},
	}
}

func (p *MultiRotateProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data MultiRotateProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if !data.Timestamp.IsNull() {
		var err error
		resp.ResourceData, err = time.Parse(time.RFC3339, data.Timestamp.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Timestamp",
				"Unable to parse timestamp: "+err.Error(),
			)
		}
	} else {
		resp.ResourceData = time.Now()
	}

	if resp.Diagnostics.HasError() {
		return
	}
}

func (p *MultiRotateProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMultiRotateSet,
	}
}

func (p *MultiRotateProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *MultiRotateProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MultiRotateProvider{
			version: version,
		}
	}
}
