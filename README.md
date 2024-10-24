# pt-tools

This is a Go command line program that will allow listing and interacting with a Pairtree without knowing anything about the Pairtree’s internal structure. When using pt-tools, we highly suggest utilizing a pairtree prefix for schemes such as `ark:/`. When not using a defined prefix, denote the pairtree ID with `pt://`, so an ID should look like `pt://ID`. This prefix will be stripped from an ID before interacting with  the pairtree. If no prefix is provided attached to the ID or `pt://` is not used, an error will be thrown.   

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

Ptcp is a cp-like tool that can copy files and folders in and out of the Pairtree structure. Unlike Linux's cp, the default is recursive. Ptcp's defualt behavior will also not overwrite files or directories if they already exist at the specificed location. Instead, it will add `.x` (x being an integer that starts from 1) to the path. 

This cp tool behaves similarly to the unix cp in relation to directories. This means that when you are copying from `SRC` to `DEST` if the folder does not exist, the folder will be created and only the contents of the src will be copied into `DEST`. If the folder does exist, the folder and its contents from `SRC` will be moved into `DEST`. When wanting to copy into a subpath or new path in the pairtree, the `-n` flag will need to be used and is detailed below. When trying to copy a single file from the pairtree into a `Dest` that does not exist, an erorr will be thrown. At the moment the ability to copy from one pairtree to another pairtree is not available. 

When an ENV PAIRTREE_ROOT is set, the basic command is
    
    ptcp [SRC] [DEST]

so to copy from a paitree ID to an output path 

    ptcp [ID] [/path/to/output]

When an ENV PAIRTREE_ROOT is not set, the command is 

    ptcp -p [PT_ROOT] [ID] [/path/to/output]

To overwrite target files that already exist in the destination use the `-d` option. It runs with the same option if ENV PAIRTREE_ROOT is set or not set.

    ptcp -d [/path/to/output/] [ID]
                                        
The `-n` option allows you to access subdirectories in the pairtree object. To modify the path of the file or directory when you are copying into the pairtree, the subpath follows `-n` and then will be added to the ID. The `-n` option should be used if you want to place the file or directory in a subpath within the ID or if you want to change the file or directory name that is copied. If the path folowing `-n` does not exist, it will be created in the pairtree. It also alows you to copy a file or directory that is in a subpath in the pairtree object. The file or directory at the end of the `-n` subpath will be the one copied into the destination source. If the file or directory does not exist an error will be returned. The command to create a new directory or place things into an existing directory would be 

    ptcp [/path/to/output] [ID] -n [newpath/in/ID]

When changing a file name copied into the `ID` 

    ptcp [/path/to/outputfile] [ID] -n [newpath/newFileName]

When specificying a file or directory that is being copied from the ID

    ptcp [ID] [/path/to/output] -n [path/to/file/or/directory]

To produce a tar/gzipped output or unpack a tar/gzipped in the pairtree structure run 

    ptcp -a [/path/to/ID.tgz] [ID]

or 

    ptcp -a [ID] [/path/to/dest]

This provides a way to archive an item from the pairtree and un-archive it again back into a pairtree structure, but it's not intended as a way to create archives within the pairtree structure. Only the entire object can be archived from the pairtree meaning the `-a` and `-n` flags should never be used together. When an object is archived, a `.tgz` file will be created and named after the Pairtree object. It will contain a folder that is named the object ID. Unless otherwise specific with the `-d` option, the `.x` pattern will be followed so as not to overwrite other existing `.tgz` files that are named the same. When unarchiving a file into the pairtree, the `.tgz` file should contain a folder named after the pairtree object. The contents of that folder will fully overwrite the contents in the pairtree object. 

## ptmv

Ptmv is a mv-like tool that can move files in and out of the Pairtree structure. Ptmv operates similarly to ptcp except it is destructive, removing the "from" source and overwriting the "to" destination (so deleting the existing directory, if there is one). Ptmv only works on the directory/Pairtree object level and not at the level of files within the Pairtree object, so all sources and targets should represent directories instead of individual files. 

When an ENV PAIRTREE_ROOT, the basic command is
    
    ptmv [ID] [/path/to/output/]

When an ENV PAIRTREE_ROOT is not set, the command is 

    ptmv [PT_ROOT] [ID] [/path/to/output/]

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
