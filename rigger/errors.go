package rigger

import "fmt"

// implment: error
func (err *Error) Error() string {
	return err.ErrStr
}

type ErrDefault string

func (e ErrDefault) Error() string {
	return string(e)
}

type ErrSpawn string

func (e ErrSpawn) Error() string {
	return string(e)
}

type ErrUnSurportedProducer string

func (u ErrUnSurportedProducer) Error() string {
	return fmt.Sprintf("unsupported producer:%s", string(u))
}

type ErrWrongProducer string

func (e ErrWrongProducer) Error() string {
	return fmt.Sprintf("wrong producer:%s", string(e))
}

type ErrSerializedSlicLenWrong struct {

}

func (e ErrSerializedSlicLenWrong) Error() string {
	return "len worng"
}

type ErrPidIsNil struct {

}

func (e ErrPidIsNil) Error() string {
	return "pid is nil"
}

type ErrNotRegister string

func (e ErrNotRegister) Error() string {
	return fmt.Sprintf("%s not registered", string(e))
}

type ErrNoParentNode string

func (e ErrNoParentNode) Error() string {
	return fmt.Sprintf("parent node: %s not existed", string(e))
}




