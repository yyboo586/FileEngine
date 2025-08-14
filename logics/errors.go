package logics

var (
	ErrFileAlreadyExists = "File with name %s already exists"
	ErrFileNotFound      = "File not found"

	ErrFileInvalidFilename     = "File name %s is invalid"
	ErrFileSizeExceeded        = "File size invalid, maximum allowed size is %d"
	ErrFileExtensionNotAllowed = "File extension invalid, extension %s is not allowed"
)
