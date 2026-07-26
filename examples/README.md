# forge examples

sample agent configs for different use cases.

## research-agent.toml

a research agent with filesystem and git MCP servers. sandboxed to ./src and ./docs. network limited to github and npm.

## code-agent.toml

a coding agent with filesystem and git MCP servers. sandboxed to source and test dirs. network limited to Go module proxy.

## usage

```bash
# copy a template
cp examples/research-agent.toml my-agent.toml

# edit it
vim my-agent.toml

# run
forge run my-agent.toml
```
