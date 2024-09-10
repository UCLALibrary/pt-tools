package error_msgs

import "errors"

var (
	Err1 = errors.New("pairtree_prefix file exists, but is empty and must be populated")
	Err2 = errors.New("the pairtree version file is empty and must be populated")
	Err3 = errors.New("the pairtree root is empty and must be populated")
	Err4 = errors.New("the pairtree id is empty and must be populated ")
	Err5 = errors.New("the pairtree id does not contain the pairtree_prefix")
	Err6 = errors.New("no ID was provided to process")
	Err7 = errors.New("--pairtree flag or PAIRTREE_ROOT environment variable must be set")
	Err8 = errors.New("too many arguments were passed into ptrm")
)
