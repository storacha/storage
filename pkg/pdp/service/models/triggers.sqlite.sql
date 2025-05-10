-- SQLite trigger functions for pdp refcounting

-- Create trigger for incrementing proofset_refcount
CREATE TRIGGER IF NOT EXISTS pdp_proofset_root_insert
AFTER INSERT ON pdp_proofset_roots
FOR EACH ROW
WHEN (NEW.pdp_piece_ref_id IS NOT NULL)
BEGIN
    UPDATE pdp_piecerefs
    SET proofset_refcount = proofset_refcount + 1
    WHERE id = NEW.pdp_piece_ref_id;
END;

-- Create trigger for decrementing proofset_refcount
CREATE TRIGGER IF NOT EXISTS pdp_proofset_root_delete
AFTER DELETE ON pdp_proofset_roots
FOR EACH ROW
WHEN (OLD.pdp_piece_ref_id IS NOT NULL)
BEGIN
    UPDATE pdp_piecerefs
    SET proofset_refcount = proofset_refcount - 1
    WHERE id = OLD.pdp_piece_ref_id;
END;

-- Create trigger for adjusting proofset_refcount on update
CREATE TRIGGER IF NOT EXISTS pdp_proofset_root_update
AFTER UPDATE ON pdp_proofset_roots
FOR EACH ROW
WHEN (OLD.pdp_piece_ref_id IS NOT NEW.pdp_piece_ref_id)
BEGIN
    -- Decrement old reference if not null
    UPDATE pdp_piecerefs
    SET proofset_refcount = proofset_refcount - 1
    WHERE id = OLD.pdp_piece_ref_id AND OLD.pdp_piece_ref_id IS NOT NULL;
    
    -- Increment new reference if not null
    UPDATE pdp_piecerefs
    SET proofset_refcount = proofset_refcount + 1
    WHERE id = NEW.pdp_piece_ref_id AND NEW.pdp_piece_ref_id IS NOT NULL;
END;

-- Create trigger for updating pdp_proofset_creates
CREATE TRIGGER IF NOT EXISTS pdp_proofset_create_message_status_change
AFTER UPDATE OF tx_status, tx_success ON message_waits_eth
FOR EACH ROW
WHEN (OLD.tx_status = 'pending' AND (NEW.tx_status = 'confirmed' OR NEW.tx_status = 'failed'))
BEGIN
    UPDATE pdp_proofset_creates
    SET ok = CASE
                WHEN NEW.tx_status = 'failed' OR NEW.tx_success = 0 THEN 0
                WHEN NEW.tx_status = 'confirmed' AND NEW.tx_success = 1 THEN 1
                ELSE ok
            END
    WHERE create_message_hash = NEW.signed_tx_hash
      AND proofset_created = 0;
END;

-- Create trigger for updating pdp_proofset_root_adds
CREATE TRIGGER IF NOT EXISTS pdp_proofset_add_message_status_change
AFTER UPDATE OF tx_status, tx_success ON message_waits_eth
FOR EACH ROW
WHEN (OLD.tx_status = 'pending' AND (NEW.tx_status = 'confirmed' OR NEW.tx_status = 'failed'))
BEGIN
    UPDATE pdp_proofset_root_adds
    SET add_message_ok = CASE
                            WHEN NEW.tx_status = 'failed' OR NEW.tx_success = 0 THEN 0
                            WHEN NEW.tx_status = 'confirmed' AND NEW.tx_success = 1 THEN 1
                            ELSE add_message_ok
                         END
    WHERE add_message_hash = NEW.signed_tx_hash;
END;