# Getting Started with Piri

This guide helps you choose the right Piri deployment path based on your specific needs and existing infrastructure.

## 🤔 First, Let's Understand Your Needs

Answer these questions to find your optimal setup path:

### Question 1: Do you have existing Filecoin infrastructure?

**No, I'm new to Filecoin storage**
→ Continue to Question 2

**Yes, I run Curio (or Lotus-Miner)**
→ You should integrate Piri with your existing infrastructure
→ 📖 **Follow**: [Piri with Curio Integration Guide](./integrations/piri-with-curio.md)

### Question 2: What's your primary goal?

**I want to join the Storacha network as a storage provider**
→ You need the full stack (UCAN + PDP servers)
→ 📖 **Follow**: [Full Stack Setup Guide](./integrations/full-stack-setup.md)

**I want to build/test storage applications**
→ Continue to Question 3

**I'm evaluating the technology**
→ 📖 **Start with**: [Architecture Overview](./architecture.md)
→ 📖 **Then try**: Local development setup in [Full Stack Guide](./integrations/full-stack-setup.md)

### Question 3: Which component do you need?

**Client-facing API for uploads (UCAN)**
→ 📖 **Follow**: [UCAN Server Setup Guide](./guides/ucan-server.md)
→ Note: You'll need access to a PDP backend

**Storage and proof generation (PDP)**
→ 📖 **Follow**: [PDP Server Setup Guide](./guides/pdp-server-piri.md)
→ Note: You'll need a funded Filecoin wallet

**Both components**
→ 📖 **Follow**: [Full Stack Setup Guide](./integrations/full-stack-setup.md)

## 📊 Quick Decision Matrix

| Your Situation | Recommended Path | Guide |
|----------------|------------------|-------|
| New storage provider | Full Piri stack | [Full Stack Setup](./integrations/full-stack-setup.md) |
| Existing Curio operator | Add UCAN server only | [Piri with Curio](./integrations/piri-with-curio.md) |
| Developer testing uploads | UCAN server + test PDP | [UCAN Server Setup](./guides/ucan-server.md) |
| Building storage backend | PDP server only | [PDP Server Setup](./guides/pdp-server-piri.md) |
| Learning the system | Read architecture first | [Architecture Overview](./architecture.md) |

## 🏗️ Deployment Architectures

### Option 1: Full Piri Stack (Recommended for New Providers)

```
Internet → UCAN Server → PDP Server → Filecoin
         (Port 3000)   (Port 3001)    Network
```

**Pros:**
- Complete control over pipeline
- Single identity management
- Simplified troubleshooting

**Requirements:**
- Single server or VM
- Domain name with DNS
- Funded Filecoin wallet
- Lotus node access

### Option 2: Piri UCAN + Existing Curio

```
Internet → UCAN Server → Curio → Filecoin
         (Piri)        (Existing) Network
```

**Pros:**
- Leverage existing investment
- Proven PDP implementation
- Gradual migration path

**Requirements:**
- Running Curio instance
- Domain for UCAN server
- Curio API access

### Option 3: Component Testing

```
Test Client → Component → Mock/Test Backend
```

**Pros:**
- Isolated testing
- Minimal requirements
- Quick iteration

**Requirements:**
- Development machine
- Basic tooling

## 🚦 Before You Begin

### Essential Prerequisites

Regardless of your path, you'll need:

1. **System Requirements**
   - Linux/macOS/Windows (WSL2)
   - 4GB+ RAM
   - 20GB+ disk space
   - See [full prerequisites](./common/prerequisites.md)

2. **Software Dependencies**
   - Go 1.23+
   - Git, Make, jq
   - See [installation guide](./common/piri-installation.md)

3. **Network Requirements**
   - Domain name (production)
   - Open ports 80/443
   - See [TLS setup](./common/tls-termination.md)

### Component-Specific Requirements

**For PDP Server:**
- Lotus node (Calibration or Mainnet)
- Funded delegated wallet
- Storage space for pieces

**For UCAN Server:**
- PDP backend endpoint
- Storacha delegation
- Ed25519 PEM key

## 🎯 Next Steps

### 1. Prepare Your Environment

```bash
# Check prerequisites
go version  # Should be 1.23+
git --version
make --version

# Install Piri
git clone https://github.com/storacha/piri
cd piri
make calibnet
```

### 2. Generate Your Identity

All deployments need a cryptographic identity:

```bash
# Generate PEM file
piri id gen -t=pem > service.pem

# Extract your DID
piri id parse service.pem | jq .did
```

See [PEM File Generation](./common/pem-file-generation.md) for details.

### 3. Follow Your Chosen Guide

Based on your decision above, proceed to the appropriate guide:

- 🏢 **Full Stack**: [Complete setup guide](./integrations/full-stack-setup.md)
- 🔧 **With Curio**: [Integration guide](./integrations/piri-with-curio.md)
- 📥 **UCAN Only**: [UCAN server guide](./guides/ucan-server.md)
- 💾 **PDP Only**: [PDP server guide](./guides/pdp-server-piri.md)

## ❓ Still Unsure?

If you're still not sure which path to take:

1. **Read** the [Architecture Overview](./architecture.md) to understand the system
2. **Try** the local development setup in the Full Stack guide
3. **Ask** questions in the community channels
4. **Start small** with a single component and expand

Remember: You can always start with one approach and migrate later. The modular architecture supports various deployment patterns.

## 🆘 Need Help?

- **Documentation**: You're here! Browse the guides
- **Issues**: [GitHub Issues](https://github.com/storacha/piri/issues)
- **Community**: Join Storacha community channels

Happy storing with Piri! 🚀