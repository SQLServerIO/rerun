package main

import (
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ivpusic/golog"
	"github.com/mitchellh/go-ps"
)

var (
	logger = golog.GetLogger("github.com/sqlserverio/rerun")
	//TestMode are we running in test mode
	TestMode = false
)

func main() {
	conf, err := loadConfiguration()
	if err != nil {
		logger.Panicf("Error while loading configuration! %s", err.Error())
	}

	// setup logger level
	if *verbose {
		logger.Level = golog.DEBUG
	} else {
		logger.Level = golog.INFO
	}

	pm := &processManager{
		conf: conf,
	}

	w := &watcher{pm: pm}

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// on ctrl+c remove build files
	go func(conf *config) {
		<-sigs

		logger.Debug("Cleanup build files generated by rerun...")
		logger.Debug("File :%s", conf.build)
		if pid, s, err := findProcess(conf.build); err == nil {
			logger.Debug("Pid:%d, Pname:%s", pid, s)
			time.Sleep(2 * time.Second)

		} else {
			logger.Debug("Process not found %s", err.Error())
		}

		err := os.Remove(conf.build)
		if err != nil && !os.IsNotExist(err) {
			logger.Warnf("Build file not removed! %s", err.Error())
		}

		os.Exit(0)
	}(conf)

	w.start()
}

func findProcess(key string) (int, string, error) {
	pname := ""
	pid := 0
	err := errors.New("not found")
	ps, _ := ps.Processes()
	for i := range ps {
		if ps[i].Executable() == key {
			pid = ps[i].Pid()
			pname = ps[i].Executable()
			err = nil
			break
		}
	}
	return pid, pname, err
}
