![Image](./logo.png?raw=true)

react
=====

Commandline tool for executing scripts on directory changes

Usage
-----

If you have a root directory structure like this

    .
    subdir_A/
    subdir_B/subdir_C/.onchange
    subdir_B/.onchange

Then activate the react tool

    cd root
    react

it will monitor directories `subdir_B/subdir_C` and `subdir_B`
for changes and execute the `.onchange` script in each directory whenever a file event eccurs.


Example `.onchange`
___________________

If working on a python module react could be used to execute tests whenever a
.py file is modified.

    #!/bin/sh

    extension="${1##*.}"
    case $extension in
        py)
            cd ~/myproject
            python setup.py test
            ;;
    esac
