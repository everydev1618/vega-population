# Vega Population

Shareable skills and personas for [Vega](https://github.com/vegaops/vega) agents.

## Quick Start

```bash
# Search for skills
vega population search kubernetes

# Install a skill
vega population install kubernetes-ops

# Install a persona
vega population install @incident-commander

# Install a complete profile (persona + skills bundle)
vega population install +platform-engineer
```

## What's Here

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

### Personas
Specialized agent personalities:

| Persona | Description |
|---------|-------------|
| `@incident-commander` | Calm, methodical incident response coordinator |
| `@security-analyst` | Security-focused code reviewer |
| `@code-reviewer` | Thorough, constructive code reviewer |
| `@devops-lead` | Infrastructure and deployment expert |
| `@technical-writer` | Documentation specialist |
| `@frontend-mentor` | UI/UX and frontend development guide |

### Profiles
Curated bundles of persona + skills:

| Profile | Description |
|---------|-------------|
| `+platform-engineer` | Full platform engineering toolkit |
| `+startup-cto` | Everything a startup CTO needs |
| `+full-stack-dev` | Frontend + backend + deployment |

## Creating Your Own

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
  You are...

  ## Your approach
  ...
```

## Contributing

1. Fork this repo
2. Add your skill/persona in the appropriate directory
3. Update the relevant `index.yaml`
4. Submit a PR

### Guidelines

- **Skills**: Focus on a specific domain, include good tool descriptions
- **Personas**: Be specific about expertise and approach, avoid generic prompts
- **Security**: No credential harvesting, no destructive commands without confirmation
- **Quality**: Test your skills before submitting

## License

MIT
