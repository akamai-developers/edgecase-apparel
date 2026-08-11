package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

const (
	Infra = "infra"
	Kube  = "kubernetes"
)

var errPulumiRunFunc = errors.New("invalid pulumi.RunFunc")

type PulumiAutoStack struct {
	Opts  PulumiAutoOpts
	Stack auto.Stack
}

type PulumiRunFuncs map[string]pulumi.RunFunc

type PulumiAutoStackMap map[string]PulumiAutoStack

type PulumiAutoOpts struct {
	DestroyOpts []optdestroy.Option
	PreviewOpts []optpreview.Option
	UpOpts      []optup.Option
}

func (p PulumiAutoStack) Preview(ctx context.Context, opts ...optpreview.Option) auto.PreviewResult {
	o := p.Opts.PreviewOpts
	if len(opts) > 0 {
		o = append(o, opts...)
	}

	res, err := p.Stack.Preview(ctx, o...)
	chkError(err, "failed pulumi preview")

	return res
}

func (p PulumiAutoStack) Deploy(ctx context.Context, opts ...optup.Option) auto.UpResult {
	o := p.Opts.UpOpts
	if len(opts) > 0 {
		o = append(o, opts...)
	}

	res, err := p.Stack.Up(ctx, o...)
	chkError(err, "failed pulumi deploy")

	return res
}

func (p PulumiAutoStack) Destroy(ctx context.Context, opts ...optdestroy.Option) auto.DestroyResult {
	o := p.Opts.DestroyOpts
	if len(opts) > 0 {
		o = append(o, opts...)
	}

	res, err := p.Stack.Destroy(ctx, o...)
	chkError(err, "failed pulumi destroy")

	return res
}

func main() {
	ctx := context.Background()

	// Get Viper configs
	if err := cfg.InitConfig(); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	var pulumiConfs cfg.PulumiConfigs

	if err := viper.UnmarshalKey("pulumi", &pulumiConfs.Projects); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	stkMap := InitStacks(ctx, pulumiConfs)

	args := os.Args[1]
	switch args {
	case "preview-kube":
		_ = stkMap[Kube].Preview(ctx)
	case "deploy-kube":
		_ = stkMap[Kube].Deploy(ctx)
	case "destroy-kube":
		_ = stkMap[Kube].Destroy(ctx)
	case "preview-infra":
		_ = stkMap[Infra].Preview(ctx)
	case "deploy-infra":
		_ = stkMap[Infra].Deploy(ctx)
	case "destroy-infra":
		_ = stkMap[Infra].Destroy(ctx)
	}
}

func InitStacks(ctx context.Context, confs cfg.PulumiConfigs) PulumiAutoStackMap {
	stkMap := PulumiAutoStackMap{}

	stkFuncs := PulumiRunFuncs{
		Infra: infra.Deploy,
		Kube:  kube.Deploy,
	}

	for _, prog := range confs.Projects {
		var stk PulumiAutoStack

		stk.Opts = PulumiAutoOpts{
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
		}

		if _, ok := stkFuncs[prog.Name]; !ok {
			fmt.Println("Did this print?")
			chkError(errPulumiRunFunc, prog.Name)
		}

		fqsn := filepath.Join("organization", prog.Project, prog.Stack)
		program := stkFuncs[prog.Name]
		workdir := filepath.Join("..", prog.Name)

		autoStk, err := auto.UpsertStackLocalSource(ctx, fqsn, workdir, auto.Program(program))
		if err != nil {
			fmt.Println("how about this")
			fmt.Println(err)
			os.Exit(1)
		}

		stk.Stack = autoStk

		stkMap[prog.Name] = stk
	}

	return stkMap
}

//nolint:err113
func chkError(err error, i ...any) {
	if err != nil {
		fmt.Printf("\n[ error ] %s: %v\n", err.Error(), i)
		os.Exit(1)
	}
}
