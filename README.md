# Vega Population

Shareable skills and personas for [Vega](https://github.com/martellcode/govega) agents.

## Quick Start

```bash
# Build the CLI
go build -o vega ./cmd/vega

# Search for personas
vega population search marketing

# Get info about a persona
vega population info @cmo

# Export a persona for your tron.vega.yaml
vega population export @cmo >> tron.vega.yaml
```

## CLI Commands

```bash
vega population search <query>     # Search skills, personas, profiles
vega population info <name>        # Show details about an item
vega population export <persona>   # Export persona as YAML for tron config
vega population install <name>     # Install to ~/.vega/
vega population list               # List installed items
vega population update             # Refresh cached indexes
```

### Export Options

```bash
vega population export @cmo                     # Uses defaults
vega population export @cmo --name=Maya         # Override agent name
vega population export @cmo --model=claude-opus-4-20250514
vega population export @cmo --temperature=0.5
vega population export @cmo --budget='$5.00'
```

## What's Here

### Personas

Specialized agent personalities:

| Persona | Description |
|---------|-------------|
| `@cmo` | Maya - data-driven growth marketer turned CMO |
| `@cto` | Strategic technical leader and engineering executive |
| `@incident-commander` | Calm, methodical incident response coordinator |
| `@security-analyst` | Security-focused code reviewer |
| `@code-reviewer` | Thorough, constructive code reviewer |
| `@devops-lead` | Infrastructure and deployment expert |
| `@technical-writer` | Documentation specialist |
| `@architect` | System design and architecture advisor |

### Skills

Tools and capabilities that agents can use:

| Skill | Description |
|-------|-------------|
| `kubernetes-ops` | Kubernetes cluster management and debugging |
| `aws-devops` | AWS infrastructure management |
| `github-actions` | GitHub Actions workflow management |
| `docker-ops` | Docker container and image management |
| `database-admin` | Database operations (PostgreSQL, MySQL) |
| `code-review` | Automated code review helpers |
| `terraform` | Terraform infrastructure as code |
| `monitoring` | Prometheus, Grafana, alerting |
| `git-advanced` | Advanced git operations and analysis |
| `npm-ops` | NPM package management and auditing |

### Profiles

Curated bundles of persona + skills:

| Profile | Description |
|---------|-------------|
| `+platform-engineer` | Full platform engineering toolkit |
| `+startup-cto` | Everything a startup CTO needs |
| `+full-stack-dev` | Frontend + backend + deployment |
| `+sre-oncall` | SRE on-call toolkit |
| `+security-reviewer` | Security-focused code review |

## Using with Tron

Export a persona and add it to your `tron.vega.yaml`:

```bash
vega population export @cmo >> ~/.tron/tron.vega.yaml
```

Then add the agent to Tony's team list and use `spawn_agent` to delegate.

## Go Library

```go
import "github.com/martellcode/vega-population/population"

client, _ := population.NewClient()

// Search for personas
results, _ := client.Search(ctx, "marketing", &population.SearchOptions{
    Kind: population.KindPersona,
})

// Get info
info, _ := client.Info(ctx, "@cmo")

// Install to ~/.vega/
client.Install(ctx, "@cmo", nil)

// List installed
items, _ := client.List(population.KindPersona)
```

## Creating Your Own

### Persona Format

```yaml
kind: persona
name: my-persona
version: 1.0.0
description: What this persona is
author: your-github-username
tags: [relevant, tags]

recommended_skills:
  - skills-this-persona-works-well-with

system_prompt: |
  You are [Name], a [role]...

  ## Your Background
  ...

  ## How You Think
  ...

  ## How You Talk
  ...
```

### Skill Format

```yaml
kind: skill
name: my-skill
version: 1.0.0
description: What this skill does
author: your-github-username
tags: [relevant, tags]

requires:
  binaries: [required-cli-tools]
  env: [REQUIRED_ENV_VARS]

tools:
  - name: tool_name
    description: What this tool does
    params:
      param_name:
        type: string
        required: true
        description: What this param is for
    run: |
      command --flag {{ param_name }}
```

## Contributing

1. Fork this repo
2. Add your skill/persona in the appropriate directory
3. Update the relevant `index.yaml`
4. Submit a PR

### Guidelines

- **Personas**: Start with "You are [Name]" for automatic name extraction
- **Skills**: Focus on a specific domain, include good tool descriptions
- **Security**: No credential harvesting, no destructive commands without confirmation
- **Quality**: Test before submitting

## License

MIT
