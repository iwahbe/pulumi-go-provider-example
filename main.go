package main

import (
	"context"
	"fmt"
	"os"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

func main() {
	provider, err := infer.NewProviderBuilder().
		WithResources(
			infer.Resource(File{}),
		).
		WithNamespace("example").
		Build()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}

	err = provider.Run(context.Background(), "file", "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}
}

// Resource definition
type File struct{}

func (f *File) Annotate(a infer.Annotator) {
	a.Describe(&f, "A file projected into a pulumi resource")
}

// Input arguments
type FileArgs struct {
	Path    string `pulumi:"path,optional"`
	Force   bool   `pulumi:"force,optional"`
	Content string `pulumi:"content"`
}

func (f *FileArgs) Annotate(a infer.Annotator) {
	a.Describe(&f.Content, "The content of the file.")
	a.Describe(&f.Force, "If existing file should be deleted if present.")
	a.Describe(&f.Path, "The file path. Defaults to resource name.")
}

// Stored state
type FileState struct {
	Path    string `pulumi:"path"`
	Force   bool   `pulumi:"force"`
	Content string `pulumi:"content"`
}

func (f *FileState) Annotate(a infer.Annotator) {
	a.Describe(&f.Content, "The content of the file.")
	a.Describe(&f.Force, "If existing file should be deleted if present.")
	a.Describe(&f.Path, "The file path.")
}

// CRUD Operations
func (File) Create(ctx context.Context, req infer.CreateRequest[FileArgs]) (infer.CreateResponse[FileState], error) {
	if !req.Inputs.Force {
		_, err := os.Stat(req.Inputs.Path)
		if !os.IsNotExist(err) {
			return infer.CreateResponse[FileState]{},
				fmt.Errorf("file exists; pass force=true to override")
		}
	}

	if req.DryRun {
		return infer.CreateResponse[FileState]{ID: req.Inputs.Path}, nil
	}

	f, err := os.Create(req.Inputs.Path)
	if err != nil {
		return infer.CreateResponse[FileState]{}, err
	}
	defer f.Close()

	n, err := f.WriteString(req.Inputs.Content)
	if err != nil {
		return infer.CreateResponse[FileState]{}, err
	}
	if n != len(req.Inputs.Content) {
		return infer.CreateResponse[FileState]{},
			fmt.Errorf("wrote %d/%d bytes", n, len(req.Inputs.Content))
	}

	return infer.CreateResponse[FileState]{
		ID: req.Inputs.Path,
		Output: FileState{
			Path:    req.Inputs.Path,
			Force:   req.Inputs.Force,
			Content: req.Inputs.Content,
		},
	}, nil
}

func (File) Delete(ctx context.Context, req infer.DeleteRequest[FileState]) (infer.DeleteResponse, error) {
	err := os.Remove(req.State.Path)
	if os.IsNotExist(err) {
		p.GetLogger(ctx).Warningf("file %q already deleted", req.State.Path)
		err = nil
	}
	return infer.DeleteResponse{}, err
}

func (File) Check(ctx context.Context, req infer.CheckRequest) (infer.CheckResponse[FileArgs], error) {
	if _, ok := req.NewInputs.GetOk("path"); !ok {
		req.NewInputs = req.NewInputs.Set("path", property.New(req.Name))
	}
	args, f, err := infer.DefaultCheck[FileArgs](ctx, req.NewInputs)
	return infer.CheckResponse[FileArgs]{Inputs: args, Failures: f}, err
}

func (File) Update(ctx context.Context, req infer.UpdateRequest[FileArgs, FileState]) (infer.UpdateResponse[FileState], error) {
	if req.DryRun {
		return infer.UpdateResponse[FileState]{}, nil
	}

	f, err := os.Create(req.State.Path)
	if err != nil {
		return infer.UpdateResponse[FileState]{}, err
	}
	defer f.Close()

	n, err := f.WriteString(req.Inputs.Content)
	if err != nil {
		return infer.UpdateResponse[FileState]{}, err
	}
	if n != len(req.Inputs.Content) {
		return infer.UpdateResponse[FileState]{},
			fmt.Errorf("wrote %d/%d bytes", n, len(req.Inputs.Content))
	}

	return infer.UpdateResponse[FileState]{
		Output: FileState{
			Path:    req.Inputs.Path,
			Force:   req.Inputs.Force,
			Content: req.Inputs.Content,
		},
	}, nil
}

func (File) Diff(ctx context.Context, req infer.DiffRequest[FileArgs, FileState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.Inputs.Content != req.State.Content {
		diff["content"] = p.PropertyDiff{Kind: p.Update}
	}
	if req.Inputs.Force != req.State.Force {
		diff["force"] = p.PropertyDiff{Kind: p.Update}
	}
	if req.Inputs.Path != req.State.Path {
		diff["path"] = p.PropertyDiff{Kind: p.UpdateReplace}
	} else {
		_, err := os.Stat(req.Inputs.Path)
		if os.IsNotExist(err) {
			diff["path"] = p.PropertyDiff{Kind: p.Add}
		}
	}
	return infer.DiffResponse{
		DeleteBeforeReplace: true,
		HasChanges:          len(diff) > 0,
		DetailedDiff:        diff,
	}, nil
}

func (File) Read(ctx context.Context, req infer.ReadRequest[FileArgs, FileState]) (infer.ReadResponse[FileArgs, FileState], error) {
	byteContent, err := os.ReadFile(req.ID)
	if err != nil {
		return infer.ReadResponse[FileArgs, FileState]{}, err
	}
	content := string(byteContent)
	return infer.ReadResponse[FileArgs, FileState]{
		ID: req.ID,
		Inputs: FileArgs{
			Path:    req.ID,
			Force:   req.State.Force,
			Content: content,
		},
		State: FileState{
			Path:    req.ID,
			Force:   req.State.Force,
			Content: content,
		},
	}, nil
}

func (File) WireDependencies(f infer.FieldSelector, args *FileArgs, state *FileState) {
	f.OutputField(&state.Content).DependsOn(f.InputField(&args.Content))
	f.OutputField(&state.Force).DependsOn(f.InputField(&args.Force))
	f.OutputField(&state.Path).DependsOn(f.InputField(&args.Path))
}
