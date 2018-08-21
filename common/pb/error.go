package pb

type ErrorCode = uint16

const (
	OK          ErrorCode = 0
	FORMATE     ErrorCode = 1
	PBUNMARSHAL ErrorCode = 2
)
