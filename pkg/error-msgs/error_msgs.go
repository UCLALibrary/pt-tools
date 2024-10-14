package error_msgs

import "errors"

var (
	Err1  = errors.New("pairtree_prefix file exists, but is empty and must be populated")
	Err2  = errors.New("the pairtree version file is empty and must be populated")
	Err3  = errors.New("the pairtree root is empty and must be populated")
	Err4  = errors.New("the pairtree id is empty and must be populated ")
	Err5  = errors.New("the pairtree id does not contain the pairtree_prefix or pt://")
	Err6  = errors.New("no ID was provided to process")
	Err7  = errors.New("--pairtree flag or PAIRTREE_ROOT environment variable must be set")
	Err8  = errors.New("too many arguments were passed")
	Err9  = errors.New("a source and destination path must be provided to ptcp")
	Err10 = errors.New("neither the source or destination are a part of the pairtree because neither contains the pairtree prefix")
	Err11 = errors.New("the -n and -a options can not be used together in ptcp")
	Err12 = errors.New("temp directory does not contain exactly one folder")
	Err13 = errors.New("folder name does not match pairtree ID")
)
