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

// pdp_services
type PDPService struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	Pubkey       []byte    `gorm:"not null;unique"`
	ServiceLabel string    `gorm:"not null;unique"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP;not null"`
}

func (PDPService) TableName() string {
	return "pdp_services"
}

// pdp_piece_uploads
type PDPPieceUpload struct {
	ID             string      `gorm:"primaryKey;type:uuid"` // or use a UUID type
	Service        string      `gorm:"not null"`             // references pdp_services(service_label)
	ServiceModel   *PDPService `gorm:"foreignKey:Service;references:ServiceLabel;constraint:OnDelete:CASCADE"`
	CheckHashCodec string      `gorm:"not null"`
	CheckHash      []byte      `gorm:"not null"`
	CheckSize      int64       `gorm:"not null"`
	PieceCID       *string     `gorm:"column:piece_cid"`
	NotifyURL      string      `gorm:"not null"`
	NotifyTaskID   *int64      // references harmony_task(id)
	PieceRef       *int64      // references parked_piece_refs(ref_id)
	CreatedAt      time.Time   `gorm:"default:CURRENT_TIMESTAMP;not null"`
}

func (PDPPieceUpload) TableName() string {
	return "pdp_piece_uploads"
}

// pdp_piecerefs
type PDPPieceRef struct {
	ID               int64       `gorm:"primaryKey;autoIncrement"`
	Service          string      `gorm:"not null"` // references pdp_services(service_label)
	ServiceModel     *PDPService `gorm:"foreignKey:Service;references:ServiceLabel;constraint:OnDelete:CASCADE"`
	PieceCID         string      `gorm:"not null;column:piece_cid"`
	PieceRef         int64       // references parked_piece_refs(ref_id)
	CreatedAt        time.Time   `gorm:"default:CURRENT_TIMESTAMP;not null"`
	ProofsetRefcount int64       `gorm:"default:0;not null"`
}

func (PDPPieceRef) TableName() string {
	return "pdp_piecerefs"
}

// pdp_piece_mh_to_commp
type PDPPieceMHToCommp struct {
	Mhash []byte `gorm:"primaryKey"` // BYTEA primary key
	Size  int64  `gorm:"not null"`
	Commp string `gorm:"not null"`
}

func (PDPPieceMHToCommp) TableName() string {
	return "pdp_piece_mh_to_commp"
}

// pdp_proof_sets
type PDPProofSet struct {
	ID                        int64 `gorm:"primaryKey"` // big int
	PrevChallengeRequestEpoch *int64
	ChallengeRequestTaskID    *int64 `gorm:"column:challenge_request_task_id"`
	ChallengeRequestTask      *Task  `gorm:"foreignKey:ChallengeRequestTaskID;references:ID;constraint:OnDelete:SET NULL"`
	ChallengeRequestMsgHash   *string
	ProvingPeriod             *int64
	ChallengeWindow           *int64
	ProveAtEpoch              *int64
	InitReady                 bool        `gorm:"default:false;not null"`
	CreateMessageHash         string      `gorm:"not null"`
	Service                   string      `gorm:"not null"` // references pdp_services(service_label)
	ServiceModel              *PDPService `gorm:"foreignKey:Service;references:ServiceLabel;constraint:OnDelete:RESTRICT"`
}

func (PDPProofSet) TableName() string {
	return "pdp_proof_sets"
}

// pdp_prove_tasks (composite PK)
type PDPProveTask struct {
	ProofsetID int64 `gorm:"primaryKey"` // references pdp_proof_sets(id)
	TaskID     int64 `gorm:"primaryKey"` // references harmony_task(id)
}

func (PDPProveTask) TableName() string {
	return "pdp_prove_tasks"
}

// pdp_proofset_creates
type PDPProofsetCreate struct {
	CreateMessageHash string      `gorm:"primaryKey"` // references message_waits_eth(signed_tx_hash)
	Ok                *bool       // NULL / TRUE / FALSE
	ProofsetCreated   bool        `gorm:"default:false;not null"`
	Service           string      `gorm:"not null"` // references pdp_services(service_label)
	ServiceModel      *PDPService `gorm:"foreignKey:Service;references:ServiceLabel;constraint:OnDelete:CASCADE"`
	CreatedAt         time.Time   `gorm:"default:CURRENT_TIMESTAMP;not null"`
}

func (PDPProofsetCreate) TableName() string {
	return "pdp_proofset_creates"
}

// pdp_proofset_roots (composite PK)
type PDPProofsetRoot struct {
	ProofsetID       int64  `gorm:"primaryKey"` // references pdp_proof_sets(id)
	RootID           int64  `gorm:"primaryKey"`
	SubrootOffset    int64  `gorm:"primaryKey"`
	Root             string `gorm:"not null"`
	AddMessageHash   string `gorm:"not null"` // references message_waits_eth(signed_tx_hash)
	AddMessageIndex  int64
	Subroot          string `gorm:"not null"`
	SubrootOffsetVal int64  // same as SubrootOffset, but for clarity if needed
	SubrootSize      int64
	PDPPieceRefID    *int64 // references pdp_piecerefs(id)
}

func (PDPProofsetRoot) TableName() string {
	return "pdp_proofset_roots"
}

// pdp_proofset_root_adds (composite PK)
type PDPProofsetRootAdd struct {
	ProofsetID      int64  `gorm:"primaryKey;column:proofset_id"` // references pdp_proof_sets(id)
	AddMessageHash  string `gorm:"primaryKey"`                    // references message_waits_eth(signed_tx_hash)
	SubrootOffset   int64  `gorm:"primaryKey"`
	Root            string `gorm:"not null"`
	AddMessageOK    *bool
	AddMessageIndex *int64
	Subroot         string `gorm:"not null"`
	SubrootSize     int64
	PDPPieceRefID   *int64 // references pdp_piecerefs(id)
}

func (PDPProofsetRootAdd) TableName() string {
	return "pdp_proofset_root_adds"
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
	SignedHash   *string    `gorm:"column:signed_hash"`
	SendTime     *time.Time `gorm:"column:send_time"`
	SendSuccess  *bool      `gorm:"column:send_success"`
	SendError    *string    `gorm:"column:send_error"`
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
	WaiterMachineID      *int64         `gorm:"column:waiter_machine_id"`
	SignedTxHash         string         `gorm:"primaryKey;column:signed_tx_hash;not null"`
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

func Ptr[T any](v T) *T {
	return &v
}
