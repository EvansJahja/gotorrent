package temprepo

type TempMetadata interface {
	Get(key interface{}) (val interface{}, ok bool)
	Set(key interface{}, val interface{})
}
