package daemon

import (
	"log"
	"os"
	"syscall"

	"github.com/sevlyar/go-daemon"
)

func TermHandler(sig os.Signal) error {
	err := syscall.Kill(pid, 9)
	if err == nil {
		log.Println("frpc subprocess stopped")
	}
	stop <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-done
	}
	return daemon.ErrStop
}

func reloadHandler(sig os.Signal) error {
	log.Println("configuration reloaded")
	return nil
}
