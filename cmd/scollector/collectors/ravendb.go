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
	var as rdbAdminStats
	if err := json.NewDecoder(res.Body).Decode(&as); err != nil {
		return nil, err
	}
	var md opentsdb.MultiDataPoint
	Add(&md, p+"totalnumberofrequests", as.TotalNumberOfRequests, nil, metadata.Counter, metadata.Request, DescRdbTotalRequests)
    Add(&md, p+"memory.databasecachesize", as.Memory.DatabaseCacheSizeInMB*megabyte, nil, metadata.Gauge, metadata.Bytes, DescRdbMemoryDatabaseCacheSize)
    Add(&md, p+"memory.managedmemorysize", as.Memory.ManagedMemorySizeInMB*megabyte, nil, metadata.Gauge, metadata.Bytes, DescRdbManagedMemorySize)
    Add(&md, p+"memory.totalprocessmemorysize", as.Memory.TotalProcessMemorySizeInMB*megabyte, nil, metadata.Gauge, metadata.Bytes, DescRdbTotalProcessMemorySize)
    for _, ldb := range as.LoadedDatabases {
        p := rdbPrefix + "database."
        if ldb.Name == "" { 
            continue // skip system database
        }
        ts := opentsdb.TagSet{"database": ldb.Name}
		Add(&md, p+"transactionalstorageallocatedsize", ldb.TransactionalStorageAllocatedSize, ts, metadata.Gauge, metadata.Bytes, DescTransactionalStorageAllocatedSize)
        Add(&md, p+"transactionalstorageusedsize", ldb.TransactionalStorageUsedSize, ts, metadata.Gauge, metadata.Bytes, DescTransactionalStorageUsedSize)
		Add(&md, p+"indexstoragesize", ldb.IndexStorageSize, ts, metadata.Gauge, metadata.Bytes, DescIndexStorageSize)
		Add(&md, p+"totaldatabasesize", ldb.TotalDatabaseSize, ts, metadata.Gauge, metadata.Bytes, DescTotalDatabaseSize)
        Add(&md, p+"countofdocuments", ldb.CountOfDocuments, ts, metadata.Counter, metadata.Document, DescCountofDocuments)
        Add(&md, p+"countofattachments", ldb.CountOfAttachments, ts, metadata.Counter, metadata.Document, DescCountOfAttachments)
        Add(&md, p+"metrics.docswritespersecond", ldb.Metrics.DocsWritesPerSecond, ts, metadata.Gauge, metadata.Document, DescMetricsDocsWritesPerSecond)
        Add(&md, p+"metrics.indexedpersecond", ldb.Metrics.IndexedPerSecond, ts, metadata.Gauge, metadata.Document, DescMetricsIndexedPerSecond)
        Add(&md, p+"metrics.reducedpersecond", ldb.Metrics.ReducedPerSecond, ts, metadata.Gauge, "reduces", DescMetricsReducedPerSecond)
        Add(&md, p+"metrics.requests.count", ldb.Metrics.Requests.Count, ts, metadata.Counter, metadata.Request, DescMetricsRequestsCount)
        Add(&md, p+"metrics.requestsduration.counter", ldb.Metrics.Requests.Count, ts, metadata.Counter, "duration", DescMetricsRequestsDurationCounter)
        Add(&md, p+"metrics.staleindexmaps.counter", ldb.Metrics.Requests.Count, ts, metadata.Counter, "index maps", DescMetricsStaleIndexMapsCounter)
        Add(&md, p+"metrics.staleindexreduces.counter", ldb.Metrics.Requests.Count, ts, metadata.Counter, "index reduces", DescMetricsStaleIndexReducesCounter)
    }
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
    LoadedDatabases []rdbLoadedDatabases `json:"LoadedDatabases"`
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
        DocsWritesPerSecond float64 `json:"DocsWritesPerSecond"`
        IndexedPerSecond    float64 `json:"IndexedPerSecond"`
        ReducedPerSecond    float64 `json:"ReducedPerSecond"`
        RequestsPerSecond   float64 `json:"RequestsPerSecond"`
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
                MaxNumberOfItems     string `json:"MaxNumberOfItems"`
                CurrentNumberOfItems string `json:"CurrentNumberOfItems"`
                InitialNumberOfItems string `json:"InitialNumberOfItems"`
            }  `json:"Raven.Database.Indexing.IndexBatchSizeAutoTuner"`
            RavenDatabaseIndexingReduceBatchSizeAutoTuner struct {
                MaxNumberOfItems     string `json:"MaxNumberOfItems"`
                CurrentNumberOfItems string `json:"CurrentNumberOfItems"`
                InitialNumberOfItems string `json:"InitialNumberOfItems"`
            }  `json:"Raven.Database.Indexing.ReduceBatchSizeAutoTuner"`
            RavenDatabaseIndexingWorkContext struct {
                RunningQueriesCount string `json:"RunningQueriesCount"`
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
	DescRdbTotalRequests = "Number of requests that have been executed against the server"
    DescRdbMemoryDatabaseCacheSize = "Size of the database cache"
    DescRdbManagedMemorySize = "Size of managed memory taken by server"
    DescRdbTotalProcessMemorySize = "Total size of memory taken by server"
    DescTransactionalStorageAllocatedSize = "The amount of storage allocated for transactions"
    DescTransactionalStorageUsedSize = "The amount of storage used for transactions"
    DescIndexStorageSize = "The amount of storage in use by indexes"
    DescTotalDatabaseSize = "The total size of the database"
    DescCountofDocuments = "The count of documents in the database"
    DescCountOfAttachments = "The count of attachments in the database"
    DescMetricsDocsWritesPerSecond = "The number of documents written to the database per second"
    DescMetricsIndexedPerSecond = "The number of documents indexes created per second"
    DescMetricsReducedPerSecond = "The nubmer of documents reduced per second"
    DescMetricsRequestsCount = "Count of requests made against the database"
    DescMetricsRequestsDurationCounter = "Duration each request took"
    DescMetricsStaleIndexMapsCounter = "Count of stale index maps"
    DescMetricsStaleIndexReducesCounter = "Count of stale index reduces"
)