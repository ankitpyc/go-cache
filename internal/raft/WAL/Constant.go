package wal

const (
	CommandBsize   = 3
	TimestampBSize = 8
	LogIndexBSize
	KlenBSize      = 1
	OperationBSize = 1
	MaxWlogSize    = 4096
)

const (
	TypeJSON byte = iota + 1
	TypeString
	TypeInt
)
