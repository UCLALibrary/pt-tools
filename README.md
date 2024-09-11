# pt-tools

This is a Go command line program that will allow listing and interacting with a Pairtree without knowing anything about the Pairtree’s internal structure. 

## Installation

First, ensure that you have Go Version 1.22 on you system. Clone this repository

    git clone https://github.com/UCLALibrary/pt-tools.git

and run 

    go build -o pt-tools main.go

or to build and run tests and checkstyles use the command 
    
    make

## ptls 

Ptls is a ls-like tool that can display the contents of the Pairtree object. The basic command is `ptls [ID]` (when an ENV PAIRTREE_ROOT is set) or `ptls [PT_ROOT] [ID]` with the output listing the contents of the Pairtree object directory. This pattern holds with all options of `ptls` except `ptls -h`. No flags need to be used, but all flags can be used depending on user needs.  

The basic command is  

    ptls "[ID]"

or when the ENV PAIRTREE_ROOT is not set 

    ptls --pairtree [PT_ROOT] "[ID]"

or 

    ptls -p [PT_ROOT] "[ID]"

For ptls help run 

    ptls -h

To list all files including . and .. directories run 

    ptls -a

To list directories of the object directory run 

    ptls -d

To return output in a JSON structure instead of a string output run 

    ptls -j

To output a recursive listing of the object direcotry, with the default being a non-recurseive listing run: 

    ptls -r

## ptcp

Ptcp is a cp-like tool that can copy files in and out of the Pairtree structure.Unlike Linux's cp, the default is recursive. 

When an ENV PAIRTREE_ROOT, the basic command is
    
    ptcp [ID] [/path/to/output/]

When an ENV PAIRTREE_ROOT is not set, the command is 

    ptcp [PT_ROOT] [ID] [/path/to/output/]

To overwrite any target files that already exist in the destination. It runs with the same option if ENV PAIRTREE_ROOT is set or not set.

    ptcp -d 

or 
    ptcp -d [PT_ROOT] [ID] [/path/to/output/]

To specify the path from which to remove files when copying into the Pairtree, use -t as a prefix 

    ptcp -t [/path/to/] [/path/to/file.ext] [ID]

To produce a tar/gzipped output or upack a tar/gzipped in the pairtree structure run 

    ptcp -a [/path/to/ID.tgz] [ID]

This provides a way to archive an item from the pairtree and un-archive it again back into a pairtree structure, but it's not intended as a way to create archives within the pairtree structure. 

## ptmv

Ptmv is a mv-like tool that can move files in and out of the Pairtree structure. Ptmv operates similarly to ptcp except it is destructive, removing the "from" source and overwriting the "to" destination (so deleting the existing directory, if there is one). Ptmv only works on the directory/Pairtree object level and not at the level of files within the Pairtree object, so all sources and targets should represent directories instead of individual files. 

When an ENV PAIRTREE_ROOT, the basic command is
    
    ptmv [ID] [/path/to/output/]

When an ENV PAIRTREE_ROOT is not set, the command is 

    ptmv [PT_ROOT] [ID] [/path/to/output/]

To specify the path from which to remove files when copying into the Pairtree, use -t as a prefix 

    ptmv -t [/path/to/] [/path/to/file.ext] [ID]

To produce a tar/gzipped output or upack a tar/gzipped in the pairtree structure run 

    ptmv -a [/path/to/ID.tgz] [ID]

## ptrm

Ptrm is a rm-like tool that can delete things from within a Pairtree object or remove a Pairtree object altogether. There is also the ability to delete files and directories in the object as long as the subpath to that file or directory is provided. 

The basic command is 

    ptrm [ID]

Or when the ENV PAIRTREE_ROOT is not set 

    ptrm [PT_ROOT] [ID]

To delete a specific file from the pairtree use 

    ptrm [ID] [subpath/to/file.txt]

To delete a specific directory from the pairtree use 

    ptrm [ID] [subpath/to/directory]

Or when the ENV PAIRTREE_ROOT is not set 

    ptrm [PT_ROOT] [ID] [subpath/to/file.txt]
