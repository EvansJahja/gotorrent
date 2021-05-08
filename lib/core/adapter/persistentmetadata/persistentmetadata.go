package persistentmetadata

type PersistentMetadata interface {
	Put(key string, value interface{}) error
	Get(key string, value interface{}) error
}
