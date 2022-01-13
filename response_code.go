package x

const (
	SuccessCode    = "00"
	SuccessMessage = "Success"

	ErrorFormatCode       = "30"
	ErrorFormatMessage    = "Format Error"
	ErrorUndefinedCode    = "05"
	ErrorUndefinedMessage = "Undefined Error"
)

var (
	ResponseMap = map[string]map[string]string{
		DomainInternal: {
			SuccessCode:     SuccessMessage,
			ErrorFormatCode: ErrorFormatMessage,
		},
		DomainProspera: {
			SuccessCode:     SuccessMessage,
			ErrorFormatCode: ErrorFormatMessage,
		},
		DomainTemenos: {
			SuccessCode:     SuccessMessage,
			ErrorFormatCode: ErrorFormatMessage,
		},
		DomainWow: {
			SuccessCode:     SuccessMessage,
			ErrorFormatCode: ErrorFormatMessage,
		},
	}
)
