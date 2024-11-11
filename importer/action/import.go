// =================================================================================
//
//		ccmm - https://www.foxhollow.cc/projects/ccmm/
//
//	 Connection Church Media Manager, aka ccmm, is a tool for managing all
//   aspects of produced media- initial import from removable media,
//   synchronization with clients and automatic data replication and backup
//
//		Copyright (c) 2024 Steve Cross <flip@foxhollow.cc>
//
//		Licensed under the Apache License, Version 2.0 (the "License");
//		you may not use this file except in compliance with the License.
//		You may obtain a copy of the License at
//
//		     http://www.apache.org/licenses/LICENSE-2.0
//
//		Unless required by applicable law or agreed to in writing, software
//		distributed under the License is distributed on an "AS IS" BASIS,
//		WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//		See the License for the specific language governing permissions and
//		limitations under the License.
//
// =================================================================================

package action

import (
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"ccmm/importer/processor"
	"ccmm/model"
	"ccmm/util"
)

// ImportStatus Enum to describe the potential state of enums
type ImportStatus int8

const (
	// Pending The import is currently pending and no action has been taken
	Pending ImportStatus = iota

	// Scanning The import process is currently scanning for supported processors and source files
	Scanning

	// Importing The import process is currently copying files from the source to the destination
	Importing

	// Completed The import process has completed
	Completed

	// Failed One or more errors occurred during the import process
	Failed
)

// ImportQueueItem Structure that defines an import job
type ImportQueueItem struct {
	Params           model.ImportVolume
	Processors       []processor.Processor
	Files            []model.SourceFile
	Status           ImportStatus
	FinishedCallback func(queueItem *ImportQueueItem)
	processCallback  func(queueItem *ImportQueueItem)
}

var (
	queueIndex  int
	importQueue = map[int]*ImportQueueItem{}
	importMutex sync.Mutex
)

func getFirstQueueIndex(importQueue *map[int]*ImportQueueItem) int {
	keys := make([]int, 0)
	for k := range *importQueue {
		keys = append(keys, k)
	}

	if len(keys) == 0 {
		return -1
	}

	sort.Ints(keys)
	for _, k := range keys {
		return k
	}

	return -1
}

// ImportWorker Routine that runs and processes the import queue sequentially
func ImportWorker() {
	// TODO: add cancel channel
	for {
		importMutex.Lock()
		firstQueueIndex := getFirstQueueIndex(&importQueue)

		// nothing pending, so we wait a bit and check again.
		if firstQueueIndex < 0 {
			importMutex.Unlock()
			time.Sleep(500 * time.Millisecond)
			continue
		}
		queueItem := importQueue[firstQueueIndex]
		importMutex.Unlock()

		// there seems to be some sort of race condition here, so this is a workaround for it
		if queueItem == nil || queueItem.processCallback == nil {
			slog.Debug(fmt.Sprintf("Process callback not set for queue item #%d, trying again..", firstQueueIndex))
			time.Sleep(500 * time.Millisecond)
			continue
		}

		slog.Info(fmt.Sprintf("Processing import queue item '%d', volume path: '%s'", firstQueueIndex, queueItem.Params.VolumePath))

		queueItem.processCallback(queueItem)

		importMutex.Lock()
		slog.Info(fmt.Sprintf("Finished processing import queue item '%d', volume path: '%s'", firstQueueIndex, queueItem.Params.VolumePath))
		delete(importQueue, firstQueueIndex)
		importMutex.Unlock()
	}
}

// Import Add a new import job to the queue by providing the params
// that describe the import job. Additionally, a finishedCallback should
// be provided and will be executed upon completion of the import job
// (regardless of a successful or failed import)
func Import(config model.ImporterConfig, params model.ImportVolume, finishedCallback func(queueItem *ImportQueueItem)) {
	if !util.DirectoryExists(params.VolumePath) {
		slog.Error(fmt.Sprintf("Cannot import because directory not found: %s", params.VolumePath))
		return
	}

	queueIndex++
	slog.Info(fmt.Sprintf("Queueing import #%d for volume '%s'", queueIndex, params.VolumePath))

	importMutex.Lock()
	importQueue[queueIndex] = &ImportQueueItem{
		Params:           params,
		Status:           Scanning,
		Processors:       make([]processor.Processor, 0),
		Files:            make([]model.SourceFile, 0),
		FinishedCallback: finishedCallback,
	}
	importMutex.Unlock()

	processors := processor.FindProcessors(config, params.VolumePath)

	importMutex.Lock()
	importQueue[queueIndex].Status = Pending
	importQueue[queueIndex].Processors = processors
	importQueue[queueIndex].processCallback = func(queueItem *ImportQueueItem) {
		importMutex.Lock()
		queueItem.Status = Scanning
		importMutex.Unlock()

		files := processor.EnumerateSources(processors, params.Dump)

		importMutex.Lock()
		queueItem.Files = files
		queueItem.Status = Importing
		importMutex.Unlock()

		// TODO: add some sort of status callback here
		processor.ImportFiles(config, files, params.DryRun)

		slog.Info(fmt.Sprintf("Finished import for volume '%s'", params.VolumePath))
		queueItem.FinishedCallback(queueItem)
	}
	importMutex.Unlock()
}
