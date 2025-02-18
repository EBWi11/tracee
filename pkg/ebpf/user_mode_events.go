// Invoked tracee-ebpf events from user mode
// This utility can prove itself useful to generate information needed by signatures that is not provided by normal
// events in the kernel.
// Because the events in the kernel are invoked by other programs behavior, we cannot anticipate which events will be
// invoked and as a result what information will be extracted.
// This is critical because tracee-rules is independent, and doesn't have to run on the same machine as tracee-ebpf.
// This means that tracee-rules might lack basic information of the operating machine needed for some signatures.
// By creating user mode events this information could be intentionally collected and passed to tracee-ebpf afterwards.
package ebpf

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aquasecurity/tracee/types/trace"
)

const InitProcNsDir = "/proc/1/ns"

// CreateInitNamespacesEvent collect the init process namespaces and create event from them.
func CreateInitNamespacesEvent() (trace.Event, error) {
	initNamespacesArgs := getInitNamespaceArguments()
	initNamespacesEvent := trace.Event{
		Timestamp:   int(time.Now().UnixNano()),
		ProcessName: "tracee-ebpf",
		EventID:     int(InitNamespacesEventID),
		EventName:   EventsDefinitions[InitNamespacesEventID].Name,
		ArgsNum:     len(initNamespacesArgs),
		Args:        initNamespacesArgs,
	}
	return initNamespacesEvent, nil
}

// getInitNamespaceArguments Fetch the namespaces of the init process and parse them into event arguments.
func getInitNamespaceArguments() []trace.Argument {
	initNamespaces := fetchInitNamespaces()
	eventDefinition := EventsDefinitions[InitNamespacesEventID]
	initNamespacesArgs := make([]trace.Argument, len(eventDefinition.Params))
	for i, arg := range initNamespacesArgs {
		arg.ArgMeta = eventDefinition.Params[i]
		arg.Value = initNamespaces[arg.Name]
		initNamespacesArgs[i] = arg
	}
	return initNamespacesArgs
}

// fetchInitNamespaces fetch the namespaces values from the /proc/1/ns directory
func fetchInitNamespaces() map[string]uint32 {
	initNamespacesMap := make(map[string]uint32)
	namespaceValueReg := regexp.MustCompile(":[[[:digit:]]*]")
	namespacesLinks, _ := ioutil.ReadDir(InitProcNsDir)
	for _, namespaceLink := range namespacesLinks {
		linkString, _ := os.Readlink(filepath.Join(InitProcNsDir, namespaceLink.Name()))
		trim := strings.Trim(namespaceValueReg.FindString(linkString), "[]:")
		namespaceNumber, _ := strconv.ParseUint(trim, 10, 32)
		initNamespacesMap[namespaceLink.Name()] = uint32(namespaceNumber)
	}
	return initNamespacesMap
}

func (t *Tracee) CreateExistingContainersEvents() []trace.Event {
	var events []trace.Event
	def := EventsDefinitions[ExistingContainerEventID]
	for _, info := range t.containers.GetContainers() {
		args := []trace.Argument{
			{ArgMeta: def.Params[0], Value: info.Runtime.String()},
			{ArgMeta: def.Params[1], Value: info.Container.ContainerId},
			{ArgMeta: def.Params[2], Value: info.Ctime.UnixNano()},
		}
		existingContainerEvent := trace.Event{
			Timestamp:   int(time.Now().UnixNano()),
			ProcessName: "tracee-ebpf",
			EventID:     int(ExistingContainerEventID),
			EventName:   EventsDefinitions[ExistingContainerEventID].Name,
			ArgsNum:     len(args),
			Args:        args,
		}
		events = append(events, existingContainerEvent)
	}
	return events
}
