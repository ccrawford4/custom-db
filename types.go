package customdb

type KV interface {
	Get(key []byte) (val []byte, ok bool)
	Set(key []byte, val []byte)
	Del(key []byte)

	// range query
	FindGreaterThan(key []byte) Iterator
}

type Iterator interface {
	HasNext() bool
	Next() (key []byte, val []byte)
}
