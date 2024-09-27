package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

var (
	_ function.Function = EmptyTrashFunction{}
)

func NewEmptyTrashFunction() function.Function {
	return EmptyTrashFunction{}
}

type EmptyTrashFunction struct{}

func (r EmptyTrashFunction) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "empty_trash"
}

func (r EmptyTrashFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             "Empty trash function",
		MarkdownDescription: "Empty volume trash asynchronously.",
		Parameters: []function.Parameter{
			function.Int64Parameter{
				Name:                "id",
				MarkdownDescription: "Volume ID to empty trash",
			},
		},
		Return: function.BoolReturn{},
	}
}

func (r EmptyTrashFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var data string

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &data))

	if resp.Error != nil {
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, data))
}
