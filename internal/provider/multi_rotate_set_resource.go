package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &MultiRotateSet{}
var _ resource.ResourceWithModifyPlan = &MultiRotateSet{}

func NewMultiRotateSet() resource.Resource {
	return &MultiRotateSet{}
}

// ExampleResource defines the resource implementation.
type MultiRotateSet struct {
}

// ExampleResourceModel describes the resource data model.
type MultiRotateSetModel struct {
	RotationPeriod  types.String `tfsdk:"rotation_period"`
	Number          types.Int64  `tfsdk:"number"`
	Version         types.String `tfsdk:"version"`
	RotationSet     types.List   `tfsdk:"rotation_set"`
	CurrentRotation types.Int64  `tfsdk:"current_rotation"`
	LastRotate      types.String `tfsdk:"last_rotate"`
	Timestamp       types.String `tfsdk:"timestamp"`
}

type MultiRotateSetItemModel struct {
	Creation   types.String `tfsdk:"creation"`
	Expiration types.String `tfsdk:"expiration"`
	Version    types.String `tfsdk:"version"`
}

func (r *MultiRotateSet) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_set"
}

func (r *MultiRotateSet) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: `
This resource allows you to control a set of objects that are rotated on a regular basis.
You can define the rotation period, as well as the number of objects to rotate.
It will create an output list with creation/expiration times as well as which one is expiring the furthest out.
`,

		Attributes: map[string]schema.Attribute{
			"rotation_period": schema.StringAttribute{
				MarkdownDescription: "Rotation period as Go duration string",
				Required:            true,
			},
			"number": schema.Int64Attribute{
				MarkdownDescription: "Number of different values to rotate",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(2),
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Version of new rotations",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"rotation_set": schema.ListNestedAttribute{
				MarkdownDescription: "List of rotation set info",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"creation": schema.StringAttribute{
							MarkdownDescription: "Creation time",
							Computed:            true,
						},
						"expiration": schema.StringAttribute{
							MarkdownDescription: "Expiration time",
							Computed:            true,
						},
						"version": schema.StringAttribute{
							MarkdownDescription: "Version",
							Computed:            true,
						},
					},
				},
				Computed: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"current_rotation": schema.Int64Attribute{
				MarkdownDescription: "Current rotation index",
				Computed:            true,
			},
			"last_rotate": schema.StringAttribute{
				MarkdownDescription: "Last rotation time",
				Computed:            true,
			},
			"timestamp": schema.StringAttribute{
				MarkdownDescription: "Current time",
				Computed:            true,
			},
		},
	}
}

func (r *MultiRotateSet) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MultiRotateSetModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rs := make([]MultiRotateSetItemModel, 0, data.Number.ValueInt64())
	rp, err := time.ParseDuration(data.RotationPeriod.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Rotation Period", "Unable to parse rotation period")
		return
	}

	lr := time.Now().Add(-rp * time.Duration(data.Number.ValueInt64()-1))

	for i := int64(0); i < data.Number.ValueInt64(); i++ {
		exp := lr.Add(time.Duration(data.Number.ValueInt64()) * rp)
		lr = lr.Add(rp)
		rs = append(rs, MultiRotateSetItemModel{
			Creation:   types.StringValue(time.Now().Format(time.RFC3339)),
			Expiration: types.StringValue(exp.Format(time.RFC3339)),
			Version:    data.Version,
		})
	}

	data.LastRotate = types.StringValue(lr.Format(time.RFC3339))

	var d diag.Diagnostics
	data.RotationSet, d = types.ListValueFrom(ctx, req.Config.Schema.GetAttributes()["rotation_set"].(schema.ListNestedAttribute).NestedObject.Type(), rs)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.CurrentRotation = types.Int64Value(data.Number.ValueInt64() - 1)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MultiRotateSet) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MultiRotateSetModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	data.Timestamp = types.StringValue(time.Now().Format(time.RFC3339))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MultiRotateSet) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var data MultiRotateSetModel

	if req.Plan.Raw.IsNull() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	now, err := time.Parse(time.RFC3339, data.Timestamp.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Current Time", "Unable to parse current time: "+err.Error())
		return
	}

	lr, err := time.Parse(time.RFC3339, data.LastRotate.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Last Rotate", "Unable to parse last rotate")
		return
	}

	rp, err := time.ParseDuration(data.RotationPeriod.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Rotation Period", "Unable to parse rotation period")
		return
	}

	for lr.Before(now.Add(time.Duration(data.Number.ValueInt64()) * -rp)) {
		lr = lr.Add(rp)
	}

	rs := make([]MultiRotateSetItemModel, 0, data.Number.ValueInt64())
	resp.Diagnostics.Append(data.RotationSet.ElementsAs(ctx, &rs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for i, r := range rs {
		exp, err := time.Parse(time.RFC3339, r.Expiration.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid Expiration", "Unable to parse expiration")
			return
		}
		if now.After(exp) {
			exp := lr.Add(time.Duration(data.Number.ValueInt64()) * rp)
			lr = lr.Add(rp)
			r.Creation = types.StringValue(now.Format(time.RFC3339))
			r.Expiration = types.StringValue(exp.Format(time.RFC3339))
			r.Version = types.StringValue(data.Version.ValueString())
			rs[i] = r
		}
	}

	data.LastRotate = types.StringValue(lr.Format(time.RFC3339))
	var d diag.Diagnostics
	data.RotationSet, d = types.ListValueFrom(ctx, req.State.Schema.GetAttributes()["rotation_set"].(schema.ListNestedAttribute).NestedObject.Type(), rs)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	furthestOut := 0
	furthestOutTime, _ := time.Parse(time.RFC3339, rs[0].Expiration.ValueString())
	for i, r := range rs {
		exp, _ := time.Parse(time.RFC3339, r.Expiration.ValueString())
		if exp.After(furthestOutTime) {
			furthestOutTime = exp
			furthestOut = i
		}
	}

	data.CurrentRotation = types.Int64Value(int64(furthestOut))

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &data)...)
}

func (r *MultiRotateSet) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MultiRotateSetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var planData MultiRotateSetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Number.Equal(planData.Number) {
		resp.Diagnostics.AddError("Number Change", "Number cannot be changed")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *MultiRotateSet) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MultiRotateSetModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}
