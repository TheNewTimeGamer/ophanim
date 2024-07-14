Ophanim is a directory and file watcher.

ophanim [directory] [deep] [action] [regex]

Directory <string>: The directory to watch.
Deep <?boolean>: Whether only the given directory is checked or also its sub-directories and their files.
Action <?string>: The command to execute when a change is detected, by default stdout will be used.
Regex <?string>: The regex that will be matched against the detected FileName before an action is performed.
