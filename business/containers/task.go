package containers

import "github.com/aws/aws-sdk-go-v2/service/ecs"

// task renders into a valid ecs task definition
type task struct {
	Name       string // name of the task
	ArchiveUrl string // where in S3 is the archive?
	taskARN    string // where the task is running
}

func (t task) render() ecs.RegisterTaskDefinitionInput {
	// family = t.Name
	return ecs.RegisterTaskDefinitionInput{}
}

// List returns a list of known instances
func List() []task {
	// list tasks
	return nil
}

func (t task) Start() error {
	// register task definition as new version
	// start task
	// store t.taskARN
	return nil
}

func (t task) Stop() error {
	// stop task
	return nil
}

func (t task) Status() error {
	// get status
	return nil
}
