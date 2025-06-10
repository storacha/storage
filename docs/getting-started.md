# Getting Started with Piri

Choose your path:

## 🚀 Running a Storage Provider

### Option 1: Full Piri Stack (New Providers)
Run both UCAN and PDP servers for complete storage provider functionality.

→ 📖 **[Full Stack Setup Guide](./integrations/full-stack-setup.md)**

### Option 2: Piri with Curio (Existing Operators)
Already running Curio? Add just the UCAN server to join Storacha network.

→ 📖 **[Piri with Curio Integration](./integrations/piri-with-curio.md)**

## 👩‍💻 Contributing to Piri

Want to contribute? Check out:
- [Architecture Overview](./architecture.md) - Understand the system
- [GitHub Issues](https://github.com/storacha/piri/issues) - Find tasks to work on
- Set up local development using the [Full Stack Guide](./integrations/full-stack-setup.md)

## Prerequisites

Before starting, ensure you have:
- Go 1.23+, Git, Make, jq
- 4GB+ RAM, 20GB+ disk
- See [detailed prerequisites](./common/prerequisites.md)

## Quick Start

```bash
# Clone and build
git clone https://github.com/storacha/piri
cd piri
make calibnet

# Generate identity
piri id gen -t=pem > service.pem
```

Then follow your chosen guide above.