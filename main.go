package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"time"

	"github.com/0xAX/notificator"
	"github.com/deckarep/gosx-notifier"
	"github.com/oleksandr/bonjour"
)

var (
	initCfg    bool
	configFile string
	cfg        configuration
)

type configuration struct {
	Service struct {
		Port int `json:"port"`
	} `json:"service"`
}

type message struct {
	Header      string
	Description string
	Icon        string
}

func (m *message) show() (err error) {
	if runtime.GOOS == "darwin" {
		// think different..
		note := gosxnotifier.NewNotification("linconnect")
		note.Title = m.Header
		note.Subtitle = m.Description
		note.Group = "com.willhauck.linconnect"
		note.AppIcon = m.Icon
		err = note.Push()
		if err != nil {
			return err
		}
		return nil
	}
	notify := notificator.New(notificator.Options{
		AppName: "linconnect",
	})
	err = notify.Push(m.Header, m.Description, m.Icon, notificator.UR_NORMAL)
	if err != nil {
		return err
	}
	return nil

}

func notif(w http.ResponseWriter, r *http.Request) {
	var msg message
	header := r.Header["Notifheader"]
	decodeHeader, _ := base64.StdEncoding.DecodeString(header[0])
	msg.Header = string(decodeHeader)
	description := r.Header["Notifdescription"]
	decodeDescription, _ := base64.StdEncoding.DecodeString(description[0])
	msg.Description = string(decodeDescription)
	log.Printf("[New Message] %s: %s", msg.Header, msg.Description)

	// Parse form data
	err := r.ParseMultipartForm(100000)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m := r.MultipartForm
	//get the *fileheaders
	files := m.File["notificon"]
	for i, _ := range files {
		//for each fileheader, get a handle to the actual file
		file, err := files[i].Open()
		defer file.Close()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//create destination file making sure the path is writeable.
		dst, err := os.Create("./" + files[i].Filename)
		defer dst.Close()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//copy the uploaded file to the destination file
		if _, err := io.Copy(dst, file); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		msg.Icon = files[i].Filename
	}
	// try to display the message
	err = msg.show()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	io.WriteString(w, "true")
}

func readConfig(file string) (c configuration, err error) {
	f, err := os.Open(file)
	if err != nil {
		return c, err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	err = decoder.Decode(&c)
	if err != nil {
		return c, err
	}

	return c, nil
}

func initConfig(file string) (err error) {
	var c configuration
	c.Service.Port = 9090

	pretty, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(file, pretty, 0644)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	flag.StringVar(&configFile, "conf", "config.json", "select config file")
	flag.BoolVar(&initCfg, "init", false, "write an example config file")
}

func main() {
	flag.Parse()
	if len(configFile) == 0 {
		log.Fatalln("Startup failed. Expected config file")
	}
	if initCfg {
		err := initConfig(configFile)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("Created default configuration file")
	}
	cfg, err := readConfig(configFile)
	if err != nil {
		log.Fatalln(err)
	}

	// Register service
	s, err := bonjour.Register("linconnect-go", "_linconnect._tcp", "", cfg.Service.Port, []string{"txtv=1", "app=com.willhauck.linconnect"}, nil)
	if err != nil {
		log.Fatalln(err.Error())
	}

	fs := http.FileServer(http.Dir("./"))
	http.Handle("/", fs)
	http.HandleFunc("/notif", notif)

	err = http.ListenAndServe(":"+strconv.Itoa(cfg.Service.Port), nil)
	if err != nil {
		log.Println(err)
	}

	// Ctrl+C
	handler := make(chan os.Signal, 1)
	signal.Notify(handler, os.Interrupt)
	for sig := range handler {
		if sig == os.Interrupt {
			s.Shutdown()
			time.Sleep(1e9)
			break
		}
	}
}
