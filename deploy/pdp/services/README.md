## Service Architecture
Below is a high-level explanation of how these services are ordered and how they depend on each other. Each service has a distinct role in bringing up a fully functional PDP node using Lotus, Curio, and YugabyteDB.

## Yugabyte
### 1. `yugabyte.service`
- Starts the single-node YugabyteDB database (yugabyted).
- Depends on `network-online.target` to ensure the instance’s network is ready.
- Runs continuously (Type=simple).
- If it crashes, systemd restarts it.

### 2. `yugabyte-ready.service`
- One-time check that polls port 5433 to confirm YugabyteDB is actually up and listening.
- Depends on `yugabyte.service` and wants it to be started first.
- Only exits successfully when YugabyteDB is fully ready.

## Lotus
### 1. `lotus-prestart.service`
- One-time setup that imports a snapshot into Lotus.
- Runs before `lotus.service` and depends only on a functioning network.
- After it completes, Lotus can start its daemon with up-to-date chain data.
- Remains in “`active` (exited)” after successful completion.

### 2. `lotus.service`
- The main Lotus daemon (Type=simple).
- Depends on `lotus-prestart.service` finishing first (via After/Wants).
- Restarts automatically if it stops unexpectedly.

### 3. `lotus-ready.service`
- One-time check that polls port 1234 to confirm Lotus is listening and ready.
- Depends on `lotus.service`, ensuring it’s already started.
- Signals “Lotus is ready” once it exits successfully.
- Remains in “`active` (exited)” after successful completion.

### 4. `lotus-poststart.service`
- One-time post-start task to import wallets.
- Runs only after Lotus is ready, depending on `lotus-ready.service`.
- Remains in “`active` (exited)” after successful completion.

## Curio
### 1. `curio-prestart.service`
- One-time initialization that configures Curio:
- Ensures Lotus and Yugabyte are ready (depends on `lotus-poststart.service` and `yugabyte-ready.service`).
- Performs commands like curio doit to set up PDP, attach keys, and import configurations.
- This must complete before the main Curio daemon starts.

### 2. `curio.service`
- The main Curio daemon (Type=simple).
- Depends on `curio-prestart.service` finishing successfully.
- Runs continuously once started.

### 3. `curio-ready.service`
- One-time check that polls port 12300 to confirm the Curio daemon is fully up.
- Depends on `curio.service` and only succeeds once Curio is actually listening.

### 4. `curio-poststart.service`
- Post-start tasks (e.g., attaching storage).
- runs after Curio is marked ready.
- Also remains in “`active` (exited)” after successful completion.

# How They Tie Together
Once all “ready” and “poststart” services exit successfully, you have a fully functional PDP node running Curio (backed by Lotus and Yugabyte), each service restarting automatically if needed.
## 1. Network & Initial Setup
- `network-online.target` becomes `active`, allowing services to begin.
- `yugabyte.service` and `lotus-prestart.service` start independently.

## 2. Database and Lotus Daemon
- Yugabyte is brought online.
- Lotus imports its snapshot `(lotus-prestart.service`) and then starts the daemon `(lotus.service`).

## 3. Readiness Checks
- `yugabyte-ready.service` confirms Yugabyte is listening.
- `lotus-ready.service` confirms Lotus is ready.

## 4. Curio Initialization
- With Lotus and Yugabyte ready, `curio-prestart.service` runs to apply Curio’s initial setup (curio doit, PDP configurations, etc.).
- `curio.service` starts to run the daemon.

## 5. Final Checks & Post-Start
- `curio-ready.service` confirms Curio is accepting requests on 12300.
- `curio-poststart.service` may do final tasks such as attaching storage or sealing.








