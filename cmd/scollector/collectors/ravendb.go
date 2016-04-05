package collectors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bosun.org/metadata"
	"bosun.org/opentsdb"
)

const (
	defaultRavenDBURL string = "http://localhost:8080"
	rdbPrefix                 = "ravendb."
)

func init() {
	collectors = append(collectors, &IntervalCollector{F: c_ravendb_adminstats, Enable: enableRavendb})
}

func RavenDB(url string) error {
	collectors = append(collectors,
		&IntervalCollector{
			F: func() (opentsdb.MultiDataPoint, error) {
				return ravendbAdminStats(url)
			},
			name: fmt.Sprintf("ravendb-adminstats-%s", url),
		})
	return nil
}

func enableRavendb() bool {
	return enableURL(defaultRavenDBURL)()
}

func c_ravendb_adminstats() (opentsdb.MultiDataPoint, error) {
	return ravendbAdminStats(defaultRavenDBURL)

}

func ravendbAdminStats(s string) (opentsdb.MultiDataPoint, error) {
    const megabyte = 1048576
	p := rdbPrefix + "admin."
	res, err := http.Get(s + "/admin/stats")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var o rdbAdminStats
	if err := json.NewDecoder(res.Body).Decode(&o); err != nil {
		return nil, err
	}
	var md opentsdb.MultiDataPoint
	ts := opentsdb.TagSet{"host": "010779455157"}
	Add(&md, p+"totalnumberofrequests", o.TotalNumberOfRequests, ts, metadata.Counter, metadata.Request, DescRdbTotalRequests)
    Add(&md, p+"memory.databasecachesize", o.Memory.DatabaseCacheSizeInMB*megabyte, ts, metadata.Gauge, metadata.Bytes, DescRdbMemoryDatabaseCacheSize)
    Add(&md, p+"memory.managedmemorysize", o.Memory.ManagedMemorySizeInMB*megabyte, ts, metadata.Gauge, metadata.Bytes, DescRdbManagedMemorySize)
    Add(&md, p+"memory.totalprocessmemorysize", o.Memory.TotalProcessMemorySizeInMB*megabyte, ts, metadata.Gauge, metadata.Bytes, DescRdbTotalProcessMemorySize)
	return md, nil
}

type rdbAdminStats struct {
	ServerName            string `json:"ServerName"`
    TotalNumberOfRequests int    `json:"TotalNumberOfRequests"`
    Memory struct {
		DatabaseCacheSizeInMB      float64 `json:"DatabaseCacheSizeInMB"`
		ManagedMemorySizeInMB      float64 `json:"ManagedMemorySizeInMB"`
		TotalProcessMemorySizeInMB float64 `json:"TotalProcessMemorySizeInMB"`
	} `json:"Memory"`
}

type rdbLoadedDatabases struct {
    Name                              string `json:"Name"`
    TransactionalStorageAllocatedSize int    `json:"TransactionalStorageAllocatedSize"`
    TransactionalStorageUsedSize      int    `json:"TransactionalStorageUsedSize"`
    IndexStorageSize                  int    `json:"IndexStorageSize"`
    TotalDatabaseSize                 int    `json:"TotalDatabaseSize"`
    CountOfDocuments                  int    `json:"CountOfDocuments"`
    CountOfAttachments                int    `json:"CountOfAttachments"`
    Metrics struct {
        DocsWritesPerSecond int `json:"DocsWritesPerSecond"`
        IndexedPerSecond    int `json:"IndexedPerSecond"`
        ReducedPerSecond    int `json:"ReducedPerSecond"`
        RequestsPerSecond   int `json:"RequestsPerSecond"`
        Requests struct {
            Count int `json:"Count"`
        } `json:"Requests"`
        RequestsDuration struct {
            Counter int `json:"Counter"`
        } `json:"RequestsDuration"`
        StaleIndexMaps struct {
            Counter int `json:"Counter"`
        } `json:"StaleIndexMaps"`
        StaleIndexReduces struct {
            Counter int `json:"Counter"`
        } `json:"StaleIndexReduces"`
        Gauges struct {
            RavenDatabaseIndexingIndexBatchSizeAutoTuner struct {
                MaxNumberOfItems     int `json:"MaxNumberOfItems"`
                CurrentNumberOfItems int `json:"CurrentNumberOfItems"`
                InitialNumberOfItems int `json:"InitialNumberOfItems"`
            }  `json:"Raven.Database.Indexing.IndexBatchSizeAutoTuner"`
            RavenDatabaseIndexingReduceBatchSizeAutoTuner struct {
                MaxNumberOfItems     int `json:"MaxNumberOfItems"`
                CurrentNumberOfItems int `json:"CurrentNumberOfItems"`
                InitialNumberOfItems int `json:"InitialNumberOfItems"`
            }  `json:"Raven.Database.Indexing.ReduceBatchSizeAutoTuner"`
            RavenDatabaseIndexingWorkContext struct {
                RunningQueriesCount int `json:"RunningQueriesCount"`
            }  `json:"Raven.Database.Indexing.WorkContext"`
        } `json:"Gauges"`
    } `json:"Metrics"`
}

type rdbDatabaseStats struct {
	CountOfIndexes                            int `json:"CountOfIndexes"`
    CountOfResultTransformers                 int `json:"CountOfResultTransformers"`
    ApproximateTaskCount                      int `json:"ApproximateTaskCount"`
    CountOfDocuments                          int `json:"CountOfDocuments"`
    CountOfAttachments                        int `json:"CountOfAttachments"`
    CurrentNumberOfItemsToIndexInSingleBatch  int `json:"CurrentNumberOfItemsToIndexInSingleBatch"`
    CurrentNumberOfItemsToReduceInSingleBatch int `json:"CurrentNumberOfItemsToReduceInSingleBatch"`
    Indexes rdbIndexes `json:"Indexes"`
}

type rdbIndexes struct {
    Name              string `json:"Name"`
    IndexingAttempts  int `json:"IndexingAttempts"`
    IndexingSuccesses int `json:"IndexingSuccesses"`
    IndexingErrors    int `json:"IndexingErrors"`
    IndexingLag       int `json:"IndexingLag"`
    TouchCount        int `json:"TouchCount"`
    DocsCount         int `json:"DocsCount"`
}

const (
	DescRdbTotalRequests  = "Number of requests that have been executed against the server"
    DescRdbMemoryDatabaseCacheSize  = "Size of the database cache"
    DescRdbManagedMemorySize  = "Size of managed memory taken by server"
    DescRdbTotalProcessMemorySize  = "Total size of memory taken by server"
)