package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"syscall"

	"golang.org/x/sys/windows"
)

func main() {

	var path, action, regex *string
	deep := false

	if len(os.Args) > 1 {
		path = &os.Args[1]
	}
	if len(os.Args) > 2 {
		if os.Args[2] == "true" {
			deep = true
		}
	}
	if len(os.Args) > 3 {
		action = &os.Args[3]
	}
	if len(os.Args) > 4 {
		regex = &os.Args[4]
	}

	watch(path, deep, action, regex)
}

func watch(directory *string, deep bool, action *string, regex *string) {
	if directory == nil {
		showHelp()
		return
	}

	if _, err := os.Stat(*directory); err != nil {
		panic(fmt.Sprintln("Directory:", *directory, "Does not exist!"))
	}

	path, err := syscall.UTF16PtrFromString(*directory)

	if err != nil {
		panic(err)
	}

	var tempHandle windows.Handle

	directoryHandle, err := windows.CreateFile(path,
		windows.FILE_LIST_DIRECTORY,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS,
		tempHandle)

	if err != nil {
		panic(err)
	}

	notifyFilter := windows.FILE_NOTIFY_CHANGE_FILE_NAME |
		windows.FILE_NOTIFY_CHANGE_SIZE |
		windows.FILE_NOTIFY_CHANGE_LAST_WRITE

	for {
		var buffer [2048]byte
		var bytesReturned uint32
		var completionRoutine uintptr

		if deep {
			fmt.Println("Watching for file changes (deep) in:", *directory)
		} else {
			fmt.Println("Watching for file changes (shallow) in:", *directory)
		}

		err := windows.ReadDirectoryChanges(directoryHandle,
			&buffer[0],
			uint32(len(buffer)),
			true,
			uint32(notifyFilter),
			&bytesReturned,
			nil,
			completionRoutine)

		if err != nil {
			panic(err)
		}

		fmt.Println("Changes detected:")

		pointer := uint32(0)

		for {
			localBuffer := buffer[pointer:]

			nextOffset := binary.LittleEndian.Uint32(localBuffer[0:4])
			fileNameLength := binary.LittleEndian.Uint32(localBuffer[8:12])
			fileName := string(localBuffer[12 : 12+fileNameLength])
			fileChangeEvent := FileChangeEvent{binary.LittleEndian.Uint32(localBuffer[4:8]), fileName}

			if regex != nil {
				match, err := regexp.MatchString(*regex, fileName)
				if err != nil {
					panic(err)
				}
				if match {
					performAction(fileChangeEvent, action)
				}
			} else {
				performAction(fileChangeEvent, action)
			}

			if nextOffset <= 0 {
				break
			}

			pointer += nextOffset
		}
	}
}

func performAction(fileChangeEvent FileChangeEvent, action *string) {
	if action == nil {
		fmt.Println(fileChangeEvent.EventType, fileChangeEvent.FileName)
		return
	}

	exec.Command(*action, fmt.Sprint(fileChangeEvent.EventType), fileChangeEvent.FileName)
}

func showHelp() {
	fmt.Println("Ophanim is a directory and file watcher.")
	fmt.Println("")
	fmt.Println("ophanim [directory] [deep] [action] [regex]")
	fmt.Println("")
	fmt.Println("Directory <string>: The directory to watch.")
	fmt.Println("Deep <?boolean>: Whether only the given directory is checked or also its sub-directories and their files.")
	fmt.Println("Action <?string>: The command to execute when a change is detected, by default stdout will be used.")
	fmt.Println("Regex <?string>: The regex that will be matched against the detected FileName before an action is performed.")
}

type FileChangeEvent struct {
	EventType uint32
	FileName  string
}
