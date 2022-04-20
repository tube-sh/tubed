package daemon

import (
	"log"
	"syscall"

	"github.com/sevlyar/go-daemon"
)

func Run(proxy string, tubedDir string, baseUrl string, token string, signal *string) {

	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, TermHandler)

	cntxt := &daemon.Context{
		PidFileName: "/var/run/tubed.pid",
		PidFilePerm: 0644,
		LogFileName: "/var/log/tubed.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		//Args:        []string{"[go-daemon tubed]"},
		Args: nil,
	}

	if len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			log.Fatalf("Unable send signal to the daemon: %s", err.Error())
		}
		daemon.SendCommands(d)
		return
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatalln(err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	log.Println("tubed daemon started")

	go Worker(proxy, tubedDir, baseUrl, token)

	err = daemon.ServeSignals()
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}

	log.Println("tubed daemon terminated")
}
