# pt-tools

This is a Go command line program that will allow listing and interacting with a Pairtree without knowing anything about the Pairtree’s internal structure. When using pt-tools, we highly suggest utilizing a pairtree prefix for schemes such as `ark:/`. When not using a defined prefix, denote the pairtree ID with `pt://`, so an ID should look like `pt://ID`. This prefix will be stripped from an ID before interacting with  the pairtree. If no prefix is provided attached to the ID or `pt://` is not used, an error will be thrown.   

## Installation

First, ensure that you have the latest Go version on you system. 

### Build from Source

Clone this repository

    git clone https://github.com/UCLALibrary/pt-tools.git

and run 

    go build -o pt main.go

or to build and run tests and checkstyles use the command 
    
    make

### Build with Homebrew

Begin by tapping into our homebrew-pt-tools respository

    brew tap UCLALibrary/homebrew-pt-tools

Next, you'll need to install `pt-tools` on your system. Please ensure that there are no other programs named `pt-tools` installed, as this could lead to conflicts.

    brew install pt-tools

Pt-tools will now be installed and can be run with 
    
    pt-tools 

## pt new

Pt new is a tool that creates a new pairtree. The PAIRTREE_ROOT must be set either with an ENV PAIRTREE_ROOT or with a flag otherwise an error will be thrown. The PAIRTREE_ROOT may contain subdirectories, and if the directories do not exist, they will be created. Setting PARITREE_ROOT to `directory/innerdirectory` would be put the pairtree into `innerdirectory` contained inside of `directory`.

The basic command when the ENV PAIRTREE_ROOT is not set

    pt new --pairtree [PT_ROOT]

or

    pt new -p [PT_ROOT]

The prefix file is left empty unless specificed by the user using the `--prefix` flag.

    pt new --prefix [PREFIX]

or 

    pt new -x [PREFIX]

## pt ls 

Pt ls is a ls-like tool that can display the contents of the Pairtree object. The basic command is `pt ls [ID]` (when an ENV PAIRTREE_ROOT is set) or `pt ls [PT_ROOT] [ID]` with the output listing the contents of the Pairtree object directory. This pattern holds with all options of `pt ls` except `pt ls -h`. No flags need to be used, but all flags can be used depending on user needs.  

The basic command is  

    pt ls "[ID]"

or when the ENV PAIRTREE_ROOT is not set 

    pt ls --pairtree [PT_ROOT] "[ID]"

or 

    pt ls -p [PT_ROOT] "[ID]"

For ls help run 

    pt ls -h

To list all files including . and .. directories run 

    pt ls -a

To list directories of the object directory run 

    pt ls -d

To return output in a JSON structure instead of a string output run 

    pt ls -j

To output a recursive listing of the object direcotry, with the default being a non-recurseive listing run: 

    pt ls -r

## pt cp

Pt cp is a cp-like tool that can copy files and folders in and out of the Pairtree structure. Unlike Linux's cp, the default is recursive. Pt cp's defualt behavior will also not overwrite files or directories if they already exist at the specificed location. Instead, it will add `.x` (x being an integer that starts from 1) to the path. 

This pt cp tool behaves similarly to the unix cp in relation to directories. This means that when you are copying from `SRC` to `DEST` if the folder does not exist, the folder will be created and only the contents of the src will be copied into `DEST`. If the folder does exist, the folder and its contents from `SRC` will be moved into `DEST`. When wanting to copy into a subpath or new path in the pairtree, the `-n` flag will need to be used and is detailed below. At the moment the ability to copy from one pairtree to another pairtree is not available. 

When an ENV PAIRTREE_ROOT is set, the basic command is
    
    pt cp [SRC] [DEST]

so to copy from a paitree ID to an output path 

    pt cp [ID] [/path/to/output]

When an ENV PAIRTREE_ROOT is not set, the command is 

    pt cp -p [PT_ROOT] [ID] [/path/to/output]

To overwrite target files that already exist in the destination use the `-d` option. It runs with the same option if ENV PAIRTREE_ROOT is set or not set.

    pt cp -d [/path/to/output/] [ID]
                                        
The `-n` option allows you to access subdirectories in the pairtree object. To modify the path of the file or directory when you are copying into the pairtree, the subpath follows `-n` and then will be added to the ID. The `-n` option should be used if you want to place the file or directory in a subpath within the ID or if you want to change the file or directory name that is copied. If the path folowing `-n` does not exist, it will be created in the pairtree. It also alows you to copy a file or directory that is in a subpath in the pairtree object. The file or directory at the end of the `-n` subpath will be the one copied into the destination source. If the file or directory does not exist an error will be returned. The command to create a new directory or place things into an existing directory would be 

    pt cp [/path/to/output] [ID] -n [newpath/in/ID]

When changing a file name copied into the `ID` 

    pt cp [/path/to/outputfile] [ID] -n [newpath/newFileName]

When specificying a file or directory that is being copied from the ID

    pt cp [ID] [/path/to/output] -n [path/to/file/or/directory]

To produce a tar/gzipped output or unpack a tar/gzipped in the pairtree structure run 

    pt cp -a [/path/to/ID.tgz] [ID]

or 

    pt cp -a [ID] [/path/to/dest]

This provides a way to archive an item from the pairtree and un-archive it again back into a pairtree structure, but it's not intended as a way to create archives within the pairtree structure. Only the entire object can be archived from the pairtree meaning the `-a` and `-n` flags should never be used together. When an object is archived, a `.tgz` file will be created and named after the Pairtree object. It will contain a folder that is named the object ID. Unless otherwise specific with the `-d` option, the `.x` pattern will be followed so as not to overwrite other existing `.tgz` files that are named the same. When unarchiving a file into the pairtree, the `.tgz` file should contain a folder named after the pairtree object. The contents of that folder will fully overwrite the contents in the pairtree object. 

## pt mv

Pt mv is a mv-like tool that can move files in and out of the Pairtree structure. Pt mv operates similarly to pt cp except it is destructive, removing the "from" source and overwriting the "to" destination (so deleting the existing directory, if there is one). Pt mv only works on the directory/Pairtree object level and not at the level of files within the Pairtree object, so all sources and targets should represent directories instead of individual files. 

When an ENV PAIRTREE_ROOT, the basic command is
    
    pt mv [ID] [/path/to/output/]

When an ENV PAIRTREE_ROOT is not set, the command is 

    pt mv [PT_ROOT] [ID] [/path/to/output/]

To produce a tar/gzipped output or upack a tar/gzipped in the pairtree structure run 

    pt mv -a [/path/to/ID.tgz] [ID]

## pt rm

Pt rm is a rm-like tool that can delete things from within a Pairtree object or remove a Pairtree object altogether. There is also the ability to delete files and directories in the object as long as the subpath to that file or directory is provided. 

The basic command is 

    pt rm [ID]

Or when the ENV PAIRTREE_ROOT is not set 

    pt rm [PT_ROOT] [ID]

To delete a specific file from the pairtree use 

    pt rm [ID] [subpath/to/file.txt]

To delete a specific directory from the pairtree use 

    pt rm [ID] [subpath/to/directory]

Or when the ENV PAIRTREE_ROOT is not set 

    pt rm [PT_ROOT] [ID] [subpath/to/file.txt]
