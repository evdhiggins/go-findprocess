// Package findprocess contains utility functions for identifying if a given process is running
package findprocess

import (
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// th32CsSnapProcess (TH32CS_SNAPPROCESS) is described in https://msdn.microsoft.com/de-de/library/windows/desktop/ms682489(v=vs.85).aspx
const th32CsSnapProcess = 0x00000002

// ProcessStatus contains basic process details
type ProcessStatus struct {
	Name      string
	ID        int
	IsRunning bool
}

// ByName checks to see if a process with a given name is running
func ByName(processName string) (*ProcessStatus, error) {
	status := ProcessStatus{Name: processName}

	procs, err := processes()
	if err != nil {
		return nil, err
	}

	process := findProcessByName(procs, processName)
	if process != nil {
		status.ID = process.ProcessID
		status.IsRunning = true
	}

	return &status, nil
}

// ByID checks to see if a process with a given pID is running
func ByID(pID int) (*ProcessStatus, error) {
	status := ProcessStatus{ID: pID}

	procs, err := processes()
	if err != nil {
		return nil, err
	}

	process := findProcessByID(procs, pID)
	if process != nil {
		status.Name = process.Filename
		status.IsRunning = true
	}

	return &status, nil
}

// WindowsProcess is an implementation of Process for Windows.
type WindowsProcess struct {
	ProcessID int
	Filename  string
}

func processes() ([]WindowsProcess, error) {
	handle, err := windows.CreateToolhelp32Snapshot(th32CsSnapProcess, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(handle)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	// get the first process
	err = windows.Process32First(handle, &entry)
	if err != nil {
		return nil, err
	}

	results := make([]WindowsProcess, 0, 50)
	for {
		results = append(results, newWindowsProcess(&entry))

		err = windows.Process32Next(handle, &entry)
		if err != nil {
			// windows sends ERROR_NO_MORE_FILES on last process
			if err == syscall.ERROR_NO_MORE_FILES {
				return results, nil
			}
			return nil, err
		}
	}
}

func findProcessByName(processes []WindowsProcess, name string) *WindowsProcess {
	for _, p := range processes {
		if strings.ToLower(p.Filename) == strings.ToLower(name) {
			return &p
		}
	}
	return nil
}

func findProcessByID(processes []WindowsProcess, pID int) *WindowsProcess {
	for _, p := range processes {
		if pID == p.ProcessID {
			return &p
		}
	}
	return nil
}

func newWindowsProcess(e *windows.ProcessEntry32) WindowsProcess {
	// Find when the string ends for decoding
	end := 0
	for {
		if e.ExeFile[end] == 0 {
			break
		}
		end++
	}

	return WindowsProcess{
		ProcessID: int(e.ProcessID),
		Filename:  syscall.UTF16ToString(e.ExeFile[:end]),
	}
}
