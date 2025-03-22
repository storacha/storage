-- Create or replace all required functions first.
CREATE OR REPLACE FUNCTION increment_proofset_refcount()
    RETURNS TRIGGER AS $$
BEGIN
    UPDATE pdp_piecerefs
    SET proofset_refcount = proofset_refcount + 1
    WHERE id = NEW.pdp_pieceref;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION decrement_proofset_refcount()
    RETURNS TRIGGER AS $$
BEGIN
    UPDATE pdp_piecerefs
    SET proofset_refcount = proofset_refcount - 1
    WHERE id = OLD.pdp_pieceref;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

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
        WHERE create_message_hash = NEW.signed_tx_hash
          AND proofset_created = FALSE;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

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

-----------------------------------------------------------------------
-- Now conditionally create each trigger only if it doesn't already exist
-----------------------------------------------------------------------
DO $$
    BEGIN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_trigger t
                     JOIN pg_class c ON t.tgrelid = c.oid
            WHERE t.tgname = 'pdp_proofset_root_insert'
              AND c.relname = 'pdp_proofset_roots'
        ) THEN
            CREATE TRIGGER pdp_proofset_root_insert
                AFTER INSERT ON pdp_proofset_roots
                FOR EACH ROW
                WHEN (NEW.pdp_piece_ref_id IS NOT NULL)
            EXECUTE FUNCTION increment_proofset_refcount();
        END IF;

        IF NOT EXISTS (
            SELECT 1
            FROM pg_trigger t
                     JOIN pg_class c ON t.tgrelid = c.oid
            WHERE t.tgname = 'pdp_proofset_root_delete'
              AND c.relname = 'pdp_proofset_roots'
        ) THEN
            CREATE TRIGGER pdp_proofset_root_delete
                AFTER DELETE ON pdp_proofset_roots
                FOR EACH ROW
                WHEN (OLD.pdp_piece_ref_id IS NOT NULL)
            EXECUTE FUNCTION decrement_proofset_refcount();
        END IF;

        IF NOT EXISTS (
            SELECT 1
            FROM pg_trigger t
                     JOIN pg_class c ON t.tgrelid = c.oid
            WHERE t.tgname = 'pdp_proofset_root_update'
              AND c.relname = 'pdp_proofset_roots'
        ) THEN
            CREATE TRIGGER pdp_proofset_root_update
                AFTER UPDATE ON pdp_proofset_roots
                FOR EACH ROW
            EXECUTE FUNCTION adjust_proofset_refcount_on_update();
        END IF;

        IF NOT EXISTS (
            SELECT 1
            FROM pg_trigger t
                     JOIN pg_class c ON t.tgrelid = c.oid
            WHERE t.tgname = 'pdp_proofset_create_message_status_change'
              AND c.relname = 'message_waits_eth'
        ) THEN
            CREATE TRIGGER pdp_proofset_create_message_status_change
                AFTER UPDATE OF tx_status, tx_success ON message_waits_eth
                FOR EACH ROW
            EXECUTE PROCEDURE update_pdp_proofset_creates();
        END IF;

        IF NOT EXISTS (
            SELECT 1
            FROM pg_trigger t
                     JOIN pg_class c ON t.tgrelid = c.oid
            WHERE t.tgname = 'pdp_proofset_add_message_status_change'
              AND c.relname = 'message_waits_eth'
        ) THEN
            CREATE TRIGGER pdp_proofset_add_message_status_change
                AFTER UPDATE OF tx_status, tx_success ON message_waits_eth
                FOR EACH ROW
            EXECUTE PROCEDURE update_pdp_proofset_roots();
        END IF;
    END
$$;
