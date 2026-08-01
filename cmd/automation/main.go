package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cfg "github.com/akamai-developers/edgecase-apparel/cmd/config"
	infra "github.com/akamai-developers/edgecase-apparel/cmd/infra/app"
	kube "github.com/akamai-developers/edgecase-apparel/cmd/kubernetes/app"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spf13/viper"
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

func buildStackArgs(s cfg.PulumiStackConfig) PulumiStackArgs {
	var stkArgs PulumiStackArgs

	name, _, _ := strings.Cut(s.Name, "-")
	project := filepath.Base(s.Project)

	stkArgs.StackName = auto.FullyQualifiedStackName("organization", project, s.Stack)
	stkArgs.WorkDir = filepath.Join("..", name)

	return stkArgs
}

func main() {
	ctx := context.Background()

	// Get Viper configs
	if err := cfg.InitConfig(); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	var pulumiConfs cfg.PulumiConfig

	if err := viper.UnmarshalKey("pulumi", &pulumiConfs); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// Make map of automation API stacks
	stkMap := make(map[string]PulumiStackArgs)
	for _, i := range pulumiConfs.Projects {
		name, _, _ := strings.Cut(i.Name, "-")
		stkArgs := buildStackArgs(i)

		switch name {
		case "infra":
			stkArgs.Program = infra.Deploy
		case "kubernetes":
			stkArgs.Program = kube.Deploy
		}

		stkMap[name] = stkArgs
	}

	// Init the infra stack
	infraStk := InitStack(ctx, stkMap["infra"])

	// Init the K8s stack
	kubeStk := InitStack(ctx, stkMap["kubernetes"])

	args := os.Args[1]
	switch args {
	case "preview-kube":
		_ = kubeStk.Preview(ctx)
	case "deploy-kube":
		_ = kubeStk.Deploy(ctx)
	case "destroy-kube":
		_ = kubeStk.Destroy(ctx)
	case "preview-infra":
		_ = infraStk.Preview(ctx)
	case "deploy-infra":
		_ = infraStk.Deploy(ctx)
	case "destroy-infra":
		_ = infraStk.Destroy(ctx)
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
			// optdestroy.Refresh(),
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
			// optup.Refresh(),
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
