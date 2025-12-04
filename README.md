# Pulumi Go Provider Example

A simple example Pulumi provider built with the [Pulumi Go Provider SDK](https://www.pulumi.com/docs/iac/guides/building-extending/providers/build-a-provider/).

## Testing

Run the example program:

```bash
make bin/pulumi-resource-file
pulumi -C examples/simple up
```

## Resources

- **File**: Creates a file with specified content at a given path.

## Reference

Built following the official guide: https://www.pulumi.com/docs/iac/guides/building-extending/providers/build-a-provider/
