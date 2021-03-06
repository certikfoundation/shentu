## certik query oracle response

Get response information

```
certik query oracle response <flags> [flags]
```

### Options

```
      --contract string   Provide the contract address
      --function string   Provide the function
      --height int        Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help              help for response
      --node string       <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --operator string   Provide the operator
  -o, --output string     Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string     The network chain ID
      --home string         directory for config and data (default "~/.certik")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

* [certik query oracle](certik_query_oracle.md)	 - Oracle staking subcommands


