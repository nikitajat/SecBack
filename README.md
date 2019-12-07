# SecBack
# Build
These are the main files, which will be used to create Executable.
In order to Build Download the GoDep Folder from the  Link : https://drive.google.com/drive/folders/15wHW1pT39IDRz0W0WBMbt3FwxLk-QRSD

Create a Go Environment path to this Folder and go inside : /src/github/chinmay/SecBack/
Run Go build command.
Executable will be created.
# Demo

  running the secback first time will create a Backup Folder in your Home directory named 'BackUp'. And also creates local                                                      clouds in the home directory Cloud_1, Cloud_2 and Cloud_3.
# 

Command 1 :
        './SecBack start '
             Add a file to Back up Folder and secBack will create shares and will upload to local cloud stores present in the home.                                     directory.
#
For each File having unique Share Id.
        You can also see '.config' in your backup folder. which will be hidden.
        Configuration file contains name of the files as well as their share ids and their SHA256 Hashes.
        Configuration file also stores information regarding Directory Map.
#
  './SecBack restore'.
        For demo you can delete files from Backup Folder and on giving restore command.
        File will be restored.
