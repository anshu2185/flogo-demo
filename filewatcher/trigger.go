import (
	"context"
	"fmt"

	"github.com/TIBCOSoftware/flogo-lib/core/trigger"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/fsnotify/fsnotify"
)

var log = logger.GetLogger("trigger-file-watcher")

type FileWatcherTrigger struct {
	metadata *trigger.Metadata
	handlers []*trigger.Handler
	config   *trigger.Config
}

func NewFactory(md *trigger.Metadata) trigger.Factory {
	return &FileWatcherFactory{metadata: md}
}

type FileWatcherFactory struct {
	metadata *trigger.Metadata
}

func (t *FileWatcherFactory) New(config *trigger.Config) trigger.Trigger {
	return &FileWatcherTrigger{metadata: t.metadata, config: config}
}

func (t *FileWatcherTrigger) Metadata() *trigger.Metadata {
	return t.metadata
}

func (t *FileWatcherTrigger) Initialize(ctx trigger.InitContext) error {

	t.handlers = ctx.GetHandlers()
	return nil
}

func (t *FileWatcherTrigger) Start() error {

	log.Debug("Start")
	handlers := t.handlers

	log.Debug("Processing handlers")

	for _, handler := range handlers {

		t.startTrigger(handler)

	}
	return nil

}

func (t *FileWatcherTrigger) startTrigger(handler *trigger.Handler) {

	fmt.Println("Starting File watching process")
	dirName := handler.GetStringSetting("dirName")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:

				if event.Op&fsnotify.Write == fsnotify.Write {
					trgData := make(map[string]interface{})
					trgData["filename"] = event.Name
					response, err := handler.Handle(context.Background(), trgData)

					fmt.Println("modified file:", event.Name)
					if err != nil {
						log.Error("Error starting action: ", err.Error())
					} else {
						log.Debugf("Action call successful: %v", response)
					}
				}
			case err := <-watcher.Errors:
				fmt.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(dirName)
	if err != nil {
		log.Error(err)
	}
	<-done

}

func (t *FileWatcherTrigger) Stop() error {

	return nil
}
