package manager

import (
	"fmt"
	"io/fs"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

func createFilesAndDirs() error {
	pid, err := getRunningInstance()
	if err == nil {
		return errors.New(fmt.Sprintf("there is a running instance of sminit with pid %d", pid))
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	err = os.Mkdir(SminitRunDir, fs.ModeDir)
	if err != nil {
		return errors.Wrapf(err, "could not create directory %s", SminitRunDir)
	}

	err = createSminitPidFile()
	if err != nil {
		return errors.Wrap(err, "could not create sminit pid file")
	}
	return nil
}

// getRunningInstance returns the pid of the running instance of sminit.
func getRunningInstance() (pid int, err error) {
	b, err := os.ReadFile(SminitPidPath)
	if err != nil {
		return 0, errors.Wrapf(err, "could not read file %s", SminitPidPath)
	}

	pid, err = strconv.Atoi(string(b))
	if err != nil {
		return 0, errors.Wrapf(err, "could not convert bytes from %s to int", SminitPidPath)
	}

	return pid, nil
}

func createSminitPidFile() error {
	f, err := os.Create(SminitPidPath)
	if err != nil {
		return errors.Wrapf(err, "could not create %s", SminitPidPath)
	}

	pidBytes := []byte(strconv.FormatInt(int64(os.Getpid()), 10))
	_, err = f.Write(pidBytes)
	if err != nil {
		return err
	}

	return nil
}

// CleanUp should delete /run/sminit directory and /run/sminit.log
func CleanUp() {
	err := os.RemoveAll(SminitRunDir)
	if err != nil {
		SminitLog.Error().Msgf("error while removing %s, you need to remove it manually. %s", SminitRunDir, err.Error())
	}
	err = os.Remove(SminitLogPath)
	if err != nil {
		SminitLog.Error().Msgf("error while removing %s, you need to remove it manually. %s", SminitLogPath, err.Error())
	}
}
