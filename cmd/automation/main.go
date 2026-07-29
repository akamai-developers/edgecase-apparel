package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/akamai-developers/edgecase-apparel/cmd/infra/app"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type PulumiStack struct {
	DestroyOpts []optdestroy.Option
	PreviewOpts []optpreview.Option
	UpOpts      []optup.Option
	Stack       auto.Stack
}

type PulumiStackArgs struct {
	Program   pulumi.RunFunc
	StackName string
	WorkDir   string
}

func (p *PulumiStack) Preview(ctx context.Context, opts ...optpreview.Option) auto.PreviewResult {
	if len(opts) > 0 {
		p.PreviewOpts = append(p.PreviewOpts, opts...)
	}
	res, err := p.Stack.Preview(ctx, p.PreviewOpts...)
	chkError("p.Preview()", err)

	return res
}

func (p *PulumiStack) Deploy(ctx context.Context, opts ...optup.Option) auto.UpResult {
	if len(opts) > 0 {
		p.UpOpts = append(p.UpOpts, opts...)
	}
	res, err := p.Stack.Up(ctx, p.UpOpts...)
	chkError("p.Deploy()", err)

	return res
}

func (p *PulumiStack) Destroy(ctx context.Context, opts ...optdestroy.Option) auto.DestroyResult {
	if len(opts) > 0 {
		p.DestroyOpts = append(p.DestroyOpts, opts...)
	}
	res, err := p.Stack.Destroy(ctx, p.DestroyOpts...)
	chkError("p.Destroy()", err)

	return res
}

func main() {
	ctx := context.Background()

	stkArgs := PulumiStackArgs{
		Program:   app.Deploy,
		StackName: auto.FullyQualifiedStackName("organization", "edgecase-apparel-infra", "dev"),
		WorkDir:   filepath.Join("..", "infra"),
	}

	infra := InitStack(ctx, stkArgs)

	//
	args := os.Args[1]
	switch args {
	case "preview-infra":
		_ = infra.Preview(ctx)
	case "deploy-infra":
		_ = infra.Deploy(ctx)
	case "destroy-infra":
		_ = infra.Destroy(ctx)
	}
}

func InitStack(ctx context.Context, st PulumiStackArgs) PulumiStack {
	stk, err := auto.UpsertStackLocalSource(ctx, st.StackName, st.WorkDir, auto.Program(st.Program))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	stkOpts := PulumiStack{
		DestroyOpts: []optdestroy.Option{
			optdestroy.Color("always"),
			optdestroy.ProgressStreams(os.Stdout),
			optdestroy.ErrorProgressStreams(os.Stderr),
			optdestroy.Parallel(1),
			optdestroy.Refresh(),
		},
		PreviewOpts: []optpreview.Option{
			optpreview.Color("always"),
			optpreview.ProgressStreams(os.Stdout),
			optpreview.ErrorProgressStreams(os.Stderr),
			optpreview.Parallel(1),
			// optpreview.Refresh(),
		},
		UpOpts: []optup.Option{
			optup.Color("always"),
			optup.ProgressStreams(os.Stdout),
			optup.ErrorProgressStreams(os.Stderr),
			optup.Parallel(1),
			optup.Refresh(),
		},
		Stack: stk,
	}
	return stkOpts
}

func chkError(funcName string, err error) {
	if err != nil {
		fmt.Printf("\n[ error ] %s: %v\n", funcName, err)
		os.Exit(1)
	}
}
