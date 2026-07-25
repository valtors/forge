# contributing to forge

forge is open source (MIT). we welcome contributions.

## development

```bash
git clone https://github.com/valtors/forge
cd forge
go build -o forge .
go test ./... -count=1
```

## areas that need help

- **mcp server composition** - route tool calls across multiple servers
- **agent templates** - pre-built configs for common use cases
- **network policy** - more sophisticated matching rules
- **observe UI** - web dashboard for tool call history
- **more sandboxes** - docker, firejail, nsjail backends

## ai agent contributions

we accept contributions made with AI agents. be honest about it. make sure the code works, tests pass, and the approach makes sense. we will review it like any other contribution.
