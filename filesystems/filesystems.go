package filesystems

import "time"

// FS is the interface that all filesystem implementations must satisfy.
// It defines the basic operations for managing files, such as uploading,
// downloading, listing, and deleting files.
type FS interface {
	Put(fileName, folder string) error
	Get(destination string, items ...string) error
	List(prefix string) ([]Listing, error)
	Delete(itemsToDelete []string) error
}

// Listing represents the metadata of a file or directory in the filesystem.
type Listing struct {
	Etag         string
	LastModified time.Time
	Key          string
	Size         float64
	IsDir        bool
}
