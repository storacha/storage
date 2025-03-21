package models

import (
	_ "embed"
	"time"

	"gorm.io/datatypes"
)

//go:embed triggers.sql
var Triggers string

// Machine represents the machines table.
type Machine struct {
	ID          int       `gorm:"primaryKey;autoIncrement;column:id"`
	LastContact time.Time `gorm:"not null;default:current_timestamp;column:last_contact"`
	HostAndPort string    `gorm:"size:300;not null;column:host_and_port"`
	CPU         int       `gorm:"not null;column:cpu"`
	RAM         int64     `gorm:"not null;column:ram"`
	GPU         float64   `gorm:"not null;column:gpu"`
}

func (Machine) TableName() string {
	return "machines"
}

// Task represents the task table.
type Task struct {
	ID           int       `gorm:"primaryKey;autoIncrement;column:id"`
	InitiatedBy  *int      `gorm:"column:initiated_by;comment:The task ID whose completion occasioned this task"`
	UpdateTime   time.Time `gorm:"not null;default:current_timestamp;column:update_time;comment:When it was last modified. not a heartbeat"`
	PostedTime   time.Time `gorm:"not null;column:posted_time"`
	OwnerID      *int      `gorm:"column:owner_id;comment:may be null if between owners or not yet taken"`
	AddedBy      int       `gorm:"not null;column:added_by"`
	PreviousTask *int      `gorm:"column:previous_task"`
	Name         string    `gorm:"size:16;not null;column:name;comment:The name of the task type"`
	// Note: The "retries" field was commented out in the original schema.
}

func (Task) TableName() string {
	return "task"
}

// TaskHistory represents the task_history table.
type TaskHistory struct {
	ID                     int       `gorm:"primaryKey;autoIncrement;column:id"`
	TaskID                 int       `gorm:"not null;column:task_id"`
	Name                   string    `gorm:"size:16;not null;column:name"`
	Posted                 time.Time `gorm:"not null;column:posted"`
	WorkStart              time.Time `gorm:"not null;column:work_start"`
	WorkEnd                time.Time `gorm:"not null;column:work_end"`
	Result                 bool      `gorm:"not null;column:result;comment:Use to detemine if this was a successful run."`
	Err                    string    `gorm:"column:err"`
	CompletedByHostAndPort string    `gorm:"size:300;not null;column:completed_by_host_and_port"`
}

func (TaskHistory) TableName() string {
	return "task_history"
}

// TaskFollow represents the task_follow table.
type TaskFollow struct {
	ID       int    `gorm:"primaryKey;autoIncrement;column:id"`
	OwnerID  int    `gorm:"not null;column:owner_id"`
	ToType   string `gorm:"size:16;not null;column:to_type"`
	FromType string `gorm:"size:16;not null;column:from_type"`
}

func (TaskFollow) TableName() string {
	return "task_follow"
}

// TaskImpl represents the task_impl table.
type TaskImpl struct {
	ID      int    `gorm:"primaryKey;autoIncrement;column:id"`
	OwnerID int    `gorm:"not null;column:owner_id"`
	Name    string `gorm:"size:16;not null;column:name"`
}

func (TaskImpl) TableName() string {
	return "task_impl"
}

// ParkedPiece represents the parked_pieces table.
type ParkedPiece struct {
	ID              int64     `gorm:"primaryKey;autoIncrement;column:id"`
	CreatedAt       time.Time `gorm:"default:current_timestamp;column:created_at"`
	PieceCID        string    `gorm:"not null;column:piece_cid;uniqueIndex:idx_piece_cid_padded_longterm_cleanup"`
	PiecePaddedSize int64     `gorm:"not null;column:piece_padded_size;uniqueIndex:idx_piece_cid_padded_longterm_cleanup"`
	PieceRawSize    int64     `gorm:"not null;column:piece_raw_size"`
	Complete        bool      `gorm:"not null;default:false;column:complete"`
	TaskID          *int64    `gorm:"column:task_id"`
	CleanupTaskID   *int64    `gorm:"column:cleanup_task_id;uniqueIndex:idx_piece_cid_padded_longterm_cleanup"`
	LongTerm        bool      `gorm:"not null;default:false;column:long_term;uniqueIndex:idx_piece_cid_padded_longterm_cleanup"`
}

func (ParkedPiece) TableName() string {
	return "parked_pieces"
}

// ParkedPieceRef represents the parked_piece_refs table.
type ParkedPieceRef struct {
	RefID       int64          `gorm:"primaryKey;autoIncrement;column:ref_id"`
	PieceID     int64          `gorm:"not null;column:piece_id"`
	DataURL     string         `gorm:"column:data_url"`
	DataHeaders datatypes.JSON `gorm:"not null;column:data_headers;default:'{}'"`
	LongTerm    bool           `gorm:"not null;default:false;column:long_term"`
}

func (ParkedPieceRef) TableName() string {
	return "parked_piece_refs"
}

// PDPService represents the pdp_services table.
type PDPService struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;column:id"`
	Pubkey       []byte    `gorm:"not null;column:pubkey;unique"`
	ServiceLabel string    `gorm:"not null;column:service_label;unique"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP;column:created_at"`
}

func (PDPService) TableName() string {
	return "pdp_services"
}

// PDPPieceUpload represents the pdp_piece_uploads table.
type PDPPieceUpload struct {
	ID             string    `gorm:"type:uuid;primaryKey;column:id"`
	Service        string    `gorm:"not null;column:service"`
	CheckHashCodec string    `gorm:"not null;column:check_hash_codec"`
	CheckHash      []byte    `gorm:"not null;column:check_hash"`
	CheckSize      int64     `gorm:"not null;column:check_size"`
	PieceCID       string    `gorm:"column:piece_cid"`
	NotifyURL      string    `gorm:"not null;column:notify_url"`
	NotifyTaskID   *int64    `gorm:"column:notify_task_id"`
	PieceRef       *int64    `gorm:"column:piece_ref"`
	CreatedAt      time.Time `gorm:"default:CURRENT_TIMESTAMP;column:created_at"`
}

func (PDPPieceUpload) TableName() string {
	return "pdp_piece_uploads"
}

// PDPPieceRef represents the pdp_piecerefs table.
type PDPPieceRef struct {
	ID               int64     `gorm:"primaryKey;autoIncrement;column:id"`
	Service          string    `gorm:"not null;column:service"`
	PieceCID         string    `gorm:"not null;column:piece_cid"`
	PieceRef         int64     `gorm:"not null;column:piece_ref;unique"`
	CreatedAt        time.Time `gorm:"default:CURRENT_TIMESTAMP;column:created_at"`
	ProofsetRefCount int64     `gorm:"not null;default:0;column:proofset_refcount"`
}

func (PDPPieceRef) TableName() string {
	return "pdp_piecerefs"
}

// PDPProofSet represents the pdp_proof_sets table.
type PDPProofSet struct {
	ID                        int64   `gorm:"primaryKey;column:id;not null"`
	PrevChallengeRequestEpoch *int64  `gorm:"column:prev_challenge_request_epoch"`
	ChallengeRequestTaskID    *int64  `gorm:"column:challenge_request_task_id"`
	ChallengeRequestMsgHash   *string `gorm:"column:challenge_request_msg_hash"`
	ProvingPeriod             int64   `gorm:"not null;column:proving_period"`
	ChallengeWindow           int64   `gorm:"not null;column:challenge_window"`
	ProveAtEpoch              *int64  `gorm:"column:prove_at_epoch"`
	InitReady                 bool    `gorm:"not null;default:false;column:init_ready"`
	CreateMessageHash         string  `gorm:"not null;column:create_message_hash"`
	Service                   string  `gorm:"not null;column:service"`
}

func (PDPProofSet) TableName() string {
	return "pdp_proof_sets"
}

// PDPProveTask represents the pdp_prove_tasks table.
type PDPProveTask struct {
	Proofset int64 `gorm:"not null;column:proofset;primaryKey"`
	TaskID   int64 `gorm:"not null;column:task_id;primaryKey"`
}

func (PDPProveTask) TableName() string {
	return "pdp_prove_tasks"
}

// PDPProofsetCreate represents the pdp_proofset_creates table.
type PDPProofsetCreate struct {
	CreateMessageHash string    `gorm:"primaryKey;column:create_message_hash"`
	Ok                *bool     `gorm:"column:ok"`
	ProofsetCreated   bool      `gorm:"not null;default:false;column:proofset_created"`
	Service           string    `gorm:"not null;column:service"`
	CreatedAt         time.Time `gorm:"default:CURRENT_TIMESTAMP;column:created_at"`
}

func (PDPProofsetCreate) TableName() string {
	return "pdp_proofset_creates"
}

// PDPProofsetRoot represents the pdp_proofset_roots table.
// Primary key is composite: (proofset, root_id, subroot_offset)
type PDPProofsetRoot struct {
	Proofset        int64  `gorm:"not null;column:proofset;primaryKey"`
	RootID          int64  `gorm:"not null;column:root_id;primaryKey"`
	SubrootOffset   int64  `gorm:"not null;column:subroot_offset;primaryKey"`
	AddMessageHash  string `gorm:"not null;column:add_message_hash"`
	AddMessageIndex int64  `gorm:"not null;column:add_message_index"`
	Root            string `gorm:"not null;column:root"`
	Subroot         string `gorm:"not null;column:subroot"`
	SubrootSize     int64  `gorm:"not null;column:subroot_size"`
	PDPPieceRef     int64  `gorm:"not null;column:pdp_pieceref"`
}

func (PDPProofsetRoot) TableName() string {
	return "pdp_proofset_roots"
}

// PDPProofsetRootAdd represents the pdp_proofset_root_adds table.
// Primary key is composite: (proofset, add_message_hash, subroot_offset)
type PDPProofsetRootAdd struct {
	AddMessageHash  string `gorm:"not null;column:add_message_hash;primaryKey"`
	SubrootOffset   int64  `gorm:"not null;column:subroot_offset;primaryKey"`
	Proofset        int64  `gorm:"not null;column:proofset"`
	Root            string `gorm:"not null;column:root"`
	AddMessageOk    *bool  `gorm:"column:add_message_ok"`
	AddMessageIndex int64  `gorm:"not null;column:add_message_index"`
	Subroot         string `gorm:"not null;column:subroot"`
	SubrootSize     int64  `gorm:"not null;column:subroot_size"`
	PDPPieceRef     int64  `gorm:"not null;column:pdp_pieceref"`
}

func (PDPProofsetRootAdd) TableName() string {
	return "pdp_proofset_root_adds"
}

// PDPPieceMHToCommp represents the pdp_piece_mh_to_commp table.
type PDPPieceMHToCommp struct {
	Mhash []byte `gorm:"primaryKey;column:mhash"`
	Size  int64  `gorm:"not null;column:size"`
	Commp string `gorm:"not null;column:commp"`
}

// TableName specifies the table name for PDPPieceMHToCommp.
func (PDPPieceMHToCommp) TableName() string {
	return "pdp_piece_mh_to_commp"
}

// EthKey represents the eth_keys table.
type EthKey struct {
	Address    string `gorm:"primaryKey;column:address;not null"`
	PrivateKey []byte `gorm:"not null;column:private_key"`
	Role       string `gorm:"not null;column:role"`
}

func (EthKey) TableName() string {
	return "eth_keys"
}

// MessageSendsEth represents the message_sends_eth table.
type MessageSendsEth struct {
	FromAddress  string     `gorm:"not null;column:from_address"`
	ToAddress    string     `gorm:"not null;column:to_address"`
	SendReason   string     `gorm:"not null;column:send_reason"`
	SendTaskID   int        `gorm:"primaryKey;autoIncrement;column:send_task_id"`
	UnsignedTx   []byte     `gorm:"not null;column:unsigned_tx"`
	UnsignedHash string     `gorm:"not null;column:unsigned_hash"`
	Nonce        *int64     `gorm:"column:nonce"`
	SignedTx     []byte     `gorm:"column:signed_tx"`
	SignedHash   string     `gorm:"column:signed_hash"`
	SendTime     *time.Time `gorm:"column:send_time"`
	SendSuccess  *bool      `gorm:"column:send_success"`
	SendError    string     `gorm:"column:send_error"`
}

func (MessageSendsEth) TableName() string {
	return "message_sends_eth"
}

// MessageSendEthLock represents the message_send_eth_locks table.
type MessageSendEthLock struct {
	FromAddress string    `gorm:"primaryKey;column:from_address;not null"`
	TaskID      int64     `gorm:"not null;column:task_id"`
	ClaimedAt   time.Time `gorm:"not null;column:claimed_at"`
}

func (MessageSendEthLock) TableName() string {
	return "message_send_eth_locks"
}

// MessageWaitsEth represents the message_waits_eth table.
type MessageWaitsEth struct {
	SignedTxHash         string         `gorm:"primaryKey;column:signed_tx_hash;not null"`
	WaiterMachineID      *int           `gorm:"column:waiter_machine_id"`
	ConfirmedBlockNumber *int64         `gorm:"column:confirmed_block_number"`
	ConfirmedTxHash      string         `gorm:"column:confirmed_tx_hash"`
	ConfirmedTxData      datatypes.JSON `gorm:"column:confirmed_tx_data"`
	TxStatus             string         `gorm:"column:tx_status"`
	TxReceipt            datatypes.JSON `gorm:"column:tx_receipt"`
	TxSuccess            *bool          `gorm:"column:tx_success"`
}

func (MessageWaitsEth) TableName() string {
	return "message_waits_eth"
}
