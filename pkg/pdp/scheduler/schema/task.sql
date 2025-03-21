CREATE TABLE machines (
                                  id SERIAL PRIMARY KEY NOT NULL,
                                  last_contact TIMESTAMP NOT NULL DEFAULT current_timestamp,
                                  host_and_port varchar(300) NOT NULL,
                                  cpu INTEGER NOT NULL,
                                  ram BIGINT NOT NULL,
                                  gpu FLOAT NOT NULL
);

CREATE TABLE task (
                              id SERIAL PRIMARY KEY NOT NULL,
                              initiated_by INTEGER,
                              update_time TIMESTAMP NOT NULL DEFAULT current_timestamp,
                              posted_time TIMESTAMP NOT NULL,
                              owner_id INTEGER REFERENCES machines (id) ON DELETE SET NULL,
                              added_by INTEGER NOT NULL,
                              previous_task INTEGER,
                              name varchar(16) NOT NULL
    -- retries INTEGER NOT NULL DEFAULT 0 -- added later
);
COMMENT ON COLUMN task.initiated_by IS 'The task ID whose completion occasioned this task.';
COMMENT ON COLUMN task.owner_id IS 'The foreign key to machines.';
COMMENT ON COLUMN task.name IS 'The name of the task type.';
COMMENT ON COLUMN task.owner_id IS 'may be null if between owners or not yet taken';
COMMENT ON COLUMN task.update_time IS 'When it was last modified. not a heartbeat';

CREATE TABLE task_history (
                                      id SERIAL PRIMARY KEY NOT NULL,
                                      task_id INTEGER NOT NULL,
                                      name VARCHAR(16) NOT NULL,
                                      posted TIMESTAMP NOT NULL,
                                      work_start TIMESTAMP NOT NULL,
                                      work_end TIMESTAMP NOT NULL,
                                      result BOOLEAN NOT NULL,
                                      err varchar,
                                      completed_by_host_and_port varchar(300) NOT NULL
);
COMMENT ON COLUMN task_history.result IS 'Use to detemine if this was a successful run.';

CREATE TABLE task_follow (
                                     id SERIAL PRIMARY KEY NOT NULL,
                                     owner_id INTEGER NOT NULL REFERENCES machines (id) ON DELETE CASCADE,
                                     to_type VARCHAR(16) NOT NULL,
                                     from_type VARCHAR(16) NOT NULL
);

CREATE TABLE task_impl (
                                   id SERIAL PRIMARY KEY NOT NULL,
                                   owner_id INTEGER NOT NULL REFERENCES machines (id) ON DELETE CASCADE,
                                   name VARCHAR(16) NOT NULL
);

CREATE INDEX task_history_task_id_result_index
    ON task_history (task_id, result);

CREATE TABLE parked_pieces (
                               id bigserial PRIMARY KEY,
                               created_at timestamp DEFAULT current_timestamp,
                               piece_cid text NOT NULL, -- v1
                               piece_padded_size bigint NOT NULL,
                               piece_raw_size bigint NOT NULL,
                               complete boolean NOT NULL DEFAULT false,
                               task_id bigint DEFAULT NULL,
                               cleanup_task_id bigint DEFAULT NULL,
                               long_term BOOLEAN NOT NULL DEFAULT FALSE,
                               FOREIGN KEY (task_id) REFERENCES task (id) ON DELETE SET NULL,
                               FOREIGN KEY (cleanup_task_id) REFERENCES task (id) ON DELETE SET NULL,
                               CONSTRAINT parked_pieces_piece_cid_cleanup_task_id_key UNIQUE (piece_cid, piece_padded_size, long_term, cleanup_task_id)
);

CREATE TABLE parked_piece_refs (
                                   ref_id bigserial PRIMARY KEY,
                                   piece_id bigint NOT NULL,
                                   data_url text,
                                   data_headers jsonb NOT NULL DEFAULT '{}',
                                   long_term BOOLEAN NOT NULL DEFAULT FALSE,
                                   FOREIGN KEY (piece_id) REFERENCES parked_pieces(id) ON DELETE CASCADE
);

/*
 * This table is used to keep track of the references to the parked pieces
 * so that we can delete them when they are no longer needed.
 *
 * All references into the parked_pieces table should be done through this table.
 *
 * data_url is optional for refs which also act as data sources.
 *
 * Refs are ADDED when:
 * 1. MK12 market accepts a non-offline deal
 *
 * Refs are REMOVED when:
 * 1. (MK12) A sector related to a pieceref: url piece is finalized
 * 2. (MK12) A deal pipeline not yet assigned to a sector is deleted
 *
 */

-- PDP tables
-- PDP services authenticate with ecdsa-sha256 keys; Allowed services here
CREATE TABLE pdp_services (
                              id BIGSERIAL PRIMARY KEY,
                              pubkey BYTEA NOT NULL,
    -- service_url TEXT NOT NULL,
                              service_label TEXT NOT NULL,
                              created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                              UNIQUE(pubkey),
                              UNIQUE(service_label)
);

CREATE TABLE pdp_piece_uploads (
                                   id UUID PRIMARY KEY NOT NULL,
                                   service TEXT NOT NULL, -- pdp_services.id
                                   check_hash_codec TEXT NOT NULL, -- hash multicodec used for checking the piece
                                   check_hash BYTEA NOT NULL, -- hash of the piece
                                   check_size BIGINT NOT NULL, -- size of the piece
                                   piece_cid TEXT, -- piece cid v2
                                   notify_url TEXT NOT NULL, -- URL to notify when piece is ready
                                   notify_task_id BIGINT, -- task task ID, moves to pdp_piecerefs and calls notify_url when piece is ready
                                   piece_ref BIGINT, -- parked_piece_refs.ref_id
                                   created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                   FOREIGN KEY (service) REFERENCES pdp_services(service_label) ON DELETE CASCADE,
                                   FOREIGN KEY (piece_ref) REFERENCES parked_piece_refs(ref_id) ON DELETE SET NULL
);

-- PDP piece references, this table tells Curio which pieces in storage are managed by PDP
CREATE TABLE pdp_piecerefs (
                               id BIGSERIAL PRIMARY KEY,
                               service TEXT NOT NULL, -- pdp_services.id
                               piece_cid TEXT NOT NULL, -- piece cid v2
                               piece_ref BIGINT NOT NULL, -- parked_piece_refs.ref_id
                               created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                               proofset_refcount BIGINT NOT NULL DEFAULT 0, -- maintained by triggers
                               UNIQUE(piece_ref),
                               FOREIGN KEY (service) REFERENCES pdp_services(service_label) ON DELETE CASCADE,
                               FOREIGN KEY (piece_ref) REFERENCES parked_piece_refs(ref_id) ON DELETE CASCADE
);

CREATE INDEX pdp_piecerefs_piece_cid_idx ON pdp_piecerefs(piece_cid);

CREATE TABLE pdp_proof_sets (
                                id BIGINT PRIMARY KEY, -- on-chain proofset id
                                prev_challenge_request_epoch BIGINT,
                                challenge_request_task_id BIGINT REFERENCES task(id) ON DELETE SET NULL,
                                challenge_request_msg_hash TEXT,
                                proving_period BIGINT,
                                challenge_window BIGINT,
                                prove_at_epoch BIGINT,
                                init_ready BOOLEAN NOT NULL DEFAULT FALSE,
                                create_message_hash TEXT NOT NULL,
                                service TEXT NOT NULL REFERENCES pdp_services(service_label) ON DELETE RESTRICT
);

CREATE TABLE pdp_prove_tasks (
                                 proofset BIGINT NOT NULL, -- pdp_proof_sets.id
                                 task_id BIGINT NOT NULL, -- task task ID
                                 PRIMARY KEY (proofset, task_id),
                                 FOREIGN KEY (proofset) REFERENCES pdp_proof_sets(id) ON DELETE CASCADE,
                                 FOREIGN KEY (task_id) REFERENCES task(id) ON DELETE CASCADE
);

CREATE TABLE pdp_proofset_creates (
                                      create_message_hash TEXT PRIMARY KEY REFERENCES message_waits_eth(signed_tx_hash) ON DELETE CASCADE,
                                      ok BOOLEAN DEFAULT NULL,
                                      proofset_created BOOLEAN NOT NULL DEFAULT FALSE,
                                      service TEXT NOT NULL REFERENCES pdp_services(service_label) ON DELETE CASCADE,
                                      created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE pdp_proofset_roots (
                                    proofset BIGINT NOT NULL, -- pdp_proof_sets.id
                                    root TEXT NOT NULL, -- root cid (piececid v2)
                                    add_message_hash TEXT NOT NULL REFERENCES message_waits_eth(signed_tx_hash) ON DELETE CASCADE,
                                    add_message_index BIGINT NOT NULL, -- index of root in the add message
                                    root_id BIGINT NOT NULL, -- on-chain index of the root in the rootCids sub-array
                                    subroot TEXT NOT NULL, -- subroot cid (piececid v2), with no aggregation this == root
                                    subroot_offset BIGINT NOT NULL, -- offset of the subroot in the root
                                    subroot_size BIGINT NOT NULL, -- size of the subroot (padded piece size)
                                    pdp_pieceref BIGINT NOT NULL, -- pdp_piecerefs.id
                                    CONSTRAINT pdp_proofset_roots_root_id_unique PRIMARY KEY (proofset, root_id, subroot_offset),
                                    FOREIGN KEY (proofset) REFERENCES pdp_proof_sets(id) ON DELETE CASCADE,
                                    FOREIGN KEY (pdp_pieceref) REFERENCES pdp_piecerefs(id) ON DELETE SET NULL
);

CREATE TABLE pdp_proofset_root_adds (
                                        proofset BIGINT NOT NULL, -- pdp_proof_sets.id
                                        root TEXT NOT NULL, -- root cid (piececid v2)
                                        add_message_hash TEXT NOT NULL REFERENCES message_waits_eth(signed_tx_hash) ON DELETE CASCADE,
                                        add_message_ok BOOLEAN,
                                        add_message_index BIGINT NOT NULL,
                                        subroot TEXT NOT NULL,
                                        subroot_offset BIGINT NOT NULL,
                                        subroot_size BIGINT NOT NULL,
                                        pdp_pieceref BIGINT NOT NULL,
                                        CONSTRAINT pdp_proofset_root_adds_root_id_unique PRIMARY KEY (proofset, add_message_hash, subroot_offset),
                                        FOREIGN KEY (proofset) REFERENCES pdp_proof_sets(id) ON DELETE CASCADE,
                                        FOREIGN KEY (pdp_pieceref) REFERENCES pdp_piecerefs(id) ON DELETE SET NULL
);

CREATE OR REPLACE FUNCTION increment_proofset_refcount()
    RETURNS TRIGGER AS $$
BEGIN
    UPDATE pdp_piecerefs
    SET proofset_refcount = proofset_refcount + 1
    WHERE id = NEW.pdp_pieceref;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER pdp_proofset_root_insert
    AFTER INSERT ON pdp_proofset_roots
    FOR EACH ROW
    WHEN (NEW.pdp_pieceref IS NOT NULL)
EXECUTE FUNCTION increment_proofset_refcount();

CREATE OR REPLACE FUNCTION decrement_proofset_refcount()
    RETURNS TRIGGER AS $$
BEGIN
    UPDATE pdp_piecerefs
    SET proofset_refcount = proofset_refcount - 1
    WHERE id = OLD.pdp_pieceref;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER pdp_proofset_root_delete
    AFTER DELETE ON pdp_proofset_roots
    FOR EACH ROW
    WHEN (OLD.pdp_pieceref IS NOT NULL)
EXECUTE FUNCTION decrement_proofset_refcount();

CREATE OR REPLACE FUNCTION adjust_proofset_refcount_on_update()
    RETURNS TRIGGER AS $$
BEGIN
    IF OLD.pdp_pieceref IS DISTINCT FROM NEW.pdp_pieceref THEN
        IF OLD.pdp_pieceref IS NOT NULL THEN
            UPDATE pdp_piecerefs
            SET proofset_refcount = proofset_refcount - 1
            WHERE id = OLD.pdp_pieceref;
        END IF;
        IF NEW.pdp_pieceref IS NOT NULL THEN
            UPDATE pdp_piecerefs
            SET proofset_refcount = proofset_refcount + 1
            WHERE id = NEW.pdp_pieceref;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER pdp_proofset_root_update
    AFTER UPDATE ON pdp_proofset_roots
    FOR EACH ROW
EXECUTE FUNCTION adjust_proofset_refcount_on_update();

CREATE OR REPLACE FUNCTION update_pdp_proofset_creates()
    RETURNS TRIGGER AS $$
BEGIN
    IF OLD.tx_status = 'pending' AND (NEW.tx_status = 'confirmed' OR NEW.tx_status = 'failed') THEN
        UPDATE pdp_proofset_creates
        SET ok = CASE
                     WHEN NEW.tx_status = 'failed' OR NEW.tx_success = FALSE THEN FALSE
                     WHEN NEW.tx_status = 'confirmed' AND NEW.tx_success = TRUE THEN TRUE
                     ELSE ok
            END
        WHERE create_message_hash = NEW.signed_tx_hash AND proofset_created = FALSE;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER pdp_proofset_create_message_status_change
    AFTER UPDATE OF tx_status, tx_success ON message_waits_eth
    FOR EACH ROW
EXECUTE PROCEDURE update_pdp_proofset_creates();

CREATE OR REPLACE FUNCTION update_pdp_proofset_roots()
    RETURNS TRIGGER AS $$
BEGIN
    IF OLD.tx_status = 'pending' AND (NEW.tx_status = 'confirmed' OR NEW.tx_status = 'failed') THEN
        UPDATE pdp_proofset_root_adds
        SET add_message_ok = CASE
                                 WHEN NEW.tx_status = 'failed' OR NEW.tx_success = FALSE THEN FALSE
                                 WHEN NEW.tx_status = 'confirmed' AND NEW.tx_success = TRUE THEN TRUE
                                 ELSE add_message_ok
            END
        WHERE add_message_hash = NEW.signed_tx_hash;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER pdp_proofset_add_message_status_change
    AFTER UPDATE OF tx_status, tx_success ON message_waits_eth
    FOR EACH ROW
EXECUTE PROCEDURE update_pdp_proofset_roots();

CREATE TABLE eth_keys (
                          address TEXT NOT NULL PRIMARY KEY,
                          private_key BYTEA NOT NULL,
                          role TEXT NOT NULL
);

CREATE TABLE message_sends_eth
(
    from_address  TEXT   NOT NULL,
    to_address    TEXT   NOT NULL,
    send_reason   TEXT   NOT NULL,
    send_task_id  SERIAL PRIMARY KEY,
    unsigned_tx   BYTEA  NOT NULL,
    unsigned_hash TEXT   NOT NULL,
    nonce         BIGINT,
    signed_tx     BYTEA,
    signed_hash   TEXT,
    send_time     TIMESTAMP DEFAULT NULL,
    send_success  BOOLEAN   DEFAULT NULL,
    send_error    TEXT
);
COMMENT ON COLUMN message_sends_eth.from_address IS 'Ethereum 0x... address';
COMMENT ON COLUMN message_sends_eth.to_address IS 'Ethereum 0x... address';
COMMENT ON COLUMN message_sends_eth.send_reason IS 'Optional description of send reason';
COMMENT ON COLUMN message_sends_eth.send_task_id IS 'Task ID of the send task';
COMMENT ON COLUMN message_sends_eth.unsigned_tx IS 'Unsigned transaction data';
COMMENT ON COLUMN message_sends_eth.unsigned_hash IS 'Hash of the unsigned transaction';
COMMENT ON COLUMN message_sends_eth.nonce IS 'Assigned transaction nonce, set while the send task is executing';
COMMENT ON COLUMN message_sends_eth.signed_tx IS 'Signed transaction data, set while the send task is executing';
COMMENT ON COLUMN message_sends_eth.signed_hash IS 'Hash of the signed transaction';
COMMENT ON COLUMN message_sends_eth.send_time IS 'Time when the send task was executed, set after pushing the transaction to the network';
COMMENT ON COLUMN message_sends_eth.send_success IS 'Whether this transaction was broadcasted to the network already, NULL if not yet attempted, TRUE if successful, FALSE if failed';
COMMENT ON COLUMN message_sends_eth.send_error IS 'Error message if send_success is FALSE';

CREATE UNIQUE INDEX message_sends_eth_success_index
    ON message_sends_eth (from_address, nonce)
    WHERE send_success IS NOT FALSE;
COMMENT ON INDEX message_sends_eth_success_index IS
    'message_sends_eth_success_index enforces sender/nonce uniqueness, it is a conditional index that only indexes rows where send_success is not false. This allows us to have multiple rows with the same sender/nonce, as long as only one of them was successfully broadcasted (true) to the network or is in the process of being broadcasted (null).';

CREATE TABLE message_send_eth_locks
(
    from_address TEXT      NOT NULL,
    task_id      BIGINT    NOT NULL,
    claimed_at   TIMESTAMP NOT NULL,
    CONSTRAINT message_send_eth_locks_pk PRIMARY KEY (from_address)
);

CREATE TABLE message_waits_eth (
                                   signed_tx_hash TEXT PRIMARY KEY,
                                   waiter_machine_id INT REFERENCES machines (id) ON DELETE SET NULL,
                                   confirmed_block_number BIGINT,
                                   confirmed_tx_hash TEXT,
                                   confirmed_tx_data JSONB,
                                   tx_status TEXT,
                                   tx_receipt JSONB,
                                   tx_success BOOLEAN
);

CREATE INDEX idx_message_waits_eth_pending
    ON message_waits_eth (waiter_machine_id)
    WHERE waiter_machine_id IS NULL AND tx_status = 'pending';
